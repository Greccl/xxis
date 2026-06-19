package parser

import (
	"testing"

	xxisExpr "github.com/Greccl/xxis/internal/expr"
	xxisToken "github.com/Greccl/xxis/internal/token"
)

func TestSplitAndOrPreservesRanges(t *testing.T) {
	tok := &Token{
		Typ: 'C',
		Toks: []*Token{
			{Typ: 'T', Buf: []rune("cmd 1 &&  cmd 2 || cmd 3"), Start: 10, End: 34},
		},
		Start: 10,
		End:   34,
	}

	got := split_and_or(tok)
	if got.Typ != 'B' {
		t.Fatalf("split_and_or() Typ = %q, want 'B'", got.Typ)
	}
	if got.Start != 10 || got.End != 34 {
		t.Fatalf("block range = %d:%d, want 10:34", got.Start, got.End)
	}
	if len(got.Toks) != 3 {
		t.Fatalf("len(block.Toks) = %d, want 3", len(got.Toks))
	}

	want := []struct {
		typ        rune
		start, end int
		text       string
	}{
		{'C', 10, 15, "cmd 1"},
		{'&', 20, 25, "cmd 2"},
		{'/', 29, 34, "cmd 3"},
	}

	for i, want := range want {
		part := got.Toks[i]
		if part.Typ != want.typ || part.Start != want.start || part.End != want.end {
			t.Fatalf("part %d = %q %d:%d, want %q %d:%d", i, part.Typ, part.Start, part.End, want.typ, want.start, want.end)
		}
		if len(part.Toks) != 1 {
			t.Fatalf("part %d child count = %d, want 1", i, len(part.Toks))
		}
		child := part.Toks[0]
		if string(child.Buf) != want.text || child.Start != want.start || child.End != want.end {
			t.Fatalf("part %d child = %q %d:%d, want %q %d:%d", i, string(child.Buf), child.Start, child.End, want.text, want.start, want.end)
		}
	}
}

func TestSplitTokensAtPreservesRangesForSplitAndOr(t *testing.T) {
	tokens := []*Token{
		{Typ: 'T', Buf: []rune("cond && other:  body || fallback"), Start: 4, End: 36},
	}

	cond, body := split_by_colon(tokens)
	condTok := split_and_or(&Token{Typ: 'C', Toks: cond})
	bodyTok := split_and_or(&Token{Typ: 'C', Toks: body})

	if condTok.Start != 4 || condTok.End != 17 {
		t.Fatalf("cond range = %d:%d, want 4:17", condTok.Start, condTok.End)
	}
	if bodyTok.Start != 20 || bodyTok.End != 36 {
		t.Fatalf("body range = %d:%d, want 20:36", bodyTok.Start, bodyTok.End)
	}
	if len(bodyTok.Toks) != 2 {
		t.Fatalf("body child count = %d, want 2", len(bodyTok.Toks))
	}
	if bodyTok.Toks[1].Start != 28 || bodyTok.Toks[1].End != 36 {
		t.Fatalf("body second part range = %d:%d, want 28:36", bodyTok.Toks[1].Start, bodyTok.Toks[1].End)
	}
}

func TestBuildAstWrapsMainAndUserFunctions(t *testing.T) {
	src := "cmd main\nfunction foo arg1 arg2\n\tcmd foo\n\tif true\n\t\tnested\ncmd after\nfunction bar\n\tcmd bar"
	next := Enumerate_tokens(Enumerate_string(src))

	got := Build_ast_from_tokens(next)
	if got.Typ != 'P' {
		t.Fatalf("root Typ = %q, want 'P'", got.Typ)
	}
	if len(got.Toks) != 3 {
		t.Fatalf("root child count = %d, want 3", len(got.Toks))
	}

	main := got.Toks[0]
	if main.Typ != 'F' || len(main.Toks) != 2 {
		t.Fatalf("main node = %q with %d children, want F with 2 children", main.Typ, len(main.Toks))
	}
	mainBlock := main.Toks[1]
	if mainBlock.Typ != 'B' || len(mainBlock.Toks) != 2 {
		t.Fatalf("main block = %q with %d children, want B with 2 commands", mainBlock.Typ, len(mainBlock.Toks))
	}
	if string(mainBlock.Toks[0].Toks[0].Buf) != "cmd main" {
		t.Fatalf("main first command = %q, want %q", string(mainBlock.Toks[0].Toks[0].Buf), "cmd main")
	}
	if string(mainBlock.Toks[1].Toks[0].Buf) != "cmd after" {
		t.Fatalf("main second command = %q, want %q", string(mainBlock.Toks[1].Toks[0].Buf), "cmd after")
	}

	foo := got.Toks[1]
	if foo.Typ != 'F' || len(foo.Toks) != 2 {
		t.Fatalf("foo node = %q with %d children, want F with 2 children", foo.Typ, len(foo.Toks))
	}
	if string(foo.Toks[0].Toks[0].Buf) != "foo arg1 arg2" {
		t.Fatalf("foo params = %q, want %q", string(foo.Toks[0].Toks[0].Buf), "foo arg1 arg2")
	}
	fooBlock := foo.Toks[1]
	if fooBlock.Typ != 'B' || len(fooBlock.Toks) != 2 {
		t.Fatalf("foo block = %q with %d children, want B with 2 commands", fooBlock.Typ, len(fooBlock.Toks))
	}
	if fooBlock.Toks[1].Typ != 'K' {
		t.Fatalf("foo nested node Typ = %q, want 'K'", fooBlock.Toks[1].Typ)
	}

	bar := got.Toks[2]
	if bar.Typ != 'F' || len(bar.Toks) != 2 {
		t.Fatalf("bar node = %q with %d children, want F with 2 children", bar.Typ, len(bar.Toks))
	}
	if string(bar.Toks[0].Toks[0].Buf) != "bar" {
		t.Fatalf("bar params = %q, want %q", string(bar.Toks[0].Toks[0].Buf), "bar")
	}
	if bar.Toks[1].Typ != 'B' || len(bar.Toks[1].Toks) != 1 {
		t.Fatalf("bar block = %q with %d children, want B with 1 command", bar.Toks[1].Typ, len(bar.Toks[1].Toks))
	}
}

func TestBuildAstParsesSharedVarScope(t *testing.T) {
	tests := []string{
		"var -s name = value",
		"var --shared name = value",
	}

	for _, src := range tests {
		t.Run(src, func(t *testing.T) {
			got := Build_ast_from_tokens(Enumerate_tokens(Enumerate_string(src)))
			varNode := got.Toks[0].Toks[1].Toks[0]
			if varNode.Typ != 'K' || len(varNode.Buf) < 4 || varNode.Buf[0] != xxisToken.VAR {
				t.Fatalf("var node = %q %v, want K VAR", varNode.Typ, varNode.Buf)
			}
			if varNode.Buf[1] != xxisToken.VarScopeShared {
				t.Fatalf("var scope = %s, want shared", xxisToken.VarScopeName(varNode.Buf[1]))
			}
		})
	}
}

func TestBuildAstUsesTabsForIndentedIfElse(t *testing.T) {
	src := "if true\n\tthen cmd\nelse\n\telse cmd\nafter"
	next := Enumerate_tokens(Enumerate_string(src))

	got := Build_ast_from_tokens(next)
	mainBlock := got.Toks[0].Toks[1]
	if len(mainBlock.Toks) != 2 {
		t.Fatalf("main block child count = %d, want 2", len(mainBlock.Toks))
	}
	if mainBlock.Toks[0].Typ != 'K' {
		t.Fatalf("first node Typ = %q, want if keyword", mainBlock.Toks[0].Typ)
	}
	ifNode := mainBlock.Toks[0]
	if ifNode.Toks[2] == nil {
		t.Fatalf("if else block is nil")
	}
	if len(ifNode.Toks[1].Toks) != 1 {
		t.Fatalf("then block child count = %d, want 1", len(ifNode.Toks[1].Toks))
	}
	if len(ifNode.Toks[2].Toks) != 1 {
		t.Fatalf("else block child count = %d, want 1", len(ifNode.Toks[2].Toks))
	}
	if string(mainBlock.Toks[1].Toks[0].Buf) != "after" {
		t.Fatalf("second main command = %q, want after", string(mainBlock.Toks[1].Toks[0].Buf))
	}
}

func TestBuildAstParsesIfExprCondition(t *testing.T) {
	src := "if $x > 0\n\tcmd ok"
	next := Enumerate_tokens(Enumerate_string(src))

	got := Build_ast_from_tokens(next)
	ifNode := got.Toks[0].Toks[1].Toks[0]
	if ifNode.Typ != 'K' {
		t.Fatalf("if node Typ = %q, want K", ifNode.Typ)
	}
	if len(ifNode.Buf) != 3 {
		t.Fatalf("if Buf len = %d, want 3", len(ifNode.Buf))
	}
	if ifNode.Buf[0] != xxisToken.IF || ifNode.Buf[1] != xxisToken.IfCondExpr {
		t.Fatalf("if Buf = %v, want IF expr", ifNode.Buf)
	}

	exprNode := xxisExpr.Get(int(ifNode.Buf[2]))
	if exprNode == nil {
		t.Fatalf("expr.Get(%d) = nil", ifNode.Buf[2])
	}
	if exprNode.Op != xxisExpr.OpGreater {
		t.Fatalf("expr root Op = %s, want %s", exprNode.Op, xxisExpr.OpGreater)
	}
}

func TestBuildAstParsesIfSuccessCommandCondition(t *testing.T) {
	src := "if success cmd 1\n\tcmd ok"
	next := Enumerate_tokens(Enumerate_string(src))

	got := Build_ast_from_tokens(next)
	ifNode := got.Toks[0].Toks[1].Toks[0]
	if ifNode.Buf[1] != xxisToken.IfCondSuccess {
		t.Fatalf("if mode = %s, want success", xxisToken.IfCondName(ifNode.Buf[1]))
	}
	if len(ifNode.Buf) != 2 {
		t.Fatalf("if Buf len = %d, want 2", len(ifNode.Buf))
	}
	if string(ifNode.Toks[0].Toks[0].Buf) != "cmd 1" {
		t.Fatalf("condition command = %q, want cmd 1", string(ifNode.Toks[0].Toks[0].Buf))
	}
}

func TestBuildAstParsesIfFailureCommandConditionWithInlineBody(t *testing.T) {
	src := "if failure cmd 1: cmd fallback"
	next := Enumerate_tokens(Enumerate_string(src))

	got := Build_ast_from_tokens(next)
	ifNode := got.Toks[0].Toks[1].Toks[0]
	if ifNode.Buf[1] != xxisToken.IfCondFailure {
		t.Fatalf("if mode = %s, want failure", xxisToken.IfCondName(ifNode.Buf[1]))
	}
	if string(ifNode.Toks[0].Toks[0].Buf) != "cmd 1" {
		t.Fatalf("condition command = %q, want cmd 1", string(ifNode.Toks[0].Toks[0].Buf))
	}
	if len(ifNode.Toks[1].Toks) != 1 {
		t.Fatalf("inline body child count = %d, want 1", len(ifNode.Toks[1].Toks))
	}
	if string(ifNode.Toks[1].Toks[0].Toks[0].Buf) != "cmd fallback" {
		t.Fatalf("inline body command = %q, want cmd fallback", string(ifNode.Toks[1].Toks[0].Toks[0].Buf))
	}
}

func TestBuildAstTreatsSuccessCallAsExprCondition(t *testing.T) {
	src := "if success($x)\n\tcmd ok"
	next := Enumerate_tokens(Enumerate_string(src))

	got := Build_ast_from_tokens(next)
	ifNode := got.Toks[0].Toks[1].Toks[0]
	if ifNode.Buf[1] != xxisToken.IfCondExpr {
		t.Fatalf("if mode = %s, want expr", xxisToken.IfCondName(ifNode.Buf[1]))
	}
	exprNode := xxisExpr.Get(int(ifNode.Buf[2]))
	if exprNode == nil {
		t.Fatalf("expr.Get(%d) = nil", ifNode.Buf[2])
	}
	if exprNode.Op != xxisExpr.OpCall || exprNode.String != "success" {
		t.Fatalf("expr root = %s %q, want call success", exprNode.Op, exprNode.String)
	}
}

func TestBuildAstRejectsIfExprCommandAndOr(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("Build_ast_from_tokens did not panic")
		}
		pe, ok := err.(*xxisExpr.ExprError)
		if !ok {
			t.Fatalf("panic = %T, want *expr.ExprError", err)
		}
		if pe.Msg != "&& is not supported in expr; use and" {
			t.Fatalf("ExprError.Msg = %q, want && error", pe.Msg)
		}
	}()

	next := Enumerate_tokens(Enumerate_string("if $a && $b\n\tcmd ok"))
	Build_ast_from_tokens(next)
}

func TestBuildAstRejectsEmptyIfCommandCondition(t *testing.T) {
	for _, src := range []string{"if success\n\tcmd ok", "if failure\n\tcmd ok"} {
		t.Run(src, func(t *testing.T) {
			defer func() {
				err := recover()
				if err == nil {
					t.Fatalf("Build_ast_from_tokens did not panic")
				}
				pe, ok := err.(ParseError)
				if !ok {
					t.Fatalf("panic = %T, want ParseError", err)
				}
				if pe.Msg != "if command condition requires a command" {
					t.Fatalf("ParseError.Msg = %q, want command condition error", pe.Msg)
				}
			}()

			next := Enumerate_tokens(Enumerate_string(src))
			Build_ast_from_tokens(next)
		})
	}
}

func TestBuildAstRejectsSpaceIndentation(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("Build_ast_from_tokens did not panic")
		}
		pe, ok := err.(ParseError)
		if !ok {
			t.Fatalf("panic = %T, want ParseError", err)
		}
		if pe.Msg != "indentation must use tabs" {
			t.Fatalf("ParseError.Msg = %q, want indentation must use tabs", pe.Msg)
		}
	}()

	next := Enumerate_tokens(Enumerate_string("if true\n  cmd"))
	Build_ast_from_tokens(next)
}

func TestBuildAstTreatsEndAsCommand(t *testing.T) {
	next := Enumerate_tokens(Enumerate_string("end"))

	got := Build_ast_from_tokens(next)
	mainBlock := got.Toks[0].Toks[1]
	if len(mainBlock.Toks) != 1 {
		t.Fatalf("main block child count = %d, want 1", len(mainBlock.Toks))
	}
	if string(mainBlock.Toks[0].Toks[0].Buf) != "end" {
		t.Fatalf("command = %q, want end", string(mainBlock.Toks[0].Toks[0].Buf))
	}
}
