import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './ConfigEditor';
import { QueryEditor } from './QueryEditor';
import { SnowflakeQuery, SnowflakeOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, SnowflakeQuery, SnowflakeOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
