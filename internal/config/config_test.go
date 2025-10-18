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

package config

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAndLoadKey(t *testing.T) {
	tmpDir := t.TempDir()
	certName := "test-cert"

	// Generate and save a key
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	err = SaveKey(tmpDir, certName, key)
	if err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}

	// Verify file exists and has correct permissions
	keyFile := filepath.Join(tmpDir, certName+".key")
	info, err := os.Stat(keyFile)
	if err != nil {
		t.Fatalf("Key file not created: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", info.Mode().Perm())
	}

	// Verify file is base64 encoded with newline
	data, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("Failed to read key file: %v", err)
	}

	if !strings.HasSuffix(string(data), "\n") {
		t.Error("Key file should end with newline")
	}

	encoded := strings.TrimSpace(string(data))
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Errorf("Key file should contain valid base64: %v", err)
	}

	if len(decoded) != 32 {
		t.Errorf("Decoded key should be 32 bytes, got %d", len(decoded))
	}

	// Load the key back
	loadedKey, err := LoadKey(tmpDir, certName)
	if err != nil {
		t.Fatalf("LoadKey failed: %v", err)
	}

	if string(loadedKey) != string(key) {
		t.Error("Loaded key does not match saved key")
	}
}

func TestGetOrCreateKey(t *testing.T) {
	tmpDir := t.TempDir()
	certName := "auto-cert"

	// First call should create new key
	key1, err := GetOrCreateKey(tmpDir, certName)
	if err != nil {
		t.Fatalf("GetOrCreateKey failed: %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key1))
	}

	// Second call should return existing key
	key2, err := GetOrCreateKey(tmpDir, certName)
	if err != nil {
		t.Fatalf("GetOrCreateKey failed on second call: %v", err)
	}

	if string(key1) != string(key2) {
		t.Error("GetOrCreateKey should return same key on subsequent calls")
	}
}
