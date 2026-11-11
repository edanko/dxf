# Spec Delta

## ADDED Requirements

### Requirement: Julian date conversion matches civil date
The system SHALL convert between civil date-time values and Julian dates using the documented astronomical convention.

#### Scenario: Known epoch
- **WHEN** the civil date 2000-01-01 12:00 UTC is converted
- **THEN** the Julian date is 2451545.0 within one second

#### Scenario: Round trip
- **WHEN** a civil date-time is converted to Julian date and back
- **THEN** the original date-time is recovered within one second
