# Changelog

## 2.0.0 (not released yet)

### ⭐ Added
- Add a query tag that includes relevant Grafana context information.
- Add support of OAuth authentication.

### 🐞 Bug Fixes
- Source query variables are replaced with hardcoded values in the query editor UI.

### 🔨 Changed
- Rewrite the datasource configuration UI (ease authentication selection).
- Support non encoded private key in the datasource configuration.
- Update deprecated APIs
- Upgrade grafana-plugin-sdk-go to version v0.265.0.
- Upgrade gosnowflake to version v1.13.0.

## 1.9.1

### 🔨 Changed
- Remove deprecated UI components

## 1.9.0

### ⭐ Added
- Template Variable: custom “display names” support with `__text` & `__value`

### 🐞 Bug Fixes
- Resolve the issue with incorrect password escaping for certain special characters.

### 🔨 Changed
- Upgrade grafana-plugin-sdk-go to version v0.260.1.
- Upgrade js dependencies.

### ❤️ Contributors
We'd like to thank all the contributors who worked on this release!
- [@a2intl](https://github.com/a2intl)
- [@MrLight](https://github.com/MrLight)

## 1.8.1

### 🐞 Bug Fixes
- Resolve the issue of reusing closed connections.

### 🔨 Changed
- Upgrade grafana-plugin-sdk-go to version v0.258.0.
- Upgrade gosnowflake to version v1.12.0.
- Upgrade js dependencies.

## 1.8.0

### ⭐ Added
- Add configuration for `MaxChunkDownloadWorkers` and `CustomJSONDecoderEnabled`.
- Add `$__timeRoundFrom()` and `$__timeRoundTo()` macros

### 🐞 Bug Fixes
- Avoid interpolation of Snowflake SYSTEM functions.
- Improve error handling when specific macro parameters are left empty.

### 🔨 Changed
- Improve unit tests coverage.
- Upgrade grafana-plugin-sdk-go to version v0.255.0.
- Upgrade go to version 1.22.
- Upgrade js dependencies.

### ❤️ Contributors
We'd like to thank all the contributors who worked on this release!
- [@MrLight](https://github.com/MrLight)

## 1.7.1

### 🐞 Bug Fixes
- Use default variable format interpolation for query variables (see documentation).

## 1.7.0

### ⭐ Added
- Add `$__timeTzFilter(column_name)` macro.

### 🐞 Bug Fixes
- Fix issue with multi-value query variables.

### 🔨 Changed
- Upgrade grafana-plugin-sdk-go to version v0.246.0.
- Upgrade gosnowflake to version v1.11.1.

### ❤️ Contributors
We'd like to thank all the contributors who worked on this release!
- [@aryder-openai](https://github.com/ryder-openai)

## 1.6.1

### 🐞 Bug Fixes
- Fix Private Data Source Connect bug, that prevents the plugin update to function correctly.

### 🔨 Changed
- Upgrade grafana-plugin-sdk-go to version v0.228.0.
- Upgrade gosnowflake to version v1.9.0.

## 1.6.0

### ⭐ Added
- Use Grafana's Code editor for query editing.
- Add query context cancellation.
- Possibility to choose the fill mode for time series queries.

### 🐞 Bug Fixes
- Fix crash when time series data are empty.
- The slice_length must be an integer greater than or equal to 1.

### 🔨 Changed
- Use FillMode 'NULL' when fillMode is not provided.
- Upgrade GitHub actions used for CI/CD.
- Upgrade grafana-plugin-sdk-go to version v0.217.0.
- Upgrade go to version 1.21.
- Upgrade gosnowflake to version v1.8.0.

### ❤️ Contributors
We'd like to thank all the contributors who worked on this release!
- [@alexnederlof](https://github.com/alexnederlof)

## 1.5.0

### 🐞 Bug Fixes
- Do not limit time series data unless query contains LIMIT clause.

### 🔨 Changed
- Improve timezone conversion in macro $__timeFilter()
- Use to_timestamp_ntz instead of to_timestamp
- Upgrade grafana-plugin-sdk-go to version v0.197.0.
- Upgrade js dependencies.
- Upgrade gosnowflake to version v1.7.1.

### ❤️ Contributors
We'd like to thank all the contributors who worked on this release!
- [@kranthikirang](https://github.com/kranthikirang)
- [@rumbin](https://github.com/rumbin)

## 1.4.1

### 🐞 Bug Fixes
- Fix issue with REAL type with Snowflake go SDK (1.6.14)
- Fix issue with NULL type

### ❤️ Contributors
We'd like to thank all the contributors who worked on this release!
- [@benesch](https://github.com/benesch)

## 1.4.0

### ⭐ Added
- Add darwin_arm64 binary
- Upgrade grafana-plugin-sdk-go to version v0.148.0.
- Upgrade js dependencies

## 1.3.0

### 🔨 Changed
- Fix LIMIT condition to avoid duplicate LIMIT keyword.
- Increase row limit to 1M
- Improve time-series wide column name
- Upgrade grafana-plugin-sdk-go to version v0.141.0.
- Upgrade js dependencies

### ❤️ Contributors
We'd like to thank all the contributors who worked on this release!
- [@kranthikirang](https://github.com/kranthikirang)

## 1.2.0
### ⭐ Added
- Add query refId in response

### 🔨 Changed
- Convert Long Frame to wide
- Fix issue with Time Formatted Columns
- Improve metadata in response
- Improve macros
- Upgrade grafana-plugin-sdk-go to version v0.139.0.
- Upgrade gosnowflake to version v1.6.13.
- Upgrade js dependencies

### ❤️ Contributors
We'd like to thank all the contributors who worked on this release!
- [@heanlan](https://github.com/heanlan)

## 1.1.0

### ⭐ Added
- Support of Key Pair Authentication.
- Add support for query variables in Snowflake data source.

### 🔨 Changed
- Rework connection string creation.
- Escape credential, segments and query parameters in connection string.
- Add LIMIT cause only for time series
- Upgrade grafana-plugin-sdk-go to version v0.134.0.
- Upgrade gosnowflake to version v1.6.9.

### ❤️ Contributors
We'd like to thank all the contributors who worked on this release!
- [@inacionery](https://github.com/inacionery)

## 1.0.0

Initial release.
