package format

// VersionProfile describes the capabilities and restrictions of one DXF release.
type VersionProfile struct {
	// Version is the $ACADVER code, e.g. AC1009, AC1015, AC1024.
	Version string
	// Release is the human-readable release name, e.g. R12, R2007.
	Release string
	// SupportsObjects true when OBJECTS section is supported (R13+).
	SupportsObjects bool
	// SupportsClasses true when CLASSES section is supported (R13+).
	SupportsClasses bool
	// SupportsBinary true when binary output is supported (R14+).
	SupportsBinary bool
	// UTF8 true when text should be encoded as UTF-8 (R2007+).
	UTF8 bool
	// SupportsTrueColor true when true color (group codes 256+) is supported.
	SupportsTrueColor bool
	// SupportsTransparency true when transparency is supported.
	SupportsTransparency bool
	// DropHeaderCodes lists group codes that must not be emitted for this
	// release (e.g. R12 cannot store true-color header variables).
	DropHeaderCodes map[int]bool
}

// VersionMap maps an $ACADVER code to its profile.
var VersionMap = map[string]*VersionProfile{
	AC1009: {Version: AC1009, Release: "R12", SupportsObjects: false, SupportsClasses: false, SupportsBinary: false, UTF8: false, DropHeaderCodes: map[int]bool{420: true}},
	AC1014: {Version: AC1014, Release: "R14", SupportsObjects: true, SupportsClasses: true, SupportsBinary: true, UTF8: false},
	AC1015: {Version: AC1015, Release: "R2000", SupportsObjects: true, SupportsClasses: true, SupportsBinary: true, UTF8: false},
	AC1018: {Version: AC1018, Release: "R2004", SupportsObjects: true, SupportsClasses: true, SupportsBinary: true, UTF8: false},
	AC1021: {Version: AC1021, Release: "R2007", SupportsObjects: true, SupportsClasses: true, SupportsBinary: true, UTF8: true, SupportsTrueColor: true},
	AC1024: {Version: AC1024, Release: "R2010", SupportsObjects: true, SupportsClasses: true, SupportsBinary: true, UTF8: true, SupportsTrueColor: true, SupportsTransparency: true},
	AC1027: {Version: AC1027, Release: "R2013", SupportsObjects: true, SupportsClasses: true, SupportsBinary: true, UTF8: true, SupportsTrueColor: true, SupportsTransparency: true},
	AC1032: {Version: AC1032, Release: "R2018", SupportsObjects: true, SupportsClasses: true, SupportsBinary: true, UTF8: true, SupportsTrueColor: true, SupportsTransparency: true},
}

// Lookup returns the profile for an $ACADVER code, or nil if unknown.
func Lookup(version string) *VersionProfile {
	return VersionMap[version]
}

// IsSupportedVersion reports whether the version has a profile.
func IsSupportedVersion(version string) bool {
	_, ok := VersionMap[version]
	return ok
}

// Profile returns the profile for the given version. It never returns nil;
// unknown versions fall back to R2000 (the lowest fully writable version).
func Profile(version string) *VersionProfile {
	if p, ok := VersionMap[version]; ok {
		return p
	}
	return VersionMap[AC1015]
}

// SupportedVersions returns all supported version codes.
func SupportedVersions() []string {
	return []string{AC1009, AC1014, AC1015, AC1018, AC1021, AC1024, AC1027, AC1032}
}

// NormalizeVersion maps unsupported versions to the nearest writable one.
// pre-R12 and R13 map to writable neighbors; everything else passes through.
func NormalizeVersion(version string) (normalized string, changed bool) {
	switch version {
	case AC1009:
		return AC1009, false
	case "AC1006", "AC1010", "AC1012", "AC1013":
		return AC1009, true // pre-R12 -> R12
	case AC1014:
		return AC1015, true // R14 -> R2000
	default:
		if _, ok := VersionMap[version]; ok {
			return version, false
		}
		return AC1015, true // unknown -> R2000
	}
}
