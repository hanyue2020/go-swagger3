# go-swagger3

`go-swagger3` scans Go source code and its documentation comments to generate an
[OpenAPI 3.0.3](https://swagger.io/specification/) specification in JSON or
YAML. It reads a Go module, API metadata, handler comments, and Go types; it
does not need to run the service.

## Requirements

- Go 1.24 or later
- A Go module (`go.mod`) for the service being documented

## Install

```sh
go install github.com/hanyue2020/go-swagger3@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH` if the `go-swagger3` command
is not found.

## Generate a specification

Run the command from the service's module directory:

```sh
go-swagger3 --module-path . --output openapi.json
```

The generator searches the module for a `main` package to read API-level
metadata. If there are multiple possible service entry points, select the
metadata file explicitly:

```sh
go-swagger3 \
  --module-path . \
  --main-file-path ./cmd/api/main.go \
  --handler-path ./internal/handlers \
  --output ./docs/openapi.json
```

`--main-file-path` is an alias for `--basic-info-path`. If neither is supplied,
the generator searches for a main file and reports an error when it cannot
choose an unambiguous one.

### Select a service by domain

For repositories with a `domain/<name>` layout, `--domain` locates the module
containing that domain and uses `domain/<name>` as the handler path and
`domain/<name>/boot/default.go` as the metadata file:

```sh
# Run from the go-swagger3 repository root.
go run main.go --domain merchant --output xxx.yaml --generate-yaml
```

For this repository, the command finds `example/rent/go.mod` and scans
`example/rent/domain/merchant` for handlers. The search starts in the current
working directory and its descendants. A positional `.` is not a module-path
argument; to select a module explicitly, use `--module-path`:

```sh
go-swagger3 \
  --module-path ./example/rent \
  --domain merchant \
  --output xxx.yaml \
  --generate-yaml
```

Explicit `--handler-path`, `--basic-info-path`, or `--main-file-path` values
take precedence over the paths inferred from `--domain`.

The `.yaml` extension alone does not select YAML output; include
`--generate-yaml`.

### Command-line options

| Option | Default | Description |
| --- | --- | --- |
| `--module-path` | Current directory | Root of the Go module to scan. |
| `--domain` | — | Infer the module and default paths for `domain/<name>`. |
| `--basic-info-path` | Auto-detected | Go file containing API metadata comments. |
| `--main-file-path` | — | Alias for `--basic-info-path`. |
| `--handler-path` | Entire module | Restrict endpoint comment scanning to this directory. |
| `--output` | `oas.json` | Output file path. |
| `--generate-yaml` | `false` | Write YAML instead of JSON. If the output ends in `.json`, its extension is changed to `.yaml`. |
| `--debug` | `false` | Enable debug logging. |
| `--strict` | `false` | Return errors encountered while parsing Go packages instead of skipping those packages. |

`--schema-without-pkg` is still accepted by the CLI for compatibility, but is
not currently connected to schema generation.

## Add API documentation

The generator reads `@` directives from Go comments. Directive names are
case-insensitive. API-level metadata is read from the selected metadata file;
operation directives are read from documented functions under the handler
path.

### API metadata

At minimum, provide a title and version:

```go
package main

// @Title Store API
// @Version 1.0.0
// @Description API for managing products and orders.
// @Server https://api.example.com Production
// @ContactName API Support
// @ContactEmail support@example.com
// @LicenseName MIT
// @LicenseURL https://opensource.org/licenses/MIT
```

Supported metadata directives include:

| Directive | Meaning |
| --- | --- |
| `@Title` | Required API title. |
| `@Version` | Required API version. |
| `@Description` / `@Desc` | API description. Repeated lines are joined. |
| `@Info.TermsOfServiceUrl` | Terms of service URL. |
| `@ContactName`, `@ContactEmail` | Contact information. |
| `@LicenseName`, `@LicenseURL` | License information. |
| `@Server <url> [description]` | Add a server. Defaults to `/` when omitted. |
| `@Tags <name>[,<description>[,<external-doc-description>,<external-doc-url>]]` | Add an API tag. |

When no API description is supplied in code, the writer uses `README.md` from
the current working directory as the OpenAPI description. Add `@Description`
if you want the specification to contain a concise API description instead.

### Operations

Document a handler function with operation directives and a route. The
following example uses the Fiber handler style and generic response types:

```go
// @Summary        用户信息
// @Title          用户信息
// @Description    用户信息
// @Tag            用户管理
// @Param          request body request.AccountInfo true "用户信息"
// @Success        200 object response.ResponseInfo[response.Account]{}
// @Failure        200 object response.AccountInfo{}
// @Router         /api/account/info [post]
// @Middlewares    auth
func (m *Logic) AccountInfo(c *fiber.Ctx, ctx context.Context, req *request.AccountInfo) error {
    // ...
}
```

| Directive | Syntax and purpose |
| --- | --- |
| `@Title` / `@Summary` | Operation summary. |
| `@Description` / `@Desc` | Operation description; repeated lines are joined. |
| `@OperationId` | Set the OpenAPI operation ID. |
| `@Deprecated [true\|false]` | Mark the operation as deprecated. |
| `@Param <name> <in> <go-type> <required> "<description>"` | Add a parameter. `in` supports `path`, `query`, `header`, `cookie`, `body`, `form`, and `file`; required accepts `true`, `false`, `required`, or `optional`. Path parameters are always required. |
| `@Header <GoType>` | Add header parameters described by a Go struct. |
| `@Success <status> <kind> <GoType> ["<description>"]` | Describe a successful response. |
| `@Failure <status> <kind> <GoType> ["<description>"]` | Describe an error response. |
| `@Response <status> <kind> <GoType> ["<description>"]` | Describe a response. |
| `@Resource <name>` / `@Tag <name>` | Assign an operation tag. |
| `@Route <path> [<method>[|<method>...]]` | Bind the operation to one or more HTTP methods. `@Router` is an alias. |

For example, `@Route /users [get|post]` assigns the operation to both methods.
Supported methods are GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD, and TRACE.

Response kinds include `object`, `array`, `string`, `integer`, `boolean`, and
`file`. Object and array responses refer to Go types. Generic Go types are
supported, for example
`response.ResponseInfo[response.Account]{}`. The response description is
optional; when provided, enclose it in double quotes. `@Middlewares` is not
currently interpreted by the generator; use `@Security` and
`@SecurityScheme` to describe API authentication in the OpenAPI document.

### Request headers

Declare reusable header parameters on a struct, then reference that struct
from an operation:

```go
// @HeaderParameters RequestHeaders
type RequestHeaders struct {
    // @required true
    Authorization string `json:"Authorization"`
}

// @Header RequestHeaders
// @Route /profile [get]
func GetProfile() {}
```

### Struct fields and schemas

JSON struct tags determine property names. The generator also recognizes these
field comments:

| Directive | Purpose |
| --- | --- |
| `@desc <text>` | Field description. Ordinary Go field comments are also used as descriptions. |
| `@example <value>` | Field example. Arrays and objects should use valid JSON. |
| `@override-example <value>` | Set the example as a string, overriding a generated example. |
| `@required true` | Mark the field as required. A `required` option in the JSON tag is also recognized. |
| `@skip true` | Omit the field from the schema. |
| `@ref <SchemaName>` | Reference another schema. |
| `@enum <value>[,<value>...]` | Set allowed values; integer ranges such as `1~5` are also accepted. |
| `@default <value>` | Set a default value. |
| `@format <format>` | Set a string format. |
| `@title <text>` | Set the schema title. |
| `@deprecated [true\|false]` | Mark the field as deprecated. |
| `@minimum`, `@maximum`, `@exclusiveMaximum` | Set numeric bounds. |
| `@minLength`, `@maxLength` | Set string length constraints. |
| `@multipleOf` | Set a numeric constraint. |

Example:

```go
type CreateUserRequest struct {
    // @desc User's display name.
    // @example Ada
    // @required true
    Name string `json:"name"`

    // @enum active,disabled
    Status string `json:"status,omitempty"`
}
```

### Security

Define a security scheme in the metadata file and apply it to the API with
`@Security`:

```go
// @SecurityScheme BearerAuth http bearer
// @Security BearerAuth
```

Supported scheme directives include:

| Scheme | Syntax |
| --- | --- |
| HTTP | `@SecurityScheme <name> http <scheme> [description]` |
| API key | `@SecurityScheme <name> apiKey <header\|query\|cookie> <key-name> [description]` |
| OpenID Connect | `@SecurityScheme <name> openIdConnect <discovery-url> [description]` |
| OAuth2 authorization code | `@SecurityScheme <name> oauth2AuthCode <authorization-url> <token-url>` |
| OAuth2 implicit | `@SecurityScheme <name> oauth2Implicit <authorization-url>` |
| OAuth2 password | `@SecurityScheme <name> oauth2ResourceOwnerCredentials <token-url>` |
| OAuth2 client credentials | `@SecurityScheme <name> oauth2ClientCredentials <token-url>` |

OAuth2 scopes can be declared with:

```go
// @SecurityScope BearerAuth read:users Read users
// @Security BearerAuth read:users
```

Security is currently configured at the API level.

## Limitations

- The target service must use Go modules.
- Inline anonymous struct fields are not supported as schema types; define a
  named struct instead.
- The generator parses source comments and Go type declarations; it does not
  infer routes from a web framework or execute the application.

## Development

Run the test suite with:

```sh
go test ./...
```

## Acknowledgements

- [yvasiyarov/swagger](https://github.com/yvasiyarov/swagger)
- [uudashr/go-module](https://github.com/uudashr/go-module)
- [mikunalpha/goas](https://github.com/mikunalpha/goas)
