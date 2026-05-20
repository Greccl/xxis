package expr

import (
	"testing"

	xxisToken "github.com/Greccl/xxis/internal/token"
)

func TestParsePrecedence(t *testing.T) {
	got := mustParseTest(t, "$foo + 1 * 2")

	if got.Op != OpAdd {
		t.Fatalf("root Op = %s, want %s", got.Op, OpAdd)
	}
	if got.Nodes[0].Op != OpVariable || got.Nodes[0].String != "foo" {
		t.Fatalf("left node = %s %q, want variable foo", got.Nodes[0].Op, got.Nodes[0].String)
	}
	right := got.Nodes[1]
	if right.Op != OpMul {
		t.Fatalf("right Op = %s, want %s", right.Op, OpMul)
	}
	if right.Nodes[0].Float != 1 || right.Nodes[1].Float != 2 {
		t.Fatalf("multiplication operands = %g and %g, want 1 and 2", right.Nodes[0].Float, right.Nodes[1].Float)
	}
}

func TestParseLogicalPrecedence(t *testing.T) {
	got := mustParseTest(t, "true or false and not $x")

	if got.Op != OpOr {
		t.Fatalf("root Op = %s, want %s", got.Op, OpOr)
	}
	if got.Nodes[1].Op != OpAnd {
		t.Fatalf("right Op = %s, want %s", got.Nodes[1].Op, OpAnd)
	}
	not := got.Nodes[1].Nodes[1]
	if not.Op != OpNot || not.Nodes[0].Op != OpVariable {
		t.Fatalf("and right node = %s, want not variable", not.Op)
	}
}

func TestParseUnaryAndPercent(t *testing.T) {
	got := mustParseTest(t, "$x * 25%")

	if got.Op != OpMul {
		t.Fatalf("root Op = %s, want %s", got.Op, OpMul)
	}
	if got.Nodes[1].Op != OpPercent {
		t.Fatalf("right Op = %s, want %s", got.Nodes[1].Op, OpPercent)
	}

	neg := mustParseTest(t, "-25%")
	if neg.Op != OpNeg {
		t.Fatalf("root Op = %s, want %s", neg.Op, OpNeg)
	}
	if neg.Nodes[0].Op != OpPercent {
		t.Fatalf("neg child Op = %s, want %s", neg.Nodes[0].Op, OpPercent)
	}
}

func TestParseFunctionCalls(t *testing.T) {
	got := mustParseTest(t, "pow($x + 1, sqrt(9))")

	if got.Op != OpCall || got.String != "pow" {
		t.Fatalf("root = %s %q, want call pow", got.Op, got.String)
	}
	if len(got.Nodes) != 2 {
		t.Fatalf("arg count = %d, want 2", len(got.Nodes))
	}
	if got.Nodes[0].Op != OpAdd {
		t.Fatalf("first arg Op = %s, want %s", got.Nodes[0].Op, OpAdd)
	}
	if got.Nodes[1].Op != OpCall || got.Nodes[1].String != "sqrt" {
		t.Fatalf("second arg = %s %q, want call sqrt", got.Nodes[1].Op, got.Nodes[1].String)
	}
}

func TestParseTokensRuntimeValues(t *testing.T) {
	cmd := &xxisToken.Token{Typ: 'S', Buf: []rune{'X', -1}, Start: 8, End: 14}
	tokens := []*xxisToken.Token{
		{Typ: 'V', Buf: []rune("count"), Start: 0, End: 6},
		{Typ: 'T', Buf: []rune(" == "), Start: 6, End: 10},
		cmd,
	}

	got, err := ParseTokens(tokens)
	if err != nil {
		t.Fatalf("ParseTokens() error = %v", err)
	}
	if got.Op != OpEqual {
		t.Fatalf("root Op = %s, want %s", got.Op, OpEqual)
	}
	if got.Nodes[0].Op != OpVariable || got.Nodes[0].String != "count" {
		t.Fatalf("left = %s %q, want variable count", got.Nodes[0].Op, got.Nodes[0].String)
	}
	if got.Nodes[1].Op != OpRuntimeInt || got.Nodes[1].Token != cmd {
		t.Fatalf("right = %s token=%v, want runtime int token", got.Nodes[1].Op, got.Nodes[1].Token)
	}
}

func TestParseRejectsCommandAndOrOperators(t *testing.T) {
	assertParseError(t, "$a && $b", "&& is not supported in expr; use and")
	assertParseError(t, "$a || $b", "|| is not supported in expr; use or")
}

func TestParseRejectsReservedFunctionName(t *testing.T) {
	assertParseError(t, "mod(1, 2)", "expected expression")
	assertParseError(t, "and(1, 2)", "expected expression")
	assertParseError(t, "true()", "unexpected token")
}

func TestParseRejectsMalformedExpressions(t *testing.T) {
	assertParseError(t, "1 +", "expected expression")
	assertParseError(t, "(1 + 2", "expected closing parenthesis")
	assertParseError(t, "pow(1,)", "expected expression after comma")
	assertParseError(t, "$", "invalid variable")
}

func mustParseTest(t *testing.T, src string) *Node {
	t.Helper()
	node, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse(%q) error = %v", src, err)
	}
	return node
}

func assertParseError(t *testing.T, src, msg string) {
	t.Helper()

	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("Parse(%q) did not fail", src)
		}
		exprErr, ok := err.(*ExprError)
		if !ok {
			t.Fatalf("Parse(%q) panic = %T, want *ExprError", src, err)
		}
		if exprErr.Msg != msg {
			t.Fatalf("Parse(%q) Msg = %q, want %q", src, exprErr.Msg, msg)
		}
	}()

	Parse(src)
}

func TestRegistryStoresNodes(t *testing.T) {
	node := mustParseTest(t, "1 + 2")
	index := Store(node)

	if got := Get(index); got != node {
		t.Fatalf("Get(%d) = %p, want %p", index, got, node)
	}
	if got := Get(-1); got != nil {
		t.Fatalf("Get(-1) = %p, want nil", got)
	}
	if got := Get(index + 1); got != nil {
		t.Fatalf("Get(%d) = %p, want nil", index+1, got)
	}
}
