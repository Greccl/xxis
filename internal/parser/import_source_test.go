package parser

import (
	"testing"

	xxisToken "github.com/Greccl/xxis/internal/token"
)

func TestBuildAstParsesQuotedImport(t *testing.T) {
	got := Build_ast_from_tokens(Enumerate_tokens(Enumerate_string(`import "./lib.xxis" as lib`)))
	mainBlock := got.Toks[0].Toks[1]
	if len(mainBlock.Toks) != 1 {
		t.Fatalf("main block child count = %d, want 1", len(mainBlock.Toks))
	}

	importNode := mainBlock.Toks[0]
	if importNode.Typ != 'K' || len(importNode.Buf) != 1 || importNode.Buf[0] != xxisToken.IMPORT {
		t.Fatalf("import node = %q %v, want K IMPORT", importNode.Typ, importNode.Buf)
	}
	if len(importNode.Toks) != 2 {
		t.Fatalf("import child count = %d, want 2", len(importNode.Toks))
	}
	if string(importNode.Toks[0].Buf) != "./lib.xxis" {
		t.Fatalf("import path = %q, want ./lib.xxis", string(importNode.Toks[0].Buf))
	}
	if string(importNode.Toks[1].Buf) != "lib" {
		t.Fatalf("import name = %q, want lib", string(importNode.Toks[1].Buf))
	}
}

func TestBuildAstParsesBareImport(t *testing.T) {
	got := Build_ast_from_tokens(Enumerate_tokens(Enumerate_string(`import ./lib.xxis as lib`)))
	importNode := got.Toks[0].Toks[1].Toks[0]

	if importNode.Typ != 'K' || importNode.Buf[0] != xxisToken.IMPORT {
		t.Fatalf("import node = %q %v, want K IMPORT", importNode.Typ, importNode.Buf)
	}
	if string(importNode.Toks[0].Buf) != "./lib.xxis" {
		t.Fatalf("import path = %q, want ./lib.xxis", string(importNode.Toks[0].Buf))
	}
	if string(importNode.Toks[1].Buf) != "lib" {
		t.Fatalf("import name = %q, want lib", string(importNode.Toks[1].Buf))
	}
}

func TestBuildAstParsesSource(t *testing.T) {
	got := Build_ast_from_tokens(Enumerate_tokens(Enumerate_string(`source "./defs.xxis"`)))
	sourceNode := got.Toks[0].Toks[1].Toks[0]

	if sourceNode.Typ != 'K' || len(sourceNode.Buf) != 1 || sourceNode.Buf[0] != xxisToken.SOURCE {
		t.Fatalf("source node = %q %v, want K SOURCE", sourceNode.Typ, sourceNode.Buf)
	}
	if len(sourceNode.Toks) != 1 {
		t.Fatalf("source child count = %d, want 1", len(sourceNode.Toks))
	}
	if string(sourceNode.Toks[0].Buf) != "./defs.xxis" {
		t.Fatalf("source path = %q, want ./defs.xxis", string(sourceNode.Toks[0].Buf))
	}
}

func TestBuildAstRejectsInvalidImport(t *testing.T) {
	tests := []struct {
		name string
		src  string
		msg  string
	}{
		{"missing_as", `import ./lib.xxis`, "import requires 'as'"},
		{"missing_name", `import ./lib.xxis as`, "import requires a name"},
		{"invalid_name", `import ./lib.xxis as 123`, "invalid import name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if err == nil {
					t.Fatalf("Build_ast_from_tokens did not panic")
				}
				pe, ok := err.(ParseError)
				if !ok {
					t.Fatalf("panic = %T, want ParseError", err)
				}
				if pe.Msg != tt.msg {
					t.Fatalf("ParseError.Msg = %q, want %q", pe.Msg, tt.msg)
				}
			}()

			Build_ast_from_tokens(Enumerate_tokens(Enumerate_string(tt.src)))
		})
	}
}

func TestBuildAstRejectsSourceAlias(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("Build_ast_from_tokens did not panic")
		}
		pe, ok := err.(ParseError)
		if !ok {
			t.Fatalf("panic = %T, want ParseError", err)
		}
		if pe.Msg != "source does not accept extra arguments" {
			t.Fatalf("ParseError.Msg = %q, want source extra arguments error", pe.Msg)
		}
	}()

	Build_ast_from_tokens(Enumerate_tokens(Enumerate_string(`source ./defs.xxis as defs`)))
}
