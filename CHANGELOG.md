# Changelog

## [0.2.0](https://github.com/flipt-io/flipt-openfeature-provider-go/compare/v0.1.5...v0.2.0) (2023-08-17)

### Changed (Breaking)

* correlates Boolean flag evaluations to Boolean flag types on the Flipt server. Upgrading to this version will require you to convert your flags that were using Boolean evaluation to the Boolean flag type on the Flipt server.

## [0.1.5](https://github.com/flipt-io/flipt-openfeature-provider-go/compare/v0.1.4...v0.1.5) (2023-05-24)

### Features

* remove transport details from user and determine the transport under the hood ([a8a04a9](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/a8a04a9fff502089a310ad6fce11ef777a3d6af5))
* use flipt sdk for openfeature interactions, and include new namespace client ([ce1fe31](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/ce1fe31ae0b0ac9b37ffdde32a9ceb79c9d0d2c6))

### Bug Fixes

* remove getFlag to subtract extra network hop ([9f6b53d](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/9f6b53d0aaf408d5b00146379b79101d2402155d))

## [0.1.4](https://github.com/flipt-io/flipt-openfeature-provider-go/compare/v0.1.3...v0.1.4) (2022-11-02)

### Miscellaneous Chores

* release 0.2.0 ([ca791ce](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/ca791ce42541a1484003818ded77017eed1ef8f9))

## [0.1.5](https://github.com/flipt-io/flipt-openfeature-provider-go/compare/v0.1.4...v0.1.5) (2023-05-24)

### Features

* remove transport details from user and determine the transport under the hood ([a8a04a9](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/a8a04a9fff502089a310ad6fce11ef777a3d6af5))
* use flipt sdk for openfeature interactions, and include new namespace client ([ce1fe31](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/ce1fe31ae0b0ac9b37ffdde32a9ceb79c9d0d2c6))

### Bug Fixes

* remove getFlag to subtract extra network hop ([9f6b53d](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/9f6b53d0aaf408d5b00146379b79101d2402155d))

## [0.1.4](https://github.com/flipt-io/flipt-openfeature-provider-go/compare/v0.1.3...v0.1.4) (2022-11-02)

### Bug Fixes

* discard unknown fields when unmarshaling json ([#14](https://github.com/flipt-io/flipt-openfeature-provider-go/issues/14)) ([4bc5504](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/4bc55047454d03bc2f40e595899f36f5fde7c7b2))

## [0.1.3](https://github.com/flipt-io/flipt-openfeature-provider-go/compare/v0.1.2...v0.1.3) (2022-11-01)

### Features

* **otel:** propagate trace context through grpc and http client ([#12](https://github.com/flipt-io/flipt-openfeature-provider-go/issues/12)) ([47897df](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/47897dfffdb1b62677399d35e129d02b1ea359e9))

## [0.1.2](https://github.com/flipt-io/flipt-openfeature-provider-go/compare/v0.1.1...v0.1.2) (2022-10-31)

### Bug Fixes

* **http:** dont use path.Join when scheme is present ([285bec6](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/285bec60cbe4bf80ad1f111af60f21a5f39817b0))

## [0.1.1](https://github.com/flipt-io/flipt-openfeature-provider-go/compare/v0.1.0...v0.1.1) (2022-10-28)

### Bug Fixes

* fix bool eval reason, fix cover task, more test coverage ([#7](https://github.com/flipt-io/flipt-openfeature-provider-go/issues/7)) ([eca4aef](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/eca4aefd2bab5188829e7bfde5c0d8234ecccb2b))

## 0.1.0 (2022-10-27)

### Miscellaneous Chores

* release 0.1.0 ([b476662](https://github.com/flipt-io/flipt-openfeature-provider-go/commit/b4766629c66465ffb9a710e4b2158a6cedea93f1))
