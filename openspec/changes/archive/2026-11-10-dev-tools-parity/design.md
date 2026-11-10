# Design

## Approach

Build developer tools on existing loader/audit APIs. `analyze` reuses section inventories; `difftags` walks raw tag trees rather than reserializing files; `debug` exposes recovery reports. Standards rules are small predicates over tables, entities, and header variables. ZIP support wraps `archive/zip` with atomic write and safe extraction.

## Testing

Add CLI integration tests for each command and unit tests for rule evaluation, tag diff paths, and ZIP traversal rejection.
