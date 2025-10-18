// Copyright 2025 Laurynas Četyrkinas <laurynas@digilol.net>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"os"
	"strings"
	"testing"
)

func TestRunCommandWithEnvVariables(t *testing.T) {
	tmpFile := t.TempDir() + "/env_test.txt"

	env := map[string]string{
		"TEST_VAR": "test_value",
	}

	// Write env var to file
	cmd := "echo $TEST_VAR > " + tmpFile
	err := RunCommandWithEnv(cmd, env)
	if err != nil {
		t.Fatalf("RunCommandWithEnv failed: %v", err)
	}

	// Read file and verify
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := strings.TrimSpace(string(data))
	if output != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", output)
	}
}
