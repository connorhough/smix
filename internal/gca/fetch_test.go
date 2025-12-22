package gca

import "testing"

func TestInferLanguage(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"main.go", "go"},
		{"app.tsx", "tsx"},
		{"Dockerfile", "dockerfile"},
		{"Makefile", "makefile"},
		{"Jenkinsfile", "jenkinsfile"},
		{"go.mod", "go"},
		{"go.sum", "go"},
		{".editorconfig", "editorconfig"},
		{"script.sh", "sh"},
		{"unknown", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := inferLanguage(tt.filename)
			if got != tt.want {
				t.Errorf("inferLanguage(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}
