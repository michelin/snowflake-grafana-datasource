import React, { ChangeEvent, PureComponent } from 'react';
import {
  Checkbox,
  ControlledCollapse,
  InlineField,
  Input,
  SecretInput,
  SecretTextArea,
  Switch
} from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { SnowflakeOptions, SnowflakeSecureOptions } from './types';

interface Props extends DataSourcePluginOptionsEditorProps<SnowflakeOptions> {}

interface State {}

const LABEL_WIDTH = 30
const INPUT_WIDTH = 40

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
    const regex = /https?:\/\//;
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

  onMaxChunkDownloadWorkersChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      maxChunkDownloadWorkers: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onCustomJSONDecoderEnabledChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      customJSONDecoderEnabled: event.target.checked,
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

  onAuthenticationChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      basicAuth: (event.target as HTMLInputElement).checked,
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

  onPrivateKeyChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
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

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as SnowflakeSecureOptions;

    return (
        <div className="gf-form-group">
          <h3 className="page-heading">Connection</h3>

          <InlineField label="Account name"
                       tooltip="All access to Snowflake is either through your account name (provided by Snowflake) or a URL that uses the following format: `xxxxx.snowflakecomputing.com`"
                       labelWidth={LABEL_WIDTH}>
            <Input
                onChange={this.onAccountChange}
                value={jsonData.account ?? ''}
                placeholder="xxxxxx.snowflakecomputing.com"
                width={INPUT_WIDTH}
            />
          </InlineField>
          <InlineField label="Username"
                       labelWidth={LABEL_WIDTH}>
            <Input
                onChange={this.onUsernameChange}
                value={jsonData.username ?? ''}
                placeholder="Username"
                width={INPUT_WIDTH}
            />
          </InlineField>
          <InlineField label="basic or key pair authentication"
                       style={{alignItems: 'center'}}
                       labelWidth={LABEL_WIDTH}>
            <Switch
                checked={jsonData.basicAuth ?? false}
                style={{ textAlign: 'center' }}
                onChange={this.onAuthenticationChange}
            />
          </InlineField>
          {!jsonData.basicAuth && (
              <InlineField label="Password"
                           labelWidth={LABEL_WIDTH}>
                <SecretInput
                    isConfigured={(secureJsonFields?.password) as boolean}
                    value={secureJsonData.password ?? ''}
                    placeholder="password"
                    width={INPUT_WIDTH}
                    onReset={this.onResetPassword}
                    onChange={this.onPasswordChange}
                />
              </InlineField>
          )}
          {jsonData.basicAuth && (
              <InlineField label="Private key"
                           tooltip="The private key must be encoded in base 64 URL encoded pkcs8 (remove PEM header '----- BEGIN PRIVATE KEY -----' and '----- END PRIVATE KEY -----', remove line space and replace '+' with '-' and '/' with '_')"
                           labelWidth={LABEL_WIDTH}>
                <SecretTextArea
                    isConfigured={(secureJsonFields?.privateKey) as boolean}
                    value={secureJsonData.privateKey ?? ''}
                    placeholder="MIIB..."
                    onReset={this.onResetPrivateKey}
                    onChange={this.onPrivateKeyChange}
                    cols={38}
                    rows={5}
                />
              </InlineField>
          )}
          <InlineField label="Role"
                       labelWidth={LABEL_WIDTH}>
            <Input
                width={INPUT_WIDTH}
                onChange={this.onRoleChange}
                value={jsonData.role ?? ''}
                placeholder="Role"
            />
          </InlineField>
          <br/>
          <h3 className="page-heading">Parameter configuration</h3>

          <InlineField label="Warehouse"
                       labelWidth={LABEL_WIDTH}>
            <Input
                width={INPUT_WIDTH}
                onChange={this.onWarehouseChange}
                value={jsonData.warehouse ?? ''}
                placeholder="Default warehouse"
            />
          </InlineField>

          <InlineField label="Database"
                       labelWidth={LABEL_WIDTH}>
            <Input
                width={INPUT_WIDTH}
                onChange={this.onDatabaseChange}
                value={jsonData.database ?? ''}
                placeholder="Default database"
            />
          </InlineField>

          <InlineField label="Schema"
                       labelWidth={LABEL_WIDTH}>
            <Input
                width={INPUT_WIDTH}
                onChange={this.onSchemaChange}
                value={jsonData.schema ?? ''}
                placeholder="Default Schema"
            />
          </InlineField>
          <br/>
          <h3 className="page-heading">Session configuration</h3>

          <InlineField label="Extra options"
                       labelWidth={LABEL_WIDTH}>
            <Input
                width={INPUT_WIDTH}
                onChange={this.onExtraOptionChange}
                value={jsonData.extraConfig ?? ''}
                placeholder="TIMESTAMP_OUTPUT_FORMAT=MM-DD-YYYY&XXXXX=yyyyy&..."
            />
          </InlineField>
          <br/>
          <ControlledCollapse label="Experimental">
            <InlineField label="Max Chunk Download Workers"
                         labelWidth={LABEL_WIDTH}>
              <Input
                  width={INPUT_WIDTH}
                  onChange={this.onMaxChunkDownloadWorkersChange}
                  value={jsonData.maxChunkDownloadWorkers ?? '10'}
              />
            </InlineField>
            <br/>
            <InlineField label="Enable Custom JSON Decoder"
                         style={{alignItems: 'center'}}
                         labelWidth={LABEL_WIDTH}>
              <Checkbox
                  value={jsonData.customJSONDecoderEnabled ?? false}
                  onChange={this.onCustomJSONDecoderEnabledChange}
              />
            </InlineField>
          </ControlledCollapse>
        </div>
    );
  }
}
