package helper

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
)

func ValidateUploadedFile(file *multipart.FileHeader, maxSize int64, allowedExtensions []string) error {
	if file.Size > maxSize {
		return fmt.Errorf("file too large (max %dMB)", maxSize/1024/1024)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return nil
		}
	}

	return fmt.Errorf("invalid file type, allowed: %v", allowedExtensions)
}
