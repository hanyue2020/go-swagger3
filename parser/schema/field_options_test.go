package schema

import (
	"go/ast"
	"testing"

	oas "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/stretchr/testify/assert"
)

func TestParseFieldOptions(t *testing.T) {
	tests := []struct {
		name         string
		field        *ast.Field
		wantName     string
		disabledName string
		required     bool
		skip         bool
		disabled     bool
		attributes   map[string]string
	}{
		{
			name: "json tag",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "UserID"}},
				Tag:   &ast.BasicLit{Value: "`json:\"user_id,required,omitempty\"`"},
			},
			wantName: "user_id",
			required: true,
		},
		{
			name: "documentation and json tag options",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "Name"}},
				Doc: &ast.CommentGroup{List: []*ast.Comment{
					{Text: "// @go-swagger3 public_name,required"},
					{Text: "// @desc display name"},
				}},
				Tag: &ast.BasicLit{Value: "`json:\"json_name,omitempty\"`"},
			},
			wantName:   "json_name",
			required:   true,
			attributes: map[string]string{"@go-swagger3": "public_name,required", "@desc": "display name"},
		},
		{
			name: "skip annotation",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "Hidden"}},
				Doc: &ast.CommentGroup{List: []*ast.Comment{
					{Text: "// @skip true"},
				}},
			},
			wantName:   "Hidden",
			skip:       true,
			attributes: map[string]string{"@skip": "true"},
		},
		{
			name: "disabled json field",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "Hidden"}},
				Tag:   &ast.BasicLit{Value: "`json:\"-\"`"},
			},
			disabledName: "Hidden",
			disabled:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := parseFieldOptions(test.field)
			assert.Equal(t, test.wantName, got.name)
			assert.Equal(t, test.disabledName, got.disabledName)
			assert.Equal(t, test.required, got.required)
			assert.Equal(t, test.skip, got.skip)
			assert.Equal(t, test.disabled, got.disabled)
			if test.attributes != nil {
				assert.Equal(t, test.attributes, got.attributes)
			}
		})
	}
}

func TestParseFieldTagAndDoc(t *testing.T) {
	structSchema := &oas.SchemaObject{}
	fieldSchema := &oas.SchemaObject{Type: "string"}
	field := &ast.Field{
		Names: []*ast.Ident{{Name: "UserID"}},
		Doc: &ast.CommentGroup{List: []*ast.Comment{
			{Text: "// @desc user identifier"},
		}},
		Tag: &ast.BasicLit{Value: "`json:\"user_id,required,omitempty\"`"},
	}

	skipped, name := (&parser{}).parseFieldTagAndDoc(field, structSchema, fieldSchema)

	assert.False(t, skipped)
	assert.Equal(t, "user_id", name)
	assert.Equal(t, []string{"user_id"}, structSchema.Required)
	assert.Equal(t, "user identifier", fieldSchema.Description)
}
