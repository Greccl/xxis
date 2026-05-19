package expr

import (
	"strconv"
	"unicode"

	xxisToken "github.com/Greccl/xxis/internal/token"
)

type lexKind int

const (
	lexEOF lexKind = iota
	lexNumber
	lexBool
	lexVariable
	lexRuntimeString
	lexRuntimeInt
	lexIdent
	lexPlus
	lexMinus
	lexStar
	lexSlash
	lexMod
	lexPercent
	lexEqualEqual
	lexBangEqual
	lexLess
	lexLessEqual
	lexGreater
	lexGreaterEqual
	lexAnd
	lexOr
	lexNot
	lexLeftParen
	lexRightParen
	lexComma
)

type lexToken struct {
	kind  lexKind
	start int
	end   int
	text  string
	num   float64
	bool  bool
	tok   *xxisToken.Token
}

type inputPart struct {
	typ   rune
	buf   []rune
	start int
	end   int
	tok   *xxisToken.Token
}

type lexer struct {
	parts []inputPart
	part  int
	pos   int
}

func newStringLexer(src string) *lexer {
	buf := []rune(src)
	return &lexer{
		parts: []inputPart{{
			typ:   'T',
			buf:   buf,
			start: 0,
			end:   len(buf),
		}},
	}
}

func newTokenLexer(tokens []*xxisToken.Token) *lexer {
	parts := make([]inputPart, 0, len(tokens))
	for _, tok := range tokens {
		if tok == nil {
			continue
		}
		parts = append(parts, inputPart{
			typ:   tok.Typ,
			buf:   tok.Buf,
			start: tok.Start,
			end:   tok.End,
			tok:   tok,
		})
	}
	return &lexer{parts: parts}
}

func (l *lexer) next() lexToken {
	for l.part < len(l.parts) {
		part := l.parts[l.part]
		if part.typ != 'T' {
			l.part++
			l.pos = 0
			if part.typ == 'V' {
				return lexToken{
					kind:  lexVariable,
					start: part.start,
					end:   part.end,
					text:  string(part.buf),
					tok:   part.tok,
				}
			}
			if part.typ == 'S' && len(part.buf) > 0 && part.buf[0] == 'X' {
				return lexToken{
					kind:  lexRuntimeInt,
					start: part.start,
					end:   part.end,
					tok:   part.tok,
				}
			}
			return lexToken{
				kind:  lexRuntimeString,
				start: part.start,
				end:   part.end,
				tok:   part.tok,
			}
		}

		if l.pos >= len(part.buf) {
			l.part++
			l.pos = 0
			continue
		}

		r := part.buf[l.pos]
		if unicode.IsSpace(r) {
			l.pos++
			continue
		}

		start := l.pos
		absStart := part.start + start

		if isDigit(r) {
			return l.scanNumber(part, start)
		}
		if isIdentStart(r) {
			return l.scanWord(part, start)
		}
		if r == '$' {
			return l.scanVariable(part, start)
		}

		l.pos++
		absEnd := part.start + l.pos
		switch r {
		case '+':
			return lexToken{kind: lexPlus, start: absStart, end: absEnd, text: "+"}
		case '-':
			return lexToken{kind: lexMinus, start: absStart, end: absEnd, text: "-"}
		case '*':
			return lexToken{kind: lexStar, start: absStart, end: absEnd, text: "*"}
		case '/':
			return lexToken{kind: lexSlash, start: absStart, end: absEnd, text: "/"}
		case '%':
			return lexToken{kind: lexPercent, start: absStart, end: absEnd, text: "%"}
		case '(':
			return lexToken{kind: lexLeftParen, start: absStart, end: absEnd, text: "("}
		case ')':
			return lexToken{kind: lexRightParen, start: absStart, end: absEnd, text: ")"}
		case ',':
			return lexToken{kind: lexComma, start: absStart, end: absEnd, text: ","}
		case '=':
			if l.match(part, '=') {
				return lexToken{kind: lexEqualEqual, start: absStart, end: part.start + l.pos, text: "=="}
			}
		case '!':
			if l.match(part, '=') {
				return lexToken{kind: lexBangEqual, start: absStart, end: part.start + l.pos, text: "!="}
			}
		case '<':
			if l.match(part, '=') {
				return lexToken{kind: lexLessEqual, start: absStart, end: part.start + l.pos, text: "<="}
			}
			return lexToken{kind: lexLess, start: absStart, end: absEnd, text: "<"}
		case '>':
			if l.match(part, '=') {
				return lexToken{kind: lexGreaterEqual, start: absStart, end: part.start + l.pos, text: ">="}
			}
			return lexToken{kind: lexGreater, start: absStart, end: absEnd, text: ">"}
		case '&':
			if l.match(part, '&') {
				panic(&ExprError{Msg: "&& is not supported in expr; use and", Start: absStart, End: part.start + l.pos})
			}
		case '|':
			if l.match(part, '|') {
				panic(&ExprError{Msg: "|| is not supported in expr; use or", Start: absStart, End: part.start + l.pos})
			}
		}

		panic(&ExprError{Msg: "unexpected character", Start: absStart, End: absEnd})
	}

	return lexToken{kind: lexEOF}
}

func (l *lexer) scanNumber(part inputPart, start int) lexToken {
	pos := start
	for pos < len(part.buf) && isDigit(part.buf[pos]) {
		pos++
	}
	if pos < len(part.buf) && part.buf[pos] == '.' {
		pos++
		if pos >= len(part.buf) || !isDigit(part.buf[pos]) {
			panic(&ExprError{Msg: "invalid number", Start: part.start + start, End: part.start + pos})
		}
		for pos < len(part.buf) && isDigit(part.buf[pos]) {
			pos++
		}
	}
	text := string(part.buf[start:pos])
	num, err := strconv.ParseFloat(text, 64)
	if err != nil {
		panic(&ExprError{Msg: "invalid number", Start: part.start + start, End: part.start + pos})
	}
	l.pos = pos
	return lexToken{
		kind:  lexNumber,
		start: part.start + start,
		end:   part.start + pos,
		text:  text,
		num:   num,
	}
}

func (l *lexer) scanWord(part inputPart, start int) lexToken {
	pos := start + 1
	for pos < len(part.buf) && isIdentChar(part.buf[pos]) {
		pos++
	}
	text := string(part.buf[start:pos])
	l.pos = pos
	switch text {
	case "true":
		return lexToken{kind: lexBool, start: part.start + start, end: part.start + pos, text: text, bool: true}
	case "false":
		return lexToken{kind: lexBool, start: part.start + start, end: part.start + pos, text: text, bool: false}
	case "mod":
		return lexToken{kind: lexMod, start: part.start + start, end: part.start + pos, text: text}
	case "and":
		return lexToken{kind: lexAnd, start: part.start + start, end: part.start + pos, text: text}
	case "or":
		return lexToken{kind: lexOr, start: part.start + start, end: part.start + pos, text: text}
	case "not":
		return lexToken{kind: lexNot, start: part.start + start, end: part.start + pos, text: text}
	}
	return lexToken{kind: lexIdent, start: part.start + start, end: part.start + pos, text: text}
}

func (l *lexer) scanVariable(part inputPart, start int) lexToken {
	pos := start + 1
	if pos >= len(part.buf) || !isIdentStart(part.buf[pos]) {
		panic(&ExprError{Msg: "invalid variable", Start: part.start + start, End: part.start + pos})
	}
	nameStart := pos
	pos++
	for pos < len(part.buf) && isIdentChar(part.buf[pos]) {
		pos++
	}
	l.pos = pos
	return lexToken{
		kind:  lexVariable,
		start: part.start + start,
		end:   part.start + pos,
		text:  string(part.buf[nameStart:pos]),
	}
}

func (l *lexer) match(part inputPart, r rune) bool {
	if l.pos >= len(part.buf) || part.buf[l.pos] != r {
		return false
	}
	l.pos++
	return true
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isIdentStart(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func isIdentChar(r rune) bool {
	return isIdentStart(r) || isDigit(r) || r == '_'
}
