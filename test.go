package main

import (
	"fmt"
	xxisCompiler "github.com/Greccl/xxis/internal/compiler"
	xxisParser "github.com/Greccl/xxis/internal/parser"
	xxisVm "github.com/Greccl/xxis/internal/vm"
)


func main() {
   defer func() {
      if err := recover(); err != nil {
         switch x := err.(type) {
         case xxisCompiler.CompileError:
            fmt.Println("error detected")
            fmt.Println(x)
         }
      }
   }()



   // read := enumerate_file("test1.txt")
	src0 := "cmd 1\nif cmd 2\n   cmd '3' $(hola)!(mundo)\n   if cmd 4   :   exit\n   last in block\nelse\n   it works\nend\ncmd 5 && cmd '6 $var6' || cmd 7"
	fmt.Println("--> source code")
	fmt.Println(src0)
	fmt.Println()

	read := xxisParser.Enumerate_string(src0)
	next := xxisParser.Enumerate_tokens(read)
	ast := xxisParser.Build_ast_from_tokens(next)

	fmt.Println("--> AST")
	ast.Dump()

	com := xxisCompiler.New_Compiler()
	for _, tok := range ast.Toks {
		com.Process(tok)
	}
	com.Finish()

   fmt.Println()
	fmt.Println("--> value table")
	for i, t := range com.Tokens {
	   fmt.Printf("%02d %s\n", i, t.Repr())
	}
	fmt.Println()

	fmt.Println("--> program")
	for i, inst := range com.Funcs["main"].Code {
		fmt.Printf("%02d: %s %d\n", i, xxisVm.OPCODES[inst.Op], inst.Arg)
	}
	fmt.Println()
}







func main2() {

	src0 := "cmd 1\nif cmd 2\n   cmd '3' $(hola)$(mundo)\n   if cmd 4   :   exit\n   last in block\nelse\n   it works\nend\ncmd 5 && cmd '6 $var6' || cmd 7"
	// src0 := "if true; echo 'hola $A $(echo $B com)mundo'\n abc $var; end"
	// src0 := "A $(B '$v0 $(C)' $(D t0 $(E$v2 t1))) t2 $(F)"
	fmt.Println("--> source code")
	fmt.Println(src0)
	fmt.Println()

	// read := enumerate_string(src0)
	// read := enumerate_file("test.txt")

	// fmt.Println("segments:")
	// segments := get_segments(read)
	/*
		for i, seg := range segments {
			fmt.Printf("%d | %c: %s\n", i, seg.typ, string(seg.buf))
		}
		fmt.Println("")
	*/

	// metas := get_metaexpressions(segments)
	/*
		fmt.Println("metaexpressions:", len(metas))
		for i, meta := range metas {
			fmt.Printf("%d | %d:", i, len(meta))
			for _, seg := range meta {
				fmt.Printf(" %c", seg.typ)
			}
			fmt.Println()
		}
		fmt.Println("")
	*/

   /*
	ast := build_ast(metas)
	fmt.Println("--> AST")
	ast.dump()

	com := New_Compiler()
	for _, tok := range ast.toks {
		com.process(tok)
	}
	com.finish()
	fmt.Println()


	fmt.Println("--> value table")
	for i, t := range com.tokens {
	   fmt.Printf("%02d %s\n", i, t.repr())
	}
	fmt.Println()

	fmt.Println("--> program")
	for i, inst := range com.f.code {
		fmt.Printf("%02d: %s %d\n", i, OPCODES[inst.op], inst.arg)
	}
	fmt.Println()
	*/
}
