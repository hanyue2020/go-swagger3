package operations

import (
	"errors"
	"testing"

	oas "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/schema"
	"github.com/stretchr/testify/assert"
)

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
