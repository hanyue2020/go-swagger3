package operations

import (
	"go/ast"
	"strings"

	. "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/model"
	"github.com/hanyue2020/go-swagger3/parser/schema"
	"github.com/spf13/cast"
)

type Parser interface {
	Parse(pkgPath, pkgName string, astComments []*ast.Comment) error
}

type parser struct {
	OpenAPI *OpenAPIObject

	model.Utils
	schema.Parser
}

func NewParser(utils model.Utils, api *OpenAPIObject, schemaParser schema.Parser) Parser {
	return &parser{
		Utils:   utils,
		OpenAPI: api,
		Parser:  schemaParser,
	}
}
func (p *parser) Parse(pkgPath, pkgName string, astComments []*ast.Comment) error {
	operation := &OperationObject{Responses: map[string]*ResponseObject{}}
	if !strings.HasPrefix(pkgPath, p.ModulePath) || (p.HandlerPath != "" && !strings.HasPrefix(pkgPath, p.HandlerPath)) {
		return nil
	}

	for _, astComment := range astComments {
		comment := strings.TrimSpace(strings.TrimLeft(astComment.Text, "/"))
		if len(comment) == 0 {
			continue
		}
		if err := p.parseOperationFromComment(pkgPath, pkgName, comment, operation); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) parseOperationFromComment(pkgPath string, pkgName string, comment string, operation *OperationObject) error {
	attribute := strings.Fields(comment)[0]
	switch strings.ToLower(attribute) {
	case "@operationid":
		operation.OperationId = strings.TrimSpace(comment[len(attribute):])
	case "@deprecated":
		operation.Deprecated = cast.ToBool(strings.TrimSpace(comment[len(attribute):]))
	case "@title", "@summary":
		operation.Summary = strings.TrimSpace(comment[len(attribute):])
	case "@description", "@desc":
		if operation.Description == "" {
			operation.Description = strings.Join([]string{operation.Description, strings.TrimSpace(comment[len(attribute):])}, " ")
		} else {
			operation.Description += "\n" + strings.TrimSpace(comment[len(attribute):])
		}
	case "@param":
		return p.parseParamComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
	case "@header":
		return p.parseHeaders(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
	case "@success", "@failure", "@response":
		return p.parseResponseComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
	case "@response.desc":
		return p.parseResponseDesc(operation, strings.TrimSpace(comment[len(attribute):]))
	case "@resource", "@tag":
		p.parseResourceAndTag(comment, attribute, operation)
	case "@route", "@router":
		return p.parseRouteComment(operation, comment)
	}
	return nil
}
