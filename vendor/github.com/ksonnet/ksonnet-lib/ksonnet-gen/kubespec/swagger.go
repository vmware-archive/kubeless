package kubespec

// APISpec represents an OpenAPI specification of an API.
type APISpec struct {
	SwaggerVersion string            `json:"swagger"`
	Info           *SchemaInfo       `json:"info"`
	Definitions    SchemaDefinitions `json:"definitions"`

	// Fields we currently ignore:
	//   - paths
	//   - securityDefinitions
	//   - security

	// Not part of the OpenAPI spec. Filled in later.
	FilePath string
	Text     []byte
}

// SchemaInfo contains information about the the API represented with
// `APISpec`. For example, `title` might be `"Kubernetes"`, and
// `version` might be `"v1.7.0"`.
type SchemaInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// SchemaDefinition is an API object definition. For example, this
// might contain a name (e.g., `v1.APIGroup`), a set of properties
// (e.g., `apiVersion`, `kind`, and so on), and the names of required
// properties.
type SchemaDefinition struct {
	Type          *SchemaType   `json:"type"`
	Description   string        `json:"description"` // nullable.
	Required      []string      `json:"required"`    // nullable.
	Properties    Properties    `json:"properties"`  // nullable.
	TopLevelSpecs TopLevelSpecs `json:"x-kubernetes-group-version-kind"`
}

// TopLevelSpec is a property that exists on `SchemaDefinition`s for
// top-level API objects.
type TopLevelSpec struct {
	Group   GroupName     `json:"Group"`
	Version VersionString `json:"Version"`
	Kind    ObjectKind    `json:"Kind"`
}
type TopLevelSpecs []*TopLevelSpec

// SchemaDefinitions is a named collection of `SchemaDefinition`s,
// represented as a collection mapping definition name ->
// `SchemaDefinition`.
type SchemaDefinitions map[DefinitionName]*SchemaDefinition

// Property represents an object property for some API object. For
// example, `v1.APIGroup` might contain a property called
// `apiVersion`, which would be specifid by a `Property`.
type Property struct {
	Description string      `json:"description"`
	Type        *SchemaType `json:"type"`
	Ref         *ObjectRef  `json:"$ref"`
	Items       Items       `json:"items"` // nil unless Type == "array".
}

// Properties is a named collection of `Properties`s, represented as a
// collection mapping definition name -> `Properties`.
type Properties map[PropertyName]*Property

// Items represents the type of an element in an array. Usually this
// is used to fully specify a `Property` object whose `type` field is
// `"array"`.
type Items struct {
	Ref *ObjectRef `json:"$ref"`

	// Ignored fields:
	// - Type *SchemaType `json:"type"`
	// - Format *string `json:"format"`
}

// SchemaType represents the type of some object in an API spec. For
// example, a property might have type `string`.
type SchemaType string

func (st SchemaType) String() string {
	return string(st)
}

// ObjectRef represents a reference to some API object. For example,
// `#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta`
type ObjectRef string

func (or ObjectRef) String() string {
	return string(or)
}

// PropertyName represents the name of a property. For example,
// `apiVersion` or `kind`.
type PropertyName string

func (pn PropertyName) String() string {
	return string(pn)
}

// DefinitionName represents the name of a definition. For example,
// `v1.APIGroup`.
type DefinitionName string

func (dn DefinitionName) String() string {
	return string(dn)
}
