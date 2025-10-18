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

package s3

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	internalConfig "github.com/digilolnet/digilol-cert-pushpuller/internal/config"
)

// NewClient creates a new S3 client from S3 configuration
func NewClient(ctx context.Context, s3Config *internalConfig.S3Config) (*s3.Client, error) {
	s3Cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(s3Config.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s3Config.AccessKey,
			s3Config.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load S3 config: %w", err)
	}

	s3Options := []func(*s3.Options){
		func(o *s3.Options) {
			if s3Config.Endpoint != "" {
				o.BaseEndpoint = aws.String(s3Config.Endpoint)
			}
			o.UsePathStyle = s3Config.ForcePathStyle
		},
	}

	return s3.NewFromConfig(s3Cfg, s3Options...), nil
}

// BuildKey builds an S3 key with optional prefix
func BuildKey(prefix, fileName string) string {
	if prefix != "" {
		return prefix + "/" + fileName
	}
	return fileName
}

// LoadHashesFromS3 downloads and parses .hashes.json from S3
func LoadHashesFromS3(ctx context.Context, client *s3.Client, bucket, prefix string) (map[string]string, error) {
	hashesKey := BuildKey(prefix, ".hashes.json")
	hashes := make(map[string]string)

	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &hashesKey,
	})
	if err != nil {
		return hashes, nil // Return empty map if file doesn't exist
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return hashes, nil
	}

	json.Unmarshal(data, &hashes)
	return hashes, nil
}
