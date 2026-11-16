package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveDomainPathsDetectsNestedModule(t *testing.T) {
	workingDir := t.TempDir()
	modulePath := filepath.Join(workingDir, "example", "rent")
	writeDomainModule(t, modulePath, "merchant")

	args := &args{domain: "merchant"}
	if err := args.resolveDomainPaths(workingDir); err != nil {
		t.Fatal(err)
	}

	if args.modulePath != modulePath {
		t.Fatalf("modulePath = %q, want %q", args.modulePath, modulePath)
	}
	if args.handlerPath != filepath.Join(modulePath, "domain", "merchant") {
		t.Fatalf("handlerPath = %q", args.handlerPath)
	}
	if args.basicInfoPath != filepath.Join(modulePath, "domain", "merchant", "boot", "default.go") {
		t.Fatalf("basicInfoPath = %q", args.basicInfoPath)
	}
}

func TestResolveDomainPathsHonorsExplicitPaths(t *testing.T) {
	modulePath := t.TempDir()
	writeDomainModule(t, modulePath, "merchant")
	handlerPath := filepath.Join(modulePath, "custom", "handlers")
	basicInfoPath := filepath.Join(modulePath, "custom", "openapi.go")

	args := &args{
		modulePath:    modulePath,
		domain:        "merchant",
		handlerPath:   handlerPath,
		basicInfoPath: basicInfoPath,
	}
	if err := args.resolveDomainPaths(modulePath); err != nil {
		t.Fatal(err)
	}

	if args.handlerPath != handlerPath || args.basicInfoPath != basicInfoPath {
		t.Fatalf("explicit paths were overwritten: handler=%q basicInfo=%q", args.handlerPath, args.basicInfoPath)
	}
}

func TestResolveDomainPathsReportsAmbiguousModules(t *testing.T) {
	workingDir := t.TempDir()
	writeDomainModule(t, filepath.Join(workingDir, "one"), "merchant")
	writeDomainModule(t, filepath.Join(workingDir, "two"), "merchant")

	args := &args{domain: "merchant"}
	err := args.resolveDomainPaths(workingDir)
	if err == nil || !strings.Contains(err.Error(), "multiple Go modules") {
		t.Fatalf("expected ambiguity error, got %v", err)
	}
}

func TestResolveDomainPathsRejectsPathTraversal(t *testing.T) {
	args := &args{domain: "../merchant"}
	err := args.resolveDomainPaths(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "single directory name") {
		t.Fatalf("expected invalid domain error, got %v", err)
	}
}

func writeDomainModule(t *testing.T, modulePath, domain string) {
	t.Helper()
	writeFile(t, filepath.Join(modulePath, "go.mod"), "module test\n")
	writeFile(t, filepath.Join(modulePath, "domain", domain, "boot", "default.go"), "package boot\n")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
