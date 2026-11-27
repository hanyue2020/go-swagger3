package schema

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"sort"
	"strings"
)

type enumValueDescription struct {
	position token.Pos
	value    int64
	comment  string
}

func (p *parser) enumValueDescriptions(pkgPath, typeName string) (string, error) {
	packages, err := p.GetPkgAst(pkgPath)
	if err != nil {
		return "", err
	}

	var enumValues []enumValueDescription
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			for _, declaration := range file.Decls {
				generalDeclaration, ok := declaration.(*ast.GenDecl)
				if !ok || generalDeclaration.Tok != token.CONST {
					continue
				}
				var previousValues []ast.Expr
				var previousType ast.Expr
				for i, spec := range generalDeclaration.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					if len(valueSpec.Values) > 0 {
						previousValues = valueSpec.Values
					}
					if valueSpec.Type != nil {
						previousType = valueSpec.Type
					}
					if !isNamedType(previousType, typeName) || len(previousValues) != len(valueSpec.Names) {
						continue
					}

					comment := ""
					if valueSpec.Comment != nil {
						comment = strings.TrimSpace(valueSpec.Comment.Text())
					} else if valueSpec.Doc != nil {
						comment = strings.TrimSpace(valueSpec.Doc.Text())
					}
					if comment == "" {
						continue
					}
					for j := range valueSpec.Names {
						value, ok := evalIntegerConst(previousValues[j], i)
						if !ok {
							continue
						}
						enumValues = append(enumValues, enumValueDescription{
							position: valueSpec.Pos(),
							value:    value,
							comment:  comment,
						})
					}
				}
			}
		}
	}

	if len(enumValues) == 0 {
		return "", nil
	}
	sort.Slice(enumValues, func(i, j int) bool {
		return enumValues[i].position < enumValues[j].position
	})
	descriptions := make([]string, 0, len(enumValues))
	for _, value := range enumValues {
		descriptions = append(descriptions, fmt.Sprintf("%d %s", value.value, value.comment))
	}
	return strings.Join(descriptions, "\n"), nil
}

func isNamedType(expr ast.Expr, typeName string) bool {
	switch typedExpr := expr.(type) {
	case *ast.Ident:
		return typedExpr.Name == typeName
	case *ast.SelectorExpr:
		return typedExpr.Sel.Name == typeName
	default:
		return false
	}
}

func evalIntegerConst(expr ast.Expr, iota int) (int64, bool) {
	value, ok := evalConst(expr, iota)
	if !ok {
		return 0, false
	}
	return constant.Int64Val(constant.ToInt(value))
}

func evalConst(expr ast.Expr, iota int) (constant.Value, bool) {
	switch typedExpr := expr.(type) {
	case *ast.Ident:
		if typedExpr.Name == "iota" {
			return constant.MakeInt64(int64(iota)), true
		}
	case *ast.BasicLit:
		value := constant.MakeFromLiteral(typedExpr.Value, typedExpr.Kind, 0)
		if value.Kind() != constant.Unknown {
			return value, true
		}
	case *ast.ParenExpr:
		return evalConst(typedExpr.X, iota)
	case *ast.UnaryExpr:
		value, ok := evalConst(typedExpr.X, iota)
		if !ok {
			return nil, false
		}
		switch typedExpr.Op {
		case token.ADD, token.SUB, token.XOR:
			return constant.UnaryOp(typedExpr.Op, value, 0), true
		}
	case *ast.BinaryExpr:
		left, leftOK := evalConst(typedExpr.X, iota)
		right, rightOK := evalConst(typedExpr.Y, iota)
		if !leftOK || !rightOK {
			return nil, false
		}
		switch typedExpr.Op {
		case token.ADD, token.SUB, token.MUL, token.QUO, token.REM,
			token.AND, token.OR, token.XOR, token.AND_NOT, token.SHL, token.SHR:
			return constant.BinaryOp(left, typedExpr.Op, right), true
		}
	}
	return nil, false
}
