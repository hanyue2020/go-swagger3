package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (a *args) resolveDomainPaths(workingDir string) error {
	if a.domain == "" {
		return nil
	}
	if a.domain == "." || a.domain == ".." || filepath.Base(a.domain) != a.domain || strings.ContainsAny(a.domain, `/\`) {
		return fmt.Errorf("--domain must be a single directory name, got %q", a.domain)
	}

	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	var modulePath string
	if a.modulePath != "" {
		modulePath = a.modulePath
		if !filepath.IsAbs(modulePath) {
			modulePath = filepath.Join(workingDir, modulePath)
		}
		modulePath, err = filepath.Abs(modulePath)
		if err != nil {
			return fmt.Errorf("resolve module path: %w", err)
		}
		if err := validateDomainModule(modulePath, a.domain); err != nil {
			return err
		}
	} else {
		modulePath, err = findDomainModule(workingDir, a.domain)
		if err != nil {
			return err
		}
	}

	domainPath := filepath.Join(modulePath, "domain", a.domain)
	a.modulePath = modulePath
	if a.handlerPath == "" {
		a.handlerPath = domainPath
	}
	if a.basicInfoPath == "" {
		a.basicInfoPath = filepath.Join(domainPath, "boot", "default.go")
	}
	return nil
}

func findDomainModule(workingDir, domain string) (string, error) {
	candidates := make(map[string]struct{})
	addCandidate := func(path string) {
		if validateDomainModule(path, domain) == nil {
			candidates[path] = struct{}{}
		}
	}

	for path := workingDir; ; path = filepath.Dir(path) {
		addCandidate(path)
		if filepath.Dir(path) == path {
			break
		}
	}

	err := filepath.WalkDir(workingDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		if path != workingDir && isDomainSearchIgnoredDir(entry.Name()) {
			return filepath.SkipDir
		}
		if entry.Name() == "domain" {
			addCandidate(filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("search for domain %q under %s: %w", domain, workingDir, err)
	}

	paths := make([]string, 0, len(candidates))
	for path := range candidates {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	switch len(paths) {
	case 0:
		return "", fmt.Errorf("cannot find a Go module containing domain/%s under %s; specify --module-path", domain, workingDir)
	case 1:
		return paths[0], nil
	default:
		return "", fmt.Errorf("multiple Go modules contain domain/%s: %s; specify --module-path", domain, strings.Join(paths, ", "))
	}
}

func validateDomainModule(modulePath, domain string) error {
	goModInfo, err := os.Stat(filepath.Join(modulePath, "go.mod"))
	if err != nil {
		return err
	}
	if goModInfo.IsDir() {
		return fmt.Errorf("%s/go.mod is not a file", modulePath)
	}

	domainPath := filepath.Join(modulePath, "domain", domain)
	domainInfo, err := os.Stat(domainPath)
	if err != nil {
		return err
	}
	if !domainInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", domainPath)
	}

	bootPath := filepath.Join(domainPath, "boot", "default.go")
	bootInfo, err := os.Stat(bootPath)
	if err != nil {
		return err
	}
	if bootInfo.IsDir() {
		return fmt.Errorf("%s is not a file", bootPath)
	}
	return nil
}

func isDomainSearchIgnoredDir(name string) bool {
	return name == "vendor" || name == "node_modules" || strings.HasPrefix(name, ".")
}
