package main

import (
   "fmt"
)



func main() {
	
	// src0 := "if true; echo 'hola $inter $(echo $inner command)mundo' abc $var; end"
	src0 := "A $(B '$v0 $(C)' $(D t0 $(E$v2 t1))) t2 $(F)"
	fmt.Println("source code:")
	fmt.Println(src0)
	fmt.Println()

	read := enumerate_string(src0)
	// read := enumerate_file("test.txt")

	fmt.Println("segments:")
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

	fmt.Println("commands:")
	com := &Compiler{}
	for _, meta := range metas {
		fmt.Println()
		sub := subcmd_by_segment(meta)
		// sub.dump()
		com.process(sub)
	}
	fmt.Println()

	

	fmt.Println("program:")
	vm := &VM{}
	for _, inst := range com.code {
		inst.Exec(vm)
	}
	fmt.Println()
}
