package operations

import (
	"strings"

	oas "github.com/hanyue2020/evadoc/internal/openapi"
	"github.com/hanyue2020/evadoc/internal/parser/utils"
)

func (p *parser) parseResourceAndTag(comment string, attribute string, operation *oas.OperationObject) {
	resource := strings.TrimSpace(comment[len(attribute):])
	if resource == "" {
		resource = "others"
	}
	if !utils.IsInStringList(operation.Tags, resource) {
		operation.Tags = append(operation.Tags, resource)
	}
}
