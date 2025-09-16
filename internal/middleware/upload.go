package middleware

import (
	"fmt"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func FileUpload() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userId := fmt.Sprintf("%v", c.Locals(constant.UserId))

		file, err := c.FormFile("image")
		if err != nil {
			return apperror.BadRequest(err.Error())
		}

		validExtensions := []string{".jpg", ".jpeg", ".png"}
		ext := filepath.Ext(file.Filename)
		valid := false
		for _, allowedExt := range validExtensions {
			if ext == allowedExt {
				valid = true
				break
			}
		}

		if !valid {
			return apperror.BadRequest(constant.InvalidImageFile)
		}

		const maxSize = 200 * 1024 // 200kb
		if file.Size >= maxSize {
			return apperror.BadRequest(constant.FileSizeIsTooLarge)
		}

		pattern := fmt.Sprintf("./storage/uploads/profile/%s_*", userId)
		oldFiles, _ := filepath.Glob(pattern)
		for _, oldFile := range oldFiles {
			_ = os.Remove(oldFile)
		}

		id := uuid.NewString()
		filename := fmt.Sprintf("%s_%s%s", userId, id, ext)
		filePath := fmt.Sprintf("./storage/uploads/profile/%s", filename)

		if err := c.SaveFile(file, filePath); err != nil {
			return apperror.Internal(constant.FailedToUploadImage, err)
		}

		c.Locals("filename", filename)

		return c.Next()
	}
}

func DocUpload() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// get the file upload and type information
		file, err := c.FormFile("file")
		tempType := c.FormValue("tempType")

		if err != nil {
			statusCode, resp := helper.GetError(err.Error())
			return c.Status(statusCode).JSON(resp)
		}

		validExtensions := []string{".csv"}
		ext := filepath.Ext(file.Filename)
		valid := false
		for _, allowedExt := range validExtensions {
			if ext == allowedExt {
				valid = true
				break
			}
		}

		if !valid {
			statusCode, resp := helper.GetError(constant.InvalidDocumentFile)
			return c.Status(statusCode).JSON(resp)
		}

		c.Locals("tempType", tempType)

		return c.Next()
	}
}

func UploadCSVFile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			statusCode, resp := helper.GetError(err.Error())
			return c.Status(statusCode).JSON(resp)
		}

		validExtensions := []string{".csv"}
		ext := filepath.Ext(file.Filename)
		valid := false
		for _, allowedExt := range validExtensions {
			if ext == allowedExt {
				valid = true
				break
			}
		}

		if !valid {
			statusCode, resp := helper.GetError(constant.InvalidDocumentFile)
			return c.Status(statusCode).JSON(resp)
		}

		return c.Next()
	}
}
