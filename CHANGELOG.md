# Changelog

All notable changes to this project are documented here, newest first.

Entries are generated from [Conventional Commits](https://www.conventionalcommits.org)
and grouped by change type. This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.2.0]: https://github.com/toaweme/log/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/toaweme/log/releases/tag/v0.1.0
