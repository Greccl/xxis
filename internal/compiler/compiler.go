package compiler

import xxisVm "xxis/internal/vm"

type Function xxisVm.Function





type CompileFrame struct {
	f *Function
	// s []*Instruction
}

type Compiler struct {
	funcs map[string]*Function
	tokens []*Token

	f *Function
	// stack []*Instruction

	frames []CompileFrame
}

func New_Compiler() *Compiler {
	self := &Compiler{}
	self.funcs = make(map[string]*Function)
	self.tokens = make([]*Token, 0)
	self.f = &Function{}
	self.f.code = make([]Instr, 0)
	self.funcs["main"] = self.f
	return self
}

func (self *Compiler) push(op, val int) int {
	self.f.code = append(self.f.code, Instr{op, val})
	return len(self.f.code) - 1
}

func (self *Compiler) finish() {
	self.push(HALT, 0)
}

func (self *Compiler) splitAndOr(tok *Token) *Token {
	if tok.typ != 'C' {
		return tok
	}

	findOp := func(buf []rune, from int) (int, rune) {
		for i := from; i < len(buf)-1; i++ {
			if buf[i] == '&' && buf[i+1] == '&' {
				return i, '&'
			}
			if buf[i] == '|' && buf[i+1] == '|' {
				return i, '|'
			}
		}
		return -1, 0
	}

	parts := make([]*Token, 0, 2)
	ops := make([]rune, 0, 1)
	curr := make([]*Token, 0)
	stripLeft := false
	found := false

	for _, t := range tok.toks {
		if t.typ != 'T' {
			if stripLeft {
				stripLeft = false
			}
			curr = append(curr, t)
			continue
		}

		buf := t.buf
		if stripLeft {
			buf = trim_buffer(buf, true, false)
			stripLeft = false
		}

		pos := 0
		for {
			opPos, op := findOp(buf, pos)
			if opPos == -1 {
				if pos == 0 {
					if len(buf) > 0 {
						curr = append(curr, &Token{typ: 'T', buf: buf})
					}
				} else if pos < len(buf) {
					frag := buf[pos:]
					if len(frag) > 0 {
						curr = append(curr, &Token{typ: 'T', buf: frag})
					}
				}
				break
			}

			found = true
			left := trim_buffer(buf[pos:opPos], false, true)
			if len(left) > 0 {
				curr = append(curr, &Token{typ: 'T', buf: left})
			}

			if len(curr) > 0 {
				parts = append(parts, &Token{typ: 'C', toks: curr})
				ops = append(ops, op)
			}
			curr = make([]*Token, 0)

			pos = opPos + 2
			for pos < len(buf) && is_space(buf[pos]) {
				pos++
			}
			if pos >= len(buf) {
				stripLeft = true
				break
			}
		}
	}

	if !found {
		return tok
	}
	if len(curr) > 0 {
		parts = append(parts, &Token{typ: 'C', toks: curr})
	}
	if len(parts) == 0 || len(ops) == 0 || len(parts) != len(ops)+1 {
		return tok
	}

	block := &Token{typ: 'B', toks: make([]*Token, len(parts))}
	for i, part := range parts {
		if i > 0 { part.typ = ops[i-1] }
		block.toks[i] = part
	}
	return block
}

func (self *Compiler) process(tok *Token) {
	switch tok.typ {
	   case 'C':
	      split := self.splitAndOr(tok)
	      if split.typ == 'B' {
	      	for _, t := range split.toks {
   		      self.process(t)
	      	}
	      } else {
	         // fmt.Printf("try to teplace %s\n", split.repr())
   		   self.replace(split, 0, true)
	      }
		case 'K':
		   switch tok.buf[0] {
		      case IF:
		         self.process(tok.toks[0])
		         i := self.push(JMPN, 0)
		         for _, t := range tok.toks[1].toks {
		            self.process(t)
		         }
   		      self.f.code[i].arg = len(self.f.code)
		         if tok.toks[2] != nil {
   		         self.f.code[i].arg++
   		         i = self.push(JMP, 0)
   		         for _, t := range tok.toks[2].toks {
   		            self.process(t)
   		         }
   		         self.f.code[i].arg = len(self.f.code)
   		      }
		   }
      case '&':
        self.push(JMPN, len(self.f.code)+2)
        tok.typ = 'C'
        self.replace(tok, 0, true)
      case '|':
        self.push(JMPZ, len(self.f.code)+2)
        tok.typ = 'C'
        self.replace(tok, 0, true)
   }
}

func (self *Compiler) replace(tok *Token, reserved int, add bool) {
   // fmt.Printf("replace %c %d %s\n", tok.typ, reserved, tok.repr())
	for i := 0; i < len(tok.toks); i++ {
		t := tok.toks[i]
		switch t.typ {
		case 'S':
			// t.buf = []rune{rune(reserved)}
			t.buf[1] = rune(reserved)
			replacement := &Token{typ: 'R'}
			replacement.buf = []rune{rune(reserved)}
			self.replace(t, reserved, true)
			reserved++
			tok.toks[i] = replacement
		case 'Q':
			self.replace(t, reserved, false)
		}
	}
	if add {
      self.push(CMD, len(self.tokens))
      self.tokens = append(self.tokens, tok)
	}
}
