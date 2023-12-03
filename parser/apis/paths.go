package apis

import (
	"fmt"
	"go/ast"
)

func (p *parser) parsePaths() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name
		// p.debug(pkgName, "->", pkgPath)

		astPkgs, err := p.schemaParser.GetPkgAst(pkgPath)
		if err != nil {
			if p.RunInStrictMode {
				return fmt.Errorf("parsePaths: parse of %s package cause error: %s", pkgPath, err)
			}

			p.Debugf("parsePaths: parse of %s package cause error: %s", pkgPath, err)
			continue
		}

		for _, astPackage := range astPkgs {
			if err := p.parsePathFromPackage(astPackage, pkgPath, pkgName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *parser) parsePathFromPackage(astPackage *ast.Package, pkgPath string, pkgName string) error {
	for _, astFile := range astPackage.Files {
		if err := p.parsePathFromFile(astFile, pkgPath, pkgName); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) parsePathFromFile(astFile *ast.File, pkgPath string, pkgName string) error {
	for _, astDeclaration := range astFile.Decls {
		if err := p.parsePathFromFuncDeclaration(astDeclaration, pkgPath, pkgName); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) parsePathFromFuncDeclaration(astDeclaration ast.Decl, pkgPath string, pkgName string) error {
	switch astDeclaration.(type) {
	case *ast.FuncDecl:
		astFuncDeclaration := astDeclaration.(*ast.FuncDecl)
		if astFuncDeclaration.Doc != nil && astFuncDeclaration.Doc.List != nil {
			if err := p.operationParser.Parse(pkgPath, pkgName, astFuncDeclaration.Doc.List); err != nil {
				return err
			}
		}
	case *ast.GenDecl:
		astGenDeclaration := astDeclaration.(*ast.GenDecl)
		for _, sp := range astGenDeclaration.Specs {
			importSpec, ok := sp.(*ast.ImportSpec)
			if !ok {
				continue
			}
			p.Debug(importSpec.Path.Value, importSpec.Name)
		}
	}
	return nil
}
