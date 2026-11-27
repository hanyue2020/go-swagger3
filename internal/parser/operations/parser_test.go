package operations

import (
	"errors"
	"testing"

	oas "github.com/hanyue2020/evadoc/internal/openapi"
	"github.com/hanyue2020/evadoc/internal/parser/schema"
	"github.com/hanyue2020/evadoc/internal/parser/schema/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_ParseResponseCommentWithAngleBracketGenericType(t *testing.T) {
	schemaParser := &mocks.SchemaParser{}
	schemaParser.
		On("RegisterType", "/test/path", "pkgName", "response.ResponseArray<response.Demo>").
		Return("pkg.response.ResponseArray__generic__Demo", nil)
	operationParser := parser{Parser: schemaParser}
	operation := &oas.OperationObject{Responses: make(map[string]*oas.ResponseObject)}

	err := operationParser.parseResponseComment(
		"/test/path",
		"pkgName",
		operation,
		`200 object response.ResponseArray<response.Demo>{}`,
	)

	assert.NoError(t, err)
	response, ok := operation.Responses["200"]
	assert.True(t, ok)
	assert.Equal(t, "#/components/schemas/pkg.response.ResponseArray__generic__Demo", response.Content[oas.ContentTypeJson].Schema.Ref)
	schemaParser.AssertExpectations(t)
}

func Test_ParseHeader(t *testing.T) {
	tests := []struct {
		name               string
		schemaParser       schema.Parser
		wantErr            bool
		errMsg             string
		expectedParameters []oas.ParameterObject
	}{
		{
			name:         "Should add parameters with ref",
			schemaParser: schema.SetupUpSchemaParseMocks(schema.GetSchemaObject(), nil),
			wantErr:      false,
			expectedParameters: []oas.ParameterObject{
				{Ref: "#/components/parameters/ContentType"},
				{Ref: "#/components/parameters/Version"},
				{Ref: "#/components/parameters/Authorization"},
			},
		},
		{
			name:         "Should return error if fails parsing the schema",
			schemaParser: schema.SetupUpSchemaParseMocks(schema.GetSchemaObject(), errors.New("someErr")),
			wantErr:      true,
			errMsg:       "someErr",
		},
		{
			name:         "Should return error schema properties are nil",
			schemaParser: schema.SetupUpSchemaParseMocks(&oas.SchemaObject{}, nil),
			wantErr:      true,
			errMsg:       "NilSchemaProperties : parseHeaders can not parse Header schema comment",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operationParser := parser{Parser: test.schemaParser}
			operationObject := &oas.OperationObject{}
			err := operationParser.parseHeaders("/test/path", "pkgName", operationObject, "comment")
			if test.wantErr {
				assert.NotNil(t, err)
				assert.EqualError(t, err, test.errMsg)
			}
			assert.Equal(t, test.expectedParameters, operationObject.Parameters)
		})
	}
}

func Test_ParseOperationFromComment_Route(t *testing.T) {
	tests := []struct {
		name         string
		comment      string
		path         string
		expectedGet  bool
		expectedPost bool
	}{
		{name: "lowercase route attr", comment: "@route /api/user [post]", path: "/api/user", expectedPost: true},
		{name: "mixed case route attr", comment: "@Route /api/group/{groupID}/users [get|post]", path: "/api/group/{groupID}/users", expectedGet: true, expectedPost: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operationParser := parser{OpenAPI: &oas.OpenAPIObject{Paths: make(oas.PathsObject)}}
			operation := &oas.OperationObject{}

			err := operationParser.parseOperationFromComment("/test/path", "pkgName", tt.comment, operation)
			assert.NoError(t, err)
			assert.Contains(t, operationParser.OpenAPI.Paths, tt.path)
			assert.NotNil(t, operationParser.OpenAPI.Paths[tt.path])
			if tt.expectedGet {
				assert.Equal(t, operation, operationParser.OpenAPI.Paths[tt.path].Get)
			} else {
				assert.Nil(t, operationParser.OpenAPI.Paths[tt.path].Get)
			}
			if tt.expectedPost {
				assert.Equal(t, operation, operationParser.OpenAPI.Paths[tt.path].Post)
			} else {
				assert.Nil(t, operationParser.OpenAPI.Paths[tt.path].Post)
			}
		})
	}
}
