package module

import (
	"encoding/json"
	"os"
)

type Metadata struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Author          string            `json:"author"`
	License         string            `json:"license"`
	Summary         string            `json:"summary"`
	Source          string            `json:"source"`
	ProjectPage     string            `json:"project_page,omitempty"`
	IssuesURL       string            `json:"issues_url,omitempty"`
	Dependencies    []Dependency      `json:"dependencies"`
	Requirements    []Requirement     `json:"requirements"`
	OperatingSystem []OperatingSystem `json:"operatingsystem_support"`
	Tags            []string          `json:"tags"`
}

type Dependency struct {
	Name               string `json:"name"`
	VersionRequirement string `json:"version_requirement"`
}

type Requirement struct {
	Name               string `json:"name"`
	VersionRequirement string `json:"version_requirement"`
}

type OperatingSystem struct {
	Name    string   `json:"operatingsystem"`
	Release []string `json:"operatingsystemrelease"`
}

func NewMetadata(name string, author string) Metadata {
	return Metadata{
		Name:         name,
		Author:       author,
		License:      "Apache-2.0",
		Summary:      "",
		Source:       "",
		ProjectPage:  "",
		IssuesURL:    "",
		Dependencies: []Dependency{},
		Requirements: []Requirement{
			{
				Name:               "puppet",
				VersionRequirement: ">= 7.0.0 < 9.0.0",
			},
		},
		OperatingSystem: []OperatingSystem{},
		Tags:            []string{},
	}
}

func ReadMetadata(path string) (Metadata, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Metadata{}, err
	}

	metadata := Metadata{}
	err = json.Unmarshal(content, &metadata)
	if err != nil {
		return Metadata{}, err
	}
	return metadata, nil
}

func (m Metadata) Write(path string) error {
	output, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, output, 0644)
}
