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
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
)

func TestParseS3ObjectReference(t *testing.T) {
	tests := []struct {
		name      string
		inputRef  string
		inputHash string
		expected  *S3ObjectReference
		errRegexp string
	}{
		{
			name:      "Parses ref from S3 URL reference",
			inputRef:  "s3://mybucket/path/to/file",
			inputHash: "testhash",
			expected: &S3ObjectReference{
				Bucket:  "mybucket",
				Key:     "path/to/file",
				Version: "",
				Hash:    "testhash",
			},
		},
		{
			name:      "Includes version ID from S3 URL reference",
			inputRef:  "s3://mybucket/path/to/file?versionId=testversion",
			inputHash: "testhash",
			expected: &S3ObjectReference{
				Bucket:  "mybucket",
				Key:     "path/to/file",
				Version: "testversion",
				Hash:    "testhash",
			},
		},
		{
			name:      "Handles multiple slashes (leading slash in key)",
			inputRef:  "s3://mybucket//file",
			inputHash: "testhash",
			expected: &S3ObjectReference{
				Bucket:  "mybucket",
				Key:     "/file",
				Version: "",
				Hash:    "testhash",
			},
		},
		{
			name:      "Requires hash to be included",
			inputRef:  "s3://mybucket/path/to/file?versionId=testversion",
			inputHash: "",
			errRegexp: "hash is required",
		},
		{
			name:      "Fails on non-s3 scheme",
			inputRef:  "s4://mybucket/path/to/file",
			inputHash: "testhash",
			errRegexp: "url does not use 's3' scheme",
		},
		{
			name:      "Error parsing URL",
			inputRef:  "s3://mybucket/path/to/file%xx",
			inputHash: "testhash",
			errRegexp: "failed to parse s3 reference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actual, err := ParseS3ObjectReference(tt.inputRef, tt.inputHash)
			if tt.errRegexp == "" {
				if !assert.NoError(t, err, "Should not return error") {
					return
				}
				assert.Equal(t, tt.expected, actual, "Parsed object should match expected")
			} else {
				if !assert.Error(t, err, "Should return an error") {
					return
				}
				assert.Regexp(t, tt.errRegexp, err.Error(), "Error message should match regexp")
			}
		})
	}
}

type mockS3Client struct {
	headFunc func(context.Context, *s3.HeadObjectInput, ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	getFunc  func(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

func (c *mockS3Client) HeadObject(ctx context.Context, obj *s3.HeadObjectInput, opts ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	return c.headFunc(ctx, obj, opts...)
}

func (c *mockS3Client) GetObject(ctx context.Context, obj *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return c.getFunc(ctx, obj, opts...)
}

func TestVerifyObjectExists(t *testing.T) {
	testRef := &S3ObjectReference{
		Bucket:  "test-bucket",
		Key:     "test-key",
		Hash:    "test-hash",
		Version: "test-version",
	}
	t.Run("Test error when HEADing object", func(t *testing.T) {
		client := &mockS3Client{
			headFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
				return nil, fmt.Errorf("test error")
			},
		}
		err := VerifyObjectExists(t.Context(), client, testRef)
		if !assert.Error(t, err, "should have returned an error") {
			return
		}
		assert.Regexp(t, "failed to HEAD object in S3 bucket.*test error", err.Error(), "Should include upstream error")
	})
	t.Run("Requires matching ETag from remote", func(t *testing.T) {
		client := &mockS3Client{
			headFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
				return &s3.HeadObjectOutput{
					ETag: stringPtrOrNil("not-test-hash"),
				}, nil
			},
		}
		err := VerifyObjectExists(t.Context(), client, testRef)
		if !assert.Error(t, err, "should have returned an error") {
			return
		}
		assert.Regexp(t, "object in s3 bucket does not match hash", err.Error(), "Should verify ETag in response")
	})
	t.Run("Handles missing ETag gracefully", func(t *testing.T) {
		client := &mockS3Client{
			headFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
				return &s3.HeadObjectOutput{}, nil
			},
		}
		err := VerifyObjectExists(t.Context(), client, testRef)
		if !assert.Error(t, err, "should have returned an error") {
			return
		}
		assert.Regexp(t, "missing ETag on object", err.Error(), "Should show useful message when missing ETag")
	})
}
