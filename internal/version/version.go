package version

import (
	"fmt"
	"runtime/debug"
)

func GetVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("idk")
	}
	return info.Main.Version
}
