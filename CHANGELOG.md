# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
- Fix `500 Invalid WRP content type` [#74](https://github.com/xmidt-org/wrp-go/pull/74)

## [v3.1.2]
- Move ParseID func and relevant consts from webpa-common to wrp-go. [#75](https://github.com/xmidt-org/wrp-go/pull/75)

## [v3.1.1]
- Fix bug so that error encoder sends a 400 when decoding fails [#70](https://github.com/xmidt-org/wrp-go/pull/70)

## [v3.1.0]
- Added enum values to `MessageTypes`. Bumped codecgen version to 1.2.6. Now install stringer and codecgen everytime `go generate` is run. [#69](https://github.com/xmidt-org/wrp-go/pull/69)
- Added missing wrp fields to header[#68](https://github.com/xmidt-org/wrp-go/pull/68)

## [v3.0.2]
- Updated references to the main branch [#52](https://github.com/xmidt-org/wrp-go/pull/52)
- Add constants for the different supported MIME types [#58](https://github.com/xmidt-org/wrp-go/pull/58)

## [v3.0.1]
- Upgrade self import paths to /v3 [#49](https://github.com/xmidt-org/wrp-go/pull/49)

## [v3.0.0]
- As a breaking change, `wrphttp.ResponseWriter`'s `WriteWRP` function now takes a `*wrphttp.Entity` type instead of an `interface{}`. 

- `Format` and `WriteWRPBytes` were introduced as additional functions to `wrphttp.ResponseWriter` to offer higher API flexibility.

All changes included in [#47](https://github.com/xmidt-org/wrp-go/pull/47)

## [v2.0.1]
- Fix bug introduced in v2.0.0 for missing logic to populate new wrp entity field [#46](https://github.com/xmidt-org/wrp-go/pull/46)]

## [v2.0.0]
- Changed folder structure to bring go files into the root directory [#32](https://github.com/xmidt-org/wrp-go/pull/32)
- Updated travis to automate releases [#40](https://github.com/xmidt-org/wrp-go/pull/40)
- Use json tag instead of wrp tag [#42](https://github.com/xmidt-org/wrp-go/pull/42)
- Extend WRPHandler Request Adapter and Decoder [#43](https://github.com/xmidt-org/wrp-go/pull/43)
- Added SessionID field to wrp SimpleEvent and Message [#45](https://github.com/xmidt-org/wrp-go/pull/45)

## [v1.3.4]
- Bumped webpa-common to v1.3.2
- Removed glide files

## [v1.3.3]
- Updated module name to be correct
- upgraded codec to v1.1.7

## [v1.3.2]
- Moved from glide to go modules

## [v1.3.1]
- Bump webpa-common to v1.3.1

## [v1.3.0]
- Enabled PartnerID and Metadata to be translated to/from HTTP headers.

## [v1.2.0]
- Updated codec version

## [v1.1.0]
- Fixed circular dependencies with webpa-common
- Added swagger doc comments
- Added Unknown message type
- Fixed imports upon move to a new github org

## [1.0.0]
- This release is exactly the same as the last version from github.com/xmidt-org/webpa-common/wrp

[Unreleased]: https://github.com/xmidt-org/wrp-go/compare/v3.1.2...HEAD
[v3.1.2]: https://github.com/xmidt-org/wrp-go/compare/v3.1.1...v3.1.2
[v3.1.1]: https://github.com/xmidt-org/wrp-go/compare/v3.1.0...v3.1.1
[v3.1.0]: https://github.com/xmidt-org/wrp-go/compare/v3.0.2...v3.1.0
[v3.0.2]: https://github.com/xmidt-org/wrp-go/compare/v3.0.1...v3.0.2
[v3.0.1]: https://github.com/xmidt-org/wrp-go/compare/v3.0.0...v3.0.1
[v3.0.0]: https://github.com/xmidt-org/wrp-go/compare/v2.0.1...v3.0.0
[v2.0.1]: https://github.com/xmidt-org/wrp-go/compare/v2.0.0...v2.0.1
[v2.0.0]: https://github.com/xmidt-org/wrp-go/compare/v1.3.4...v2.0.0
[v1.3.4]: https://github.com/xmidt-org/wrp-go/compare/v1.3.3...v1.3.4
[v1.3.3]: https://github.com/xmidt-org/wrp-go/compare/v1.3.2...v1.3.3
[v1.3.2]: https://github.com/xmidt-org/wrp-go/compare/v1.3.1...v1.3.2
[v1.3.1]: https://github.com/xmidt-org/wrp-go/compare/v1.3.0...v1.3.1
[v1.3.0]: https://github.com/xmidt-org/wrp-go/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/xmidt-org/wrp-go/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/xmidt-org/wrp-go/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/xmidt-org/wrp-go/compare/v0.0.0...v1.0.0
