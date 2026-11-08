# Tasks

## 1. Colors

- [ ] 1.1 Add the generated ACI table and nearest-index lookup; verify documented samples
- [ ] 1.2 Add the tagged color value with ACI, true color, BYLAYER, and BYBLOCK; verify round trips
- [ ] 1.3 Implement effective color resolution with block and layer context; verify BYBLOCK/BYLAYER tests
- [ ] 1.4 Implement transparency representation and preservation; verify alpha round trip

## 2. Enums

- [ ] 2.1 Add typed enums with name/numeric conversion for entity, table, and object values; verify round trips
- [ ] 2.2 Preserve and report unknown enum values; verify invalid fixture

## 3. Units

- [ ] 3.1 Add unit definitions and conversion factors; verify mm-to-m and other conversions
- [ ] 3.2 Map units to `$INSUNITS` and handle unitless explicitly; verify header round trip
- [ ] 3.3 Add linetype, lineweight, and text-style defaults and validation; verify defaults

## 4. Integration

- [ ] 4.1 Use the shared color/enum/unit types in entity, table, render, and export; verify existing tests pass
- [ ] 4.2 Add boundary validation with reporting; verify invalid ACI and enum values
- [ ] 4.3 Update `docs/parity/README.md` color/enum/unit matrix

## Workflow follow-up

- Archive the change after review.
