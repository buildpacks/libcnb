package internal

// LayerAPI5 is used for backwards compatibility with serialization/deserialization of the layer toml
type LayerAPI5 struct {
	// Build indicates that a layer should be used for builds.
	Build bool `toml:"build"`

	// Cache indicates that a layer should be cached.
	Cache bool `toml:"cache"`

	// Launch indicates that a layer should be used for launch.
	Launch bool `toml:"launch"`

	// Metadata is the metadata associated with the layer.
	Metadata map[string]interface{} `toml:"metadata"`
}
