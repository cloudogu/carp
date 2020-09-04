# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- base-url configuration option
- resource-path configuration option

### Changed
- Deliver resources if browsers request them directly and they are available anonymously; #9

## [v1.0.0] - 2020-07-01
### Changed
- Changed logger to go-logging instead of glog
- Make log-level configurable
