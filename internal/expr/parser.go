package expr

import xxisToken "github.com/Greccl/xxis/internal/token"

type parser struct {
	lex     *lexer
	tok     lexToken
	prevEnd int
}

func Parse(src string) (node *Node, err error) {
	return parse(newStringLexer(src))
}

func MustParse(src string) *Node {
	node, err := Parse(src)
	if err != nil {
		panic(err)
	}
	return node
}

func ParseTokens(tokens []*xxisToken.Token) (node *Node, err error) {
	return parse(newTokenLexer(tokens))
}

func MustParseTokens(tokens []*xxisToken.Token) *Node {
	node, err := ParseTokens(tokens)
	if err != nil {
		panic(err)
	}
	return node
}

func parse(lex *lexer) (node *Node, err error) {
	p := &parser{lex: lex}
	p.advance()
	node = p.parseExpr()
	if p.tok.kind != lexEOF {
		p.errorAt(p.tok, "unexpected token")
	}
	return node, nil
}

func (p *parser) advance() {
	p.tok = p.lex.next()
	if p.tok.kind != lexEOF {
		p.prevEnd = p.tok.end
	}
}

func (p *parser) parseExpr() *Node {
	return p.parseOr()
}

func (p *parser) parseOr() *Node {
	node := p.parseAnd()
	for p.tok.kind == lexOr {
		op := p.tok
		p.advance()
		node = binaryNode(OpOr, op, node, p.parseAnd())
	}
	return node
}

func (p *parser) parseAnd() *Node {
	node := p.parseEquality()
	for p.tok.kind == lexAnd {
		op := p.tok
		p.advance()
		node = binaryNode(OpAnd, op, node, p.parseEquality())
	}
	return node
}

func (p *parser) parseEquality() *Node {
	node := p.parseComparison()
	for p.tok.kind == lexEqualEqual || p.tok.kind == lexBangEqual {
		op := p.tok
		p.advance()
		if op.kind == lexEqualEqual {
			node = binaryNode(OpEqual, op, node, p.parseComparison())
		} else {
			node = binaryNode(OpNotEqual, op, node, p.parseComparison())
		}
	}
	return node
}

func (p *parser) parseComparison() *Node {
	node := p.parseAdditive()
	for {
		op := p.tok
		var nodeOp Op
		switch op.kind {
		case lexLess:
			nodeOp = OpLess
		case lexLessEqual:
			nodeOp = OpLessEq
		case lexGreater:
			nodeOp = OpGreater
		case lexGreaterEqual:
			nodeOp = OpGreaterEq
		default:
			return node
		}
		p.advance()
		node = binaryNode(nodeOp, op, node, p.parseAdditive())
	}
}

func (p *parser) parseAdditive() *Node {
	node := p.parseMultiplicative()
	for p.tok.kind == lexPlus || p.tok.kind == lexMinus {
		op := p.tok
		p.advance()
		if op.kind == lexPlus {
			node = binaryNode(OpAdd, op, node, p.parseMultiplicative())
		} else {
			node = binaryNode(OpSub, op, node, p.parseMultiplicative())
		}
	}
	return node
}

func (p *parser) parseMultiplicative() *Node {
	node := p.parseUnary()
	for p.tok.kind == lexStar || p.tok.kind == lexSlash || p.tok.kind == lexMod {
		op := p.tok
		p.advance()
		switch op.kind {
		case lexStar:
			node = binaryNode(OpMul, op, node, p.parseUnary())
		case lexSlash:
			node = binaryNode(OpDiv, op, node, p.parseUnary())
		case lexMod:
			node = binaryNode(OpMod, op, node, p.parseUnary())
		}
	}
	return node
}

func (p *parser) parseUnary() *Node {
	if p.tok.kind == lexMinus || p.tok.kind == lexNot {
		op := p.tok
		p.advance()
		if op.kind == lexMinus {
			return unaryNode(OpNeg, op, p.parseUnary())
		}
		return unaryNode(OpNot, op, p.parseUnary())
	}
	return p.parsePostfix()
}

func (p *parser) parsePostfix() *Node {
	node := p.parsePrimary()
	for p.tok.kind == lexPercent {
		op := p.tok
		p.advance()
		node = &Node{
			Op:    OpPercent,
			Start: node.Start,
			End:   op.end,
			Nodes: []*Node{node},
		}
	}
	return node
}

func (p *parser) parsePrimary() *Node {
	tok := p.tok
	switch tok.kind {
	case lexNumber:
		p.advance()
		return &Node{Op: OpNumber, Start: tok.start, End: tok.end, Float: tok.num}
	case lexBool:
		p.advance()
		return &Node{Op: OpBool, Start: tok.start, End: tok.end, Bool: tok.bool}
	case lexVariable:
		p.advance()
		return &Node{Op: OpVariable, Start: tok.start, End: tok.end, String: tok.text, Token: tok.tok}
	case lexRuntimeString:
		p.advance()
		return &Node{Op: OpRuntimeString, Start: tok.start, End: tok.end, Token: tok.tok}
	case lexRuntimeInt:
		p.advance()
		return &Node{Op: OpRuntimeInt, Start: tok.start, End: tok.end, Token: tok.tok}
	case lexIdent:
		return p.parseCall()
	case lexLeftParen:
		p.advance()
		node := p.parseExpr()
		if p.tok.kind != lexRightParen {
			p.errorAt(p.tok, "expected closing parenthesis")
		}
		p.advance()
		return node
	}
	p.errorAt(tok, "expected expression")
	return nil
}

func (p *parser) parseCall() *Node {
	ident := p.tok
	p.advance()
	if p.tok.kind != lexLeftParen {
		p.errorAt(ident, "expected function call")
	}
	p.advance()

	args := make([]*Node, 0)
	if p.tok.kind != lexRightParen {
		for {
			args = append(args, p.parseExpr())
			if p.tok.kind != lexComma {
				break
			}
			p.advance()
			if p.tok.kind == lexRightParen {
				p.errorAt(p.tok, "expected expression after comma")
			}
		}
	}
	if p.tok.kind != lexRightParen {
		p.errorAt(p.tok, "expected closing parenthesis")
	}
	end := p.tok.end
	p.advance()

	return &Node{
		Op:     OpCall,
		Start:  ident.start,
		End:    end,
		String: ident.text,
		Nodes:  args,
	}
}

func binaryNode(op Op, tok lexToken, left, right *Node) *Node {
	return &Node{
		Op:    op,
		Start: left.Start,
		End:   right.End,
		Nodes: []*Node{left, right},
	}
}

func unaryNode(op Op, tok lexToken, child *Node) *Node {
	return &Node{
		Op:    op,
		Start: tok.start,
		End:   child.End,
		Nodes: []*Node{child},
	}
}

func (p *parser) errorAt(tok lexToken, msg string) {
	start := tok.start
	end := tok.end
	if tok.kind == lexEOF {
		start = p.prevEnd
		end = p.prevEnd
	}
	panic(ExprError{Msg: msg, Start: start, End: end})
}
