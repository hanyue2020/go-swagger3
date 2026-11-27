package schema

import (
	"go/ast"
	goParser "go/parser"
	"go/token"
	"testing"

	oas "github.com/hanyue2020/evadoc/internal/openapi"
	"github.com/hanyue2020/evadoc/internal/parser/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterGenericStructType(t *testing.T) {
	const responsePkg = "example/response"
	const logicPkg = "example/logic"
	const responsePath = "/example/response"

	file, err := goParser.ParseFile(token.NewFileSet(), "response.go", `
package response

type Account struct {
	Name string
}

type ResponseItems[T any] struct {
	Status int
	Data struct {
		Items []T
	}
}

type ResponseEnvelope struct {
	Data any
	Items []any
}
`, goParser.ParseComments)
	require.NoError(t, err)

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

	utils := model.Utils{
		PkgAndSpecs: &model.PkgAndSpecs{
			KnownPkgs: []model.Pkg{{Name: responsePkg, Path: responsePath}},
			KnownNamePkg: map[string]*model.Pkg{
				responsePkg: {Name: responsePkg, Path: responsePath},
			},
			KnownPathPkg:            map[string]*model.Pkg{},
			KnownIDSchema:           map[string]*oas.SchemaObject{},
			TypeSpecs:               map[string]map[string]*ast.TypeSpec{responsePkg: typeSpecs},
			PkgPathAstPkgCache:      map[string]map[string]*ast.Package{},
			PkgNameImportedPkgAlias: map[string]map[string][]string{logicPkg: {"response": {responsePkg}}},
		},
	}
	openAPI := &oas.OpenAPIObject{
		Components: oas.ComponentsObject{Schemas: map[string]*oas.SchemaObject{}},
	}
	p := &parser{Utils: utils, OpenAPI: openAPI}

	schemaID, err := p.RegisterType("/example/logic", logicPkg, "response.ResponseItems<response.Account>", nil)
	require.NoError(t, err)
	assert.Equal(t, "example.response.ResponseItems__generic__Account", schemaID)

	responseSchema, ok := openAPI.Components.Schemas[schemaID]
	require.True(t, ok)
	dataValue, ok := responseSchema.Properties.Get("Data")
	require.True(t, ok)
	dataSchema, ok := dataValue.(*oas.SchemaObject)
	require.True(t, ok)
	itemsValue, ok := dataSchema.Properties.Get("Items")
	require.True(t, ok)
	itemsSchema, ok := itemsValue.(*oas.SchemaObject)
	require.True(t, ok)
	assert.Equal(t, "array", itemsSchema.Type)
	require.NotNil(t, itemsSchema.Items)
	assert.Equal(t, "object", itemsSchema.Items.Type)
	_, hasName := itemsSchema.Items.Properties.Get("Name")
	assert.True(t, hasName)

	envelopeID, err := p.RegisterType("/example/logic", logicPkg, "response.ResponseEnvelope<response.Account>", nil)
	require.NoError(t, err)
	assert.Equal(t, "example.response.ResponseEnvelope__generic__Account", envelopeID)
	envelopeSchema := openAPI.Components.Schemas[envelopeID]
	dataValue, ok = envelopeSchema.Properties.Get("Data")
	require.True(t, ok)
	dataSchema, ok = dataValue.(*oas.SchemaObject)
	require.True(t, ok)
	assert.Equal(t, "#/components/schemas/example.response.Account", dataSchema.Ref)
	itemsValue, ok = envelopeSchema.Properties.Get("Items")
	require.True(t, ok)
	itemsSchema, ok = itemsValue.(*oas.SchemaObject)
	require.True(t, ok)
	require.NotNil(t, itemsSchema.Items)
	assert.Equal(t, "#/components/schemas/example.response.Account", itemsSchema.Items.Ref)
}
