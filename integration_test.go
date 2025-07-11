package plate_test

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// TestExamplesIntegration runs the examples tests
func TestExamplesIntegration(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Run tests in examples directory
	// Note: examples/main_test.go will automatically run go generate before tests
	examplesDir := filepath.Join(".", "examples")
	
	t.Log("Running tests in examples...")
	testCmd := exec.Command("go", "test", "-v")
	testCmd.Dir = examplesDir
	output, err := testCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run tests: %v\nOutput:\n%s", err, output)
	}
	
	t.Logf("Examples tests passed:\n%s", output)
}