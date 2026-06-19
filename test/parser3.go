package main

import (
	xxisParser "github.com/Greccl/xxis/internal/parser"
	xxisDebug "github.com/Greccl/xxis/internal/debug"
	"github.com/spf13/pflag"
)

func main() {
	pflag.Parse()
	args := pflag.Args()
	if len(args) < 1 {
		return
	}
	path := args[0]

	ast := xxisParser.BuildAstFromPath(path)

   xxisDebug.DebugAst(ast)
}