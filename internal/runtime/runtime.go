package runtime

import (
	"errors"
	"fmt"
	"path/filepath"

	xxisParser "github.com/Greccl/xxis/internal/parser"
	xxisToken "github.com/Greccl/xxis/internal/token"
)

func normalizePath(path string) (string, error) {
	if path == "" {
		return "", errors.New("source path is empty")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

type Value interface {
	Expand() string
	Get(string) string
	Set(string, string)
}

type StringValue string

func (v StringValue) Expand() string {
	return string(v)
}

func (v StringValue) Get(string) string {
	return ""
}

func (v StringValue) Set(string, string) {
}

type ModuleValue struct {
	Module *Module
}

func (v ModuleValue) Expand() string {
	if v.Module == nil {
		return ""
	}
	return v.Module.Path
}

func (v ModuleValue) Get(name string) string {
	if v.Module == nil {
		return ""
	}
	if item, ok := v.Module.Vars[name]; ok {
		return item.Expand()
	}
	return ""
}

func (v ModuleValue) Set(name, value string) {
	if v.Module == nil {
		return
	}
	v.Module.Vars[name] = StringValue(value)
}

type Module struct {
	Path  string
	Ast   *xxisToken.Token
	Vars  map[string]Value
	Funcs map[string]*xxisToken.Token
}

var modules = make(map[string]*Module)

func New_Module() *Module {
	return &Module{
		Vars:  make(map[string]Value),
		Funcs: make(map[string]*xxisToken.Token),
	}
}

func GetOrImport(path string) (*Module, error) {
	normalized, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	if mod, ok := modules[normalized]; ok {
		return mod, nil
	}

	ast := xxisParser.BuildAstFromPath(normalized)
	mod := New_Module()
	mod.Path = normalized
	mod.Ast = ast
	modules[normalized] = mod

	if err := Exec(ast, mod); err != nil {
		return nil, err
	}
	return mod, nil
}

func Source(parent *Module, path string) error {
	if parent == nil {
		return errors.New("source requires a parent module")
	}
	normalized, err := normalizePath(path)
	if err != nil {
		return err
	}

	ast := xxisParser.BuildAstFromPath(normalized)
	return Exec(ast, parent)
}

func Exec(tok *xxisToken.Token, mod *Module) error {
	if tok == nil {
		return nil
	}
	if mod == nil {
		return errors.New("exec requires a module")
	}
	if mod.Vars == nil {
		mod.Vars = make(map[string]Value)
	}
	if mod.Funcs == nil {
		mod.Funcs = make(map[string]*xxisToken.Token)
	}

	switch tok.Typ {
	case 'P':
		return execProgram(tok, mod)
	case 'F':
		return execFunction(tok, mod)
	case 'B':
		for _, child := range tok.Toks {
			if err := Exec(child, mod); err != nil {
				return err
			}
		}
	case 'K':
		return execKeyword(tok, mod)
	case 'C', '&', '/':
		return nil
	}
	return nil
}

func execProgram(tok *xxisToken.Token, mod *Module) error {
	for i, child := range tok.Toks {
		if i == 0 {
			continue
		}
		name := functionName(child)
		if name != "" {
			mod.Funcs[name] = child
		}
	}
	if len(tok.Toks) == 0 {
		return nil
	}
	return Exec(tok.Toks[0], mod)
}

func execFunction(tok *xxisToken.Token, mod *Module) error {
	if len(tok.Toks) < 2 {
		return nil
	}
	return Exec(tok.Toks[1], mod)
}

func execKeyword(tok *xxisToken.Token, mod *Module) error {
	if len(tok.Buf) == 0 {
		return nil
	}

	switch tok.Buf[0] {
	case xxisToken.IMPORT:
		if len(tok.Toks) != 2 {
			return fmt.Errorf("invalid import token")
		}
		imported, err := GetOrImport(tokenText(tok.Toks[0]))
		if err != nil {
			return err
		}
		mod.Vars[tokenText(tok.Toks[1])] = ModuleValue{Module: imported}
	case xxisToken.SOURCE:
		if len(tok.Toks) != 1 {
			return fmt.Errorf("invalid source token")
		}
		return Source(mod, tokenText(tok.Toks[0]))
	case xxisToken.IF:
		return nil
	case xxisToken.VAR:
		return nil
	}
	return nil
}

func functionName(tok *xxisToken.Token) string {
	if tok == nil || tok.Typ != 'F' || len(tok.Toks) == 0 {
		return ""
	}
	params := tok.Toks[0]
	if params == nil || len(params.Toks) == 0 {
		return ""
	}
	text := []rune(tokenText(params.Toks[0]))
	for i, r := range text {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return string(text[:i])
		}
	}
	return string(text)
}

func tokenText(tok *xxisToken.Token) string {
	if tok == nil {
		return ""
	}
	if len(tok.Buf) > 0 {
		return string(tok.Buf)
	}
	var out []rune
	for _, child := range tok.Toks {
		out = append(out, []rune(tokenText(child))...)
	}
	return string(out)
}
