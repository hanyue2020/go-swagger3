package schema

import (
	"fmt"
	"go/ast"
	"strings"

	. "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/utils"
	"github.com/iancoleman/orderedmap"
	"github.com/spf13/cast"
)

func (p *parser) parseCustomTypeSchemaObject(pkgPath string, pkgName string, typeName string, astExpr ast.Expr) (*SchemaObject, error) {
	var typeSpec *ast.TypeSpec
	var exist bool
	var schemaObject SchemaObject

	// handler other type
	typeNameParts := strings.Split(typeName, ".")
	if len(typeNameParts) == 1 {
		typeSpec, exist = p.getTypeSpec(pkgName, typeName)
		if !exist {
			panic(fmt.Sprintf("Can not find definition of %v ast.TypeSpec. Current package %v", typeName, pkgName))
		}
		schemaObject.PkgName = pkgName
		schemaObject.ID = utils.GenSchemaObjectID(pkgName, typeName)
		p.KnownIDSchema[schemaObject.ID] = &schemaObject
	} else {
		guessPkgName := strings.Join(typeNameParts[:len(typeNameParts)-1], "/")
		guessPkgPath := ""
		for i := range p.KnownPkgs {
			if guessPkgName == p.KnownPkgs[i].Name {
				guessPkgPath = p.KnownPkgs[i].Path
				break
			}
		}
		guessTypeName := typeNameParts[len(typeNameParts)-1]
		typeSpec, exist = p.getTypeSpec(guessPkgName, guessTypeName)
		if !exist {
			found := false
			aliases := p.PkgNameImportedPkgAlias[pkgName][guessPkgName]
			for k := range p.PkgNameImportedPkgAlias[pkgName] {
				if k == guessPkgName && len(aliases) != 0 {
					found = true
					break
				}
			}
			if !found {
				p.Debugf("unknown guess %s ast.TypeSpec in package %s", guessTypeName, guessPkgName)
				return &schemaObject, nil
			}
			for index, currentAliasName := range aliases {
				guessPkgName = currentAliasName
				guessPkgPath = ""
				for i := range p.KnownPkgs {
					if guessPkgName == p.KnownPkgs[i].Name {
						guessPkgPath = p.KnownPkgs[i].Path
						break
					}
				}
				typeSpec, exist = p.getTypeSpec(guessPkgName, guessTypeName)
				if exist {
					break
				}
				if !exist && index == len(aliases)-1 {
					p.Debugf("can not find definition of guess %s ast.TypeSpec in package %s", guessTypeName, guessPkgName)
					return &schemaObject, nil
				}
			}
			schemaObject.PkgName = guessPkgName
			schemaObject.ID = utils.GenSchemaObjectID(guessPkgName, guessTypeName)
			p.KnownIDSchema[schemaObject.ID] = &schemaObject
		}
		pkgPath, pkgName = guessPkgPath, guessPkgName
	}
	switch typeSpec.Type.(type) {
	case *ast.Ident:
		if astIdent := typeSpec.Type.(*ast.Ident); astIdent != nil {
			schemaObject.Type = astIdent.Name
		}
	case *ast.StructType:
		if astStructType := typeSpec.Type.(*ast.StructType); astStructType != nil {
			schemaObject.Type = "object"
			if astStructType.Fields != nil {
				p.parseSchemaPropertiesFromStructFields(pkgPath, pkgName, &schemaObject, astStructType.Fields.List)
			}
			typeNameParts := strings.Split(typeName, ".")
			if len(typeNameParts) > 1 {
				typeName = typeNameParts[len(typeNameParts)-1]
			}
			if !utils.IsBasicGoType(typeName) {
				_, err := p.RegisterType(pkgPath, pkgName, typeName, astExpr)
				if err != nil {
					p.Debugf("ParseSchemaObject parse array items err: %s", err.Error())
				}
			}
		}
	case *ast.ArrayType:
		if astArrayType := typeSpec.Type.(*ast.ArrayType); astArrayType != nil {
			schemaObject.Type = "array"
			schemaObject.Items = &SchemaObject{}
			typeAsString := p.getTypeAsString(astArrayType.Elt)
			typeAsString = strings.TrimLeft(typeAsString, "*")
			if !utils.IsBasicGoType(typeAsString) {
				schemaItemsSchemeaObjectID, err := p.RegisterType(pkgPath, pkgName, typeAsString, astExpr)
				if err != nil {
					p.Debugf("ParseSchemaObject parse array items err: %s", err.Error())
				} else {
					schemaObject.Items.Ref = utils.AddSchemaRefLinkPrefix(schemaItemsSchemeaObjectID)
				}
			} else if utils.IsGoTypeOASType(typeAsString) {
				schemaObject.Items.Type = utils.GoTypesOASTypes[typeAsString]
			}
		}
	case *ast.MapType:
		if astMapType := typeSpec.Type.(*ast.MapType); astMapType != nil {
			schemaObject.Type = "object"
			schemaObject.Properties = orderedmap.New()
			propertySchema := &SchemaObject{}
			schemaObject.Properties.Set("key", propertySchema)
			typeAsString := p.getTypeAsString(astMapType.Value)
			typeAsString = strings.TrimLeft(typeAsString, "*")
			if !utils.IsBasicGoType(typeAsString) {
				schemaItemsSchemeaObjectID, err := p.RegisterType(pkgPath, pkgName, typeAsString, astExpr)
				if err != nil {
					p.Debugf("ParseSchemaObject parse array items err: %s", err.Error())
				} else {
					propertySchema.Ref = utils.AddSchemaRefLinkPrefix(schemaItemsSchemeaObjectID)
				}
			} else if utils.IsGoTypeOASType(typeAsString) {
				propertySchema.Type = utils.GoTypesOASTypes[typeAsString]
			}
		}
	}
	return &schemaObject, nil
}

func (p *parser) getTypeSpec(pkgName, typeName string) (*ast.TypeSpec, bool) {
	pkgTypeSpecs, exist := p.TypeSpecs[pkgName]
	if !exist {
		return nil, false
	}
	astTypeSpec, exist := pkgTypeSpecs[typeName]
	if !exist {
		return nil, false
	}
	return astTypeSpec, true
}
func (p *parser) parseStructField(pkgPath, pkgName string, structSchema *SchemaObject, astExpr ast.Expr) (fieldSchema *SchemaObject, err error) {
	fieldSchema = &SchemaObject{}
	typeAsString := p.getTypeAsString(astExpr)
	typeAsString = strings.TrimLeft(typeAsString, "*")

	switch typeAsString {
	case "time.Time", "decimal.Decimal", "struct{}", "interface{}":
		fieldSchema, err = p.ParseSchemaObject(pkgPath, pkgName, typeAsString, astExpr)
		if err != nil {
			p.Debug(err)
		}
		return
	}
	if strings.HasPrefix(typeAsString, "[]") {
		fieldSchema, err = p.ParseSchemaObject(pkgPath, pkgName, typeAsString, astExpr)
		if err != nil {
			p.Debug(err)
			return
		}
	} else if strings.HasPrefix(typeAsString, "map[]") {
		fieldSchema, err = p.ParseSchemaObject(pkgPath, pkgName, typeAsString, astExpr)
		if err != nil {
			p.Debug(err)
			return
		}
	} else if !utils.IsBasicGoType(typeAsString) {
		fieldSchemaSchemeaObjectID, err := p.RegisterType(pkgPath, pkgName, typeAsString, astExpr)
		if err != nil {
			p.Debug("parseSchemaPropertiesFromStructFields err:", err)
		} else {
			fieldSchema.ID = fieldSchemaSchemeaObjectID
			schema, ok := p.KnownIDSchema[fieldSchemaSchemeaObjectID]
			if ok {
				fieldSchema.Type = schema.Type
				if schema.Items != nil {
					fieldSchema.Items = schema.Items
				}
				fieldSchema.Ref = utils.AddSchemaRefLinkPrefix(fieldSchemaSchemeaObjectID)
			}
		}
	} else if utils.IsGoTypeOASType(typeAsString) {
		fieldSchema.Type = utils.GoTypesOASTypes[typeAsString]
	}
	return
}

func (p *parser) parseSchemaPropertiesFromStructFields(pkgPath, pkgName string, structSchema *SchemaObject, astFields []*ast.Field) {
	if astFields == nil {
		return
	}
	structSchema.Properties = orderedmap.New()
	if structSchema.DisabledFieldNames == nil {
		structSchema.DisabledFieldNames = map[string]struct{}{}
	}
astFieldsLoop:
	for _, astField := range astFields {
		if len(astField.Names) == 0 {
			continue
		}
		fieldSchema, err := p.parseStructField(pkgPath, pkgName, structSchema, astField.Type)
		if err != nil {
			return
		}

		name := astField.Names[0].Name
		fieldSchema.FieldName = name
		continueLoop := false
		_, disabled := structSchema.DisabledFieldNames[name]
		if disabled {
			continue
		}

		continueLoop, name = p.parseFieldTagAndDoc(astField, structSchema, fieldSchema)

		if continueLoop {
			continue astFieldsLoop
		}
		if fieldSchema.Description == "" && fieldSchema.Ref == "" {
			if astField.Comment != nil {
				fieldSchema.Description = strings.TrimSpace(strings.Trim(astField.Comment.List[0].Text, "//"))
			}
		}
		if fieldSchema.Ref != "" {
			fieldSchema.Description = ""
			fieldSchema.Type = ""
		}
		structSchema.Properties.Set(name, fieldSchema)
	}
	// embedded type
	for _, astField := range astFields {
		if len(astField.Names) > 0 {
			continue
		}
		fieldSchema, err := p.parseStructField(pkgPath, pkgName, structSchema, astField.Type)
		if err != nil {
			return
		}
		// 需要将 内嵌结构的 required 上移
		structSchema.Required = append(structSchema.Required, fieldSchema.Required...)

		if fieldSchema.Properties != nil {
			for _, propertyName := range fieldSchema.Properties.Keys() {
				_, exist := structSchema.Properties.Get(propertyName)
				if exist {
					continue
				}
				propertySchema, _ := fieldSchema.Properties.Get(propertyName)
				structSchema.Properties.Set(propertyName, propertySchema)
			}
		} else if len(fieldSchema.Ref) != 0 && len(fieldSchema.ID) != 0 {
			refSchema, ok := p.KnownIDSchema[fieldSchema.ID]
			if ok {
				structSchema.Required = append(structSchema.Required, refSchema.Required...)
				for _, propertyName := range refSchema.Properties.Keys() {
					refPropertySchema, _ := refSchema.Properties.Get(propertyName)
					_, disabled := structSchema.DisabledFieldNames[refPropertySchema.(*SchemaObject).FieldName]
					if disabled {
						continue
					}
					_, exist := structSchema.Properties.Get(propertyName)
					if exist {
						continue
					}
					structSchema.Properties.Set(propertyName, refPropertySchema)
				}
			}
		}
		continue
	}
}

func (p *parser) getTypeAsString(fieldType ast.Expr) string {
	switch fieldType.(type) {
	case *ast.ArrayType:
		return fmt.Sprintf("[]%v", p.getTypeAsString(fieldType.(*ast.ArrayType).Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[]%v", p.getTypeAsString(fieldType.(*ast.MapType).Value))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StarExpr:
		return fmt.Sprintf("%v", p.getTypeAsString(fieldType.(*ast.StarExpr).X))
	case *ast.StructType:
		return "struct{}"
	case *ast.SelectorExpr:
		packageNameIdent, _ := fieldType.(*ast.SelectorExpr).X.(*ast.Ident)
		if packageNameIdent != nil && packageNameIdent.Obj != nil && packageNameIdent.Obj.Decl != nil {
			a, ok := packageNameIdent.Obj.Decl.(DECL)
			if ok {
				fmt.Println(a)
			}
		}
		return packageNameIdent.Name + "." + fieldType.(*ast.SelectorExpr).Sel.Name
	default:
		return fmt.Sprint(fieldType)
	}
}

func parseEnumValues(enumString string) interface{} {
	var result []interface{}
	seg := strings.Split(enumString, "~")

	// 对于区间范围 1~5 表示 enum[1,2,3,4,5]
	if len(seg) == 2 {
		for i := cast.ToInt(seg[0]); i <= cast.ToInt(seg[1]); i++ {
			result = append(result, i)
		}
		return result
	}
	for _, currentEnumValue := range strings.Split(enumString, EnumValueSeparator) {
		result = append(result, currentEnumValue)
	}
	return result
}

type DECL struct {
	Type struct {
		Name string
	}
}

const EnumValueSeparator = ","
