import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { DataQueryRequest, DataFrame, MetricFindValue, DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { SnowflakeQuery, SnowflakeOptions } from './types';
import { switchMap, map } from 'rxjs/operators';
import { firstValueFrom } from 'rxjs';
import { uniqBy } from 'lodash';

export class DataSource extends DataSourceWithBackend<SnowflakeQuery, SnowflakeOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<SnowflakeOptions>) {
    super(instanceSettings);
    this.annotations = {};
  }

  applyTemplateVariables(query: SnowflakeQuery, scopedVars: ScopedVars): SnowflakeQuery {
    query.queryText = getTemplateSrv().replace(query.queryText, scopedVars);
    return query;
  }

  filterQuery(query: SnowflakeQuery): boolean {
    return query.queryText !== '' && !query.hide;
  }

  async metricFindQuery(queryText: string): Promise<MetricFindValue[]> {
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
