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

package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// Generate a random 32-byte key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	testData := []byte("Hello, this is a test message!")

	// Encrypt the data
	encrypted, err := EncryptData(testData, key)
	if err != nil {
		t.Fatalf("EncryptData failed: %v", err)
	}

	// Encrypted data should be different from original
	if bytes.Equal(encrypted, testData) {
		t.Error("Encrypted data should be different from original")
	}

	// Decrypt the data
	decrypted, err := DecryptData(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptData failed: %v", err)
	}

	// Decrypted data should match original
	if !bytes.Equal(decrypted, testData) {
		t.Errorf("Decrypted data doesn't match original.\nExpected: %s\nGot: %s", testData, decrypted)
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	_, err := rand.Read(key1)
	if err != nil {
		t.Fatalf("Failed to generate key1: %v", err)
	}
	_, err = rand.Read(key2)
	if err != nil {
		t.Fatalf("Failed to generate key2: %v", err)
	}

	testData := []byte("Secret data")

	// Encrypt with key1
	encrypted, err := EncryptData(testData, key1)
	if err != nil {
		t.Fatalf("EncryptData failed: %v", err)
	}

	// Try to decrypt with key2
	_, err = DecryptData(encrypted, key2)
	if err == nil {
		t.Error("DecryptData should fail with wrong key")
	}
}
