import {DataSourceWithBackend, getTemplateSrv, TemplateSrv} from '@grafana/runtime';
import {DataFrame, DataQueryRequest, DataSourceInstanceSettings, MetricFindValue, ScopedVars} from '@grafana/data';
import {SnowflakeOptions, SnowflakeQuery} from './types';
import {map, switchMap} from 'rxjs/operators';
import {firstValueFrom} from 'rxjs';
import {uniqBy} from 'lodash';

export class DataSource extends DataSourceWithBackend<SnowflakeQuery, SnowflakeOptions> {

  constructor(instanceSettings: DataSourceInstanceSettings<SnowflakeOptions>,
      private readonly templateSrv: TemplateSrv = getTemplateSrv()
  ) {
    super(instanceSettings);
    this.annotations = {};
  }

  applyTemplateVariables(query: SnowflakeQuery, scopedVars: ScopedVars): SnowflakeQuery {
    return {
      ...query,
      queryText: this.templateSrv.replace(query.queryText, scopedVars),
    }
  }

  filterQuery(query: SnowflakeQuery): boolean {
    return query.queryText !== '' && !query.hide;
  }

  async metricFindQuery(queryText: string): Promise<MetricFindValue[]> {
    if (!queryText) {
      return [];
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
          if (response.errors) {
            console.log('Error: ' + response.errors?.map(value => value.message ?? '').join(','));
            throw new Error(response.errors?.map(value => value.message ?? '').join(','));
          }
          return response.data;
        }),
        map((data: DataFrame) => {
          const values: MetricFindValue[] = [];
          const textField = data.fields.find((f) => f.name.toLowerCase() === '__text');
          const valueField = data.fields.find((f) => f.name.toLowerCase() === '__value');

          if (textField && valueField) {
            for (let i = 0; i < textField.values.length; i++) {
              values.push({ text: '' + textField.values[i], value: '' + valueField.values[i] });
            }
          } else {
            for (const field of data.fields) {
              for (const value of field.values) {
                values.push({ text: value });
              }
            }
          }

          return uniqBy(values, 'text');
        })
      ));
  }
}
