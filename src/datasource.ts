import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { SnowflakeQuery, SnowflakeOptions } from './types';

export class DataSource extends DataSourceWithBackend<SnowflakeQuery, SnowflakeOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<SnowflakeOptions>) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: SnowflakeQuery, scopedVars: ScopedVars): SnowflakeQuery {
    query.queryText = getTemplateSrv().replace(query.queryText, scopedVars);
    return query;
  }

  filterQuery(query: SnowflakeQuery): boolean {
    return query.queryText !== '' && !query.hide;
  }
}
