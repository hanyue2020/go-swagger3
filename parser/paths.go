package parser

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hanyue2020/go-swagger3/parser/utils"
)

func (p *parser) verifyAndSetPaths() error {
	// check modulePath is exist
	if err := p.verifyModulePath(); err != nil {
		return err
	}

	// check go.mod file is exist
	if err := p.setGoModFilePath(); err != nil {
		return err
	}

	// check basicInfoPath is exist
	if err := p.verifyBasicInfoPath(); err != nil {
		return err
	}

	// get module name from go.mod file
	if err := p.getModuleNameFromGoModFile(); err != nil {
		return err
	}

	// check go module cache path is exist ($GOPATH/pkg/mod)
	if err := p.setGoModCachePath(); err != nil {
		return err
	}

	// check handlerPath is exist
	if err := p.setHandlerPath(); err != nil {
		return err
	}
	return nil
}

func (p *parser) setHandlerPath() error {
	var err error
	if p.HandlerPath != "" {
		p.HandlerPath, err = filepath.Abs(p.HandlerPath)
		if err != nil {
			return err
		}
		_, err := os.Stat(p.HandlerPath)
		if err != nil {
			if os.IsNotExist(err) {
				return err
			}
			return fmt.Errorf("cannot get information of %s: %s", p.HandlerPath, err)
		}
	}
	p.Debugf("handler path: %s", p.HandlerPath)
	return nil
}

func (p *parser) setGoModCachePath() error {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		current, err := user.Current()
		if err != nil {
			return fmt.Errorf("cannot get current user: %s", err)
		}
		goPath = filepath.Join(current.HomeDir, "go")
	}
	goModCachePath := filepath.Join(goPath, "pkg", "mod")
	goModCacheInfo, err := os.Stat(goModCachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("cannot get information of %s: %s", goModCachePath, err)
	}
	if !goModCacheInfo.IsDir() {
		return fmt.Errorf("%s should be a directory", goModCachePath)
	}
	p.GoModCachePath = goModCachePath
	p.Debugf("go module cache path: %s", p.GoModCachePath)
	return nil
}

func (p *parser) getModuleNameFromGoModFile() error {
	moduleName := utils.GetModuleNameFromGoMod(p.GoModFilePath)
	if moduleName == "" {
		return fmt.Errorf("cannot get module name from %s", p.GoModFilePath)
	}
	p.ModuleName = moduleName
	p.Debugf("module name: %s", p.ModuleName)
	return nil
}

func (p *parser) verifyBasicInfoPath() error {
	if p.BasicInfoPath == "" {
		mainFile, err := findBestMainFile(p.ModulePath)
		if err != nil {
			return err
		}
		p.BasicInfoPath = mainFile
	} else {
		basicInfoPath, err := os.Stat(p.BasicInfoPath)
		if err != nil {
			if os.IsNotExist(err) {
				return err
			}
			return fmt.Errorf("cannot get information of %s: %s", p.BasicInfoPath, err)
		}
		if basicInfoPath.IsDir() {
			return fmt.Errorf("basicInfoPath should not be a directory")
		}
	}
	p.Debugf("main file path: %s", p.BasicInfoPath)
	return nil
}

func findBestMainFile(modulePath string) (string, error) {
	var candidates []string
	walkErr := filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if isIgnoredDir(path, modulePath) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if utils.IsMainFile(path) {
			candidates = append(candidates, path)
		}
		return nil
	})
	if walkErr != nil {
		return "", walkErr
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("cannot find a valid main file in %s, please specify --basic-info-path or --main-file-path explicitly", modulePath)
	}

	sort.Strings(candidates)
	best := candidates[0]
	bestScore := scoreMainCandidate(modulePath, best)
	bestCount := 1
	for _, candidate := range candidates[1:] {
		score := scoreMainCandidate(modulePath, candidate)
		if score < bestScore {
			best = candidate
			bestScore = score
			bestCount = 1
			continue
		}
		if score == bestScore {
			bestCount++
		}
	}
	if bestCount > 1 {
		return "", fmt.Errorf("multiple main files found under %s: %s; please specify --basic-info-path or --main-file-path explicitly", modulePath, strings.Join(candidates, ", "))
	}
	return best, nil
}

func scoreMainCandidate(modulePath, candidate string) int {
	rel, err := filepath.Rel(modulePath, candidate)
	if err != nil {
		return 1 << 30
	}
	if rel == "" || rel == "." {
		return 0
	}
	depth := strings.Count(filepath.ToSlash(rel), "/")
	if filepath.Base(candidate) == "main.go" {
		depth--
	}
	if depth < 0 {
		depth = 0
	}
	return depth
}

func isIgnoredDir(path, modulePath string) bool {
	if path == modulePath {
		return false
	}
	base := filepath.Base(path)
	return base == ".git" || base == "vendor" || base == "node_modules" || strings.HasPrefix(base, ".")
}

func (p *parser) setGoModFilePath() error {
	goModFilePath := filepath.Join(p.ModulePath, "go.mod")
	goModFileInfo, err := os.Stat(goModFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("cannot get information of %s: %s", goModFilePath, err)
	}
	if goModFileInfo.IsDir() {
		return fmt.Errorf("%s should be a file", goModFilePath)
	}
	p.GoModFilePath = goModFilePath
	p.Debugf("go.mod file path: %s", p.GoModFilePath)
	return err
}

func (p *parser) verifyModulePath() error {
	var err error
	p.ModulePath, err = filepath.Abs(p.ModulePath)
	if err != nil {
		return err
	}
	moduleInfo, err := os.Stat(p.ModulePath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("cannot get information of %s: %s", p.ModulePath, err)
	}
	if !moduleInfo.IsDir() {
		return fmt.Errorf("modulePath should be a directory")
	}
	p.Debugf("module path: %s", p.ModulePath)
	return err
}
