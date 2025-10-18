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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/digilolnet/digilol-cert-pushpuller/internal/command"
	"github.com/digilolnet/digilol-cert-pushpuller/internal/config"
	"github.com/digilolnet/digilol-cert-pushpuller/internal/crypto"
	s3client "github.com/digilolnet/digilol-cert-pushpuller/internal/s3"
)

func pull(cfg *config.PullConfig) error {
	ctx := context.Background()

	// Create S3 client
	s3Client, err := s3client.NewClient(ctx, &cfg.S3)
	if err != nil {
		return err
	}

	// Download hashes from S3
	s3Hashes, err := s3client.LoadHashesFromS3(ctx, s3Client, cfg.S3.Bucket, cfg.S3.Prefix)
	if err != nil {
		return err
	}

	// Read all available keys
	keyFiles, err := filepath.Glob(filepath.Join(cfg.KeyDir, "*.key"))
	if err != nil {
		return fmt.Errorf("list key files: %w", err)
	}

	if len(keyFiles) == 0 {
		return nil
	}

	// List all .enc files in S3
	var prefix *string
	if cfg.S3.Prefix != "" {
		p := cfg.S3.Prefix + "/"
		prefix = &p
	}

	listOutput, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &cfg.S3.Bucket,
		Prefix: prefix,
	})
	if err != nil {
		return fmt.Errorf("list S3 objects: %w", err)
	}

	if len(listOutput.Contents) == 0 {
		return nil
	}

	// Create certificate directory
	if err := os.MkdirAll(cfg.CertDir, 0755); err != nil {
		return fmt.Errorf("create certificate directory %s: %w", cfg.CertDir, err)
	}

	// Download and decrypt each .enc file
	for _, obj := range listOutput.Contents {
		fileName := filepath.Base(*obj.Key)

		// Skip non-.enc files
		if !strings.HasSuffix(fileName, ".enc") {
			continue
		}

		// Extract cert name from filename (e.g., _.domain.com from _.domain.com.crt.enc)
		nameWithoutEnc := strings.TrimSuffix(fileName, ".enc")
		certName, ok := config.ExtractCertName(nameWithoutEnc)
		if !ok {
			continue
		}

		// Check if we have the key for this certificate
		key, err := config.LoadKey(cfg.KeyDir, certName)
		if err != nil {
			continue
		}

		// Check if local file exists and compare hash with S3 hashes
		filePath := filepath.Join(cfg.CertDir, nameWithoutEnc)
		if localData, err := os.ReadFile(filePath); err == nil {
			localHash := sha256.Sum256(localData)
			localHashStr := hex.EncodeToString(localHash[:])

			// Compare with hash from S3 .hashes.json
			if s3Hash, ok := s3Hashes[nameWithoutEnc]; ok && s3Hash == localHashStr {
				continue
			}
		}

		// Get object from S3
		getOutput, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &cfg.S3.Bucket,
			Key:    obj.Key,
		})
		if err != nil {
			return fmt.Errorf("download %s from S3: %w", *obj.Key, err)
		}

		encryptedData, err := io.ReadAll(getOutput.Body)
		getOutput.Body.Close()
		if err != nil {
			return fmt.Errorf("read %s: %w", *obj.Key, err)
		}

		// Decrypt the data
		decrypted, err := crypto.DecryptData(encryptedData, key)
		if err != nil {
			return fmt.Errorf("decrypt %s: %w", *obj.Key, err)
		}

		// Write to local file (without .enc extension)
		if err := os.WriteFile(filePath, decrypted, 0600); err != nil {
			return fmt.Errorf("write %s: %w", filePath, err)
		}

		log.Printf("downloaded %s", nameWithoutEnc)
	}

	// Run reload command if specified
	if cfg.ReloadCmd != "" {
		if err := command.RunCommandWithEnv(cfg.ReloadCmd, nil); err != nil {
			return fmt.Errorf("run reload command: %w", err)
		}
	}

	return nil
}
