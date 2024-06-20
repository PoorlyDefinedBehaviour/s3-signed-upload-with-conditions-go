## About

The AWS Go SDK does not support custom policies which makes it not possible to limit the file size for presigned uploads to [S3](https://aws.amazon.com/s3/).

This library can be used to create a S3 presigned url with conditions (policy) and use it to upload objects to S3.

## Usage

```go
func TestPresignUpload(t *testing.T) {
	t.Parallel()

	input := Input{
		Region:          os.Getenv("AWS_REGION"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Bucket:          os.Getenv("AWS_BUCKET"),
		StartsWithKey:   "key_4",
		Expiration:      1 * time.Minute,
	}

	filePath := "./test_file.jpeg"

	presignedUpload, err := PresignUpload(input)
	require.NoError(t, err)

	buffer := bytes.NewBuffer([]byte{})
	writer := multipart.NewWriter(buffer)
	defer writer.Close()

	formField, err := writer.CreateFormField("policy")
	require.NoError(t, err)
	_, err = formField.Write([]byte(presignedUpload.ConditionsBase64))
	require.NoError(t, err)

	formField, err = writer.CreateFormField("X-Amz-Credential")
	require.NoError(t, err)
	_, err = formField.Write([]byte(presignedUpload.AmzCredential))
	require.NoError(t, err)

	formField, err = writer.CreateFormField("X-Amz-Signature")
	require.NoError(t, err)
	_, err = formField.Write([]byte(presignedUpload.AmzSignature))
	require.NoError(t, err)

	formField, err = writer.CreateFormField("X-Amz-Date")
	require.NoError(t, err)
	_, err = formField.Write([]byte(presignedUpload.AmzDate))
	require.NoError(t, err)

	formField, err = writer.CreateFormField("X-Amz-Algorithm")
	require.NoError(t, err)
	_, err = formField.Write([]byte(presignedUpload.AmzAlgorithm))
	require.NoError(t, err)

	formField, err = writer.CreateFormField("acl")
	require.NoError(t, err)
	_, err = formField.Write([]byte(presignedUpload.ACL))
	require.NoError(t, err)

	formField, err = writer.CreateFormField("key")
	require.NoError(t, err)
	_, err = formField.Write([]byte(input.StartsWithKey))
	require.NoError(t, err)

	formField, err = writer.CreateFormFile("file", filePath)
	require.NoError(t, err)

	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	_, err = io.Copy(formField, file)
	require.NoError(t, err)

	// Required to insert terminating boundary.
	require.NoError(t, writer.Close())

	request, err := http.NewRequest(http.MethodPost, presignedUpload.URL, buffer)
	require.NoError(t, err)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	assert.Equal(t, http.StatusNoContent, response.StatusCode)
}
```
