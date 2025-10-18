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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type S3Config struct {
	Bucket         string `toml:"bucket"`
	Endpoint       string `toml:"endpoint"`
	Region         string `toml:"region"`
	Prefix         string `toml:"prefix"`
	ForcePathStyle bool   `toml:"force_path_style"`
	AccessKey      string `toml:"access_key"`
	SecretKey      string `toml:"secret_key"`
}

type LegoCommand struct {
	Command string            `toml:"command"`
	Env     map[string]string `toml:"env"`
}

type DaemonConfig struct {
	Enabled      bool `toml:"enabled"`
	IntervalSecs int  `toml:"interval_secs"`
	JitterSecs   int  `toml:"jitter_secs"`
}

type PushConfig struct {
	KeyDir       string        `toml:"key_dir"`
	CertDir      string        `toml:"cert_dir"`
	LegoCommands []LegoCommand `toml:"lego_commands"`
	ReloadCmd    string        `toml:"reload_cmd"`
	S3           S3Config      `toml:"s3"`
	Daemon       DaemonConfig  `toml:"daemon"`
}

type PullConfig struct {
	KeyDir    string       `toml:"key_dir"`
	CertDir   string       `toml:"cert_dir"`
	ReloadCmd string       `toml:"reload_cmd"`
	S3        S3Config     `toml:"s3"`
	Daemon    DaemonConfig `toml:"daemon"`
}

// LoadPush loads the push configuration from a TOML file
func LoadPush(configPath string) (*PushConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg PushConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// LoadPull loads the pull configuration from a TOML file
func LoadPull(configPath string) (*PullConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg PullConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// GenerateKey generates a random 32-byte key for encryption
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate random key: %w", err)
	}
	return key, nil
}

// SaveKey saves the encryption key to a .key file as base64
func SaveKey(keyDir, certName string, key []byte) error {
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("create key directory: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(key) + "\n"
	keyFile := filepath.Join(keyDir, certName+".key")
	if err := os.WriteFile(keyFile, []byte(encoded), 0600); err != nil {
		return fmt.Errorf("write key file %s: %w", keyFile, err)
	}

	return nil
}

// LoadKey loads the encryption key from a .key file (base64 encoded)
func LoadKey(keyDir, certName string) ([]byte, error) {
	keyFile := filepath.Join(keyDir, certName+".key")
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("read key file %s: %w", keyFile, err)
	}

	// Trim whitespace (including newline)
	encoded := strings.TrimSpace(string(data))

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode key from %s: %w", keyFile, err)
	}

	if len(decoded) != 32 {
		return nil, fmt.Errorf("key file %s: invalid size %d bytes (expected 32)", keyFile, len(decoded))
	}

	return decoded, nil
}

// GetOrCreateKey gets an existing key or creates a new one if it doesn't exist
func GetOrCreateKey(keyDir, certName string) ([]byte, error) {
	key, err := LoadKey(keyDir, certName)
	if err == nil {
		return key, nil
	}

	// Key doesn't exist, create a new one
	key, err = GenerateKey()
	if err != nil {
		return nil, err
	}

	if err := SaveKey(keyDir, certName, key); err != nil {
		return nil, err
	}

	return key, nil
}
