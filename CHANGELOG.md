# Changelog

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
