package version

var (
	// Version of the tool
	Version = "develop"
	// Time is build time, injected at compile
	Time = "today"
	// Commit is git commit hash for a given checked out version, injected at compile time
	Commit = "git.commit"
)
