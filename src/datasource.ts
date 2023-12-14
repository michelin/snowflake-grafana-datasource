import { DataSourceWithBackend, getTemplateSrv, TemplateSrv } from '@grafana/runtime';
import { DataQueryRequest, DataQueryResponse, TypedVariableModel, DataFrame, MetricFindValue, DataSourceInstanceSettings, ScopedVars, DataFrameView, vectorator } from '@grafana/data';
import { SnowflakeQuery, SnowflakeOptions } from './types';
import { switchMap, map } from 'rxjs/operators';
import { firstValueFrom } from 'rxjs';
import AdHocFilter from './adHocFilter';

export class DataSource extends DataSourceWithBackend<SnowflakeQuery, SnowflakeOptions> {
  templateSrv: TemplateSrv;
  adHocFilter: AdHocFilter;
  skipAdHocFilter = false;
  adHocFiltersStatus = AdHocFilterStatus.none;

  constructor(instanceSettings: DataSourceInstanceSettings<SnowflakeOptions>) {
    super(instanceSettings);
    this.annotations = {};
    this.templateSrv = getTemplateSrv();
    this.adHocFilter = new AdHocFilter();
  }

  applyTemplateVariables(query: SnowflakeQuery, scopedVars: ScopedVars): SnowflakeQuery {
    console.log('applyTemplateVariables with query: '+query.queryText);
    let rawQuery = query.queryText || '';
    const templateSrv = getTemplateSrv();
    if(!this.skipAdHocFilter) {
      const adHocFilters = (templateSrv as any)?.getAdhocFilters(this.name);
      if (this.adHocFiltersStatus === AdHocFilterStatus.disabled && adHocFilters.length > 0) {
        throw new Error(`unable to appply ad hoc filters`);
      }
      rawQuery = this.adHocFilter.apply(rawQuery, adHocFilters);
    }
    this.skipAdHocFilter = false;
    rawQuery = this.applyConditionalAll(rawQuery, getTemplateSrv().getVariables());
    return {
      ...query,
      queryText: getTemplateSrv().replace(rawQuery, scopedVars) || '',
    };
  }

  applyConditionalAll(rawQuery: string, templateVars: TypedVariableModel[]): string {
    console.log('applyConditionalAll with query: '+rawQuery);
    if (!rawQuery) {
      return rawQuery;
    }
    const macro = '$__conditionalAll(';
    let macroIndex = rawQuery.lastIndexOf(macro);

    while (macroIndex !== -1) {
      const params = this.getMacroArgs(rawQuery, macroIndex + macro.length - 1);
      if (params.length !== 2) {
        return rawQuery;
      }
      const templateVar = params[1].trim();
      const key = templateVars.find( (x) => x.name === templateVar.substring(1, templateVar.length)) as any;
      let phrase = params[0];
      let value = key?.current.value.toString();
      if (value === '' || value === '$__all') {
        phrase = '1=1';
      }
      rawQuery = rawQuery.replace(`${macro}${params[0]},${params[1]})`, phrase);
      macroIndex = rawQuery.lastIndexOf(macro);
    }
    return rawQuery;
  }

  private getMacroArgs(query: string, argsIndex: number): string[] {
    const args = [] as string[];
    const re = /\(|\)|,/g;
    let bracketCount = 0;
    let lastArgEndIndex = 1;
    let regExpArray: RegExpExecArray | null;
    const argsSubstr = query.substring(argsIndex, query.length);
    while ((regExpArray = re.exec(argsSubstr)) !== null) {
      const foundNode = regExpArray[0];
      if (foundNode === '(') {
        bracketCount++;
      } else if (foundNode === ')') {
        bracketCount--;
      }
      if (foundNode === ',' && bracketCount === 1) {
        args.push(argsSubstr.substring(lastArgEndIndex, re.lastIndex - 1));
        lastArgEndIndex = re.lastIndex;
      }
      if (bracketCount === 0) {
        args.push(argsSubstr.substring(lastArgEndIndex, re.lastIndex - 1));
        console.log('args: '+args)
        return args;
      }
    }
    return [];
  }

  filterQuery(query: SnowflakeQuery): boolean {
    console.log('filterQuery called with: '+query.queryText+' and result: '+(query.queryText !== '' && !query.hide))
    return query.queryText !== '' && !query.hide;
  }

  runQuery(request: Partial<SnowflakeQuery>): Promise<DataFrame> {
    console.log('runQuery called')
    return new Promise( (resolve) => {
      const req = {
        targets: [{ ...request, refId: String(Math.random()) }]
      } as DataQueryRequest<SnowflakeQuery>;
      this.query(req).subscribe((res: DataQueryResponse) => {
        resolve(res.data[0] || { fields: [] });
      });
    });
  }

  async metricFindQuery(queryText: string): Promise<MetricFindValue[]> {
    console.log('metricFindQuery with query: '+queryText);
    if (!queryText) {
      return Promise.resolve([]);
    }

    return firstValueFrom(this.query({
      targets: [
        {
          queryText: queryText,
          refId: 'search',
        },
      ],
      maxDataPoints: 0,
    } as DataQueryRequest<SnowflakeQuery>)
      .pipe(
        switchMap((response) => {
          if (response.error) {
            console.log('Error: ' + response.error.message);
            throw new Error(response.error.message);
          }
          return response.data;
        }),
        switchMap((data: DataFrame) => {
          return data.fields;
        }),
        map((field) =>
          field.values.toArray().map((value) => {
            return { text: value };
          })
        )
      ));
  }

  async getTagKeys(): Promise<MetricFindValue[]> {
    const { type, frame } = await this.fetchTags();
    if (type === TagType.query) {
      return frame.fields.map((f) => ({ text: f.name }));
    }
    const view = new DataFrameView(frame);
    return view.map((item) => ({
      text: `${item[2]}.${item[0]}`,
    }));
  }

  async getTagValues({ key }: any): Promise<MetricFindValue[]> {
    const { type } = this.getTagSource();
    this.skipAdHocFilter = true;
    if (type === TagType.query) {
      return this.fetchTagValuesFromQuery(key);
    }
    return this.fetchTagValuesFromSchema(key);
  }

  private async fetchTagValuesFromSchema(key: string): Promise<MetricFindValue[]> {
    const { from } = this.getTagSource();
    const [table, col] = key.split('.');
    const source = from?.includes('.') ? `${from.split('.')[0]}.${table}` : table;
    const rawSql = `select distinct ${col} from ${source} limit 1000`;
    const frame = await this.runQuery({ queryText: rawSql });
    if (frame.fields?.length === 0) {
      return [];
    }
    const field = frame.fields[0];
    // Convert to string to avoid https://github.com/grafana/grafana/issues/12209
    return vectorator(field.values)
      .filter((value) => value !== null)
      .map((value) => {
        return { text: String(value) };
      });
  }

  private async fetchTagValuesFromQuery(key: string): Promise<MetricFindValue[]> {
    const { frame } = await this.fetchTags();
    const field = frame.fields.find((f) => f.name === key);
    if (field) {
      // Convert to string to avoid https://github.com/grafana/grafana/issues/12209
      return vectorator(field.values)
        .filter((value) => value !== null)
        .map((value) => {
          return { text: String(value) };
        });
    }
    return [];
  }

  async fetchTags(): Promise<Tags> {
    const tagSource = this.getTagSource();
    this.skipAdHocFilter = true;

    if (tagSource.source === undefined) {
      const sql = 'SELECT COLUMN_NAME, DATA_TYPE, TABLE_NAME FROM information_schema.columns';
      const results = await this.runQuery({ queryText: sql });
      return { type: TagType.schema, frame: results };
    }

    if (tagSource.type === TagType.query) {
      this.adHocFilter.setTargetTableFromQuery(tagSource.source);
    } else {
      let table = tagSource.from;
      this.adHocFilter.setTargetTable(table || '');
    }

    const results = await this.runQuery({ queryText: tagSource.source });
    return { type: tagSource.type, frame: results }
  }

  private getTagSource() {
    const ADHOC_VAR = '$snowflake_adhoc_query';
    let source = getTemplateSrv().replace(ADHOC_VAR);
    console.log('called getTagSource: '+source)
    if (source === ADHOC_VAR) {
      return { type: TagType.schema, source: undefined };
    }
    if (source.toLowerCase().startsWith('select')) {
      return { type: TagType.query, source };
    }
    const sql = `SELECT COLUMN_NAME, DATA_TYPE, TABLE_NAME FROM information_schema.columns WHERE table_name='${source}'`
    return { type: TagType.schema, source: sql, from: source }
  }
}

enum TagType {
  query,
  schema,
}

enum AdHocFilterStatus {
  none = 0,
  enabled,
  disabled,
}

interface Tags {
  type?: TagType;
  frame: DataFrame;
}
