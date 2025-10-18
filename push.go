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

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/digilolnet/digilol-cert-pushpuller/internal/command"
	"github.com/digilolnet/digilol-cert-pushpuller/internal/config"
	"github.com/digilolnet/digilol-cert-pushpuller/internal/crypto"
	s3client "github.com/digilolnet/digilol-cert-pushpuller/internal/s3"
)

func push(cfg *config.PushConfig) error {
	ctx := context.Background()

	// Run all lego commands if configured
	for i, legoCmd := range cfg.LegoCommands {
		if err := command.RunCommandWithEnv(legoCmd.Command, legoCmd.Env); err != nil {
			log.Printf("lego command %d/%d failed: %v", i+1, len(cfg.LegoCommands), err)
			// Continue anyway to try other commands and push existing certs
		}
	}

	// Create S3 client
	s3Client, err := s3client.NewClient(ctx, &cfg.S3)
	if err != nil {
		return err
	}

	// Download existing hashes from S3
	existingHashes, err := s3client.LoadHashesFromS3(ctx, s3Client, cfg.S3.Bucket, cfg.S3.Prefix)
	if err != nil {
		return err
	}

	// Find all certificate files
	entries, err := os.ReadDir(cfg.CertDir)
	if err != nil {
		return fmt.Errorf("read certificate directory %s: %w", cfg.CertDir, err)
	}

	// Group files by certificate name (e.g., _.domain.com)
	// Only care about .crt and .key files
	certFiles := make(map[string][]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		certName, ok := config.ExtractCertName(name)
		if !ok {
			continue
		}

		certFiles[certName] = append(certFiles[certName], name)
	}

	newHashes := make(map[string]string)
	uploaded := false

	// Process each certificate
	for certName, files := range certFiles {
		// Get or create encryption key for this certificate
		key, err := config.GetOrCreateKey(cfg.KeyDir, certName)
		if err != nil {
			return fmt.Errorf("get encryption key for %s: %w", certName, err)
		}

		for _, fileName := range files {
			filePath := filepath.Join(cfg.CertDir, fileName)
			data, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("failed to read %s: %v", filePath, err)
				continue
			}

			// Calculate SHA256 of unencrypted file
			localHash := sha256.Sum256(data)
			localHashStr := hex.EncodeToString(localHash[:])

			// Build S3 key with .enc extension
			s3FileName := fileName + ".enc"
			s3Key := s3client.BuildKey(cfg.S3.Prefix, s3FileName)

			newHashes[s3FileName] = localHashStr

			// Check if hash matches
			if existingHash, ok := existingHashes[s3FileName]; ok && existingHash == localHashStr {
				continue
			}

			// Encrypt the certificate data
			encrypted, err := crypto.EncryptData(data, key)
			if err != nil {
				return fmt.Errorf("encrypt %s: %w", filePath, err)
			}

			// Upload to S3
			_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
				Bucket: &cfg.S3.Bucket,
				Key:    &s3Key,
				Body:   bytes.NewReader(encrypted),
			})
			if err != nil {
				return fmt.Errorf("upload %s to S3: %w", s3Key, err)
			}

			uploaded = true
			log.Printf("uploaded %s", s3Key)
		}
	}

	// Upload updated hashes file if anything changed
	if uploaded {
		hashesJSON, err := json.Marshal(newHashes)
		if err != nil {
			return fmt.Errorf("marshal hashes: %w", err)
		}
		hashesJSON = append(hashesJSON, '\n')

		hashesKey := s3client.BuildKey(cfg.S3.Prefix, ".hashes.json")
		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &cfg.S3.Bucket,
			Key:    &hashesKey,
			Body:   bytes.NewReader(hashesJSON),
		})
		if err != nil {
			log.Printf("failed to upload hashes file: %v", err)
		}
	}

	// Run reload command if specified
	if cfg.ReloadCmd != "" {
		if err := command.RunCommandWithEnv(cfg.ReloadCmd, nil); err != nil {
			return fmt.Errorf("run reload command: %w", err)
		}
	}

	return nil
}
