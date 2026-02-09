// Copyright 2026 The KitOps Authors.
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
//
// SPDX-License-Identifier: Apache-2.0

package s3api

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3config "github.com/aws/aws-sdk-go-v2/config"
	s3credentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kitops-ml/kitops/pkg/output"
)

type S3ObjectReference struct {
	Bucket  string
	Key     string
	Hash    string
	Version string
}

// ParseS3ObjectReference parses a Kit-spec 's3://' reference and hash into an S3ObjectReference. For filling
// endpoint and region, environment variables are read.
// The format for an s3:// reference is
//
//	s3://<bucket-name>/<object-key>[?versionId=<version-ID>]
func ParseS3ObjectReference(ref string, hash string) (*S3ObjectReference, error) {
	if hash == "" {
		return nil, fmt.Errorf("hash is required for s3 references")
	}
	s3url, err := url.Parse(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to parse s3 reference: %w", err)
	}
	if s3url.Scheme != "s3" {
		return nil, fmt.Errorf("url does not use 's3' scheme")
	}
	bucketName := s3url.Host
	bucketKey := strings.TrimPrefix(s3url.Path, "/")
	version := s3url.Query().Get("versionId")

	return &S3ObjectReference{
		Bucket:  bucketName,
		Key:     bucketKey,
		Hash:    hash,
		Version: version,
	}, nil
}

func VerifyObjectExists(ctx context.Context, client *s3.Client, ref *S3ObjectReference) error {
	var refVersion *string
	if ref.Version != "" {
		refVersion = &ref.Version
	}
	obj, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:    &ref.Bucket,
		Key:       &ref.Key,
		VersionId: refVersion,
	})
	if err != nil {
		return fmt.Errorf("failed to HEAD object in S3 bucket: %w", err)
	}
	if obj.ETag == nil {
		return fmt.Errorf("missing ETag on object")
	}

	// ETags as returned by the API are strings containing quotation marks, which we need to strip
	if strings.Trim(*obj.ETag, `"`) != ref.Hash {
		return fmt.Errorf("object in s3 bucket does not match hash: ETag = %s, hash = %s", *obj.ETag, ref.Hash)
	}

	return nil
}

func SetUpClient(ctx context.Context) (*s3.Client, error) {
	var cfgOpts []func(*s3config.LoadOptions) error
	var clientOpts []func(*s3.Options)

	output.Debugf("Setting up S3 client")
	// Parse ref and env into AWS options
	accessKeyID := os.Getenv(AccessKeyIDEnvVar)
	secretAccessKey := os.Getenv(SecretAccessKeyEnvVar)
	if accessKeyID != "" && secretAccessKey != "" {
		output.Logf(output.LogLevelTrace, "Using access key credentials")
		cfgOpts = append(cfgOpts, s3config.WithCredentialsProvider(s3credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)))
	}
	if region := os.Getenv(RegionEnvVar); region != "" {
		output.Logf(output.LogLevelTrace, "Using region %s", region)
		cfgOpts = append(cfgOpts, s3config.WithRegion(region))
	}
	if endpoint := os.Getenv(APIEndpointEnvVar); endpoint != "" {
		output.Logf(output.LogLevelTrace, "Using API endpoint %s", endpoint)
		clientOpts = append(clientOpts, func(o *s3.Options) { o.BaseEndpoint = aws.String(endpoint) })
		// For now, use path style URLs (endpoint.com/bucket/key) instead of virtual-hosted style (bucket.endpoint.com/key) when
		// an alternate endpoint is specified; this is used/supported by most S3-compatible APIs
		clientOpts = append(clientOpts, func(o *s3.Options) { o.UsePathStyle = true })
	}

	s3cfg, err := s3config.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load S3 configuration: %w", err)
	}

	client := s3.NewFromConfig(s3cfg, clientOpts...)

	return client, nil
}
