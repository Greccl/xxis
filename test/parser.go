package main

import (
	"fmt"
	// "os"
	// "path/filepath"
	// xxisCompiler "github.com/Greccl/xxis/internal/compiler"
	xxisParser "github.com/Greccl/xxis/internal/parser"
	// xxisVm "github.com/Greccl/xxis/internal/vm"
)


func main() {
	fmt.Println("-- TEST PARSER --")

	// home, _ := os.UserHomeDir()
	// path := filepath.Join(home, "dev", "xxis", "test1.txt")
	read := xxisParser.Enumerate_file("test/src0.txt")

	// src0 := "cmd 1\nif cmd 2\n   cmd '3' $(hola)!(mundo)\n   if cmd 4   :   exit\n   last in block\nelse\n   it works\nend\ncmd 5 && cmd '6 $var6' || cmd 7"
	// read := xxisParser.Enumerate_string(src0)

	next := xxisParser.Enumerate_tokens(read)
	ast := xxisParser.Build_ast_from_tokens(next)

	ast.Dump()
}