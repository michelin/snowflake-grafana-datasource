import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps} from '@grafana/data';
import { SnowflakeOptions, SnowflakeSecureOptions } from './types';

const { SecretFormField, FormField, Switch } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<SnowflakeOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onAccountChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;

    var value;
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

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as SnowflakeSecureOptions;

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
          <FormField
            label="Username"
            labelWidth={10}
            inputWidth={20}
            onChange={this.onUsernameChange}
            value={jsonData.username || ''}
            placeholder="Username"
          />
        </div>

        <div className="gf-form">
          <Switch label="basic or key pair authentication"
                  checked={jsonData.basicAuth}
                  onChange={this.onAuthenticationChange} />
        </div>
        <div className="gf-form">
        { !jsonData.basicAuth &&
              <SecretFormField
                  isConfigured={(secureJsonFields && secureJsonFields.password) as boolean}
                  value={secureJsonData.password || ''}
                  label="Password"
                  placeholder="password"
                  labelWidth={10}
                  inputWidth={20}
                  onReset={this.onResetPassword}
                  onChange={this.onPasswordChange}
              />
        }
        { jsonData.basicAuth &&
              <SecretFormField
                  isConfigured={(secureJsonFields && secureJsonFields.privateKey) as boolean}
                  value={secureJsonData.privateKey || ''}
                  tooltip="The private key must be encoded in base 64 URL encoded pkcs8 (remove PEM header '----- BEGIN PRIVATE KEY -----' and '----- END PRIVATE KEY -----', remove line space and replace '+' with '-' and '/' with '_')"
                  label="Private key"
                  placeholder="MIIB..."
                  labelWidth={10}
                  inputWidth={20}
                  onReset={this.onResetPrivateKey}
                  onChange={this.onPrivateKeyChange}
              />
        }
        </div>

        <div className="gf-form">
          <FormField
            label="Role"
            labelWidth={10}
            inputWidth={20}
            onChange={this.onRoleChange}
            value={jsonData.role || ''}
            placeholder="Role"
          />
        </div>
        <br />
        <h3 className="page-heading">Parameter configuration</h3>

        <div className="gf-form">
          <FormField
            label="Warehouse"
            labelWidth={10}
            inputWidth={20}
            onChange={this.onWarehouseChange}
            value={jsonData.warehouse || ''}
            placeholder="Default warehouse"
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
          />
        </div>
        <br />
        <h3 className="page-heading">Session configuration</h3>

        <div className="gf-form">
          <FormField
            label="Extra options"
            labelWidth={10}
            inputWidth={30}
            onChange={this.onExtraOptionChange}
            value={jsonData.extraConfig || ''}
            placeholder="TIMESTAMP_OUTPUT_FORMAT=MM-DD-YYYY&XXXXX=yyyyy&..."
          />
        </div>
      </div>
    );
  }
}
