package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
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

var (
	errFileTooLarge = errors.New("file too large (max 5MB)")
	errBadImageType = errors.New("only jpg/png/webp allowed")
)

// validateAndReadImage reads r fully (bounded), then validates size and the
// server-sniffed content type. Returns the bytes, detected MIME type, and ext.
func validateAndReadImage(r io.Reader, declaredSize int64) (data []byte, contentType, ext string, err error) {
	if declaredSize > maxUploadSize {
		return nil, "", "", errFileTooLarge
	}
	// Read up to one byte past the limit so an understated declaredSize can't slip through.
	data, err = io.ReadAll(io.LimitReader(r, maxUploadSize+1))
	if err != nil {
		return nil, "", "", err
	}
	if int64(len(data)) > maxUploadSize {
		return nil, "", "", errFileTooLarge
	}
	contentType = http.DetectContentType(data) // tolerates inputs shorter than 512 bytes
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return nil, "", "", errBadImageType
	}
	return data, contentType, ext, nil
}

// UploadImage accepts a multipart "file" field, stores it in Cloudflare R2,
// and returns the public URL. Used for merchant product images.
func UploadImage(c *gin.Context) {
	if config.R2Client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "upload not configured"})
		return
	}

	// Hard-bound the request body regardless of the declared part size.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize+8192)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read file"})
		return
	}
	defer f.Close()

	data, contentType, ext, err := validateAndReadImage(f, fileHeader.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := "products/" + randomHex(16) + ext
	_, err = config.R2Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(config.AppConfig.R2.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
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
