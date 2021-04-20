package internal

// LayerAPI5 is used for backwards compatibility with serialization/deserialization of the layer toml
type LayerAPI5 struct {
	// bool's are inlined to instead of using libcnb.LayerTypes to break a circular package reference, internal -> libcnb -> internal -> ...

	// Build indicates that a layer should be used for builds.
	Build bool `toml:"build"`

	// Cache indicates that a layer should be cached.
	Cache bool `toml:"cache"`

	// Launch indicates that a layer should be used for launch.
	Launch bool `toml:"launch"`

	// Metadata is the metadata associated with the layer.
	Metadata map[string]interface{} `toml:"metadata"`
}
