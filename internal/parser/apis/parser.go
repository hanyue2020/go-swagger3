package apis

import (
	oas "github.com/hanyue2020/evadoc/internal/openapi"
	"github.com/hanyue2020/evadoc/internal/parser/model"
	"github.com/hanyue2020/evadoc/internal/parser/operations"
	"github.com/hanyue2020/evadoc/internal/parser/schema"
	log "github.com/sirupsen/logrus"
)

type Parser interface {
	Parse() error
}

type parser struct {
	OpenAPI *oas.OpenAPIObject

	model.Utils
	schemaParser    schema.Parser
	operationParser operations.Parser
}

func NewParser(utils model.Utils, api *oas.OpenAPIObject, schemaParser schema.Parser) Parser {
	return &parser{
		Utils:           utils,
		OpenAPI:         api,
		schemaParser:    schemaParser,
		operationParser: operations.NewParser(utils, api, schemaParser),
	}
}

// Parse parse APIs info
func (p *parser) Parse() error {
	log.Info("Parsing APIs ...")
	err := p.parseImportStatements()
	if err != nil {
		return err
	}

	err = p.parseTypeSpecs()
	if err != nil {
		return err
	}

	err = p.parseParameters()
	if err != nil {
		return err
	}

	return p.parsePaths()
}
