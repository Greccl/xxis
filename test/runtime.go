package main

import (
   "os"
   "fmt"
	"github.com/spf13/pflag"
	// xxisError "github.com/Greccl/xxis/internal/errors"
	xxisRuntime "github.com/Greccl/xxis/internal/runtime"
	// xxisParser "github.com/Greccl/xxis/internal/parser"
	// xxisDebug "github.com/Greccl/xxis/internal/debug"
)

func main() {
	pflag.Parse()
   /*
	args := pflag.Args()
	if len(args) < 1 {
	   fmt.Fprintln(os.Stderr, "no file provided")
		os.Exit(1)
	}
	path := args[0]
   xxisRuntime.Main("test/src/single_cmd.xxis")
	*/
   mod, err := xxisRuntime.DefaultGlobalContext().GetOrImport("test/src/single_cmd.xxis")
   if err != nil {
      fmt.Println("error cargando modulo")
      fmt.Println(err)
      os.Exit(1)
   }
   ctx := xxisRuntime.NewThreadContext()
   for _, tok := range mod.Init.Toks {
      if tok.Typ == 'C' {
         fmt.Printf(">> Expand command\n")
         args := xxisRuntime.Expand(tok, ctx)
         for _, arg := range args {
            fmt.Printf("- %v\n", arg)
         }
      }
   }
}