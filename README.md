# Snowflake Grafana Data Source
[![Build](https://github.com/michelin/snowflake-grafana-datasource/actions/workflows/test.yml/badge.svg)](https://github.com/michelin/snowflake-grafana-datasource/actions/workflows/test.yml) 
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=michelin_snowflake-grafana-datasource&metric=ncloc)](https://sonarcloud.io/summary/new_code?id=michelin_snowflake-grafana-datasource) [![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=michelin_snowflake-grafana-datasource&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=michelin_snowflake-grafana-datasource)[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=michelin_snowflake-grafana-datasource&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=michelin_snowflake-grafana-datasource) [![Coverage](https://sonarcloud.io/api/project_badges/measure?project=michelin_snowflake-grafana-datasource&metric=coverage)](https://sonarcloud.io/summary/new_code?id=michelin_snowflake-grafana-datasource)

With the Snowflake plugin, you can visualize your Snowflake data in Grafana and build awesome chart.

## Get started with the plugin

### Set up the Snowflake Data Source
#### Install the Data Source

1. Install the plugin into the grafana plugin folder:

**With grafana-cli**
```shell
grafana cli --pluginUrl https://github.com/michelin/snowflake-grafana-datasource/releases/latest/download/snowflake-grafana-datasource.zip plugins install michelin-snowflake-datasource
```
`--pluginsDir` option can be used to specify a custom plugin directory

**Manually**
```shell
cd /var/lib/grafana/plugins/
wget https://github.com/michelin/snowflake-grafana-datasource/releases/latest/download/snowflake-grafana-datasource.zip
unzip snowflake-grafana-datasource.zip
```

2. Edit the grafana configuration file `grafana.ini` to allow unsigned plugins:
* Linux：/etc/grafana/grafana.ini
* macOS：/usr/local/etc/grafana/grafana.ini
```shell
[plugins]
allow_loading_unsigned_plugins = michelin-snowflake-datasource
```
Or with docker
```shell
docker run -d \
-p 3000:3000 \
-v "$(pwd)"/grafana-plugins:/var/lib/grafana/plugins \
--name=grafana \
-e "GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=michelin-snowflake-datasource" \
grafana/grafana
```

> [!NOTE]
> Please refer to the documentation for more details.
https://grafana.com/docs/grafana/latest/administration/plugin-management/#allow-unsigned-plugins

3. Restart grafana
Restart the Grafana server to apply the changes:
``` bash
service grafana-server restart
```

#### Configure the Datasource

* Open the side menu by clicking the Grafana icon in the top header.
* In the side menu under the Configuration icon you should find a link named Data Sources.
* Click the `+ Add data source` button in the top header.
* Select Snowflake.

Add your authentication and [configuration details](https://docs.snowflake.com/en/user-guide/jdbc-configure.html#connection-parameters). <br/>

![Setting datasources](./img/configuration.png)

Available configuration fields are as follows:

| Name                      | Description                                                                                                                                                                                                    |
|---------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Account Name              | Specifies the full name of your account (provided by Snowflake)                                                                                                                                                |
| Username                  | Specifies the login name of the user for the connection.                                                                                                                                                       |
| Password                  | Specifies the password for the specified user.                                                                                                                                                                 |
| Private key               | Specifies the private key in PEM (PKCS8 format).                                                                                                                                                               |
| Programmatic Access Token | Specifies the Programmatic Access Token for token-based authentication.                                                                                                                                        |
| Client Id                 | Specifies the Oauth client ID.                                                                                                                                                                                 |
| Client Secret             | Specifies the Oauth client Secret.                                                                                                                                                                             |
| Token Endpoint            | Specifies the Oauth Token endpoint.                                                                                                                                                                            |
| Role (Optional)           | Specifies the default access control role to use in the Snowflake session initiated by Grafana. With Oauth, it's used to limit the access token to a single role that the user can consent to for the session. |
| Warehouse (Optional)      | Specifies the virtual warehouse to use once connected.                                                                                                                                                         |
| Database (Optional)       | Specifies the default database to use once connected.                                                                                                                                                          |
| Schema (Optional)         | Specifies the default schema to use for the specified database once connected.                                                                                                                                 |
| Extra Options (Optional)  | Specifies a series of one or more parameters, in the form of `<param>=<value>`, with each parameter separated by the ampersand character (&), and no spaces anywhere in the connection string.                 |

**External OAuth authentication**

> [!NOTE]
> Snowflake oauth authentication is not supported without external service (like Okta, Azure Entra, Keycloak ...) because of the lack of support for oauth Client credentials flow in snowflake.
https://docs.snowflake.com/en/user-guide/oauth-intro

The plugin supports OAuth authentication with snowflake only with external_service.<br/>
To use OAuth, you need to create an [external OAuth](https://docs.snowflake.com/en/user-guide/oauth-ext-custom) integration in your Snowflake account.
```sql
-- Create a security integration for external OAuth flow
CREATE OR REPLACE SECURITY INTEGRATION OAUTH_INTEGRATION
TYPE = EXTERNAL_OAUTH
ENABLED = TRUE
EXTERNAL_OAUTH_TYPE = CUSTOM
EXTERNAL_OAUTH_SCOPE_MAPPING_ATTRIBUTE = 'scope'
EXTERNAL_OAUTH_TOKEN_USER_MAPPING_CLAIM = 'name'
EXTERNAL_OAUTH_SNOWFLAKE_USER_MAPPING_ATTRIBUTE = 'login_name'
EXTERNAL_OAUTH_ALLOWED_ROLES_LIST = ('<xxxxx>')
EXTERNAL_OAUTH_AUDIENCE_LIST =('https://xxxxx')
EXTERNAL_OAUTH_RSA_PUBLIC_KEY = 'MIIBIj'
EXTERNAL_OAUTH_ISSUER = 'https://xxxxx';
```

#### Supported Macros

Macros can be used within a query to simplify syntax and allow for dynamic parts.

Macro example                                          | Description
------------------------------------------------------ | -------------
`$__time(dateColumn)`                                  | Will be replaced by an expression to convert to a UNIX timestamp and rename the column to `time`. For example, *TRY_TO_TIMESTAMP_NTZ(dateColumn) as time*
`$__timeEpoch(dateColumn)`                             | Will be replaced by an expression to convert to a UNIX timestamp and rename the column to `time`.
`$__timeFilter(dateColumn)`                            | Will be replaced by a time range filter using the specified column name, which is expected to be a timestamp without time zone. For example, *time > CONVERT_TIMEZONE('UTC', 'UTC', '2023-12-13T23:21:44Z'::timestamp_ntz) AND time < CONVERT_TIMEZONE('UTC', 'UTC', '2023-12-13T23:22:44Z'::timestamp_ntz)*
`$__timeFilter(dateColumn, timezone)`                  | Will be replaced by a time range filter using the specified column name, which is expected to be a timestamp without time zone. For example, *time > CONVERT_TIMEZONE('UTC', 'America/New_York', '2023-12-13T23:22:46Z'::timestamp_ntz) AND time < CONVERT_TIMEZONE('UTC', 'America/New_York', '2023-12-13T23:23:46Z'::timestamp_ntz)*
`$__timeTzFilter(dateColumn)`                          | Will be replaced by a time range filter using the specified column name, which is expected to be a timestamp with time zone. For example, *time > '2023-12-13T23:21:44Z'::timestamp_tz AND time < '2023-12-13T23:22:44Z'::timestamp_tz*
`$__timeFrom()`                                        | Will be replaced by the start of the currently active time selection. For example, *1494410783*
`$__timeTo()`                                          | Will be replaced by the end of the currently active time selection. For example, *1494410983*
`$__timeGroup(dateColumn,'5m')`                        | Will be replaced by an expression usable in GROUP BY clause. For example, *floor(extract(epoch from dateColumn)/120)*120*
`$__timeGroup(dateColumn,'5m', 0)`                     | Same as above but with a fill parameter so missing points in that series will be added by grafana and 0 will be used as value.
`$__timeGroup(dateColumn,'5m', NULL)`                  | Same as above but NULL will be used as value for missing points.
`$__timeGroup(dateColumn,'5m', previous)`              | Same as above but the previous value in that series will be used as fill value if no value has been seen yet NULL will be used (only available in Grafana 5.3+).
`$__timeGroup(dateColumn, interval, fill, timezone)`   | Same as above but with a timezone parameter, which will be used to convert the timestamp in dateColumn to the specified timezone before slicing.
`$__timeGroupAlias(dateColumn,'5m')`                   | Will be replaced identical to $__timeGroup but with an added column alias `time` (only available in Grafana 5.3+).
`$__unixEpochFilter(dateColumn)`                       | Will be replaced by a time range filter using the specified column name with times represented as Unix timestamp. For example, *dateColumn > 1494410783 AND dateColumn < 1494497183*
`$__unixEpochNanoFilter(dateColumn)`                   | Will be replaced by a time range filter using the specified column name with times represented as nanosecond timestamp. For example, *dateColumn > 1494410783152415214 AND dateColumn < 1494497183142514872*
`$__unixEpochNanoFrom()`                               | Will be replaced by the start of the currently active time selection as nanosecond timestamp. For example, *1494410783152415214*
`$__unixEpochNanoTo()`                                 | Will be replaced by the end of the currently active time selection as nanosecond timestamp. For example, *1494497183142514872*
`$__unixEpochGroup(dateColumn,'5m', [fillmode])`       | Same as $__timeGroup but for times stored as Unix timestamp (only available in Grafana 5.3+).
`$__unixEpochGroupAlias(dateColumn,'5m', [fillmode])`  | Same as above but also adds a column alias (only available in Grafana 5.3+).
`$__timeRoundFrom(d duration in minutes)`              | The result of rounding __timeFrom() down to a multiple of d. [default d: 15] -- Will round the time to the last full quarter. $__timeRoundFrom(5) will round time to the last full 5 minutes.
`$__timeRoundTo(d duration in minutes)`                | The result of rounding __timeTo() up to a multiple of d. [default d: 15] -- Will round the time to the next full quarter. $__timeRoundTo(5) will round time to the next full 5 minutes.

#### Write Queries

Create a panel in a dashboard and select a Snowflake Data Source to start using the query editor.

Select a query type 'Time Series' or 'Table'.

For Time series query:
* Date / time column can appear anywhere in the query as long as it is included (you must define the time formatted columns name)
* A numerical column must be included.

![Query editor](img/query.png)

> [!CAUTION]
> This plugin cannot identify malicious code in queries executed on Snowflake and assumes no responsibility for their execution. As a precaution, use a ROLE with minimal privileges, configured to grant read-only access

##### Query Variables

You can use query variable in your Snowflake queries by using [variable syntax](https://grafana.com/docs/grafana/latest/dashboards/variables/variable-syntax/).<br/>
You can also set the [interpolation format](https://grafana.com/docs/grafana/latest/dashboards/variables/variable-syntax/#general-syntax) in the variable.<br/>

Single-value variables usage:
```sql
-- $variable = 'xxxxxxx'  
SELECT column FROM table WHERE column = ${variable:sqlstring}
-- Interpolation result
SELECT column FROM table WHERE column = 'xxxxxxx'
```

Single-value variables with format usage:
```sql
-- $variable = 'xxxxxxx'       
SELECT column FROM table WHERE column = ${variable:raw}
-- nterpolation result
SELECT column FROM table WHERE column = xxxxxxx
```

Multiple-value variables usage:
```sql
-- $variable = ['xxxxxxx','yyyyyy']        
SELECT column FROM table WHERE column in (${variable:sqlstring})
-- Interpolation result
SELECT column FROM table WHERE column in ('xxxxxxx','yyyyyy')
```

Multiple-value variables with format usage:
```sql
-- $variable = ['xxxxxxx','yyyyyy']  
SELECT column FROM table WHERE column in ${variable:regex}
-- Interpolation result
SELECT column FROM table WHERE column in (test1|test2)
```

Add a Template Variable:<br/>
You can use a SQL Query to define a [Template Variable](https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#add-a-query-variable)<br/>
```sql
-- Add __text and __value columns to your query to support custom "display names"
SELECT value_column as __value, text_column as __text FROM table 
```

##### Layout of a query

*Simple query*
```sql
SELECT
  <time_column> as time,
  <numerical_column>
FROM
  <table>
WHERE
  $__timeFilter(<time_column>) 
  AND $<variable> = 'xxxxx' -- custom grafana variables start with dollar sign
```

*SQL Query Group By Interval*

```sql
SELECT
  $__timeGroup(<time_column>, $__interval), -- group time by interval
  <numerical_column>
FROM
  <table>
WHERE
  $__timeFilter(<time_column>)
  AND $<variable> = 'xxxxx'
GROUP BY 
  <time_column>
```


#### Create an annotation

Annotations allow you to overlay events on a graph.
To create an annotation, in the dashboard settings click "Annotations", and "New".

#### Oauth Configuration

To use Oauth, you need to create an Oauth custom integration in your Snowflake account.<
You can follow the steps in the [Snowflake documentation](https://docs.snowflake.com/en/user-guide/oauth-custom).

## Caching
### Snowflake caching

Snowflake caches queries with the same footprint / hash in its own query-cache. Since a Grafana query mostly has a now() component the cache will never be used.
To get more queries with the same hash use the two macros `$__timeRoundFrom(d)` and `$__timeRoundTo(d)` to create wider truncated timestamps. This is no problem for timeseries charts. Grafana cuts it's x-Axis to the selected dashboard time window. If a table is displayed the whole result will be presented and it could be slightly out of the time window.\
More info about snowflake-side caching: https://docs.snowflake.com/en/user-guide/querying-persisted-results#retrieval-optimization

## Supported Grafana Versions
This plugin supports only version with [Active Support from Grafana](https://grafana.com/docs/grafana/next/upgrade-guide/when-to-upgrade/?pg=blog&plcmt=body-txt#what-to-know-about-version-support).

## Development

The snowflake datasource is a data source backend plugin composed of both frontend and backend components.

To build the project you must have the following tools installed:
* Go
* Node
* yarn


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

1. Install dependencies

   ```bash
   go mod tidy
   ```

2. Build backend plugin binaries for Linux, Windows and Darwin:

   ```bash
   mage -v
   ```

3. List all available Mage targets for additional commands:

   ```bash
   mage -l
   ```
   
## License

Snowflake grafana plugin has been released under Apache License 2.0. Please, refer to the LICENSE.txt file for further information.
