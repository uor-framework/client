package v1alpha1

// DataSetConfigurationKind object kind of DataSetConfiguration.
const DataSetConfigurationKind = "DataSetConfiguration"

// DataSetConfiguration configures a dataset
type DataSetConfiguration struct {
	Kind       string `mapstructure:"kind,omitempty"`
	APIVersion string `mapstructure:"apiVersion,omitempty"`
	// Files defines custom attributes to add the the files in the
	// workspaces when publishing content/
	Files []File `mapstructure:"files,omitempty"`
	// SchemaAddress is the remote location for the default schema of the
	// collection.
	SchemaAddress string `mapstructure:"schemaAddress,omitempty"`
	// LinkedCollections are the remote addresses of collection that are
	// linked to the collection.
	LinkedCollections []string `mapstructure:"linkedCollections,omitempty"`
}

// File associates attributes with file names
type File struct {
	File       string            `mapstructure:"file,omitempty"`
	Attributes map[string]string `mapstructure:"attributes,omitempty"`
}
