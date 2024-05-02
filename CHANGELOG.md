# Changelog

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
