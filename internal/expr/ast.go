package expr

import (
	"fmt"

	xxisToken "github.com/Greccl/xxis/internal/token"
)

type Op int

const (
	OpInvalid Op = iota
	OpNumber
	OpBool
	OpVariable
	OpRuntimeString
	OpRuntimeInt
	OpCall
	OpNeg
	OpNot
	OpPercent
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpLess
	OpLessEq
	OpGreater
	OpGreaterEq
	OpEqual
	OpNotEqual
	OpAnd
	OpOr
)

type Node struct {
	Start  int
	End    int
	
	Op     Op
	Int    int64
	Float  float64
	String string
	Bool   bool

	Nodes  []*Node

	Token  *xxisToken.Token
}

func (op Op) String() string {
	switch op {
	case OpNumber:
		return "number"
	case OpBool:
		return "bool"
	case OpVariable:
		return "variable"
	case OpRuntimeString:
		return "runtime_string"
	case OpRuntimeInt:
		return "runtime_int"
	case OpCall:
		return "call"
	case OpNeg:
		return "neg"
	case OpNot:
		return "not"
	case OpPercent:
		return "percent"
	case OpAdd:
		return "add"
	case OpSub:
		return "sub"
	case OpMul:
		return "mul"
	case OpDiv:
		return "div"
	case OpMod:
		return "mod"
	case OpLess:
		return "less"
	case OpLessEq:
		return "less_eq"
	case OpGreater:
		return "greater"
	case OpGreaterEq:
		return "greater_eq"
	case OpEqual:
		return "equal"
	case OpNotEqual:
		return "not_equal"
	case OpAnd:
		return "and"
	case OpOr:
		return "or"
	}
	return "invalid"
}

func (n *Node) StringRepr() string {
	if n == nil {
		return "<nil>"
	}
	switch n.Op {
	case OpNumber:
		return fmt.Sprintf("%g", n.Float)
	case OpBool:
		return fmt.Sprintf("%t", n.Bool)
	case OpVariable:
		return "$" + n.String
	case OpCall:
		return fmt.Sprintf("%s(...)", n.String)
	}
	return n.Op.String()
}

type ExprError struct {
	Msg   string
	Start int
	End   int
}

func (e *ExprError) Error() string {
	return fmt.Sprintf("ExprError: %s", e.Msg)
}
