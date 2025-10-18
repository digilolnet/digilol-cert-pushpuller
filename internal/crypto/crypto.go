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
	"fmt"
	"io"

	"github.com/minio/sio"
)

// EncryptData encrypts data using the provided key
func EncryptData(data []byte, key []byte) ([]byte, error) {
	var buf bytes.Buffer

	config := sio.Config{
		MinVersion: sio.Version20,
		Key:        key,
	}

	encWriter, err := sio.EncryptWriter(&buf, config)
	if err != nil {
		return nil, fmt.Errorf("create encryption writer: %w", err)
	}

	if _, err := encWriter.Write(data); err != nil {
		return nil, fmt.Errorf("encrypt data: %w", err)
	}

	if err := encWriter.Close(); err != nil {
		return nil, fmt.Errorf("close encryption writer: %w", err)
	}

	return buf.Bytes(), nil
}

// DecryptData decrypts data using the provided key
func DecryptData(encryptedData []byte, key []byte) ([]byte, error) {
	config := sio.Config{
		MinVersion: sio.Version20,
		Key:        key,
	}

	decReader, err := sio.DecryptReader(bytes.NewReader(encryptedData), config)
	if err != nil {
		return nil, fmt.Errorf("create decryption reader: %w", err)
	}

	decrypted, err := io.ReadAll(decReader)
	if err != nil {
		return nil, fmt.Errorf("decrypt data: %w", err)
	}

	return decrypted, nil
}
