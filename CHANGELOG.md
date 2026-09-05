# Changelog

All notable changes to this project are documented here, newest first.

Entries are generated from [Conventional Commits](https://www.conventionalcommits.org)
and grouped by change type. This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-09-05

### Fixes

- Pin care action by commit sha, not tag object by [@iberflow](https://github.com/iberflow) in [da2f42a](https://github.com/toaweme/log/commit/da2f42a3fddd8e92fa747203439d971d68e7e2fc).

### Documentation

- Link to docs site from README by [@iberflow](https://github.com/iberflow) in [8a703d7](https://github.com/toaweme/log/commit/8a703d74cf0403b188c5615dabfedb7a3020e4ca).
- Add Contributing section to README by [@iberflow](https://github.com/iberflow) in [3e8cebf](https://github.com/toaweme/log/commit/3e8cebf837007a1467ee71b011cd776e10857520).

### Refactors

- Collapse the level onto one process-wide source and drop the filter handler by [@iberflow](https://github.com/iberflow) in [#6](https://github.com/toaweme/log/pull/6).

### CI & Build

- Pin release build to go 1.26 instead of stable by [@iberflow](https://github.com/iberflow) in [5f629e2](https://github.com/toaweme/log/commit/5f629e2e2cb327c23fd0f8e5178a166fc238c351).
- Pin quality gate to go 1.26 instead of stable by [@iberflow](https://github.com/iberflow) in [e75f6a1](https://github.com/toaweme/log/commit/e75f6a15d91fdc541cba52559439d87b69f39794).
- Publish code.json via codereport action by [@iberflow](https://github.com/iberflow) in [#2](https://github.com/toaweme/log/pull/2).
- Bump care action to v0.9.3 by [@iberflow](https://github.com/iberflow) in [712d2a4](https://github.com/toaweme/log/commit/712d2a4f72f5f2ac302c048925bb164ad02b3eaa).
- Move action version to inline comment so dependabot maintains it by [@iberflow](https://github.com/iberflow) in [d202e69](https://github.com/toaweme/log/commit/d202e69b32cb30c3cfbd390419141a0529850ba7).
- Drop gomod dependabot block from dependency-free module by [@iberflow](https://github.com/iberflow) in [a94f0b1](https://github.com/toaweme/log/commit/a94f0b12968cd46b0a65236695afde4e2fdb145a).
- Add governance workflows and contributor docs by [@iberflow](https://github.com/iberflow) in [cd5b9bf](https://github.com/toaweme/log/commit/cd5b9bf7327a8ce73d3060225a8cadf35fa098d3).

### Chores & Other

- Linter fix by [@iberflow](https://github.com/iberflow) in [246d275](https://github.com/toaweme/log/commit/246d275bfe5b6afd6dbb8f734654ac1a18b1ce30).
- Relicense from MIT to Apache 2.0 by [@iberflow](https://github.com/iberflow) in [d3624a2](https://github.com/toaweme/log/commit/d3624a2680a54bd446f24ca781155491bf797f88).

## [0.2.1] - 2026-07-01

### CI & Build

- Bump care to v0.8.0 by [@iberflow](https://github.com/iberflow) in [85429c8](https://github.com/toaweme/log/commit/85429c86f7660d558568b8cd0815551744e5938c).
- Bump care to v0.7.1 and pin to commit sha by [@iberflow](https://github.com/iberflow) in [cca3fe8](https://github.com/toaweme/log/commit/cca3fe8f17bdeb2e18172167c1080250735d6c35).
- Bump care to v0.6.0 and fix card-svg dark/light wiring by [@iberflow](https://github.com/iberflow) in [c5b8d4c](https://github.com/toaweme/log/commit/c5b8d4cfce337c80e577f2d5d6b7cb8364b2c7d8).

### Chores & Other

- Align README, CHANGELOG, and quality workflow with org standards by [@iberflow](https://github.com/iberflow) in [47b55fe](https://github.com/toaweme/log/commit/47b55fee7096107611ee7d7e0a9fd1b25cabd165).

## [0.2.0] - 2026-06-15

### Fixes

- Guard filter race, clamp negative shorten limit, simplify wildcard match by [@iberflow](https://github.com/iberflow) in [6c71f54](https://github.com/toaweme/log/commit/6c71f545920be5cff7fe6fb660ddfdb6a3ff1873).

### Refactors

- Add Discard logger, rename Options->HandlerOptions and FilteredLogger->FilterHandler by [@iberflow](https://github.com/iberflow) in [15a1b3b](https://github.com/toaweme/log/commit/15a1b3bb11aae648b56602cd90ba985662470e8c).

### Chores & Other

- Update readme and old naming in tests by [@iberflow](https://github.com/iberflow) in [cfb1d8b](https://github.com/toaweme/log/commit/cfb1d8ba0385ef2e4d5ff4ec0ba4171deb510035).

## [0.1.0] - 2026-06-13

### Features

- Http tracing header constants by [@iberflow](https://github.com/iberflow) in [ea15d43](https://github.com/toaweme/log/commit/ea15d4316a13c802ea2980524f82dcd8dbf89c53).
- Trace and fatal levels by [@iberflow](https://github.com/iberflow) in [08602eb](https://github.com/toaweme/log/commit/08602ebb4ec55e5faf4ef93cc57cb7123b61f9be).
- NewMultiHandler constructor by [@iberflow](https://github.com/iberflow) in [abf302c](https://github.com/toaweme/log/commit/abf302c388440dd18246dc58bf813b558569baa3).
- Filtered logger by [@iberflow](https://github.com/iberflow) in [3fb7e76](https://github.com/toaweme/log/commit/3fb7e76c486cf99544b8f1a40549acff4d139543).
- Extended logger by [@iberflow](https://github.com/iberflow) in [e35bade](https://github.com/toaweme/log/commit/e35bade2e9f5713f0e2c9863d7497f2ca64a62c3).
- Filters can modify attributes by [@iberflow](https://github.com/iberflow) in [da5cdc0](https://github.com/toaweme/log/commit/da5cdc0886645cbce29f67be125aa86f735f6d28).
- Filter logs with * suffix by [@iberflow](https://github.com/iberflow) in [9d49cbb](https://github.com/toaweme/log/commit/9d49cbb58efbbdfcdd138ae059b6fd27a2fc738f).
- Filter logs by [@iberflow](https://github.com/iberflow) in [4506ca7](https://github.com/toaweme/log/commit/4506ca7f93bdd755204ee4ccb820f5709b323184).
- WithLevel by [@iberflow](https://github.com/iberflow) in [c657186](https://github.com/toaweme/log/commit/c65718604628b20d49871acf6149addedbdd4b1d).

### Fixes

- Filtered logger by [@iberflow](https://github.com/iberflow) in [a91f3b7](https://github.com/toaweme/log/commit/a91f3b77b7c5926e0c066de3835945627b26c954).

### Refactors

- Move gin and desktop logging code to separate packages by [@iberflow](https://github.com/iberflow) in [57b5258](https://github.com/toaweme/log/commit/57b5258f9fa66a8c674ddf17c372c4dc4909c475).
- Unexport filterAction by [@iberflow](https://github.com/iberflow) in [fd91c9f](https://github.com/toaweme/log/commit/fd91c9f125cb17225dee64ef3a1c2d3fd4eabbe2).

### Chores & Other

- Initial commit :) by [@iberflow](https://github.com/iberflow) in [3757101](https://github.com/toaweme/log/commit/3757101ac5cf70cd836e036743f8a1980b03ebb7).
- Remove header names by [@iberflow](https://github.com/iberflow) in [8a64ef1](https://github.com/toaweme/log/commit/8a64ef109ff2b1372970cf7dace1340246d1ded6).
- Cleanup module by [@iberflow](https://github.com/iberflow) in [42ce4ab](https://github.com/toaweme/log/commit/42ce4ab975ff8a31b0256fd3146f3eaf74229e90).

[0.3.0]: https://github.com/toaweme/log/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/toaweme/log/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/toaweme/log/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/toaweme/log/releases/tag/v0.1.0
