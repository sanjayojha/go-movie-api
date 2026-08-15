package vcs

import "runtime/debug"

func Version() string {
	buildInfo, ok := debug.ReadBuildInfo()

	if ok {
		return buildInfo.Main.Version
	}
	return ""
}
