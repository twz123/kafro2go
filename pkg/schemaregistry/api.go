package schemaregistry

// API of the Schema Registry.
type API interface {
	GetSchemaByID(id int64) (string, error)
}

func NewRegistry(host string) API {
	return &restAPI{host}
}
