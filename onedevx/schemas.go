package main

type OneDevxCRD struct {
	ApiVersion string               `yaml:"apiVersion"`
	Kind       string               `yaml:"kind"`
	Metadata   OneDevxCRDMetadata   `yaml:"metadata"`
	Spec       OneDevxComponentSpec `yaml:"spec"`
}

type OneDevxCRDMetadata struct {
	Name string `yaml:"name"`
}

type OneDevxComponentHelm struct {
	ChartName string `yaml:"chartName"`
}

type OneDevxComponentImage struct {
	ImageName string `yaml:"imageName"`
}
type OneDevxComponentSpec struct {
	ComponentType string                `yaml:"componentType"`
	HelmInfo      OneDevxComponentHelm  `yaml:"helmInfo"`
	ImageInfo     OneDevxComponentImage `yaml:"imageInfo"`

	RestSchema   []string `yaml:"restSchema"`
	Dependencies []string `yaml:"dependencies"`
}
