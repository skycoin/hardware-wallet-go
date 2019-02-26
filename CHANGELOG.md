# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Add an `entropy` field in `GenerateMnemonic`
- Mnemonic and recovery functions support `--wordCount` argument for the seed size (default `12`) .
- Add `-deviceType` flag and `DEVICE_TYPE` env var to set devicetype, options are `USB` or `EMULATOR`.

### Fixed

- Change protobuf messages for check signature to be consistent with [harware-wallet](https://github.com/skycoin/hardware-wallet/blob/2648cf384b5455c994ba54acf6a31cd1272c6f66/tiny-firmware/protob/messages.options#L21).

### Changed

### Removed
- Remove `GetEntropy` and `Entropy` msg.

### Fixed

### Security

