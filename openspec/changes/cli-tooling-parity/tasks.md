# Tasks

## 1. CLI foundation

- [ ] 1.1 Add the CLI package, root command, help, version, and exit-code mapping; verify `--help` and `version`
- [ ] 1.2 Add text/JSON output writers and config precedence; verify precedence tests

## 2. Inspection commands

- [ ] 2.1 Implement info; verify counts on a fixture
- [ ] 2.2 Implement audit with `--fix`; verify clean and repairable files
- [ ] 2.3 Implement header and layers; verify listing and setting

## 3. Transformation commands

- [ ] 3.1 Implement convert; verify target version output
- [ ] 3.2 Implement explode; verify composite entities are exploded
- [ ] 3.3 Implement strip; verify comments and proxy options

## 4. Rendering and query

- [ ] 4.1 Implement draw for SVG/PDF/PNG; verify output files
- [ ] 4.2 Implement query with text/JSON; verify valid JSON output

## 5. Safety and completion

- [ ] 5.1 Implement atomic writes and backups; verify original preservation
- [ ] 5.2 Add shell completion generation; verify completion output

## 6. Verification

- [ ] 6.1 Add CLI integration tests running the built binary against fixtures; verify exit codes and output
- [ ] 6.2 Document the CLI in `docs/` and update `docs/parity/README.md`

## Workflow follow-up

- Archive the change after review.
