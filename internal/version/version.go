package version

import (
	"fmt"
	"runtime/debug"
)

func PrintVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("idk")
	}
	fmt.Println(info.Main.Version)
}
