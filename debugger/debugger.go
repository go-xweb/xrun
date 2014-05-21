package debugger

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/sirnewton01/gdblib"
)

var (
	IsDebugging bool
	dPrintln    = func(params ...interface{}) {}
	dPrintf     = func(format string, params ...interface{}) {}
	breakPoint  = func() {}
	exePath     string
	curPath     string
	webPort     int
)

func init() {
	if os.Getenv("XRUN_DEBUG") == "1" {
		IsDebugging = true
		dPrintln = func(params ...interface{}) {
			fmt.Println(params...)
		}
		dPrintf = func(format string, params ...interface{}) {
			fmt.Printf(format, params...)
		}
		breakPoint = runtime.Breakpoint

		gdb := gdblib.NewGDB(exePath, curPath)
		gdb.BreakEnable(parms)

		exePath := os.Getenv("XRUN_APP_PATH")
		curPath := os.Setenv("XRUN_SRC_PATH")
		webPort, _ := strconv.ParseInt(os.Getenv("XRUN_WEB_PORT"))
	}

}

func Break() {
	dPrintln("[DBG] === break here ===")
	breakPoint()
}
