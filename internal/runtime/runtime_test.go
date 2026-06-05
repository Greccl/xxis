package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	xxisParser "github.com/Greccl/xxis/internal/parser"
)

func resetModules() {
	modules = make(map[string]*Module)
}

func writeSource(t *testing.T, dir, name, src string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	return path
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

	if err := Exec(ast, parent); err != nil {
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
