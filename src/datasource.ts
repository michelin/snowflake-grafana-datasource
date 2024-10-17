import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { DataQueryRequest, DataFrame, MetricFindValue, DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { SnowflakeQuery, SnowflakeOptions } from './types';
import { switchMap, map } from 'rxjs/operators';
import { firstValueFrom } from 'rxjs';

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
        switchMap((data: DataFrame) => {
          return data.fields;
        }),
        map((field) =>
          field.values.map((value) => {
            return { text: value };
          })
        )
      ));
  }
}
