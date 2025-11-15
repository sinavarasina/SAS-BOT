package uploader

// ImgbbUploader untuk gambar
type ImgbbUploader struct {
	APIKey string
}
func NewImgbbUploader(apiKey string) *ImgbbUploader {
	return &ImgbbUploader{APIKey: apiKey}
}

// R2Uploader untuk PDF
type R2Uploader struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
}
func NewR2Uploader(accountID, accessKey, secret, bucket, url string) *R2Uploader {
	return &R2Uploader{
		AccountID:       accountID,
		AccessKeyID:     accessKey,
		SecretAccessKey: secret,
		BucketName:      bucket,
		PublicURL:       url,
	}
}
