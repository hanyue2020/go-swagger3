package schema

import (
	"go/ast"
	"reflect"
	"strings"
)

type fieldOptions struct {
	name         string
	disabledName string
	required     bool
	skip         bool
	disabled     bool
	attributes   map[string]string
}

func parseFieldOptions(field *ast.Field) fieldOptions {
	options := fieldOptions{attributes: make(map[string]string)}
	if field == nil || len(field.Names) == 0 {
		return options
	}
	options.name = field.Names[0].Name

	if field.Doc == nil {
		if field.Tag != nil {
			options = applyFieldJSONTag(options, reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("json"))
		}
		return options
	}

	options.attributes = parseFieldDoc(field.Doc)
	if goSwagger3 := options.attributes["@go-swagger3"]; goSwagger3 != "" {
		options = applyFieldJSONTag(options, goSwagger3)
		if options.disabled {
			return options
		}
	}

	if options.attributes["@skip"] == "true" {
		options.skip = true
		return options
	}

	if options.attributes["@json"] == "" && field.Tag != nil {
		options = applyFieldJSONTag(options, reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("json"))
	}
	return options
}

func applyFieldJSONTag(options fieldOptions, jsonTag string) fieldOptions {
	disabledName := options.name
	options.name = ""
	for _, value := range strings.Split(jsonTag, ",") {
		switch {
		case value == "-":
			options.disabled = true
			options.disabledName = disabledName
			return options
		case value == "required":
			options.required = true
		case value != "" && value != "omitempty":
			options.name = value
		}
	}
	return options
}

func parseFieldDoc(doc *ast.CommentGroup) map[string]string {
	attributes := make(map[string]string)
	if doc == nil {
		return attributes
	}
	for _, comment := range strings.Split(doc.Text(), "\n") {
		if comment == "" {
			continue
		}
		key, value, ok := parseDocAttributeAndValue(strings.TrimSpace(comment))
		if !ok {
			continue
		}
		if key == "@desc" {
			if _, exists := attributes[key]; exists {
				attributes[key] += "\n" + value
				continue
			}
		}
		attributes[key] = value
	}
	return attributes
}

func parseDocAttributeAndValue(comment string) (string, string, bool) {
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
