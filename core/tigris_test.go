// Copyright 2024 Tigris Data, Inc.
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

package core

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	s4 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigrisfs/core/cfg"
)

var (
	triedDetect bool
	detected    bool
	localTigris bool
)

func tigrisDetected(flags *cfg.FlagStorage) (bool, bool) {
	if triedDetect {
		return detected, localTigris
	}

	endpoint := flags.Endpoint
	if endpoint == "" {
		endpoint = os.Getenv("AWS_ENDPOINT_URL")
	}

	local := strings.Contains(endpoint, "localhost") || strings.Contains(endpoint, "127.0.0.1")

	r, err := http.Get(endpoint + "/")
	if err != nil {
		return false, local
	}

	triedDetect = true
	localTigris = local
	detected = r.StatusCode == http.StatusOK && strings.Contains(r.Header.Get("Server"), "Tigris")

	return detected, local
}

func LocalTigrisDetected(flags *cfg.FlagStorage) bool {
	t, local := tigrisDetected(flags)
	return t && local
}

func TigrisDetected(flags *cfg.FlagStorage) bool {
	t, _ := tigrisDetected(flags)
	return t
}

func TestListIncludeMetadataAndContent(t *testing.T) {
	flags := cfg.DefaultFlags()

	conf := selectTestConfig(flags)
	conf.EnableSpecials = true
	flags.Backend = &conf
	flags.TigrisListContent = true

	if !TigrisDetected(flags) {
		t.Skip("Tigris not detected")
	}

	bucket := fmt.Sprintf("test-metadata-bucket-1-%x", rand.Int63())

	s3, err := NewS3(bucket, flags, &conf)
	require.NoError(t, err)

	_, err = s3.CreateBucket(&s4.CreateBucketInput{Bucket: aws.String(bucket)})
	require.NoError(t, err)

	blobs := []struct {
		name     string
		content  []byte
		metadata map[string]*string
	}{
		{"blob1", []byte("content of blob1"), nil},
		{"blob2", make([]byte, 16384+2048), map[string]*string{"key2": aws.String("value2"), "key5": aws.String("value5")}},
		{"blob3", []byte("content of blob3"), map[string]*string{"key3": aws.String("value3")}},
	}

	for _, blob := range blobs {
		input := &PutBlobInput{
			Key:      blob.name,
			Body:     bytes.NewReader(blob.content),
			Metadata: blob.metadata,
		}
		_, err := s3.PutBlob(input)
		require.NoError(t, err)
	}

	input := &ListBlobsInput{}
	listedBlobs, err := s3.ListBlobs(input)
	require.NoError(t, err)

	require.Equal(t, len(blobs), len(listedBlobs.Items))
	for i, blob := range blobs {
		require.Equal(t, blob.name, *listedBlobs.Items[i].Key, i)
		require.Equal(t, blob.metadata, listedBlobs.Items[i].Metadata, i)
		if len(blob.content) < 1024 {
			require.Equal(t, blob.content, listedBlobs.Items[i].Content, i)
		} else {
			require.Nil(t, listedBlobs.Items[i].Content, i)
		}
	}
}
