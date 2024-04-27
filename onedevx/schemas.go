package main

// Common headers
type OneDevxCRD struct {
	ApiVersion string             `yaml:"apiVersion"`
	Kind       string             `yaml:"kind"`
	Metadata   OneDevxCRDMetadata `yaml:"metadata"`
}

type OneDevxCRDMetadata struct {
	Name string `yaml:"name"`
}

// Components

type OneDevxComponentCRD struct {
	OneDevxCRD `yaml:",inline"`
	Spec       OneDevxComponentSpec `yaml:"spec"`
}

type OneDevxComponentSpec struct {
	ComponentType string                `yaml:"componentType"`
	HelmInfo      OneDevxComponentHelm  `yaml:"helmInfo"`
	ImageInfo     OneDevxComponentImage `yaml:"imageInfo"`

	RestSchema   OneDevxRestSchema `yaml:"restSchema"`
	Dependencies []string          `yaml:"dependencies"`
}

type OneDevxRestSchema struct {
	Port    int      `yaml:"port"`
	Entries []string `yaml:"entries"`
}
type OneDevxComponentHelm struct {
	ChartName    string `yaml:"chartName"`
	ChartRepo    string `yaml:"chartRepo"`
	ChartVersion string `yaml:"chartVersion"`
}

type OneDevxComponentImage struct {
	ImageName string `yaml:"imageName"`
}

// Workspec

type OneDevxWorkspecSpec struct {
	ComponentList []struct {
		Type string `yaml:"type"`
		Path string `yaml:"path"`
	} `yaml:"componentList"`
}

type OneDevxWorkspecCRD struct {
	OneDevxCRD `yaml:",inline"`
	Spec       OneDevxWorkspecSpec `yaml:"spec"`
}
