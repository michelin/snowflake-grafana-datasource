# Changelog

## [Unreleased]
- Fix issue with Time Formatted Columns
- Improve metadata in response
- Add query refId in response
- Upgrade grafana-plugin-sdk-go to version v0.139.0.
- Upgrade gosnowflake to version v1.6.13.

## 1.1.0

### Added
- Support of Key Pair Authentication.
- Add support for query variables in Snowflake data source.

### Changed
- Rework connection string creation.
- Escape credential, segments and query parameters in connection string.
- Add LIMIT cause only for time series
- Upgrade grafana-plugin-sdk-go to version v0.134.0.
- Upgrade gosnowflake to version v1.6.9.

## 1.0.0

Initial release.
