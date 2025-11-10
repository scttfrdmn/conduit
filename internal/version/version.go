package version

import "fmt"

// These variables are set at build time via ldflags
var (
	// Version is the semantic version
	Version = "dev"
	// Commit is the git commit hash
	Commit = "unknown"
	// Date is the build date
	Date = "unknown"
	// BuiltBy is who built the binary
	BuiltBy = "unknown"
)

// Info represents version information
type Info struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

// Get returns the version information
func Get() Info {
	return Info{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
		BuiltBy: BuiltBy,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("conduit version %s (commit: %s, built: %s by %s)",
		i.Version, i.Commit, i.Date, i.BuiltBy)
}
