import { DataSourceWithBackend, getTemplateSrv, TemplateSrv } from '@grafana/runtime';
import { DataQueryRequest, DataQueryResponse, TypedVariableModel, DataFrame, MetricFindValue, DataSourceInstanceSettings, ScopedVars, ArrayDataFrame, vectorator } from '@grafana/data';
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
    console.log('applyTemplateVariables');
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
    console.log('applyConditionalAll');
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
        return args;
      }
    }
    return [];
  }

  filterQuery(query: SnowflakeQuery): boolean {
    return query.queryText !== '' && !query.hide;
  }

  runQuery(request: Partial<SnowflakeQuery>): Promise<DataFrame> {
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
    console.log('metricFindQuery');
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

  async getTagKeys(options?: any): Promise<MetricFindValue[]> {
    const frame = await this.fetchTags();
    return frame.fields.map( (f) => ({ text: f.name }) );
  }

  async getTagValues({ key }: any): Promise<MetricFindValue[]> {
    const frame = await this.fetchTags();
    const field = frame.fields.find( (f) => f.name === key);
    if (field) {
      return vectorator(field.values)
        .filter( (value) => value !== null )
        .map( (value) => { return { text: String(value) }; });
    }
    return [];
  }

  async fetchTags(): Promise<DataFrame> {
    const rawQuery = this.templateSrv.replace('$snowflake_adhoc_query');
    if (rawQuery === '$snowflake_adhoc_query') {
      return new ArrayDataFrame([]);
    } else {
      this.skipAdHocFilter = true;
      // this.adHocFilter.setTargetTable(rawQuery)
      return await this.runQuery({ queryText: rawQuery });
    }
  }
}

// enum TagType {
//   query,
//   schema,
// }

enum AdHocFilterStatus {
  none = 0,
  enabled,
  disabled,
}
