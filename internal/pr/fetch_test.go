package pr

import "testing"

func TestAddLineNumbers(t *testing.T) {
	input := `func main() {
	fmt.Println("Hello")
}`
	want := `1: func main() {
2: 	fmt.Println("Hello")
3: }`

	got := addLineNumbers(input, 1)
	if got != want {
		t.Errorf("addLineNumbers() =\n%s\nwant:\n%s", got, want)
	}
}

func TestAddLineNumbersWithOffset(t *testing.T) {
	input := `fmt.Println("Hello")
return nil`
	want := `42: fmt.Println("Hello")
43: return nil`

	got := addLineNumbers(input, 42)
	if got != want {
		t.Errorf("addLineNumbers() =\n%s\nwant:\n%s", got, want)
	}
}

func TestAddLineNumbersWithTrailingNewline(t *testing.T) {
	input := "line 1\nline 2\n"
	want := "1: line 1\n2: line 2"
	got := addLineNumbers(input, 1)
	if got != want {
		t.Errorf("addLineNumbers() with trailing newline =\n%s\nwant:\n%s", got, want)
	}
}

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
