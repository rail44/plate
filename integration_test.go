package plate_test

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// TestExamplesIntegration runs the examples tests with fresh code generation
func TestExamplesIntegration(t *testing.T) {
	examplesDir := filepath.Join(".", "examples")
	
	// Step 1: Generate fresh code BEFORE compiling tests
	t.Log("Generating fresh code...")
	generateCmd := exec.Command("go", "generate", "./...")
	generateCmd.Dir = examplesDir
	output, err := generateCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to generate code: %v\nOutput:\n%s", err, output)
	}
	
	// Step 2: Run tests with the newly generated code
	// This ensures the tests compile and run with the latest generated code
	t.Log("Running tests with fresh generated code...")
	testCmd := exec.Command("go", "test", "-v")
	testCmd.Dir = examplesDir
	output, err = testCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run tests: %v\nOutput:\n%s", err, output)
	}
	
	t.Logf("Examples tests passed:\n%s", output)
}
