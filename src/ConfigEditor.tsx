import React, { ChangeEvent, PureComponent } from 'react';
import {Checkbox, ControlledCollapse, LegacyForms, RadioButtonGroup, InlineLabel } from '@grafana/ui';
import {DataSourcePluginOptionsEditorProps} from '@grafana/data';
import { SnowflakeOptions, SnowflakeSecureOptions } from './types';

const { SecretFormField, FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<SnowflakeOptions> {}

interface State {
  authMethod: string;
}

const authOptions = [
  { label: 'Password', value: 'password' },
  { label: 'Key Pair', value: 'keyPair' },
  { label: 'OAuth', value: 'oauth' },
];

export class ConfigEditor extends PureComponent<Props, State> {

  state: State = {
    authMethod: authOptions[0].value,
  };

  onAuthMethodChange = (value: string) => {
    const { onOptionsChange, options } = this.props;
    const authMethod = value || 'password';
    this.setState({ authMethod: authMethod });
    const jsonData = {
      ...options.jsonData,
      authMethod,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onAccountChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;

    let value = event.target.value.trim();
    if (!value.includes('.snowflakecomputing.com')) {
      value += '.snowflakecomputing.com';
    }

    // Sanitize value to avoid error
    value = value.replace(/^https?:\/\//, '');

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
        token: ''
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
    let privateKey = event.target.value;

    // If the private key is not in the correct format, try to convert it
    if (!/^[A-Za-z0-9\-_]+$/.test(privateKey) && privateKey !== '') {

      // Remove the PEM header and footer
      privateKey = privateKey.replace(/-----BEGIN PRIVATE KEY-----|-----END PRIVATE KEY-----/g, '');

      // Remove all newline and space characters
      privateKey = privateKey.replace(/\n|\r|\s/g, '');

      // Replace + with - and / with _
      privateKey = privateKey.replace(/\+/g, '-').replace(/\//g, '_');
    }
    onOptionsChange({
      ...options,
      secureJsonData: {
        privateKey: privateKey,
        password: '',
        token: ''
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

  onTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        token: event.target.value,
        privateKey: '',
        password: ''
      },
    });
  };

  onResetToken = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        token: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        token: ''
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as SnowflakeSecureOptions;
    const { authMethod } = this.state;

    return (
        <div className="gf-form-group">
          <h3 className="page-heading">Connection</h3>

          <div className="gf-form">
            <FormField
                label="Account name"
                labelWidth={10}
                inputWidth={30}
                onChange={this.onAccountChange}
                tooltip="All access to Snowflake is either through your account name (provided by Snowflake) or a URL that uses the following format: `xxxxx.snowflakecomputing.com`"
                value={jsonData.account || ''}
                placeholder="xxxxxx.snowflakecomputing.com"
            />
          </div>

          <div className="gf-form">
            <InlineLabel width={20}>Authentications method</InlineLabel>
            <RadioButtonGroup
                options={authOptions}
                value={authOptions.find((option) => option.value === authMethod)?.value}
                onChange={this.onAuthMethodChange}
            />
          </div>

          { authMethod !== 'oauth' && (
              <div className="gf-form">
                    <FormField
                    label="Username"
                    labelWidth={10}
                    inputWidth={20}
                    onChange={this.onUsernameChange}
                    value={jsonData.username || ''}
                    placeholder="Username"
                    tooltip="The snowflake account username"
                />
              </div>
          )}
          <div className="gf-form">
            {authMethod === 'password' && (
                  <SecretFormField
                      isConfigured={(secureJsonFields && secureJsonFields.password) as boolean}
                      value={secureJsonData.password || ''}
                      label="Password"
                      placeholder="password"
                      labelWidth={10}
                      inputWidth={20}
                      onReset={this.onResetPassword}
                      onChange={this.onPasswordChange}
                      tooltip="The snowflake account password"
                  />
            )}
            {authMethod === 'keyPair' && (
                <SecretFormField
                    isConfigured={(secureJsonFields && secureJsonFields.privateKey) as boolean}
                    value={secureJsonData.privateKey || ''}
                    tooltip="The private key"
                    label="Private key"
                    placeholder="-----BEGIN PRIVATE KEY-----"
                    labelWidth={10}
                    inputWidth={20}
                    onReset={this.onResetPrivateKey}
                    onChange={this.onPrivateKeyChange}
                />
            )}
            {authMethod === 'oauth' && (
                <SecretFormField
                    isConfigured={(secureJsonFields && secureJsonFields.token) as boolean}
                    value={secureJsonData.token || ''}
                    tooltip="Oauth token"
                    label="Oauth token"
                    placeholder="eyJhbGciOiJ..."
                    labelWidth={10}
                    inputWidth={20}
                    onReset={this.onResetToken}
                    onChange={this.onTokenChange}
                />
            )}
          </div>
          <div className="gf-form">
            <FormField
                label="Role"
                labelWidth={10}
                inputWidth={20}
                onChange={this.onRoleChange}
                value={jsonData.role || ''}
                placeholder="Role"
                tooltip="Global role to use for the connection"
            />
          </div>
          <br/>
          <h3 className="page-heading">Parameter configuration</h3>

          <div className="gf-form">
            <FormField
                label="Warehouse"
                labelWidth={10}
                inputWidth={20}
                onChange={this.onWarehouseChange}
                value={jsonData.warehouse || ''}
                placeholder="Default warehouse"
                tooltip="Warehouse to use for the connection"
            />
          </div>

          <div className="gf-form">
            <FormField
                label="Database"
                labelWidth={10}
                inputWidth={20}
                onChange={this.onDatabaseChange}
                value={jsonData.database || ''}
                placeholder="Default database"
                tooltip="Database to use for the connection"
            />
          </div>

          <div className="gf-form">
            <FormField
                label="Schema"
                labelWidth={10}
                inputWidth={20}
                onChange={this.onSchemaChange}
                value={jsonData.schema || ''}
                placeholder="Default Schema"
                tooltip="Schema to use for the connection"
            />
          </div>
          <br/>
          <h3 className="page-heading">Session configuration</h3>

          <div className="gf-form">
            <FormField
                label="Extra options"
                labelWidth={10}
                inputWidth={30}
                onChange={this.onExtraOptionChange}
                value={jsonData.extraConfig || ''}
                placeholder="TIMESTAMP_OUTPUT_FORMAT=MM-DD-YYYY&XXXXX=yyyyy&..."
                tooltip="Extra connection parameters to use for the connection"
            />
          </div>
          <br/>
          <ControlledCollapse label="Experimental">
            <div className="gf-form">
              <FormField
                  label="Max Chunk Download Workers"
                  labelWidth={15}
                  inputWidth={3}
                  onChange={this.onMaxChunkDownloadWorkersChange}
                  value={jsonData.maxChunkDownloadWorkers || '10'}
              />
            </div>
            <br/>
            <div className="gf-form">
                <Checkbox
                    value={jsonData.customJSONDecoderEnabled}
                    onChange={this.onCustomJSONDecoderEnabledChange}
                    label="Enable Custom JSON Decoder"
                    />
            </div>
          </ControlledCollapse>
        </div>
    );
  }
}
