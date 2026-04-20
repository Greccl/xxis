package vm

import (
	"unicode"
)


type Value interface {
	Get() string
}


type Context struct {
	vars map[string]Value
}





type Instr struct {
   op int
   arg int
}

type Function struct {
	code []Instr
}


const (
   NOOP int = iota
   CMD
   JMP
   JMPZ
   JMPN
   HALT
   CALL
)

var OPCODES = []string{
   "NOOP",
   "CMD ",
   "JMP ",
   "JMPZ",
   "JMPN",
   "HALT",
   "CALL",
}



type Program struct {
	// code []Instruction
	funcs []Function
}



type VM struct {
	program *Program

	code []Instr
	pc int
	halted bool
	ctx *Context
}


func (vm *VM) run() {
   LOOP:
	for {
   	ins := vm.code[vm.pc]
   	// instr.Exec(vm)
   	switch ins.op {
   	   case NOOP:
   	      continue // USELESS
   	   case HALT:
   	      break LOOP
   	}
	}
}

/*
func (vm *VM) runExternal(cmd Command) error {
    c := exec.Command(cmd.Name, cmd.Args...)
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr
    c.Stdin = os.Stdin
    return c.Run()
}
*/






type ArgBuilder struct {
	curr []rune
	args []string
}

func (self *ArgBuilder) AddString(str string) {
	self.curr = append(self.curr, []rune(str)...)
}

func (self *ArgBuilder) AddRunes(buf []rune) {
	self.curr = append(self.curr, buf...)
}

func (self *ArgBuilder) ParseRunes(buf []rune) {
	for _, r := range buf {
		if unicode.IsSpace(r) {
			self.End()
		} else {
			self.curr = append(self.curr, r)
		}
	}
}

func (self *ArgBuilder) ParseString(str string) {
	for _, r := range str {
		if unicode.IsSpace(r) {
			self.End()
		} else {
			self.curr = append(self.curr, r)
		}
	}
}

func (self *ArgBuilder) End() {
	if len(self.curr) > 0 {
		self.args = append(self.args, string(self.curr))
		self.curr = nil
	}
}