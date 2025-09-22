package main

import (
	"testing"
)

func TestParseImageInfo(t *testing.T) {
	// 测试用例
	tests := []struct {
		name               string
		input              string
		expectedRegistry   string
		expectedNamespace  string
		expectedRepository string
		expectedTag        string
	}{
		{
			name:               "official image with tag",
			input:              "nginx:latest",
			expectedRegistry:   "registry-1.docker.io",
			expectedNamespace:  "library",
			expectedRepository: "nginx",
			expectedTag:        "latest",
		},
		{
			name:               "official image without tag",
			input:              "nginx",
			expectedRegistry:   "registry-1.docker.io",
			expectedNamespace:  "library",
			expectedRepository: "nginx",
			expectedTag:        "latest",
		},
		{
			name:               "image with namespace and tag",
			input:              "library/nginx:1.20",
			expectedRegistry:   "registry-1.docker.io",
			expectedNamespace:  "library",
			expectedRepository: "nginx",
			expectedTag:        "1.20",
		},
		{
			name:               "full qualified image name",
			input:              "docker.io/library/nginx:latest",
			expectedRegistry:   "registry-1.docker.io",
			expectedNamespace:  "library",
			expectedRepository: "nginx",
			expectedTag:        "latest",
		},
		{
			name:               "private registry image",
			input:              "myregistry.com/myproject/myapp:v1.0",
			expectedRegistry:   "myregistry.com",
			expectedNamespace:  "myproject",
			expectedRepository: "myapp",
			expectedTag:        "v1.0",
		},
		{
			name:               "private registry with port",
			input:              "myregistry.com:5000/myproject/myapp:v1.0",
			expectedRegistry:   "myregistry.com:5000",
			expectedNamespace:  "myproject",
			expectedRepository: "myapp",
			expectedTag:        "v1.0",
		},
		{
			name:               "private registry image without tag",
			input:              "myregistry.com:5000/myproject/myapp",
			expectedRegistry:   "myregistry.com:5000",
			expectedNamespace:  "myproject",
			expectedRepository: "myapp",
			expectedTag:        "latest",
		},
		{
			name:               "private registry image without tag and namespace",
			input:              "myregistry.com:5000/myapp",
			expectedRegistry:   "myregistry.com:5000",
			expectedNamespace:  "library",
			expectedRepository: "myapp",
			expectedTag:        "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseImageInfo(tt.input)
			if err != nil {
				t.Errorf("ParseImageInfo(%q) error = %v", tt.input, err)
				return
			}

			if result.Registry != tt.expectedRegistry {
				t.Errorf("ParseImageInfo(%q) Registry = %v, want %v", tt.input, result.Registry, tt.expectedRegistry)
			}

			if result.Namespace != tt.expectedNamespace {
				t.Errorf("ParseImageInfo(%q) Namespace = %v, want %v", tt.input, result.Namespace, tt.expectedNamespace)
			}

			if result.Repository != tt.expectedRepository {
				t.Errorf("ParseImageInfo(%q) Repository = %v, want %v", tt.input, result.Repository, tt.expectedRepository)
			}

			if result.Tag != tt.expectedTag {
				t.Errorf("ParseImageInfo(%q) Tag = %v, want %v", tt.input, result.Tag, tt.expectedTag)
			}
		})
	}
}
