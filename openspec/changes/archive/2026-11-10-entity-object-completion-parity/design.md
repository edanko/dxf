# Design

## Approach

Extend the existing loader registry rather than adding a parallel parser. Each typed loader consumes `lldxf.Tags`, calls the shared base loader, and stores unhandled tags for `Format` output. ACAD_TABLE parsing follows ezdxf's row/cell group-code layout but only exposes stable fields; unknown cell data stays raw.

ACIS decoding is intentionally shallow: detect SAT versus SAB, read the version header when present, and record body/history presence. Full geometric modeling of ACIS entities is out of scope because it requires a SAT kernel.

## Risks

- ACAD_TABLE variants differ across versions and third-party writers; raw-tag retention is the fallback.
- Typed objects-as-entities must not change their section on save.
- Collection members are handles, not embedded objects; resolution failures must be reported, not silently dropped.

## Testing

Use existing fixtures plus new ACAD_TABLE, polyface, mesh, ACIS, and collection fixtures. Extend the ezdxf cross-check harness to cover typed object inventories.
