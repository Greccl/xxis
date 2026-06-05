package main

import (
   "os"
	"github.com/spf13/pflag"
	xxisError "github.com/Greccl/xxis/internal/errors"
	// xxisRuntime "github.com/Greccl/xxis/internal/runtime"
	xxisParser "github.com/Greccl/xxis/internal/parser"
	xxisDebug "github.com/Greccl/xxis/internal/debug"
)

func main() {
   defer xxisError.Recover()

	pflag.Parse()
	args := pflag.Args()
	if len(args) < 1 {
	   fmt.Fprintln(os.Stderr, "no file provided")
		os.Exit(1)
	}
	path := args[0]

	ast := xxisParser.BuildAstFromPath(path)

   xxisDebug.DebugScreen(ast)

   // var globals, locals Context

	// Exec(ast, &globals, &locals)
}