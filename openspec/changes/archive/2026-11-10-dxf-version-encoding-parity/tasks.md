# Tasks

## 1. Version model

- [x] 1.1 Add a version profile table for R12 through R2018 and verify each profile with table-driven tests
- [x] 1.2 Detect `$ACADVER` on load and expose the source and normalized version; verify with fixtures per version
- [x] 1.3 Implement normalization of pre-R12 and R13/R14 input and verify the load report

## 2. Encoding

- [x] 2.1 Add codepage-aware ASCII decoding on load and verify non-ASCII fixtures
- [x] 2.2 Add UTF-8 output for R2007+ and codepage output for older targets; verify round trips
- [x] 2.3 Complete the binary writer (OBJECTS and EOF) and verify ASCII/binary equivalence tests

## 3. Conversion

- [x] 3.1 Implement save to a requested target version and verify the declared `$ACADVER`
- [x] 3.2 Implement the R12 sanitizer and conversion report; verify downgrade of an R2018 fixture
- [x] 3.3 Implement upgrade/downgrade round-trip checks and record unsupported content

## 4. Verification

- [x] 4.1 Add cross-version fixtures for R12, R2000, R2007, R2013, and binary; verify load and save
- [x] 4.2 Run the full suite and document remaining version differences in `docs/parity/README.md`

## Workflow follow-up

- Update the parity matrix for versions and encoding.
- Archive the change after review.
