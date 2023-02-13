package debug

import (
	_debug "runtime/debug"
)

// Run will only call fn if DEBUG is enabled.
func Run(fn func()) {
	if DEBUG {
		fn()
	}
}

// BuildInfo will return a useful new-line separated build info string for current binary, setting name as given value.
func BuildInfo(name string) string {
	// Read build info from current binary
	build, ok := _debug.ReadBuildInfo()
	if !ok {
		return "name=" + name + "\n"
	}

	var flags, vcs, commit, time string

	// Parse build information from BuildInfo.Settings
	for i := 0; i < len(build.Settings); i++ {
		switch build.Settings[i].Key {
		case "-gcflags":
			flags += ` -gcflags="` + build.Settings[i].Value + `"`
		case "-ldflags":
			flags += ` -ldflags="` + build.Settings[i].Value + `"`
		case "-tags":
			flags += ` -tags="` + build.Settings[i].Value + `"`
		case "vcs":
			vcs = build.Settings[i].Value
		case "vcs.revision":
			commit = build.Settings[i].Value
			if len(commit) > 8 {
				commit = commit[:8]
			}
		case "vcs.time":
			time = build.Settings[i].Value
		}
	}

	return "" +
		"name=" + name + "\n" +
		"vcs=" + vcs + "\n" +
		"commit=" + commit + "\n" +
		"version=" + build.Main.Version + "\n" +
		"path=" + build.Path + "\n" +
		"build=" + build.GoVersion + flags + "\n" +
		"time=" + time + "\n"
}
