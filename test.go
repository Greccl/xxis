package main

import (
	"fmt"
)

func main() {

	src0 := "cmd 1\nif cmd 2\n   cmd '3' $(hola)\n   cmd 4\nend\ncmd 5"
	// src0 := "if true; echo 'hola $A $(echo $B com)mundo'\n abc $var; end"
	// src0 := "A $(B '$v0 $(C)' $(D t0 $(E$v2 t1))) t2 $(F)"
	fmt.Println("source code:")
	fmt.Println(src0)
	fmt.Println()

	read := enumerate_string(src0)
	// read := enumerate_file("test.txt")

	// fmt.Println("segments:")
	segments := get_segments(read)
	/*
		for i, seg := range segments {
			fmt.Printf("%d | %c: %s\n", i, seg.typ, string(seg.buf))
		}
		fmt.Println("")
	*/

	metas := get_metaexpressions(segments)
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

	com := New_Compiler()
	ast := build_ast(metas)
	fmt.Println("compiling:")
	for _, tok := range ast.toks {
		com.process(tok)
		tok.dump()
	}
	fmt.Println()

	fmt.Println("program:")
	vm := &VM{}
	for addr, inst := range com.f.code {
		fmt.Printf("%d: ", addr)
		inst.Exec(vm)
		fmt.Println()
	}
	fmt.Println()
}
