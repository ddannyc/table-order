package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/example/table-order/config"
	"github.com/gin-gonic/gin"
)

const maxUploadSize = 5 << 20 // 5MB

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// UploadImage accepts a multipart "file" field, stores it in Cloudflare R2,
// and returns the public URL. Used for merchant product images.
func UploadImage(c *gin.Context) {
	if config.R2Client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "upload not configured"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	if fileHeader.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 5MB)"})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read file"})
		return
	}
	defer f.Close()

	// Sniff content type from the first 512 bytes (don't trust the client header)
	head := make([]byte, 512)
	n, _ := f.Read(head)
	contentType := http.DetectContentType(head[:n])
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only jpg/png/webp allowed"})
		return
	}

	// Reassemble the full file (head + remainder) for upload
	rest := new(bytes.Buffer)
	if _, err := rest.ReadFrom(f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read file failed"})
		return
	}
	body := bytes.NewReader(append(head[:n], rest.Bytes()...))

	key := "products/" + randomHex(16) + ext
	_, err = config.R2Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(config.AppConfig.R2.Bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}

	url := strings.TrimRight(config.AppConfig.R2.PublicBase, "/") + "/" + key
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func randomHex(nBytes int) string {
	b := make([]byte, nBytes)
	rand.Read(b)
	return hex.EncodeToString(b)
}
