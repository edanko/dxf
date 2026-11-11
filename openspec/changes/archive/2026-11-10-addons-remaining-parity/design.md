# Design

## Approach

Implement pure-Go inspection and generator add-ons directly. External-tool add-ons follow the existing `odafc` pattern: a `CommandRunner` interface, explicit executable configuration, and no partial output on failure. Browser add-ons return plain data structures so CLI and GUI layers can render them.

Generators emit normal entities through the graphics factory so they inherit handle/owner/layer behavior.

## Testing

Test generator counts and bounding boxes, browser data shape, missing-tool errors, and mixin attribute behavior.
