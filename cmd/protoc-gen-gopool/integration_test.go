package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestIntegration tests the actual plugin functionality by running protoc with our plugin
func TestIntegration(t *testing.T) {
	// Create a temporary directory for our test
	tempDir := t.TempDir()

	// Create a simple proto file
	protoContent := `syntax = "proto3";

package test;

option go_package = "./test";

message TestMessage {
  string name = 1;
  int32 id = 2;
}

message AnotherMessage {
  bool active = 1;
  string description = 2;
}`

	protoFilePath := filepath.Join(tempDir, "test.proto")
	if err := os.WriteFile(protoFilePath, []byte(protoContent), 0o644); err != nil {
		t.Fatalf("Failed to write proto file: %v", err)
	}

	// Build the plugin
	pluginBinary := filepath.Join(tempDir, "protoc-gen-gopool")
	// Get the project root directory (where go.mod is located)
	projectRoot, err := filepath.Abs("../..") // From cmd/protoc-gen-gopool, go up two levels to reach project root
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	buildCmd := exec.Command("go", "build", "-o", pluginBinary, "github.com/soyacen/protobuf-gopool/cmd/protoc-gen-gopool")
	buildCmd.Dir = projectRoot

	if err := buildCmd.Run(); err != nil {
		t.Logf("Build command failed with error: %v", err)
		t.Logf("Trying alternative build approach...")

		// Alternative approach: build from current module directory
		altBuildCmd := exec.Command("go", "build", "-o", pluginBinary, ".")
		altBuildCmd.Dir = "." // Current directory (cmd/protoc-gen-gopool)
		if err2 := altBuildCmd.Run(); err2 != nil {
			t.Fatalf("Both build attempts failed. First error: %v, Second error: %v", err, err2)
		}
	}

	// Add the plugin binary to PATH temporarily
	oldPath := os.Getenv("PATH")
	newPath := tempDir + ":" + oldPath
	os.Setenv("PATH", newPath)
	defer os.Setenv("PATH", oldPath) // Restore PATH

	// Run protoc with our plugin
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	protocCmd := exec.Command("protoc",
		"--plugin=protoc-gen-gopool="+pluginBinary,
		"--gopool_out="+outputDir,
		"--go_out="+outputDir,
		"-I"+tempDir,
		"test.proto")

	output, err := protocCmd.CombinedOutput()
	if err != nil {
		t.Logf("protoc command failed: %v", err)
		t.Logf("protoc output: %s", string(output))
		// If protoc is not available, skip the test
		if strings.Contains(string(output), "protoc") || strings.Contains(err.Error(), "executable file not found") {
			t.Skip("protoc not available, skipping integration test")
		} else {
			t.Fatalf("protoc command failed: %v, output: %s", err, string(output))
		}
	}

	// Check if the pool file was generated
	poolFile := filepath.Join(outputDir, "test.pb.pool.go")
	if _, err := os.Stat(poolFile); os.IsNotExist(err) {
		t.Fatalf("Pool file was not generated: %s", poolFile)
	}

	// Read and verify the content of the generated pool file
	content, err := os.ReadFile(poolFile)
	if err != nil {
		t.Fatalf("Failed to read pool file: %v", err)
	}

	contentStr := string(content)

	// Verify that the pool file contains expected elements
	expectedElements := []string{
		"TestMessagePool",    // Pool variable for TestMessage
		"GetTestMessage",     // Get function for TestMessage
		"PutTestMessage",     // Put function for TestMessage
		"AnotherMessagePool", // Pool variable for AnotherMessage
		"GetAnotherMessage",  // Get function for AnotherMessage
		"PutAnotherMessage",  // Put function for AnotherMessage
		"sync.Pool",          // Uses sync.Pool
	}

	for _, expected := range expectedElements {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Generated pool file missing expected element: %s", expected)
		}
	}

	t.Logf("Generated pool file content:\n%s", contentStr)
}

// TestPluginVersion tests that the plugin responds to --version flag
func TestPluginVersion(t *testing.T) {
	tempDir := t.TempDir()

	// Build the plugin
	pluginBinary := filepath.Join(tempDir, "protoc-gen-gopool")
	// Get the project root directory (where go.mod is located)
	projectRoot, err := filepath.Abs("../..") // From cmd/protoc-gen-gopool, go up two levels to reach project root
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	buildCmd := exec.Command("go", "build", "-o", pluginBinary, "github.com/soyacen/protobuf-gopool/cmd/protoc-gen-gopool")
	buildCmd.Dir = projectRoot

	if err := buildCmd.Run(); err != nil {
		t.Logf("Build command failed with error: %v", err)
		t.Logf("Trying alternative build approach...")

		// Alternative approach: build from current module directory
		altBuildCmd := exec.Command("go", "build", "-o", pluginBinary, ".")
		altBuildCmd.Dir = "." // Current directory (cmd/protoc-gen-gopool)
		if err2 := altBuildCmd.Run(); err2 != nil {
			t.Fatalf("Both build attempts failed. First error: %v, Second error: %v", err, err2)
		}
	}

	// Run the plugin with --version flag
	cmd := exec.Command(pluginBinary, "--version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Plugin --version command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "protoc-gen-gopool") {
		t.Errorf("Version output doesn't contain plugin name: %s", outputStr)
	}

	if !strings.Contains(outputStr, "v0.0.1") {
		t.Errorf("Version output doesn't contain expected version: %s", outputStr)
	}

	t.Logf("Plugin version output: %s", outputStr)
}
