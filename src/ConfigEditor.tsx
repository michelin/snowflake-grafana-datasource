import React, { ChangeEvent, PureComponent } from 'react';
import { InlineField, InlineSwitch, SecretInput, Input, FieldSet } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { SnowflakeOptions, SnowflakeSecureOptions } from './types';



interface Props extends DataSourcePluginOptionsEditorProps<SnowflakeOptions> { }

interface State { }

export class ConfigEditor extends PureComponent<Props, State> {
  
  onAccountChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;

    let value;
    if (event.target.value.includes('.snowflakecomputing.com')) {
      value = event.target.value;
    } else {
      value = event.target.value + '.snowflakecomputing.com';
    }

    // Sanitize value to avoid error
    const regex = new RegExp('https?://');
    value = value.replace(regex, '');

    const jsonData = {
      ...options.jsonData,
      account: value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onUsernameChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      username: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onRoleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      role: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onWarehouseChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      warehouse: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onExtraOptionChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      extraConfig: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onDatabaseChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      database: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onAuthenticationChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      basicAuth: event.target.checked,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onSchemaChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      schema: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secure field (only sent to the backend)
  onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        password: event.target.value,
        privateKey: '',
      },
    });
  };

  onResetPassword = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  onPrivateKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        privateKey: event.target.value,
        password: '',
      },
    });
  };

  onResetPrivateKey = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        privateKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        privateKey: '',
      },
    });
  };

  onMaxOpenConnectionsChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      maxOpenConnections: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };
  onMaxQueuedQueriesChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      maxQueuedQueries: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };
  onConnectionLifetimeChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      conectionLifetime: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };
  onUseCachingChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      useCaching: event.target.checked,
    };
    onOptionsChange({ ...options, jsonData });
  };
  onUseCacheByDefaultChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      useCacheByDefault: event.target.checked,
    };
    onOptionsChange({ ...options, jsonData });
  };
  onCacheSizeChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      cacheSize:  event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };
  onCacheRetentionChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      cacheRetention:  event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };
  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as SnowflakeSecureOptions;
    return (
      <FieldSet>
        <h3 className="page-heading">Connection</h3>
        <InlineField
          labelWidth={30}
          label="Account name"
          tooltip="All access to Snowflake is either through your account name (provided by Snowflake) or a URL that uses the following format: `xxxxx.snowflakecomputing.com`" >
          <Input
            placeholder="xxxxxx.snowflakecomputing.com"
            type="string"
            className="width-30"
            value={jsonData.account || ''}
            onChange={this.onAccountChange}
          />
        </InlineField>
        <InlineField
          labelWidth={30}
          label="Username"
          tooltip="" >
          <Input
            placeholder="Username"
            type="string"
            className="width-20"
            onChange={this.onUsernameChange}
            value={jsonData.username || ''}
          />
        </InlineField>
        <InlineField
          labelWidth={30}
          label="basic or key pair authentication"
          tooltip="" >
          <InlineSwitch
            name="keyorpairauth"
            required
            value={jsonData.basicAuth ?? false}
            autoComplete="off"
            onChange={this.onAuthenticationChange}
          />
        </InlineField>

        {!jsonData.basicAuth && (
          <InlineField
            labelWidth={30}
            label="Password"
            tooltip="" >
            <SecretInput
              type="string"
              className="width-20"
              placeholder="password"
              isConfigured={(secureJsonFields && secureJsonFields.password) as boolean}
              value={secureJsonData.password || ''}
              onReset={this.onResetPassword}
              onChange={this.onPasswordChange}
            />
          </InlineField>
        )}
        {jsonData.basicAuth && (
          <InlineField
            labelWidth={30}
            label="Private key"
            tooltip="The private key must be encoded in base 64 URL encoded pkcs8 (remove PEM header '----- BEGIN PRIVATE KEY -----' and '----- END PRIVATE KEY -----', remove line space and replace '+' with '-' and '/' with '_')" >
            <SecretInput
              type="string"
              className="width-20"
              placeholder="MIIB..."
              isConfigured={(secureJsonFields && secureJsonFields.privateKey) as boolean}
              value={secureJsonData.privateKey || ''}
              onReset={this.onResetPrivateKey}
              onChange={this.onPrivateKeyChange}
            />
          </InlineField>
        )}
        <InlineField
          labelWidth={30}
          label="Role"
          tooltip="" >
          <Input
            type="string"
            className="width-20"
            onChange={this.onRoleChange}
            value={jsonData.role || ''}
            placeholder="Role"
          />
        </InlineField>

        <br />
        <h3 className="page-heading">Parameter configuration</h3>
        <InlineField
          labelWidth={30}
          label="Warehouse"
          tooltip="" >
          <Input
            type="string"
            className="width-20"
            onChange={this.onWarehouseChange}
            value={jsonData.warehouse || ''}
            placeholder="Default warehouse"
          />
        </InlineField>
        <InlineField
          labelWidth={30}
          label="Database"
          tooltip="" >
          <Input
            type="string"
            className="width-20"
            onChange={this.onDatabaseChange}
            value={jsonData.database || ''}
            placeholder="Default database"
          />
        </InlineField>
        <InlineField
          labelWidth={30}
          label="Schema"
          tooltip="" >
          <Input
            type="string"
            className="width-20"
            onChange={this.onSchemaChange}
            value={jsonData.schema || ''}
            placeholder="Default Schema"
          />
        </InlineField>
        <br />
        <h3 className="page-heading">Session configuration</h3>
        <InlineField
          labelWidth={30}
          label="Extra options"
          tooltip="" >
          <Input
            type="string"
            className="width-30"
            onChange={this.onExtraOptionChange}
            value={jsonData.extraConfig || ''}
            placeholder="TIMESTAMP_OUTPUT_FORMAT=MM-DD-YYYY&XXXXX=yyyyy&..."
          />
        </InlineField>
        <br />
        <h3 className="page-heading">Connection Pool configuration</h3>
        <InlineField
          labelWidth={30}
          label="max. open Connections"
          tooltip="How many connections will be opend from the datasource to snowflake (default: 100)" >
          <Input
            type="number"
            className="width-20"
            onChange={this.onMaxOpenConnectionsChange}
            value={jsonData.maxOpenConnections}
            placeholder="100"
          />
        </InlineField>
        <InlineField
          labelWidth={30}
          label="max. queued Queries"
          tooltip='How many queries will be put into the query queue. This should be higher as "max. open Connections" when more queries as set are waiting to be executed a "too many open queries" error will be thrown. (default: 400 | 0 = no limit)' >
          <Input
            type="number"
            className="width-20"
            onChange={this.onMaxQueuedQueriesChange}
            value={jsonData.maxQueuedQueries}
            placeholder="400"
          />
        </InlineField>
        <InlineField
          labelWidth={30}
          label="Connection lifetime [min]"
          tooltip="How long open connections are hold to be reused in minutes. (default=60 | 0=never close)" >
          <Input
            type="number"
            className="width-20"
            onChange={this.onConnectionLifetimeChange}
            value={jsonData.connectionLifetime}
            placeholder="60"
          />
        </InlineField>
        <br />
        <h3 className="page-heading">local Caching configuration</h3>
        <InlineField
          labelWidth={30}
          label="enable Caching"
          tooltip="Enable the Caching Backend in the Datasource. If similar sql-statements are queried thee result will be delivered out of cache." >
            <InlineSwitch
                name="useCaching"
                required
                value={jsonData.useCaching ?? false}
                autoComplete="off"
                onChange={this.onUseCachingChange}
              />
        </InlineField>
         <InlineField
          labelWidth={30}
          label="useCaching by default"
          tooltip="Always use caching for every Queries be default. No config Statement needed in the Query." >
          <InlineSwitch
                name="useCacheByDefault"
                required
                value={jsonData.useCacheByDefault ?? false}
                autoComplete="off"
                onChange={this.onUseCacheByDefaultChange}
              />
        </InlineField>
        <InlineField
          labelWidth={30}
          label="max CacheSize"
          tooltip="Size of the cache in MB. If exceed oldest queries are dropped. (default=2048)" >
          <Input
            type="number"
            className="width-20"
            onChange={this.onCacheSizeChange}
            value={jsonData.cacheSize}
            placeholder="2048"
          />
        </InlineField>
        <InlineField
          labelWidth={30}
          label="Cache Retention"
          tooltip="How long a query is hold in the cache in minutes. (default=60)" >
          <Input
            type="number"
            className="width-20"
            onChange={this.onCacheRetentionChange}
            value={jsonData.cacheRetention}
            placeholder="60"
          />
        </InlineField>
      </FieldSet >
    )

  }
}
