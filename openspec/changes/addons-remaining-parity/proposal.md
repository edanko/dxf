# Proposal

## Why

Several ezdxf bundled add-ons are still absent. Some are inspection helpers, some generate geometry, and some delegate to external tools. Users migrating from ezdxf expect the same add-on entry points.

## What Changes

- Add browser/acisbrowser data extraction helpers.
- Add DWG conversion add-on through an external ODA converter with actionable errors.
- Add dimension-line, r12export, and xqt add-ons.
- Add generative add-ons: genetic algorithm, Menger sponge, Sierpinski pyramid.
- Add mixins that match ezdxf behavior.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `addons`: add the missing bundled add-ons.

## Impact

Affected packages: `addons/*`, command runner abstractions. No existing add-on APIs change.
