package schema

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	goParser "go/parser"
	"go/token"
	"strings"
)

func (p *parser) instantiateGenericType(currentPkgName, typeName string) (pkgPath, pkgName, specializedName string, handled bool, err error) {
	if strings.HasPrefix(typeName, "map[]") {
		return "", "", "", false, nil
	}
	if strings.Contains(typeName, "<") {
		typeName = strings.NewReplacer("<", "[", ">", "]").Replace(typeName)
	}
	expr, err := goParser.ParseExpr(typeName)
	if err != nil {
		return "", "", "", false, fmt.Errorf("parse generic type %q: %w", typeName, err)
	}

	var base ast.Expr
	var typeArgs []ast.Expr
	switch genericExpr := expr.(type) {
	case *ast.IndexExpr:
		base = genericExpr.X
		typeArgs = []ast.Expr{genericExpr.Index}
	case *ast.IndexListExpr:
		base = genericExpr.X
		typeArgs = genericExpr.Indices
	default:
		return "", "", "", false, nil
	}

	basePackage, baseType, ok := splitQualifiedType(base)
	if !ok {
		return "", "", "", false, nil
	}

	pkgName, pkgPath, typeSpec, ok := p.resolveGenericType(currentPkgName, basePackage, baseType)
	if !ok {
		return "", "", "", false, fmt.Errorf("cannot find generic type %q in package %q", baseType, basePackage)
	}
	var typeParams []string
	if typeSpec.TypeParams == nil || typeSpec.TypeParams.NumFields() == 0 {
		if !hasTypeParameter(typeSpec.Type, "any") {
			return "", "", "", false, fmt.Errorf("type %q is not generic", baseType)
		}
		typeParams = append(typeParams, "any")
	} else {
		for _, field := range typeSpec.TypeParams.List {
			for _, name := range field.Names {
				typeParams = append(typeParams, name.Name)
			}
		}
	}
	if len(typeArgs) != len(typeParams) {
		return "", "", "", false, fmt.Errorf("generic type %q expects %d type arguments, got %d", baseType, len(typeParams), len(typeArgs))
	}

	substitutions := make(map[string]ast.Expr, len(typeParams))
	argNames := make([]string, 0, len(typeArgs))
	for i, arg := range typeArgs {
		normalizedArg, argName, normalizeErr := p.normalizeGenericArgument(currentPkgName, pkgName, arg)
		if normalizeErr != nil {
			return "", "", "", false, normalizeErr
		}
		substitutions[typeParams[i]] = normalizedArg
		argNames = append(argNames, argName)
	}

	specializedName = baseType + "__generic__" + strings.Join(argNames, "_")
	typeSpecs, exists := p.TypeSpecs[pkgName]
	if !exists {
		typeSpecs = make(map[string]*ast.TypeSpec)
		p.TypeSpecs[pkgName] = typeSpecs
	}
	if _, exists := typeSpecs[specializedName]; !exists {
		typeSpecs[specializedName] = &ast.TypeSpec{
			Name: ast.NewIdent(specializedName),
			Type: substituteTypeParameters(typeSpec.Type, substitutions),
		}
	}
	return pkgPath, pkgName, specializedName, true, nil
}

func hasTypeParameter(expr ast.Expr, name string) bool {
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if ok && ident.Name == name {
			found = true
			return false
		}
		return !found
	})
	return found
}

func splitQualifiedType(expr ast.Expr) (packageName, typeName string, ok bool) {
	switch typedExpr := expr.(type) {
	case *ast.Ident:
		return "", typedExpr.Name, true
	case *ast.SelectorExpr:
		packageIdent, ok := typedExpr.X.(*ast.Ident)
		if !ok {
			return "", "", false
		}
		return packageIdent.Name, typedExpr.Sel.Name, true
	default:
		return "", "", false
	}
}

func (p *parser) resolveGenericType(currentPkgName, packageAlias, typeName string) (pkgName, pkgPath string, typeSpec *ast.TypeSpec, ok bool) {
	candidates := []string{currentPkgName}
	if packageAlias != "" {
		candidates = p.importedPackageNames(currentPkgName, packageAlias)
		if _, exists := p.TypeSpecs[packageAlias]; exists {
			candidates = append([]string{packageAlias}, candidates...)
		}
	}

	for _, candidate := range candidates {
		spec, exists := p.getTypeSpec(candidate, typeName)
		if !exists {
			continue
		}
		pkg, exists := p.KnownNamePkg[candidate]
		if exists {
			return candidate, pkg.Path, spec, true
		}
		for _, knownPkg := range p.KnownPkgs {
			if knownPkg.Name == candidate {
				return candidate, knownPkg.Path, spec, true
			}
		}
		if candidate == currentPkgName {
			return candidate, "", spec, true
		}
	}
	return "", "", nil, false
}

func (p *parser) importedPackageNames(pkgName, alias string) []string {
	importedPaths := p.PkgNameImportedPkgAlias[pkgName][alias]
	names := make([]string, 0, len(importedPaths))
	for _, importedPath := range importedPaths {
		if _, exists := p.TypeSpecs[importedPath]; exists {
			names = append(names, importedPath)
			continue
		}
		if pkg, exists := p.KnownNamePkg[importedPath]; exists {
			names = append(names, pkg.Name)
		}
	}
	return names
}

func (p *parser) normalizeGenericArgument(currentPkgName, genericPkgName string, arg ast.Expr) (ast.Expr, string, error) {
	selector, isSelector := arg.(*ast.SelectorExpr)
	if !isSelector {
		return substituteTypeParameters(arg, nil), genericTypeName(arg), nil
	}
	packageIdent, ok := selector.X.(*ast.Ident)
	if !ok {
		return nil, "", fmt.Errorf("unsupported generic argument %q", formatTypeExpr(arg))
	}

	importedNames := p.importedPackageNames(currentPkgName, packageIdent.Name)
	if len(importedNames) == 0 {
		return nil, "", fmt.Errorf("cannot resolve package alias %q in generic argument %q", packageIdent.Name, formatTypeExpr(arg))
	}
	argPkgName := importedNames[0]
	if argPkgName == genericPkgName {
		return ast.NewIdent(selector.Sel.Name), selector.Sel.Name, nil
	}

	targetAliases := p.PkgNameImportedPkgAlias[genericPkgName]
	for alias, paths := range targetAliases {
		for _, path := range paths {
			if importedNames[0] == path {
				return &ast.SelectorExpr{X: ast.NewIdent(alias), Sel: ast.NewIdent(selector.Sel.Name)}, alias + "_" + selector.Sel.Name, nil
			}
			if pkg, exists := p.KnownNamePkg[importedNames[0]]; exists && pkg.Path == path {
				return &ast.SelectorExpr{X: ast.NewIdent(alias), Sel: ast.NewIdent(selector.Sel.Name)}, alias + "_" + selector.Sel.Name, nil
			}
		}
	}

	importPath, ok := p.knownPackagePath(importedNames[0])
	if !ok {
		return nil, "", fmt.Errorf("cannot find import path for generic argument package %q", importedNames[0])
	}
	if targetAliases == nil {
		targetAliases = make(map[string][]string)
		if p.PkgNameImportedPkgAlias == nil {
			p.PkgNameImportedPkgAlias = make(map[string]map[string][]string)
		}
		p.PkgNameImportedPkgAlias[genericPkgName] = targetAliases
	}
	alias := packageIdent.Name
	for suffix := 0; ; suffix++ {
		paths, exists := targetAliases[alias]
		if !exists || containsString(paths, importPath) {
			break
		}
		alias = fmt.Sprintf("goSwagger3GenericArg%d", suffix)
	}
	if !containsString(targetAliases[alias], importPath) {
		targetAliases[alias] = append(targetAliases[alias], importPath)
	}
	return &ast.SelectorExpr{X: ast.NewIdent(alias), Sel: ast.NewIdent(selector.Sel.Name)}, alias + "_" + selector.Sel.Name, nil
}

func (p *parser) knownPackagePath(pkgName string) (string, bool) {
	if pkg, ok := p.KnownNamePkg[pkgName]; ok {
		return pkg.Path, true
	}
	for _, pkg := range p.KnownPkgs {
		if pkg.Name == pkgName {
			return pkg.Path, true
		}
	}
	return "", false
}

func containsString(values []string, value string) bool {
	for _, current := range values {
		if current == value {
			return true
		}
	}
	return false
}

func substituteTypeParameters(expr ast.Expr, substitutions map[string]ast.Expr) ast.Expr {
	switch typedExpr := expr.(type) {
	case *ast.Ident:
		if replacement, ok := substitutions[typedExpr.Name]; ok {
			return substituteTypeParameters(replacement, nil)
		}
		return ast.NewIdent(typedExpr.Name)
	case *ast.StarExpr:
		return &ast.StarExpr{X: substituteTypeParameters(typedExpr.X, substitutions)}
	case *ast.ArrayType:
		return &ast.ArrayType{Len: typedExpr.Len, Elt: substituteTypeParameters(typedExpr.Elt, substitutions)}
	case *ast.MapType:
		return &ast.MapType{Key: substituteTypeParameters(typedExpr.Key, substitutions), Value: substituteTypeParameters(typedExpr.Value, substitutions)}
	case *ast.SelectorExpr:
		return &ast.SelectorExpr{X: substituteTypeParameters(typedExpr.X, substitutions), Sel: ast.NewIdent(typedExpr.Sel.Name)}
	case *ast.IndexExpr:
		return &ast.IndexExpr{X: substituteTypeParameters(typedExpr.X, substitutions), Index: substituteTypeParameters(typedExpr.Index, substitutions)}
	case *ast.IndexListExpr:
		indices := make([]ast.Expr, len(typedExpr.Indices))
		for i, index := range typedExpr.Indices {
			indices[i] = substituteTypeParameters(index, substitutions)
		}
		return &ast.IndexListExpr{X: substituteTypeParameters(typedExpr.X, substitutions), Indices: indices}
	case *ast.ParenExpr:
		return &ast.ParenExpr{X: substituteTypeParameters(typedExpr.X, substitutions)}
	case *ast.StructType:
		return &ast.StructType{Fields: substituteFieldList(typedExpr.Fields, substitutions)}
	case *ast.InterfaceType:
		return &ast.InterfaceType{Methods: substituteFieldList(typedExpr.Methods, substitutions)}
	default:
		return expr
	}
}

func substituteFieldList(fields *ast.FieldList, substitutions map[string]ast.Expr) *ast.FieldList {
	if fields == nil {
		return nil
	}
	substituted := &ast.FieldList{List: make([]*ast.Field, len(fields.List))}
	for i, field := range fields.List {
		names := make([]*ast.Ident, len(field.Names))
		for j, name := range field.Names {
			names[j] = ast.NewIdent(name.Name)
		}
		substituted.List[i] = &ast.Field{
			Names:   names,
			Type:    substituteTypeParameters(field.Type, substitutions),
			Tag:     field.Tag,
			Doc:     field.Doc,
			Comment: field.Comment,
		}
	}
	return substituted
}

func genericTypeName(expr ast.Expr) string {
	switch typedExpr := expr.(type) {
	case *ast.Ident:
		return typedExpr.Name
	case *ast.SelectorExpr:
		if packageIdent, ok := typedExpr.X.(*ast.Ident); ok {
			return packageIdent.Name + "_" + typedExpr.Sel.Name
		}
	}
	return sanitizeGenericTypeName(formatTypeExpr(expr))
}

func sanitizeGenericTypeName(name string) string {
	var sanitized strings.Builder
	lastUnderscore := false
	for _, r := range name {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			sanitized.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			sanitized.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(sanitized.String(), "_")
}

func formatTypeExpr(expr ast.Expr) string {
	var buffer bytes.Buffer
	if err := format.Node(&buffer, token.NewFileSet(), expr); err != nil {
		return fmt.Sprint(expr)
	}
	return buffer.String()
}
