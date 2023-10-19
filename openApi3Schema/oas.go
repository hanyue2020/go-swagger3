package openApi3Schema

import (
	"github.com/iancoleman/orderedmap"
)

const (
	OpenAPIVersion = "3.0.3"

	ContentTypeText = "text/plain"
	ContentTypeJson = "application/json"
	ContentTypeForm = "multipart/form-data"
)

type OpenAPIObject struct {
	Version string         `json:"openapi" yaml:"openapi"` // Required
	Info    InfoObject     `json:"info" yaml:"info"`       // Required
	Servers []ServerObject `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths   PathsObject    `json:"paths" yaml:"paths"` // Required

	Components ComponentsObject      `json:"components,omitempty" yaml:"components,omitempty"` // Required for Authorization header
	Security   []map[string][]string `json:"security,omitempty" yaml:"security,omitempty"`

	// Tags
	// ExternalDocs
}

type ServerObject struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Variables
}

type InfoObject struct {
	Title          string         `json:"title" yaml:"title"`
	Description    string         `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string         `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *ContactObject `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *LicenseObject `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string         `json:"version" yaml:"version"`
}

type ContactObject struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

type LicenseObject struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

type PathsObject map[string]*PathItemObject

type PathItemObject struct {
	Ref         string           `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string           `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string           `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *OperationObject `json:"get,omitempty" yaml:"get,omitempty"`
	Post        *OperationObject `json:"post,omitempty" yaml:"post,omitempty"`
	Patch       *OperationObject `json:"patch,omitempty" yaml:"patch,omitempty"`
	Put         *OperationObject `json:"put,omitempty" yaml:"put,omitempty"`
	Delete      *OperationObject `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options     *OperationObject `json:"options,omitempty" yaml:"options,omitempty"`
	Head        *OperationObject `json:"head,omitempty" yaml:"head,omitempty"`
	Trace       *OperationObject `json:"trace,omitempty" yaml:"trace,omitempty"`

	// Servers
	// Parameters
}

type OperationObject struct {
	Responses ResponsesObject `json:"responses" yaml:"responses"` // Required

	Tags        []string           `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string             `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Parameters  []ParameterObject  `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBodyObject `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`

	// Tags
	// ExternalDocs
	// OperationID
	// Callbacks
	// Deprecated
	// Security
	// Servers
}

type ParameterObject struct {
	Name        string        `json:"name,omitempty" yaml:"name,omitempty"` // Required
	In          string        `json:"in,omitempty" yaml:"in,omitempty"`     // Required. Possible values are "query", "header", "path" or "cookie"
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool          `json:"required,omitempty" yaml:"required,omitempty"`
	Example     interface{}   `json:"example,omitempty" yaml:"example,omitempty"`
	Schema      *SchemaObject `json:"schema,omitempty" yaml:"schema,omitempty"`

	// Ref is used when ParameterOjbect is as a ReferenceObject
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	// Deprecated
	// AllowEmptyValue
	// Style
	// Explode
	// AllowReserved
	// Examples
	// Content
}

type RequestBodyObject struct {
	Content map[string]*MediaTypeObject `json:"content" yaml:"content"` // Required

	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool   `json:"required,omitempty" yaml:"required,omitempty"`

	// Ref is used when RequestBodyObject is as a ReferenceObject
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

type MediaTypeObject struct {
	Schema SchemaObject `json:"schema,omitempty" yaml:"schema,omitempty"`
	// Example string       `json:"example,omitempty" yaml:"example,omitempty"`

	// Examples
	// Encoding
}

type SchemaObject struct {
	ID                 string                 `json:"-" yaml:"-"` // For go-swagger3
	PkgName            string                 `json:"-" yaml:"-"` // For go-swagger3
	FieldName          string                 `json:"-" yaml:"-"` // For go-swagger3
	DisabledFieldNames map[string]struct{}    `json:"-" yaml:"-"` // For go-swagger3
	Type               string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Format             string                 `json:"format,omitempty" yaml:"format,omitempty"`
	Required           []string               `json:"required,omitempty" yaml:"required,omitempty"`
	Properties         *orderedmap.OrderedMap `json:"properties,omitempty" yaml:"properties,omitempty"`
	Description        string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Items              *SchemaObject          `json:"items,omitempty" yaml:"items,omitempty"` // use ptr to prevent recursive error
	Example            interface{}            `json:"example,omitempty" yaml:"example,omitempty"`
	Deprecated         bool                   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Ref                string                 `json:"$ref,omitempty" yaml:"$ref,omitempty"` // Ref is used when SchemaObject is as a ReferenceObject
	Enum               interface{}            `json:"enum,omitempty" yaml:"enum,omitempty"`

	// Title
	// MultipleOf
	// Maximum
	// ExclusiveMaximum
	// Minimum
	// ExclusiveMinimum
	// MaxLength
	// MinLength
	// Pattern
	// MaxItems
	// MinItems
	// UniqueItems
	// MaxProperties
	// MinProperties
	// AllOf
	// OneOf
	// AnyOf
	// Not
	// AdditionalProperties
	// Description
	// Default
	// Nullable
	// ReadOnly
	// WriteOnly
	// XML
	// ExternalDocs
}

type ResponsesObject map[string]*ResponseObject // [status]ResponseObject

type ResponseObject struct {
	Description string `json:"description" yaml:"description"` // Required

	Headers map[string]*HeaderObject    `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content map[string]*MediaTypeObject `json:"content,omitempty" yaml:"content,omitempty"`

	// Ref is for ReferenceObject
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	// Links
}

type HeaderObject struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`

	// Ref is used when HeaderObject is as a ReferenceObject
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

type ComponentsObject struct {
	Schemas         map[string]*SchemaObject         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	SecuritySchemes map[string]*SecuritySchemeObject `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Parameters      map[string]*ParameterObject      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	// Responses
	// Examples
	// RequestBodies
	// Headers
	// Links
	// Callbacks
}

type SecuritySchemeObject struct {
	// Generic fields
	Type        string `json:"type" yaml:"type"` // Required
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// http
	Scheme string `json:"scheme,omitempty" yaml:"scheme,omitempty"`

	// apiKey
	In   string `json:"in,omitempty" yaml:"in,omitempty"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// OpenID
	OpenIdConnectUrl string `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`

	// OAuth2
	OAuthFlows *SecuritySchemeOauthObject `json:"flows,omitempty" yaml:"flows,omitempty"`

	// BearerFormat
}

type SecuritySchemeOauthObject struct {
	Implicit              *SecuritySchemeOauthFlowObject `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	AuthorizationCode     *SecuritySchemeOauthFlowObject `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
	ResourceOwnerPassword *SecuritySchemeOauthFlowObject `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials     *SecuritySchemeOauthFlowObject `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
}

func (s *SecuritySchemeOauthObject) ApplyScopes(scopes map[string]string) {
	if s.Implicit != nil {
		s.Implicit.Scopes = scopes
	}

	if s.AuthorizationCode != nil {
		s.AuthorizationCode.Scopes = scopes
	}

	if s.ResourceOwnerPassword != nil {
		s.ResourceOwnerPassword.Scopes = scopes
	}

	if s.ClientCredentials != nil {
		s.ClientCredentials.Scopes = scopes
	}
}

type SecuritySchemeOauthFlowObject struct {
	AuthorizationUrl string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenUrl         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes" yaml:"scopes"`
}
