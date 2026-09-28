// Copyright 2026 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigrisfs/core/cfg"
)

// publicBucketTransport answers the detection HEAD the way a bucket that
// serves unauthenticated requests does.
type publicBucketTransport struct{}

func (publicBucketTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     http.Header{},
		Body:       http.NoBody,
	}, nil
}

// backendForDetection builds just enough of an S3Backend to run the startup
// probe against a canned response: no live endpoint, no real credentials.
func backendForDetection(t *testing.T, useIAM bool) *S3Backend {
	t.Helper()

	// The probe reads these two as evidence that anonymous access was intended.
	t.Setenv("AWS_ACCESS_KEY_ID", "")

	awsConfig := (&aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String("https://example.invalid"),
		Credentials: credentials.NewStaticCredentials("key", "secret", ""),
		HTTPClient:  &http.Client{Transport: publicBucketTransport{}},
	})

	sess, err := session.NewSession(awsConfig)
	require.NoError(t, err)

	return &S3Backend{
		S3:        s3.New(sess),
		bucket:    "some-public-bucket",
		awsConfig: awsConfig,
		flags:     &cfg.FlagStorage{},
		config:    &cfg.S3Config{UseIAM: useIAM},
	}
}

// A 200 from the unsigned probe only proves the bucket serves anonymous
// requests. Treating that as a request for anonymous access silently drops an
// IAM-configured mount to unauthenticated calls for the rest of its life.
func TestDetectAnonymousDoesNotOverrideIAM(t *testing.T) {
	t.Run("iam mount keeps its credentials", func(t *testing.T) {
		s := backendForDetection(t, true)

		_, _ = s.detectBucketLocationByHEAD()

		require.NotEqual(t, credentials.AnonymousCredentials, s.awsConfig.Credentials,
			"an --iam mount must not be switched to anonymous credentials")
	})

	t.Run("without iam the detection still applies", func(t *testing.T) {
		s := backendForDetection(t, false)

		_, _ = s.detectBucketLocationByHEAD()

		require.Equal(t, credentials.AnonymousCredentials, s.awsConfig.Credentials,
			"detection should still recognise an anonymous bucket when IAM is off")
	})
}
