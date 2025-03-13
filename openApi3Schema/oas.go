package openApi3Schema

import (
	"github.com/iancoleman/orderedmap"
)

const (
	OpenAPIVersion = "3.0.3"

	ContentTypeText = "text/plain"
	ContentTypeXML  = "application/xml"
	ContentTypePDF  = "application/pdf"
	ContentTypeFile = "application/octet-stream"
	// Text
	ContentTypePlain = "text/plain"
	ContentTypeHtml  = "text/html"
	ContentTypeCss   = "text/css"
	ContentTypeCsv   = "text/csv"
	ContentTypeJs    = "text/javascript"
	ContentTypeXml   = "text/xml"
	ContentTypeMd    = "text/markdown"
	// Image
	ContentTypeJpeg = "image/jpeg"
	ContentTypePng  = "image/png"
	ContentTypeGif  = "image/gif"
	ContentTypeSvg  = "image/svg+xml"
	ContentTypeWebp = "image/webp"
	ContentTypeBmp  = "image/bmp"
	ContentTypeTiff = "image/tiff"

	// Audio
	ContentTypeMp3  = "audio/mpeg"
	ContentTypeOgg  = "audio/ogg"
	ContentTypeWav  = "audio/wav"
	ContentTypeWebm = "audio/webm"
	ContentTypeAac  = "audio/aac"
	ContentTypeMidi = "audio/midi"

	// Video
	ContentTypeMp4   = "video/mp4"
	ContentTypeOggV  = "video/ogg"
	ContentTypeWebmV = "video/webm"
	ContentTypeAvi   = "video/x-msvideo"
	ContentTypeMov   = "video/quicktime"
	ContentTypeMkv   = "video/x-matroska"

	// Application
	ContentTypeJson  = "application/json"
	ContentTypeXmlA  = "application/xml"
	ContentTypePdf   = "application/pdf"
	ContentTypeZip   = "application/zip"
	ContentTypeGzip  = "application/gzip"
	ContentTypeTar   = "application/x-tar"
	ContentTypeOctet = "application/octet-stream"
	ContentTypeForm  = "application/x-www-form-urlencoded"

	// Office Documents
	ContentTypeXls  = "application/vnd.ms-excel"
	ContentTypeXlsx = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	ContentTypeDoc  = "application/msword"
	ContentTypeDocx = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	ContentTypePpt  = "application/vnd.ms-powerpoint"
	ContentTypePptx = "application/vnd.openxmlformats-officedocument.presentationml.presentation"

	// Font
	ContentTypeTtf   = "font/ttf"
	ContentTypeOtf   = "font/otf"
	ContentTypeWoff  = "font/woff"
	ContentTypeWoff2 = "font/woff2"

	// Multipart
	ContentTypeFormData = "multipart/form-data"

	// Message
	ContentTypeRfc822 = "message/rfc822"
	ContentTypeHttp   = "message/http"

	// Other
	ContentTypeRtf    = "application/rtf"
	ContentTypeSh     = "application/x-sh"
	ContentTypePython = "application/x-python-code"
	ContentTypeJar    = "application/x-java-archive"
)

var (
	ContentTypeMap = map[string]string{
		"text":      ContentTypeText,
		"xml":       ContentTypeXML,
		"file":      ContentTypeFile,
		"plain":     ContentTypePlain,
		"html":      ContentTypeHtml,
		"css":       ContentTypeCss,
		"csv":       ContentTypeCsv,
		"js":        ContentTypeJs,
		"md":        ContentTypeMd,
		"jpeg":      ContentTypeJpeg,
		"png":       ContentTypePng,
		"gif":       ContentTypeGif,
		"svg":       ContentTypeSvg,
		"webp":      ContentTypeWebp,
		"bmp":       ContentTypeBmp,
		"tiff":      ContentTypeTiff,
		"mp3":       ContentTypeMp3,
		"ogg":       ContentTypeOgg,
		"wav":       ContentTypeWav,
		"webm":      ContentTypeWebm,
		"aac":       ContentTypeAac,
		"midi":      ContentTypeMidi,
		"mp4":       ContentTypeMp4,
		"oggv":      ContentTypeOggV,
		"webmv":     ContentTypeWebmV,
		"avi":       ContentTypeAvi,
		"mov":       ContentTypeMov,
		"mkv":       ContentTypeMkv,
		"json":      ContentTypeJson,
		"xmla":      ContentTypeXmlA,
		"pdf":       ContentTypePdf,
		"zip":       ContentTypeZip,
		"gzip":      ContentTypeGzip,
		"tar":       ContentTypeTar,
		"octet":     ContentTypeOctet,
		"form":      ContentTypeForm,
		"xls":       ContentTypeXls,
		"xlsx":      ContentTypeXlsx,
		"doc":       ContentTypeDoc,
		"docx":      ContentTypeDocx,
		"ppt":       ContentTypePpt,
		"pptx":      ContentTypePptx,
		"ttf":       ContentTypeTtf,
		"otf":       ContentTypeOtf,
		"woff":      ContentTypeWoff,
		"woff2":     ContentTypeWoff2,
		"form_data": ContentTypeFormData,
		"rfc822":    ContentTypeRfc822,
		"http":      ContentTypeHttp,
		"rtf":       ContentTypeRtf,
		"sh":        ContentTypeSh,
		"python":    ContentTypePython,
		"jar":       ContentTypeJar,
	}
)

type ExternalDocs struct {
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}
type Tags struct {
	Name         string        `json:"name,omitempty"`
	Description  string        `json:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
}

type OpenAPIObject struct {
	Version    string                `json:"openapi"` // Required
	Info       InfoObject            `json:"info"`    // Required
	Servers    []ServerObject        `json:"servers,omitempty"`
	Paths      PathsObject           `json:"paths"`                // Required
	Components ComponentsObject      `json:"components,omitempty"` // Required for Authorization header
	Security   []map[string][]string `json:"security,omitempty"`
	Tags       []*Tags               `json:"tags,omitempty"`
	// ExternalDocs
}

type ServerObject struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`

	// Variables
}

type InfoObject struct {
	Title          string         `json:"title"`
	Description    string         `json:"description,omitempty"`
	TermsOfService string         `json:"termsOfService,omitempty"`
	Contact        *ContactObject `json:"contact,omitempty"`
	License        *LicenseObject `json:"license,omitempty"`
	Version        string         `json:"version"`
}

type ContactObject struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type LicenseObject struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type PathsObject map[string]*PathItemObject

type PathItemObject struct {
	Ref         string           `json:"$ref,omitempty"`
	Summary     string           `json:"summary,omitempty"`
	Description string           `json:"description,omitempty"`
	Get         *OperationObject `json:"get,omitempty"`
	Post        *OperationObject `json:"post,omitempty"`
	Patch       *OperationObject `json:"patch,omitempty"`
	Put         *OperationObject `json:"put,omitempty"`
	Delete      *OperationObject `json:"delete,omitempty"`
	Options     *OperationObject `json:"options,omitempty"`
	Head        *OperationObject `json:"head,omitempty"`
	Trace       *OperationObject `json:"trace,omitempty"`
	// Servers
	// Parameters
}

type OperationObject struct {
	Responses ResponsesObject `json:"responses"` // Required

	Tags        []string           `json:"tags,omitempty"`
	Summary     string             `json:"summary,omitempty"`
	Description string             `json:"description,omitempty"`
	Parameters  []ParameterObject  `json:"parameters,omitempty"`
	RequestBody *RequestBodyObject `json:"requestBody,omitempty"`
	OperationId string             `json:"operationId,omitempty"`
	Deprecated  bool               `json:"deprecated,omitempty"`
	// Tags
	// ExternalDocs
	// Callbacks
	// Security
	// Servers
}

type ParameterObject struct {
	Name        string        `json:"name,omitempty"` // Required
	In          string        `json:"in,omitempty"`   // Required. Possible values are "query", "header", "path" or "cookie"
	Description string        `json:"description,omitempty"`
	Required    bool          `json:"required,omitempty"`
	Example     interface{}   `json:"example,omitempty"`
	Schema      *SchemaObject `json:"schema,omitempty"`

	// Ref is used when ParameterOjbect is as a ReferenceObject
	Ref string `json:"$ref,omitempty"`

	// Deprecated
	// AllowEmptyValue
	// Style
	// Explode
	// AllowReserved
	// Examples
	// Content
}

type RequestBodyObject struct {
	Content map[string]*MediaTypeObject `json:"content"` // Required

	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`

	// Ref is used when RequestBodyObject is as a ReferenceObject
	Ref string `json:"$ref,omitempty"`
}

type MediaTypeObject struct {
	Schema SchemaObject `json:"schema,omitempty"`
	// Example string       `json:"example,omitempty"`

	// Examples
	// Encoding
}

type SchemaObject struct {
	ID                 string                 `json:"-"` // For go-swagger3
	PkgName            string                 `json:"-"` // For go-swagger3
	FieldName          string                 `json:"-"` // For go-swagger3
	DisabledFieldNames map[string]struct{}    `json:"-"` // For go-swagger3
	Type               string                 `json:"type,omitempty"`
	Format             string                 `json:"format,omitempty"`
	Required           []string               `json:"required,omitempty"`
	Properties         *orderedmap.OrderedMap `json:"properties,omitempty"`
	Description        string                 `json:"description,omitempty"`
	Minimum            any                    `json:"minimum,omitempty"`
	Maximum            any                    `json:"maximum,omitempty"`
	ExclusiveMaximum   any                    `json:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum   any                    `json:"exclusiveMinimum,omitempty"`
	Items              *SchemaObject          `json:"items,omitempty"` // use ptr to prevent recursive error
	Example            interface{}            `json:"example,omitempty"`
	Deprecated         bool                   `json:"deprecated,omitempty"`
	Ref                string                 `json:"$ref,omitempty"` // Ref is used when SchemaObject is as a ReferenceObject
	Enum               interface{}            `json:"enum,omitempty"`
	Title              string                 `json:"title,omitempty"`
	Default            any                    `json:"default,omitempty"`
	MultipleOf         any                    `json:"multipleOf,omitempty"`
	MaxLength          int32                  `json:"maxLength,omitempty"`
	MinLength          int32                  `json:"minLength,omitempty"`
	ReadOnly           bool                   `json:"read_only,omitempty"`
	WriteOnly          bool                   `json:"write_only,omitempty"`
	Pattern            string                 `json:"pattern,omitempty"`
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
	// Nullable
	// XML
	// ExternalDocs
}

type ResponsesObject map[string]*ResponseObject // [status]ResponseObject

type ResponseObject struct {
	Description string `json:"description"` // Required

	Headers map[string]*HeaderObject    `json:"headers,omitempty"`
	Content map[string]*MediaTypeObject `json:"content,omitempty"`

	// Ref is for ReferenceObject
	Ref string `json:"$ref,omitempty"`

	// Links
}

type HeaderObject struct {
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`

	// Ref is used when HeaderObject is as a ReferenceObject
	Ref string `json:"$ref,omitempty"`
}

type ComponentsObject struct {
	Schemas         map[string]*SchemaObject         `json:"schemas,omitempty"`
	SecuritySchemes map[string]*SecuritySchemeObject `json:"securitySchemes,omitempty"`
	Parameters      map[string]*ParameterObject      `json:"parameters,omitempty"`
	// Responses
	// Examples
	// RequestBodies
	// Headers
	// Links
	// Callbacks
}

type SecuritySchemeObject struct {
	// Generic fields
	Type        string `json:"type"` // Required
	Description string `json:"description,omitempty"`

	// http
	Scheme string `json:"scheme,omitempty"`

	// apiKey
	In   string `json:"in,omitempty"`
	Name string `json:"name,omitempty"`

	// OpenID
	OpenIdConnectUrl string `json:"openIdConnectUrl,omitempty"`

	// OAuth2
	OAuthFlows *SecuritySchemeOauthObject `json:"flows,omitempty"`

	// BearerFormat
}

type SecuritySchemeOauthObject struct {
	Implicit              *SecuritySchemeOauthFlowObject `json:"implicit,omitempty"`
	AuthorizationCode     *SecuritySchemeOauthFlowObject `json:"authorizationCode,omitempty"`
	ResourceOwnerPassword *SecuritySchemeOauthFlowObject `json:"password,omitempty"`
	ClientCredentials     *SecuritySchemeOauthFlowObject `json:"clientCredentials,omitempty"`
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
	AuthorizationUrl string            `json:"authorizationUrl,omitempty"`
	TokenUrl         string            `json:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}
