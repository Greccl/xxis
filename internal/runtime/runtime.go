package runtime

import (
	"errors"
	"path/filepath"
	"io"
   "os"
   "fmt"
   "sync"
	xxisParser "github.com/Greccl/xxis/internal/parser"
	xxisToken "github.com/Greccl/xxis/internal/token"
)

type Token = xxisToken.Token

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

//
// Functions
//

type Function struct {
   Name string
   Body []*Token
}

//
// Modules
//

type Module struct {
	Path   string
	Funcs  map[string]*Token
	Init   *Token
	Locals map[string]Value
	Shared map[string]SyncValue
}

func NewModule() *Module {
	return &Module{
		Funcs:  make(map[string]*Token),
		Shared: make(map[string]SyncValue),
		Locals: make(map[string]Value),
	}
}

func (mod *Module) Load(tok *Token) {
   mod.Init = tok.Toks[0].Toks[1]
   for _, t := range tok.Toks[1:] {
      name := string(t.Buf)
      mod.Funcs[name] = t.Toks[1]
   }
}

//
// ThreadContext
//

type Frame struct {
   Module *Module
   Locals map[string]Value
   Stdin  io.Reader
   Stdout io.Writer
   Stderr io.Writer
}

type ThreadContext struct {
   Stack   []*Frame
   Current *Frame
}

func NewThreadContext() *ThreadContext {
   ctx := &ThreadContext{}
   ctx.Stack = make([]*Frame, 0)
   return ctx
}

func (ctx *ThreadContext) PushFrame(mod *Module) *Frame {
   frame := &Frame{}
   frame.Locals = make(map[string]Value)
   frame.Module = mod
   ctx.Stack = append(ctx.Stack, frame)
   ctx.Current = frame
   return frame
}

func (ctx *ThreadContext) ExpandVar(name string) string {
   return fmt.Sprintf("<%s>", name)
}

//
// GlobalContext
//

type GlobalContext struct {
   Globals map[string]SyncValue
   Modules map[string]*Module
   mu sync.Mutex
}

var globalContext = &GlobalContext{
	Globals: make(map[string]SyncValue),
	Modules: make(map[string]*Module),
}

func DefaultGlobalContext() *GlobalContext {
	return globalContext
}

func (global *GlobalContext) GetOrImport(path string) (*Module, error) {
	normalized, err := normalizePath(path)
	if err != nil {
		return nil, err
	}

	global.mu.Lock()
	if mod, ok := global.Modules[normalized]; ok {
		global.mu.Unlock()
		return mod, nil
	}
	global.mu.Unlock()

	ast := xxisParser.BuildAstFromPath(normalized)
	mod := NewModule()
	mod.Path = normalized
	mod.Load(ast)

	global.mu.Lock()
	global.Modules[normalized] = mod
	global.mu.Unlock()

	ctx := NewThreadContext()
	ctx.PushFrame(mod)
	if err := Exec(mod.Init, ctx); err != nil {
		return nil, err
	}
	return mod, nil
}

func (global *GlobalContext) Source(parent *Module, path string) error {
	if parent == nil {
		return errors.New("source requires a parent module")
	}
	normalized, err := normalizePath(path)
	if err != nil {
		return err
	}

	ast := xxisParser.BuildAstFromPath(normalized)
	return Exec(ast, NewThreadContext())
}

//
// Extra
//

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

//
// Main
//

func Main(path string) error {
   defer func() {
      if err := recover(); err != nil {
         fmt.Println("! RUNTIME ERROR ! ", err)
         os.Exit(1)
      }
   }()
   _, err := globalContext.GetOrImport(path)
   if err != nil {
      fmt.Println("! MAIN ERROR ! ", err)
      os.Exit(1)
   }
   return nil
}