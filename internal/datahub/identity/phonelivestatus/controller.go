package phonelivestatus

import (
	"bytes"
	"fmt"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func NewController(
	svc Service,
) Controller {
	return &controller{svc}
}

type controller struct {
	svc Service
}

type Controller interface {
	SingleSearch(c *fiber.Ctx) error
	BulkSearch(c *fiber.Ctx) error
	GetJobs(c *fiber.Ctx) error
	GetJobDetails(c *fiber.Ctx) error
	ExportJobDetails(c *fiber.Ctx) error
	GetJobsSummary(c *fiber.Ctx) error
	ExportJobsSummary(c *fiber.Ctx) error
}

func (ctrl *controller) SingleSearch(c *fiber.Ctx) error {
	reqBody := c.Locals(constant.Request).(*phoneLiveStatusRequest)

	apiKey := fmt.Sprintf("%v", c.Locals(constant.APIKey))
	memberId := fmt.Sprintf("%v", c.Locals(constant.UserId))
	companyId := fmt.Sprintf("%v", c.Locals(constant.CompanyId))

	err := ctrl.svc.PhoneLiveStatus(apiKey, memberId, companyId, reqBody)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"success",
		nil,
	))
}

func (ctrl *controller) BulkSearch(c *fiber.Ctx) error {
	apiKey := fmt.Sprintf("%v", c.Locals(constant.APIKey))
	memberId := fmt.Sprintf("%v", c.Locals(constant.UserId))
	companyId := fmt.Sprintf("%v", c.Locals(constant.CompanyId))
	quotaType := fmt.Sprintf("%v", c.Locals(constant.QuotaType))

	file, err := c.FormFile("file")
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	err = ctrl.svc.BulkPhoneLiveStatus(apiKey, memberId, companyId, quotaType, file)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"success",
		nil,
	))
}

func (ctrl *controller) GetJobs(c *fiber.Ctx) error {
	filter := &phoneLiveStatusFilter{
		Page:        c.Query(constant.Page, "1"),
		Size:        c.Query(constant.Size, "10"),
		StartDate:   c.Query(constant.StartDate, ""),
		EndDate:     c.Query(constant.EndDate, ""),
		ProductSlug: constant.SlugPhoneLiveStatus,
		MemberId:    fmt.Sprintf("%v", c.Locals(constant.UserId)),
		CompanyId:   fmt.Sprintf("%v", c.Locals(constant.CompanyId)),
		TierLevel:   fmt.Sprintf("%v", c.Locals(constant.RoleId)),
	}

	jobs, err := ctrl.svc.GetJobs(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		"succeeded to get phone live status jobs",
		jobs,
	))
}

func (ctrl *controller) GetJobDetails(c *fiber.Ctx) error {
	filter := &phoneLiveStatusFilter{
		Page:        c.Query(constant.Page, "1"),
		Size:        c.Query(constant.Size, "10"),
		Keyword:     c.Query(constant.Keyword),
		JobId:       c.Params("id"),
		ProductSlug: constant.SlugPhoneLiveStatus,
		MemberId:    fmt.Sprintf("%v", c.Locals(constant.UserId)),
		CompanyId:   fmt.Sprintf("%v", c.Locals(constant.CompanyId)),
		TierLevel:   fmt.Sprintf("%v", c.Locals(constant.RoleId)),
	}

	if filter.JobId == "" {
		return apperror.BadRequest("missing job ID")
	}

	jobDetail, err := ctrl.svc.GetJobDetails(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		"succeeded to get phone live status job details",
		jobDetail,
	))
}

func (ctrl *controller) ExportJobDetails(c *fiber.Ctx) error {
	masked, _ := strconv.ParseBool(c.Query("masked"))
	filter := &phoneLiveStatusFilter{
		JobId:       c.Params("id"),
		ProductSlug: constant.SlugPhoneLiveStatus,
		StartDate:   c.Query(constant.StartDate, ""),
		EndDate:     c.Query(constant.EndDate, ""),
		MemberId:    fmt.Sprintf("%v", c.Locals(constant.UserId)),
		CompanyId:   fmt.Sprintf("%v", c.Locals(constant.CompanyId)),
		TierLevel:   fmt.Sprintf("%v", c.Locals(constant.RoleId)),
		Size:        constant.SizeUnlimited,
		Masked:      masked,
	}

	var buf bytes.Buffer

	filename, err := ctrl.svc.ExportJobDetails(filter, &buf)
	if err != nil {
		return err
	}

	c.Set(constant.HeaderContentType, constant.TextOrCSVContentType)
	c.Set(constant.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))

	return c.SendStream(bytes.NewReader(buf.Bytes()))
}

func (ctrl *controller) GetJobsSummary(c *fiber.Ctx) error {
	filter := &phoneLiveStatusFilter{
		ProductSlug: constant.SlugPhoneLiveStatus,
		StartDate:   c.Query(constant.StartDate, ""),
		EndDate:     c.Query(constant.EndDate, ""),
		MemberId:    fmt.Sprintf("%v", c.Locals(constant.UserId)),
		CompanyId:   fmt.Sprintf("%v", c.Locals(constant.CompanyId)),
		TierLevel:   fmt.Sprintf("%v", c.Locals(constant.RoleId)),
		Size:        constant.SizeUnlimited,
	}

	if filter.StartDate == "" || filter.EndDate == "" {
		return apperror.BadRequest("start_date and end_date are required")
	}

	jobsSummary, err := ctrl.svc.GetJobsSummary(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		"succeeded to get phone live status jobs summary",
		jobsSummary,
	))
}

func (ctrl *controller) ExportJobsSummary(c *fiber.Ctx) error {
	masked, _ := strconv.ParseBool(c.Query("masked"))
	filter := &phoneLiveStatusFilter{
		ProductSlug: constant.SlugPhoneLiveStatus,
		StartDate:   c.Query(constant.StartDate, ""),
		EndDate:     c.Query(constant.EndDate, ""),
		MemberId:    fmt.Sprintf("%v", c.Locals(constant.UserId)),
		CompanyId:   fmt.Sprintf("%v", c.Locals(constant.CompanyId)),
		TierLevel:   fmt.Sprintf("%v", c.Locals(constant.RoleId)),
		Size:        constant.SizeUnlimited,
		Masked:      masked,
	}

	if filter.StartDate == "" || filter.EndDate == "" {
		return apperror.BadRequest("start_date and end_date are required")
	}

	var buf bytes.Buffer
	filename, err := ctrl.svc.ExportJobsSummary(filter, &buf)
	if err != nil {
		return err
	}

	c.Set(constant.HeaderContentType, constant.TextOrCSVContentType)
	c.Set(constant.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))

	return c.SendStream(bytes.NewReader(buf.Bytes()))
}
