package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	xxisParser "github.com/Greccl/xxis/internal/parser"
	xxisToken "github.com/Greccl/xxis/internal/token"
)

func resetModules() {
	modules = make(map[string]*Module)
	defaultGlobal = &GlobalContext{
		Globals: NewVariableStore(true),
		Modules: modules,
	}
}

func writeSource(t *testing.T, dir, name, src string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	return path
}

func astFrom(src string) *xxisToken.Token {
	return xxisParser.Build_ast_from_tokens(xxisParser.Enumerate_tokens(xxisParser.Enumerate_string(src)))
}

func TestGetOrImportCachesByNormalizedPath(t *testing.T) {
	resetModules()
	dir := t.TempDir()
	path := writeSource(t, dir, "lib.xxis", "cmd lib\n")

	first, err := GetOrImport(path)
	if err != nil {
		t.Fatalf("GetOrImport first: %v", err)
	}
	second, err := GetOrImport(path)
	if err != nil {
		t.Fatalf("GetOrImport second: %v", err)
	}
	if first != second {
		t.Fatalf("GetOrImport returned different modules for same path")
	}

	normalized, err := normalizePath(path)
	if err != nil {
		t.Fatalf("normalizePath: %v", err)
	}
	if _, ok := modules[normalized]; !ok {
		t.Fatalf("modules does not contain normalized path %q", normalized)
	}
}

func TestExecImportRegistersModuleAlias(t *testing.T) {
	resetModules()
	dir := t.TempDir()
	path := writeSource(t, dir, "lib.xxis", "cmd lib\n")

	ast := xxisParser.Build_ast_from_tokens(xxisParser.Enumerate_tokens(xxisParser.Enumerate_string(fmt.Sprintf("import %q as lib", path))))
	parent := New_Module()
	parent.Path = filepath.Join(dir, "main.xxis")
	ctx := NewThreadContext(DefaultGlobalContext(), parent)

	if err := Exec(ast, ctx); err != nil {
		t.Fatalf("Exec: %v", err)
	}

	value, ok := parent.Vars["lib"]
	if !ok {
		t.Fatalf("parent module does not contain import alias")
	}
	moduleValue, ok := value.(ModuleValue)
	if !ok {
		t.Fatalf("alias value = %T, want ModuleValue", value)
	}

	normalized, err := normalizePath(path)
	if err != nil {
		t.Fatalf("normalizePath: %v", err)
	}
	if moduleValue.Module == nil || moduleValue.Module.Path != normalized {
		t.Fatalf("alias module path = %q, want %q", moduleValue.Expand(), normalized)
	}
}

func TestExecRequiresThreadContext(t *testing.T) {
	err := Exec(astFrom("cmd noop\n"), nil)
	if err == nil {
		t.Fatalf("Exec did not reject nil context")
	}
	if !strings.Contains(err.Error(), "thread context") {
		t.Fatalf("Exec error = %q, want thread context error", err)
	}
}

func TestExecVarScopes(t *testing.T) {
	resetModules()
	global := DefaultGlobalContext()
	mod := New_Module()
	ctx := NewThreadContext(global, mod)

	if err := Exec(astFrom("var -g g = G\nvar -s s = S\nvar m = M\n"), ctx); err != nil {
		t.Fatalf("Exec: %v", err)
	}

	if value, ok := global.Globals.Get("g"); !ok || value.Expand() != "G" {
		t.Fatalf("global g = %v %v, want G", value, ok)
	}
	if value, ok := mod.Shared.Get("s"); !ok || value.Expand() != "S" {
		t.Fatalf("shared s = %v %v, want S", value, ok)
	}
	if value, ok := mod.Vars["m"]; !ok || value.Expand() != "M" {
		t.Fatalf("module local m = %v %v, want M", value, ok)
	}

	funcCtx := ctx.FunctionContext()
	if value, ok := funcCtx.Resolve("m"); !ok || value.Expand() != "M" {
		t.Fatalf("function resolve module local m = %v %v, want M", value, ok)
	}
	if value, ok := funcCtx.Resolve("s"); !ok || value.Expand() != "S" {
		t.Fatalf("function resolve shared s = %v %v, want S", value, ok)
	}
	if value, ok := funcCtx.Resolve("g"); !ok || value.Expand() != "G" {
		t.Fatalf("function resolve global g = %v %v, want G", value, ok)
	}
}

func TestExecFunctionLocalDoesNotWriteModuleLocal(t *testing.T) {
	resetModules()
	mod := New_Module()
	ctx := NewThreadContext(DefaultGlobalContext(), mod)
	funcCtx := ctx.FunctionContext()

	if err := Exec(astFrom("var x = function\n"), funcCtx); err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if _, ok := mod.Vars["x"]; ok {
		t.Fatalf("function local x leaked into module locals")
	}
	if value, ok := funcCtx.Resolve("x"); !ok || value.Expand() != "function" {
		t.Fatalf("function local x = %v %v, want function", value, ok)
	}
}

func TestExecVarOperations(t *testing.T) {
	resetModules()
	mod := New_Module()
	ctx := NewThreadContext(DefaultGlobalContext(), mod)

	if err := Exec(astFrom("var a = hola\nvar a < mundo\nvar a > start-\n"), ctx); err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if value, ok := mod.Vars["a"]; !ok || value.Expand() != "start-holamundo" {
		t.Fatalf("a = %v %v, want start-holamundo", value, ok)
	}

	if err := Exec(astFrom("var a delete\n"), ctx); err != nil {
		t.Fatalf("Exec delete: %v", err)
	}
	if _, ok := mod.Vars["a"]; ok {
		t.Fatalf("a still exists after delete")
	}
}

func TestExecUniversalScopeIsReserved(t *testing.T) {
	resetModules()
	mod := New_Module()
	ctx := NewThreadContext(DefaultGlobalContext(), mod)

	err := Exec(astFrom("var -u name = value\n"), ctx)
	if err == nil {
		t.Fatalf("Exec did not reject universal scope")
	}
	if !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("Exec error = %q, want reserved scope error", err)
	}
}
