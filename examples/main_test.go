package main

import (
	"log"
	"os"
	"os/exec"
	"testing"
)

// TestMain runs before all tests in this package
func TestMain(m *testing.M) {
	// Run code generation before tests
	log.Println("Regenerating code before tests...")
	cmd := exec.Command("go", "generate", "./...")
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}
	
	// Run tests
	code := m.Run()
	
	// Exit with the test result code
	os.Exit(code)
}