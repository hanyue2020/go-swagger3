package schema

import (
	"go/ast"
	goParser "go/parser"
	"go/token"
	"testing"

	oas "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldDescriptionIncludesNamedTypeComment(t *testing.T) {
	const typesPkg = "example/types"
	const requestPkg = "example/request"
	const typesPath = "/example/types"
	const requestPath = "/example/request"

	typesFile, err := goParser.ParseFile(token.NewFileSet(), "bills.go", `
package types

type BillType int32 // BillType 类型说明

const (
	BillTypeDeposit BillType = iota + 1 // 后台上分
)

type MerchantBillType int32 // 1 类型说明 2 类型说明

const (
	MerchantBillTypeDeposit MerchantBillType = iota + 1 // 充值
	MerchantBillTypeCredit                              // 上分
)
`, goParser.ParseComments)
	require.NoError(t, err)
	requestFile, err := goParser.ParseFile(token.NewFileSet(), "merchant_bill.go", `
package request

type Filter struct {
	Type types.MerchantBillType // 帐单类型
	Types []types.MerchantBillType // 帐单类型组合
	BillType types.BillType // BillType 字段
}
`, goParser.ParseComments)
	require.NoError(t, err)

	typeSpecs := map[string]map[string]*ast.TypeSpec{
		typesPkg:   collectTypeSpecs(typesFile),
		requestPkg: collectTypeSpecs(requestFile),
	}
	utils := model.Utils{
		PkgAndSpecs: &model.PkgAndSpecs{
			KnownPkgs: []model.Pkg{
				{Name: typesPkg, Path: typesPath},
				{Name: requestPkg, Path: requestPath},
			},
			KnownNamePkg: map[string]*model.Pkg{
				typesPkg:   {Name: typesPkg, Path: typesPath},
				requestPkg: {Name: requestPkg, Path: requestPath},
			},
			KnownPathPkg:  map[string]*model.Pkg{},
			KnownIDSchema: map[string]*oas.SchemaObject{},
			TypeSpecs:     typeSpecs,
			PkgPathAstPkgCache: map[string]map[string]*ast.Package{
				typesPath: {
					"types": {Name: "types", Files: map[string]*ast.File{"bills.go": typesFile}},
				},
			},
			PkgNameImportedPkgAlias: map[string]map[string][]string{requestPkg: {"types": {typesPkg}}},
		},
	}
	openAPI := &oas.OpenAPIObject{
		Components: oas.ComponentsObject{Schemas: map[string]*oas.SchemaObject{}},
	}
	p := &parser{Utils: utils, OpenAPI: openAPI}

	filterType := typeSpecs[requestPkg]["Filter"]
	filterSchema, err := p.ParseSchemaObject(requestPath, requestPkg, "Filter", filterType.Type)
	require.NoError(t, err)
	fieldValue, ok := filterSchema.Properties.Get("Type")
	require.True(t, ok)
	fieldSchema, ok := fieldValue.(*oas.SchemaObject)
	require.True(t, ok)
	assert.Equal(t, "帐单类型\n1 充值\n2 上分", fieldSchema.Description)
	assert.Empty(t, fieldSchema.Ref)
	require.Len(t, fieldSchema.AllOf, 1)
	assert.Equal(t, "#/components/schemas/example.types.MerchantBillType", fieldSchema.AllOf[0].Ref)

	arrayValue, ok := filterSchema.Properties.Get("Types")
	require.True(t, ok)
	arraySchema, ok := arrayValue.(*oas.SchemaObject)
	require.True(t, ok)
	assert.Equal(t, "帐单类型组合\n1 充值\n2 上分", arraySchema.Description)
	require.NotNil(t, arraySchema.Items)
	assert.Equal(t, "1 充值\n2 上分", arraySchema.Items.Description)

	billTypeValue, ok := filterSchema.Properties.Get("BillType")
	require.True(t, ok)
	billTypeSchema, ok := billTypeValue.(*oas.SchemaObject)
	require.True(t, ok)
	assert.Equal(t, "BillType 字段\n1 后台上分", billTypeSchema.Description)
}

func collectTypeSpecs(file *ast.File) map[string]*ast.TypeSpec {
	typeSpecs := make(map[string]*ast.TypeSpec)
	for _, declaration := range file.Decls {
		generalDeclaration, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range generalDeclaration.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if ok {
				typeSpecs[typeSpec.Name.Name] = typeSpec
			}
		}
	}
	return typeSpecs
}
