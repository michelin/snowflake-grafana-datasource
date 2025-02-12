import React, {ChangeEvent, PureComponent} from 'react';
import {
  Checkbox,
  ControlledCollapse,
  InlineField,
  Input,
  RadioButtonGroup,
  SecretInput,
  SecretTextArea,
} from '@grafana/ui';
import {DataSourcePluginOptionsEditorProps} from '@grafana/data';
import {SnowflakeOptions, SnowflakeSecureOptions} from './types';

interface Props extends DataSourcePluginOptionsEditorProps<SnowflakeOptions> {}

interface State {
  authMethod: string;
}

const authOptions = [
  { label: 'Password', value: 'password' },
  { label: 'Key Pair', value: 'keyPair' },
  { label: 'OAuth', value: 'oauth' },
];

const LABEL_WIDTH = 30
const INPUT_WIDTH = 50

export class ConfigEditor extends PureComponent<Props, State> {
  searchParams = new URLSearchParams(location.search);

  constructor(props: Props) {
    super(props);
    this.state = {
      authMethod: this.props.options.jsonData.authMethod ?? authOptions[0].value,
    };
  }

  onAuthMethodChange = (value: string) => {
    const { onOptionsChange, options } = this.props;
    const authMethod = value ?? 'password';
    this.setState({ authMethod: authMethod });
    const jsonData = {
      ...options.jsonData,
      authMethod,
    };

    onOptionsChange({
      ...options,
      jsonData,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
        privateKey: false,
        clientSecret: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
        privateKey: '',
        clientSecret: '',
      },
    });
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
        ...options.secureJsonData,
        password: event.target.value,
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
    let privateKey = event.target.value;

    // If the private key is not in the correct format, try to convert it
    if (!/^[A-Za-z0-9\-_=]+$/.test(privateKey) && privateKey !== '') {

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
        ...options.secureJsonData,
        privateKey: privateKey,
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

  onClientIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      clientId: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onTokenEndpointChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      tokenEndpoint: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onClientSecretChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        clientSecret: event.target.value,
      },
    });
  };

  onResetClientSecret = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        clientSecret: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        clientSecret: ''
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData ?? {}) as SnowflakeSecureOptions;
    const { authMethod } = this.state;

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

        <InlineField label="Authentications method"
                     labelWidth={LABEL_WIDTH}>
            <RadioButtonGroup
                  options={authOptions}
                  value={authMethod}
                  onChange={this.onAuthMethodChange}
              />
        </InlineField>

        { authMethod !== 'oauth' && (
            <InlineField label="Username"
                         tooltip="The snowflake account username"
                         labelWidth={LABEL_WIDTH}>
                <Input
                    onChange={this.onUsernameChange}
                    value={jsonData.username ?? ''}
                    placeholder="Username"
                    width={INPUT_WIDTH}
                />
            </InlineField>
        )}
        {authMethod === 'password' && (
            <InlineField label="Password"
                         labelWidth={LABEL_WIDTH}>
                <SecretInput
                    isConfigured={secureJsonFields?.password}
                    value={secureJsonData.password ?? ''}
                    placeholder="password"
                    width={INPUT_WIDTH}
                    onReset={this.onResetPassword}
                    onChange={this.onPasswordChange}
                />
            </InlineField>
        )}
        {authMethod === 'keyPair' && (
            <InlineField label="Private key"
                         tooltip="The private key must be encoded in base 64"
                         labelWidth={LABEL_WIDTH}>
                <SecretTextArea
                    isConfigured={secureJsonFields?.privateKey}
                    value={secureJsonData.privateKey ?? ''}
                    placeholder="MIIB..."
                    onReset={this.onResetPrivateKey}
                    onChange={this.onPrivateKeyChange}
                    cols={38}
                    rows={5}
                />
            </InlineField>
        )}
        {authMethod === 'oauth' && (
          <div>
              <InlineField label="Client ID"
                           tooltip="Oauth client ID"
                           labelWidth={LABEL_WIDTH}>
                  <Input
                      value={jsonData.clientId ?? ''}
                      width={INPUT_WIDTH}
                      onChange={this.onClientIdChange}
                  />
              </InlineField>
              <InlineField label="Client Secret"
                           tooltip="Oauth Client Secret"
                           labelWidth={LABEL_WIDTH}>
                <SecretInput
                    isConfigured={secureJsonFields?.clientSecret}
                    value={secureJsonData.clientSecret ?? ''}
                    width={INPUT_WIDTH}
                    onReset={this.onResetClientSecret}
                    onChange={this.onClientSecretChange}
                />
              </InlineField>
              <InlineField label="Token endpoint"
                           tooltip="Oauth token endpoint"
                           labelWidth={LABEL_WIDTH}>
                <Input
                    value={jsonData.tokenEndpoint ?? ''}
                    width={INPUT_WIDTH}
                    onChange={this.onTokenEndpointChange}
                />
              </InlineField>
          </div>
        )}
        <InlineField label="Role"
                     tooltip="Global role to use for the connection. With Oauth, it's used to limit the access token to a single role that the user can consent to for the session."
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
                     tooltip="Warehouse to use for the connection"
                     labelWidth={LABEL_WIDTH}>
            <Input
                width={INPUT_WIDTH}
                onChange={this.onWarehouseChange}
                value={jsonData.warehouse ?? ''}
                placeholder="Default warehouse"
            />
        </InlineField>

        <InlineField label="Database"
                     tooltip="Database to use for the connection"
                     labelWidth={LABEL_WIDTH}>
            <Input
                width={INPUT_WIDTH}
                onChange={this.onDatabaseChange}
                value={jsonData.database ?? ''}
                placeholder="Default database"
            />
        </InlineField>

        <InlineField label="Schema"
                     tooltip="Schema to use for the connection"
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
                     tooltip="Extra connection parameters to use for the connection"
                     labelWidth={LABEL_WIDTH}>
          <Input
              width={INPUT_WIDTH}
              onChange={this.onExtraOptionChange}
              value={jsonData.extraConfig ?? ''}
              placeholder="TIMESTAMP_OUTPUT_FORMAT=MM-DD-YYYY&..."
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
