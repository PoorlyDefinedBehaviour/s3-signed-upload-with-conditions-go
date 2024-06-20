package s3signedupload

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

const amzAlgorithm = "AWS4-HMAC-SHA256"

type PresignedUpload struct {
	URL              string
	ConditionsBase64 string
	AmzSignature     string
	AmzDate          string
	AmzCredential    string
	AmzAlgorithm     string
	ACL              string
}

type Input struct {
	Region           string
	AccessKeyID      string
	SecretAccessKey  string
	Bucket           string
	StartsWithKey    string
	MaxFileSizeBytes uint64
	Expiration       time.Duration
}

func PresignUpload(input Input) (PresignedUpload, error) {
	if input.Region == "" {
		return PresignedUpload{}, fmt.Errorf("region is required")
	}
	if input.AccessKeyID == "" {
		return PresignedUpload{}, fmt.Errorf("access key id is required")
	}
	if input.SecretAccessKey == "" {
		return PresignedUpload{}, fmt.Errorf("secret access key is required")
	}
	if input.Bucket == "" {
		return PresignedUpload{}, fmt.Errorf("bucket is required")
	}
	if input.StartsWithKey == "" {
		return PresignedUpload{}, fmt.Errorf("starts with key is required")
	}
	if input.MaxFileSizeBytes == 0 {
		return PresignedUpload{}, fmt.Errorf("max file size is required")
	}
	if input.Expiration == 0 {
		return PresignedUpload{}, fmt.Errorf("expiration is required")
	}
	expiresAt := time.Now().Add(input.Expiration).UTC()

	expiresAtRFC3339 := expiresAt.Format("2006-01-02T15:04:05.000Z")
	dateYYYYMMDD := expiresAt.Format("20060102")
	dateIso8601 := expiresAt.Format("20060102T150405000Z")

	amzCredential := fmt.Sprintf("%s/%s/%s/s3/aws4_request",
		input.AccessKeyID,
		dateYYYYMMDD,
		input.Region,
	)

	acl := "private"

	conditions := map[string]any{
		"expiration": expiresAtRFC3339,
		"conditions": []any{
			map[string]string{"bucket": input.Bucket},
			[]string{"starts-with", "$key", input.StartsWithKey},
			map[string]string{"acl": acl},
			map[string]string{"x-amz-credential": amzCredential},
			map[string]string{"x-amz-algorithm": amzAlgorithm},
			map[string]string{"x-amz-date": dateIso8601},
			[]any{"content-length-range", 0, input.MaxFileSizeBytes},
		},
	}

	conditionsBytes, err := json.Marshal(conditions)
	if err != nil {
		return PresignedUpload{}, fmt.Errorf("json marshalling conditions: %w", err)
	}

	conditionsBase64 := base64.StdEncoding.EncodeToString(conditionsBytes)

	dateHash := hmacHash([]byte(fmt.Sprintf("AWS4%s", input.SecretAccessKey)), []byte(dateYYYYMMDD))
	regionHash := hmacHash(dateHash, []byte(input.Region))
	serviceHash := hmacHash(regionHash, []byte("s3"))
	signingHash := hmacHash(serviceHash, []byte("aws4_request"))
	signatureHash := hmacHash(signingHash, []byte(conditionsBase64))

	signature := hex.EncodeToString(signatureHash)

	return PresignedUpload{
		URL:              fmt.Sprintf("https://%s.s3.amazonaws.com", input.Bucket),
		ConditionsBase64: conditionsBase64,
		AmzSignature:     signature,
		AmzDate:          dateIso8601,
		AmzCredential:    amzCredential,
		AmzAlgorithm:     amzAlgorithm,
		ACL:              acl,
	}, nil
}

func hmacHash(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	// Write never returns error in this case.
	_, _ = hash.Write(data)
	return hash.Sum(nil)
}
