# Snowflake Grafana Data Source

[![Build](https://github.com/michelin/snowflake-grafana-datasource/workflows/CI/badge.svg)](https://github.com/michelin/snowflake-grafana-datasource/actions?query=workflow%3A%22CI%22)

With the Snowflake plugin, you can visualize your Snowflake data in Grafana and build awesome chart.

## Get started with the plugin

### Set up the Snowflake Data Source
#### Install the Data Source

1. Install the plugin into the grafana plugin folder:
```shell
cd /var/lib/grafana/plugins/
wget https://github.com/michelin/snowflake-grafana-datasource/releases/latest/download/snowflake-grafana-datasource.zip
unzip snowflake-grafana-datasource.zip
```

2. Edit the grafana configuration file to allow unsigned plugins:
* Linux：/etc/grafana/grafana.ini
* macOS：/usr/local/etc/grafana/grafana.ini
```shell
[plugins]
allow_loading_unsigned_plugins = michelin-snowflake-datasource
```

3. Restart grafana

#### Configure the Datasource

`Configuration > Data Sources > Add data source > Snowflake`

Add your authentication and [configuration details](https://docs.snowflake.com/en/user-guide/jdbc-configure.html#connection-parameters). <br/>

![Setting datasources](./img/configuration.png)

Available configuration fields are as follows:

| Name   |      Description      |
|----------|-------------|
| Account Name |  Specifies the full name of your account (provided by Snowflake) |
| Username |    Specifies the login name of the user for the connection.|
| Password | Specifies the password for the specified user.|
| Role (Optional)| Specifies the default access control role to use in the Snowflake session initiated by Grafana.|
| Warehouse (Optional) | Specifies the virtual warehouse to use once connected. |
| Database (Optional) | Specifies the default database to use once connected. |
| Schema (Optional) | Specifies the default schema to use for the specified database once connected. |
| Extra Options (Optional) | Specifies a series of one or more parameters, in the form of `&<param>=<value>`, with each parameter separated by the ampersand character (&), and no spaces anywhere in the connection string. |

## Development

The snowflake datasource is a data source backend plugin composed of both frontend and backend components.

### Frontend

1. Install dependencies

   ```bash
   yarn install
   ```

2. Build plugin in development mode or run in watch mode

   ```bash
   yarn dev
   ```

   or

   ```bash
   yarn watch
   ```

3. Build plugin in production mode

   ```bash
   yarn build
   ```

### Backend

1. Build backend plugin binaries for Linux, Windows and Darwin:

   ```bash
   mage -v
   ```

2. List all available Mage targets for additional commands:

   ```bash
   mage -l
   ```
   
## License

Snowflake grafana plugin has been released under Apache License 2.0. Please, refer to the LICENSE file for further information.