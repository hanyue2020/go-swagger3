package operations

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	oas "github.com/hanyue2020/evadoc/openApi3Schema"
)

func (p *parser) parseRouteComment(operation *oas.OperationObject, comment string) error {
	sourceString := strings.TrimSpace(comment)
	if strings.HasPrefix(sourceString, "@") {
		if idx := strings.IndexAny(sourceString, " \t"); idx > 0 {
			sourceString = strings.TrimSpace(sourceString[idx:])
		} else {
			return fmt.Errorf("Can not parse router comment \"%s\", skipped", comment)
		}
	}

	// /path [method]
	re := regexp.MustCompile(`^([^\[]+?)\s*\[([^\]]+)\]\s*$`)
	matches := re.FindStringSubmatch(sourceString)
	if len(matches) != 3 {
		return fmt.Errorf("Can not parse router comment \"%s\", skipped", comment)
	}

	path := strings.TrimSpace(matches[1])
	if path == "" {
		return fmt.Errorf("Can not parse router comment \"%s\", skipped", comment)
	}

	_, ok := p.OpenAPI.Paths[path]
	if !ok {
		p.OpenAPI.Paths[path] = &oas.PathItemObject{}
	}

	methods := strings.Split(matches[2], "|")
	for _, v := range methods {
		switch strings.ToUpper(strings.TrimSpace(v)) {
		case http.MethodGet:
			p.OpenAPI.Paths[path].Get = operation
		case http.MethodPost:
			p.OpenAPI.Paths[path].Post = operation
		case http.MethodPatch:
			p.OpenAPI.Paths[path].Patch = operation
		case http.MethodPut:
			p.OpenAPI.Paths[path].Put = operation
		case http.MethodDelete:
			p.OpenAPI.Paths[path].Delete = operation
		case http.MethodOptions:
			p.OpenAPI.Paths[path].Options = operation
		case http.MethodHead:
			p.OpenAPI.Paths[path].Head = operation
		case http.MethodTrace:
			p.OpenAPI.Paths[path].Trace = operation
		}
	}
	return nil
}
