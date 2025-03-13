package schema

import (
	"encoding/json"
	"go/ast"
	goParser "go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"

	. "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/model"
	"github.com/hanyue2020/go-swagger3/parser/utils"
	"github.com/spf13/cast"
)

type Parser interface {
	GetPkgAst(pkgPath string) (map[string]*ast.Package, error)
	RegisterType(pkgPath, pkgName, typeName string, astExpr ast.Expr) (string, error)
	ParseSchemaObject(pkgPath, pkgName, typeName string, astExpr ast.Expr) (*SchemaObject, error)
}

type parser struct {
	model.Utils
	OpenAPI *OpenAPIObject
}

func NewParser(utils model.Utils, openAPIObject *OpenAPIObject) Parser {
	return &parser{
		Utils:   utils,
		OpenAPI: openAPIObject,
	}
}

func (p *parser) GetPkgAst(pkgPath string) (map[string]*ast.Package, error) {
	if cache, ok := p.PkgPathAstPkgCache[pkgPath]; ok {
		return cache, nil
	}
	ignoreFileFilter := func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}
	astPackages, err := goParser.ParseDir(token.NewFileSet(), pkgPath, ignoreFileFilter, goParser.ParseComments)
	if err != nil {
		return nil, err
	}
	p.PkgPathAstPkgCache[pkgPath] = astPackages
	return astPackages, nil
}

func (p *parser) RegisterType(pkgPath, pkgName, typeName string, astExpr ast.Expr) (string, error) {
	var registerTypeName string

	if utils.IsBasicGoType(typeName) || utils.IsInterfaceType(typeName) {
		registerTypeName = typeName
	} else if schemaObject, ok := p.KnownIDSchema[utils.GenSchemaObjectID(pkgName, typeName)]; ok {
		_, ok := p.OpenAPI.Components.Schemas[schemaObject.ID]
		if !ok {
			// p.OpenAPI.Components.Schemas[utils.ReplaceBackslash(typeName)] = schemaObject
			p.OpenAPI.Components.Schemas[schemaObject.ID] = schemaObject
		}
		return utils.GenSchemaObjectID(pkgName, typeName), nil
	} else {
		schemaObject, err := p.ParseSchemaObject(pkgPath, pkgName, typeName, astExpr)
		if err != nil {
			return "", err
		}
		registerTypeName = schemaObject.ID
		_, ok := p.OpenAPI.Components.Schemas[utils.ReplaceBackslash(registerTypeName)]
		if !ok {
			p.OpenAPI.Components.Schemas[utils.ReplaceBackslash(registerTypeName)] = schemaObject
		}
	}
	return registerTypeName, nil
}

func (p *parser) ParseSchemaObject(pkgPath, pkgName, typeName string, astExpr ast.Expr) (*SchemaObject, error) {
	schemaObject, err, isBasicType := p.parseBasicTypeSchemaObject(pkgPath, pkgName, typeName, astExpr)
	if err != nil {
		return nil, err
	}
	if isBasicType {
		return schemaObject, err
	}
	return p.parseCustomTypeSchemaObject(pkgPath, pkgName, typeName, astExpr)
}

func (p *parser) parseDocAttributeAndValue(comment string) (string, string, bool) {
	attribute := strings.ToLower(strings.Split(comment, " ")[0])
	if len(attribute) == 0 || attribute[0] != '@' {
		return "", "", false
	}
	value := strings.TrimSpace(comment[len(attribute):])
	if len(value) == 0 {
		return "", "", false
	}
	return attribute, value, true
}

func (p *parser) parseFieldDoc(doc *ast.CommentGroup) map[string]string {
	attrs := make(map[string]string)
	if doc == nil {
		return attrs
	}
	for _, comment := range strings.Split(doc.Text(), "\n") {
		if comment == "" {
			continue
		}
		key, value, ok := p.parseDocAttributeAndValue(strings.TrimSpace(comment))
		if !ok {
			continue
		}
		if key == "@desc" {
			if _, ok := attrs[key]; ok {
				attrs[key] = attrs[key] + "\n" + value
				continue
			}
		}
		attrs[key] = value
	}
	return attrs
}
func (p *parser) parseFileName(jsonTag, name string, structSchema, fieldSchema *SchemaObject) (filedName string, isRequired, astFieldsLoop bool) {
	tagValues := strings.Split(jsonTag, ",")
	for _, v := range tagValues {
		if v == "-" {
			structSchema.DisabledFieldNames[name] = struct{}{}
			fieldSchema.Deprecated = true
			astFieldsLoop = true
			return
		} else if v == "required" {
			isRequired = true
		} else if v != "" && v != "required" && v != "omitempty" {
			filedName = v
		}
	}
	return
}
func (p *parser) parseFieldTagAndDoc(astField *ast.Field, structSchema, fieldSchema *SchemaObject) (astFieldsLoop bool, name string) {
	isRequired := false
	name = astField.Names[0].Name
	if astField.Doc == nil {
		if astField.Tag == nil {
			return
		}
		astFieldTag := reflect.StructTag(strings.Trim(astField.Tag.Value, "`"))
		name, isRequired, astFieldsLoop = p.parseFileName(astFieldTag.Get("json"), name, structSchema, fieldSchema)
		if isRequired {
			structSchema.Required = append(structSchema.Required, name)
		}
		if astFieldsLoop {
			return
		}
		return
	}

	doc := p.parseFieldDoc(astField.Doc)

	if goSwagger3 := doc["@go-swagger3"]; goSwagger3 != "" {
		name, isRequired, astFieldsLoop = p.parseFileName(goSwagger3, name, structSchema, fieldSchema)
		if astFieldsLoop {
			return
		}
	}

	if skip := doc["@skip"]; skip == "true" {
		astFieldsLoop = true
		return
	}

	if tag := doc["@json"]; tag == "" && astField.Tag != nil {
		astFieldTag := reflect.StructTag(strings.Trim(astField.Tag.Value, "`"))
		tag = astFieldTag.Get("json")
		name, isRequired, astFieldsLoop = p.parseFileName(tag, name, structSchema, fieldSchema)
		if astFieldsLoop {
			return
		}
	}
	// 解析example
	if example := doc["@example"]; example != "" {
		switch fieldSchema.Type {
		case "boolean":
			fieldSchema.Example, _ = strconv.ParseBool(example)
		case "integer":
			fieldSchema.Example, _ = strconv.Atoi(example)
		case "number":
			fieldSchema.Example, _ = strconv.ParseFloat(example, 64)
		case "array":
			b, err := json.RawMessage(example).MarshalJSON()
			if err != nil {
				fieldSchema.Example = "invalid example"
			} else {
				sliceOfInterface := []interface{}{}
				err := json.Unmarshal(b, &sliceOfInterface)
				if err != nil {
					fieldSchema.Example = "invalid example"
				} else {
					fieldSchema.Example = sliceOfInterface
				}
			}
		case "object":
			b, err := json.RawMessage(example).MarshalJSON()
			if err != nil {
				fieldSchema.Example = "invalid example"
			} else {
				mapOfInterface := map[string]interface{}{}
				err := json.Unmarshal(b, &mapOfInterface)
				if err != nil {
					fieldSchema.Example = "invalid example"
				} else {
					fieldSchema.Example = mapOfInterface
				}
			}
		default:
			fieldSchema.Example = example
		}

		if fieldSchema.Example != nil && len(fieldSchema.Ref) != 0 {
			fieldSchema.Ref = ""
		}
	}

	if overrideExample := doc["@override-example"]; overrideExample != "" {
		fieldSchema.Example = overrideExample

		if fieldSchema.Example != nil && len(fieldSchema.Ref) != 0 {
			fieldSchema.Ref = ""
		}
	}

	if data, ok := doc["@required"]; ok || isRequired || cast.ToBool(data) {
		structSchema.Required = append(structSchema.Required, name)
	}
	// 解析备注
	desc := doc["@desc"]

	if desc != "" {
		fieldSchema.Description = desc
	} else if astField.Comment != nil {
		desc = strings.Split(astField.Comment.Text(), "\n")[0]
		fieldSchema.Description = desc
	}
	// 解析ref
	if ref := doc["@ref"]; ref != "" {
		fieldSchema.Ref = utils.AddSchemaRefLinkPrefix(ref)
		fieldSchema.Type = "" // remove default type in case of reference link
		fieldSchema.Description = ""
	}
	// 解析枚举
	if enumValues := doc["@enum"]; enumValues != "" {
		if fieldSchema.Type == "array" {
			fieldSchema.Items.Enum = parseEnumValues(enumValues)
		} else {
			fieldSchema.Enum = parseEnumValues(enumValues)
		}
	}
	// 解析multipleOf
	if v := doc["@multipleOf"]; v != "" {
		switch fieldSchema.Type {
		case "integer":
			fieldSchema.MultipleOf = cast.ToInt64(v)
		case "number":
			fieldSchema.MultipleOf = cast.ToFloat64(v)
		}
	}
	// 解析maxLength
	if v := doc["@maxLength"]; v != "" {
		switch fieldSchema.Type {
		case "string":
			fieldSchema.MaxLength = cast.ToInt32(v)
		}
	}
	// 解析minLength
	if v := doc["@minLength"]; v != "" {
		switch fieldSchema.Type {
		case "string":
			fieldSchema.MinLength = cast.ToInt32(v)
		}
	}
	// 解析exclusiveMaximum
	if v := doc["@exclusiveMaximum"]; v != "" {
		switch fieldSchema.Type {
		case "integer":
			fieldSchema.ExclusiveMaximum = cast.ToInt64(v)
		case "number":
			fieldSchema.ExclusiveMaximum = cast.ToFloat64(v)
		}
	}
	// 解析exclusiveMinimum
	if v := doc["@exclusiveMaximum"]; v != "" {
		switch fieldSchema.Type {
		case "integer":
			fieldSchema.ExclusiveMinimum = cast.ToInt64(v)
		case "number":
			fieldSchema.ExclusiveMinimum = cast.ToFloat64(v)
		}
	}
	// 解析maximum
	if v := doc["@maximum"]; v != "" {
		switch fieldSchema.Type {
		case "integer":
			fieldSchema.Maximum = cast.ToInt64(v)
		case "number":
			fieldSchema.Maximum = cast.ToFloat64(v)
		}
	}
	// 解析minimum
	if v := doc["@minimum"]; v != "" {
		fieldSchema.Minimum = cast.ToInt64(v)
		switch fieldSchema.Type {
		case "integer":
			fieldSchema.Minimum = cast.ToInt64(v)
		case "number":
			fieldSchema.Minimum = cast.ToFloat64(v)
		}
	}
	// 解析title
	if v := doc["@title"]; v != "" {
		fieldSchema.Title = v
	}
	// 解析deprecated
	if v, ok := doc["@deprecated"]; ok {
		fieldSchema.Deprecated = true
		if v != "" {
			fieldSchema.Deprecated = cast.ToBool(v)
		}
	}
	// 解析format
	if v := doc["@format"]; v != "" {
		switch fieldSchema.Type {
		case "string":
			fieldSchema.Format = v
		}
	}
	// 解析default
	if v := doc["@default"]; v != "" {
		switch fieldSchema.Type {
		case "boolean":
			fieldSchema.Default = cast.ToBool(v)
		case "integer":
			fieldSchema.Default = cast.ToInt64(v)
		case "number":
			fieldSchema.Default = cast.ToFloat64(v)
		default:
			fieldSchema.Default = v
		}
	}
	return
}
