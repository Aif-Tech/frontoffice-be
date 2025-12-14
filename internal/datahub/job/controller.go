package job

import (
	"bytes"
	"errors"
	"fmt"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func NewController(svc Service) Controller {
	return &controller{Svc: svc}
}

type controller struct {
	Svc Service
}

type Controller interface {
	GetJobs(c *fiber.Ctx) error
	GetGenRetailJobs(c *fiber.Ctx) error
	GetJobDetails(c *fiber.Ctx) error
	ExportJobDetails(c *fiber.Ctx) error
	GetJobDetailsByDateRange(c *fiber.Ctx) error
	ExportJobDetailsByDateRange(c *fiber.Ctx) error
}

func (ctrl *controller) GetJobs(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	slug := c.Params("product_slug")

	productSlug, err := mapProductSlug(slug)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	filter := &logFilter{
		Page:        c.Query(constant.Page, "1"),
		Size:        c.Query(constant.Size, "10"),
		StartDate:   c.Query(constant.StartDate, ""),
		EndDate:     c.Query(constant.EndDate, ""),
		ProductSlug: productSlug,
		AuthCtx:     authCtx,
	}

	result, err := ctrl.Svc.GetJobs(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (ctrl *controller) GetGenRetailJobs(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	filter := &logFilter{
		Page:        c.Query(constant.Page, "1"),
		Size:        c.Query(constant.Size, "10"),
		StartDate:   c.Query(constant.StartDate, ""),
		EndDate:     c.Query(constant.EndDate, ""),
		ProductSlug: constant.SlugGenRetailV3,
		AuthCtx:     authCtx,
	}

	result, err := ctrl.Svc.GetGenRetailJobs(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (ctrl *controller) GetJobDetails(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	slug := c.Params("product_slug")

	productSlug, err := mapProductSlug(slug)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	filter := &logFilter{
		AuthCtx:     authCtx,
		Page:        c.Query(constant.Page, ""),
		Size:        c.Query(constant.Size, ""),
		Keyword:     c.Query("keyword"),
		JobId:       c.Params("job_id"),
		ProductSlug: productSlug,
	}

	result, err := ctrl.Svc.GetJobDetails(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (ctrl *controller) GetJobDetailsByDateRange(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	slug := c.Params("product_slug")

	productSlug, err := mapProductSlug(slug)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	startDate := c.Query(constant.StartDate)
	endDate := c.Query(constant.EndDate)
	if startDate != "" && endDate == "" {
		endDate = startDate
	}

	filter := &logFilter{
		AuthCtx:     authCtx,
		Page:        c.Query(constant.Page, "1"),
		Size:        c.Query(constant.Size, "10"),
		Keyword:     c.Query("keyword"),
		ProductSlug: productSlug,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	result, err := ctrl.Svc.GetJobDetailsByDateRange(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (ctrl *controller) ExportJobDetails(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	slug := c.Params("product_slug")
	masked, _ := strconv.ParseBool(c.Query("masked"))

	productSlug, err := mapProductSlug(slug)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	filter := &logFilter{
		AuthCtx:     authCtx,
		ProductSlug: productSlug,
		JobId:       c.Params("job_id"),
		Size:        constant.SizeUnlimited,
		IsMasked:    masked,
	}

	var buf bytes.Buffer
	filename, err := ctrl.Svc.ExportJobDetails(filter, &buf)
	if err != nil {
		return err
	}

	c.Set(constant.HeaderContentType, constant.TextOrCSVContentType)
	c.Set(constant.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))

	return c.SendStream(bytes.NewReader(buf.Bytes()))
}

func (ctrl *controller) ExportJobDetailsByDateRange(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	slug := c.Params("product_slug")
	masked, _ := strconv.ParseBool(c.Query("masked"))

	productSlug, err := mapProductSlug(slug)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	startDate := c.Query(constant.StartDate)
	endDate := c.Query(constant.EndDate)
	if startDate != "" && endDate == "" {
		endDate = startDate
	}

	filter := &logFilter{
		AuthCtx:     authCtx,
		ProductSlug: productSlug,
		StartDate:   startDate,
		EndDate:     endDate,
		Size:        constant.SizeUnlimited,
		IsMasked:    masked,
	}

	var buf bytes.Buffer
	filename, err := ctrl.Svc.ExportJobDetailsByDateRange(filter, &buf)
	if err != nil {
		return err
	}

	c.Set(constant.HeaderContentType, constant.TextOrCSVContentType)
	c.Set(constant.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))

	return c.SendStream(bytes.NewReader(buf.Bytes()))
}

var productSlugMap = map[string]string{
	"loan-record-checker":     constant.SlugLoanRecordChecker,
	"7d-multiple-loan":        constant.Slug7DaysMultipleLoan,
	"30d-multiple-loan":       constant.Slug30DaysMultipleLoan,
	"90d-multiple-loan":       constant.Slug90DaysMultipleLoan,
	"tax-compliance-status":   constant.SlugTaxComplianceStatus,
	"tax-score":               constant.SlugTaxScore,
	"tax-verification-detail": constant.SlugTaxVerificationDetail,
	"npwp-verification":       constant.SlugNPWPVerification,
	"gen-retail":              constant.SlugGenRetailV3,
	"recycle-number":          constant.SlugRecycleNumber,
}

func mapProductSlug(slug string) (string, error) {
	if mapped, ok := productSlugMap[slug]; ok {
		return mapped, nil
	}

	return "", errors.New("unsupported product slug")
}
