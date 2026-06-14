// Package version provides build-time version information.
package version

var (
	// Version is the semantic version, set at build time via -ldflags.
	Version = "dev"
	// Commit is the git commit hash, set at build time.
	Commit = "none"
	// Date is the build date, set at build time.
	Date = "unknown"
)
