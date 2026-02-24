package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	mc        *minio.Client
	bucket    string
	publicURL string
}

func NewMinIOClient(endpoint, accessKey, secretKey, bucket, publicURL string, useSSL bool) (*Client, error) {
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Client{mc: mc, bucket: bucket, publicURL: publicURL}, nil
}

// EnsureBucket creates the bucket if it doesn't exist and sets a public-read
// policy on the avatars/ prefix.
func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := c.mc.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}

	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect":    "Allow",
				"Principal": map[string]any{"AWS": []string{"*"}},
				"Action":    []string{"s3:GetObject"},
				"Resource":  []string{fmt.Sprintf("arn:aws:s3:::%s/avatars/*", c.bucket)},
			},
		},
	}
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	return c.mc.SetBucketPolicy(ctx, c.bucket, string(policyJSON))
}

// UploadAvatar uploads an image file for the given user and returns the public URL.
// The key is avatars/{userID}.{ext}, overwriting any previous avatar.
func (c *Client) UploadAvatar(ctx context.Context, userID string, r io.Reader) (string, error) {
	// Sniff the first 512 bytes for content-type detection.
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	buf = buf[:n]
	contentType := http.DetectContentType(buf)

	ext, ok := imageExtension(contentType)
	if !ok {
		return "", fmt.Errorf("unsupported image type: %s", contentType)
	}

	// Reconstruct the full stream with the sniffed bytes prepended.
	full := io.MultiReader(bytes.NewReader(buf), r)
	key := fmt.Sprintf("avatars/%s%s", userID, ext)

	_, err = c.mc.PutObject(ctx, c.bucket, key, full, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", c.publicURL, c.bucket, key), nil
}

func imageExtension(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}
