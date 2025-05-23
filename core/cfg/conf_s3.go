// Copyright 2019 Databricks
// Copyright 2021 Yandex LLC
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

package cfg

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/tigrisdata/tigrisfs/log"
)

type S3Config struct {
	Profile         string
	SharedConfig    []string
	AccessKey       string
	SecretKey       string
	RoleArn         string
	RoleExternalId  string
	RoleSessionName string
	StsEndpoint     string

	SDKMaxRetries       int
	SDKMinRetryDelay    time.Duration
	SDKMaxRetryDelay    time.Duration
	SDKMinThrottleDelay time.Duration
	SDKMaxThrottleDelay time.Duration

	RequesterPays bool
	Region        string
	RegionSet     bool
	ProjectId     string

	StorageClass string
	ColdMinSize  uint64
	MultipartAge time.Duration

	NoExpireMultipart bool

	MultipartCopyThreshold uint64

	NoDetect   bool
	UseSSE     bool
	UseKMS     bool
	KMSKeyID   string
	SseC       string
	SseCDigest string
	ACL        string
	NoChecksum bool
	ListV2     bool

	EnableSpecials bool

	Subdomain bool

	UseIAM    bool
	IAMFlavor string
	IAMUrl    string
	IAMHeader string

	Credentials *credentials.Credentials
	Session     *session.Session

	BucketOwner string
}

var (
	s3Session *session.Session
	cfgLog    = log.GetLogger("config")
)

func (c *S3Config) Init() *S3Config {
	if c.Region == "" {
		c.Region = "us-east-1"
	}
	if c.StorageClass == "" {
		c.StorageClass = "STANDARD"
	}
	if c.IAMHeader == "" {
		c.IAMHeader = "X-YaCloud-SubjectToken"
	}
	if c.IAMFlavor == "" {
		c.IAMFlavor = "gcp"
	}
	if c.MultipartCopyThreshold == 0 {
		c.MultipartCopyThreshold = 128
	}
	return c
}

func (c *S3Config) logSourceOfCredentials(flags *FlagStorage) {
	if c.AccessKey != "" {
		cfgLog.Info().Msgf("internal config: %v", c.AccessKey)
	} else if c.Profile != "" {
		cfgLog.Info().Msgf("command line profile: %v", c.Profile)
	} else if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		cfgLog.Info().Msgf("env AWS_ACCESS_KEY_ID = %v", os.Getenv("AWS_ACCESS_KEY_ID"))
	} else if os.Getenv("AWS_PROFILE") != "" {
		cfgLog.Info().Msgf("env AWS_PROFILE = %v", os.Getenv("AWS_PROFILE"))
	}
	if flags.Endpoint != "" {
		cfgLog.Info().Msgf("command line endpoint: %v", flags.Endpoint)
	} else if os.Getenv("AWS_ENDPOINT_URL") != "" {
		cfgLog.Info().Msgf("env AWS_ENDPOINT_URL = %v", os.Getenv("AWS_ENDPOINT_URL"))
	} else if DefaultEndpoint != "" {
		cfgLog.Info().Msgf("default endpoint = %v", DefaultEndpoint)
	}
}

func (c *S3Config) ToAwsConfig(flags *FlagStorage) (*aws.Config, error) {
	c.logSourceOfCredentials(flags)

	tr := &defaultHTTPTransport
	if flags.NoVerifySSL {
		if tr.TLSClientConfig != nil {
			tr.TLSClientConfig.InsecureSkipVerify = true
		} else {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}
	awsConfig := (&aws.Config{
		Region: &c.Region,
		Logger: log.GetLogger("s3"),
	}).WithHTTPClient(&http.Client{
		Transport: tr,
		Timeout:   flags.HTTPTimeout,
	})
	if flags.DebugS3 {
		awsConfig.LogLevel = aws.LogLevel(aws.LogDebug | aws.LogDebugWithRequestErrors)
	}
	if c.NoChecksum {
		awsConfig.S3DisableContentMD5Validation = aws.Bool(true)
	}

	if c.Credentials == nil {
		if c.AccessKey != "" {
			c.Credentials = credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, "")
		}
	}
	if flags.Endpoint != "" {
		awsConfig.Endpoint = &flags.Endpoint
	} else if os.Getenv("AWS_ENDPOINT_URL") != "" {
		awsConfig.Endpoint = aws.String(os.Getenv("AWS_ENDPOINT_URL"))
	} else {
		awsConfig.Endpoint = &DefaultEndpoint
	}

	awsConfig.S3ForcePathStyle = aws.Bool(!c.Subdomain)

	awsConfig.Retryer = client.DefaultRetryer{
		NumMaxRetries:    c.SDKMaxRetries,
		MinRetryDelay:    c.SDKMinRetryDelay,
		MaxRetryDelay:    c.SDKMaxRetryDelay,
		MinThrottleDelay: c.SDKMinThrottleDelay,
		MaxThrottleDelay: c.SDKMaxThrottleDelay,
	}

	if c.Session == nil {
		if s3Session == nil {
			var err error
			cfg := c.SharedConfig
			if len(cfg) == 0 {
				// aws-sdk doesn't ignore empty slices
				cfg = nil
			}
			s3Session, err = session.NewSessionWithOptions(session.Options{
				Profile:           c.Profile,
				SharedConfigFiles: cfg,
				SharedConfigState: session.SharedConfigEnable,
			})
			if err != nil {
				return nil, err
			}
		}
		c.Session = s3Session
	}

	if c.RoleArn != "" {
		c.Credentials = stscreds.NewCredentials(stsConfigProvider{c}, c.RoleArn,
			func(p *stscreds.AssumeRoleProvider) {
				if c.RoleExternalId != "" {
					p.ExternalID = &c.RoleExternalId
				}
				p.RoleSessionName = c.RoleSessionName
			})
	}

	if c.Credentials != nil {
		awsConfig.Credentials = c.Credentials
	}

	if c.SseC != "" {
		key, err := base64.StdEncoding.DecodeString(c.SseC)
		if err != nil {
			return nil, fmt.Errorf("sse-c is not base64-encoded: %v", err)
		}

		c.SseC = string(key)
		m := md5.Sum(key)
		c.SseCDigest = base64.StdEncoding.EncodeToString(m[:])
	}

	return awsConfig, nil
}

type stsConfigProvider struct {
	*S3Config
}

func (c stsConfigProvider) ClientConfig(serviceName string, cfgs ...*aws.Config) client.Config {
	config := c.Session.ClientConfig(serviceName, cfgs...)
	if c.Credentials != nil {
		config.Config.Credentials = c.Credentials
	}
	if c.StsEndpoint != "" {
		config.Endpoint = c.StsEndpoint
	}

	return config
}
