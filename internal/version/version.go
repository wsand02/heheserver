package version

import (
	"runtime/debug"
)

// unknownVersion is returned when build info is unavailable (for example when
// the binary was not built as a module).
const unknownVersion = "unknown"

func GetVersion() string {
	info, ok := debug.ReadBuildInfo()
	return versionFrom(info, ok)
}

// versionFrom derives the version string from the result of
// debug.ReadBuildInfo. It is split out so the ok == false / nil-info path can
// be tested without depending on how the test binary was built.
func versionFrom(info *debug.BuildInfo, ok bool) string {
	if !ok || info == nil || info.Main.Version == "" {
		return unknownVersion
	}
	return info.Main.Version
}
