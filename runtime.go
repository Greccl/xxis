package main

import (
	"unicode"
	"fmt"
	// "os/exec"
)


type Value interface {
	Get() string
}


type Context struct {
	vars map[string]Value
}







type Instruction interface {
	Exec(*VM)
	// Dump() string
}

type Function2 struct {
	addr int
}

type Program struct {
	code []Instruction
	funcs []Function
}



type VM struct {
	program *Program

	code []Instruction
	pc int
	halted bool
	ctx *Context
}


func (vm *VM) run() {
	for {
   	instr := vm.code[vm.pc]
   	instr.Exec(vm)

   	if vm.halted {
			break
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









type InstrTest struct {
	str string
}

func (self InstrTest) Exec(vm *VM) {
	fmt.Print(self.str)
}








type InstrCmd struct {
	tok *Token
	out int
}

func (self *InstrCmd) Exec(vm *VM) {
	// 1. construir tokens
	// b := &ArgBuilder{}
	for _, t := range self.tok.toks {
		// t.appendArgsTo(b)
		t.dump()
	}

	// 2. definir quien ejecuta
	// 3. ejecutar
	// 4. guardar resultado en registros
}







type InstrJump struct {
	// -1 continue
	// -2 break
	addr int
}

func (self *InstrJump) Exec(vm *VM) {
	// vm.pc = self.addr
	fmt.Printf("JMP %d", self.addr)
}







type InstrHalt struct {
	reason int
}

func (self *InstrHalt) Exec(vm *VM) {
	// vm.pc = self.addr
	fmt.Printf("HLT %d", self.reason)
}
