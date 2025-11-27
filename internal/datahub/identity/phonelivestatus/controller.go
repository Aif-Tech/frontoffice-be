package phonelivestatus

import (
	"bytes"
	"fmt"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"mime/multipart"
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
	reqBody, ok := c.Locals(constant.Request).(*phoneLiveStatusRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	if err := ctrl.svc.PhoneLiveStatus(authCtx, reqBody); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"success",
		nil,
	))
}

func (ctrl *controller) BulkSearch(c *fiber.Ctx) error {
	file, ok := c.Locals(constant.ValidatedFile).(*multipart.FileHeader)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	if err := ctrl.svc.BulkPhoneLiveStatus(authCtx, file); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"success",
		nil,
	))
}

func (ctrl *controller) GetJobs(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	filter := &phoneLiveStatusFilter{
		Page:        c.Query(constant.Page, "1"),
		Size:        c.Query(constant.Size, "10"),
		StartDate:   c.Query(constant.StartDate, ""),
		EndDate:     c.Query(constant.EndDate, ""),
		ProductSlug: constant.SlugPhoneLiveStatus,
		MemberId:    authCtx.UserIdStr(),
		CompanyId:   authCtx.CompanyIdStr(),
		TierLevel:   authCtx.RoleIdStr(),
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
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	filter := &phoneLiveStatusFilter{
		Page:        c.Query(constant.Page, "1"),
		Size:        c.Query(constant.Size, "10"),
		Keyword:     c.Query(constant.Keyword),
		JobId:       c.Params("id"),
		ProductSlug: constant.SlugPhoneLiveStatus,
		MemberId:    authCtx.UserIdStr(),
		CompanyId:   authCtx.CompanyIdStr(),
		TierLevel:   authCtx.RoleIdStr(),
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
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	masked, _ := strconv.ParseBool(c.Query("masked"))
	filter := &phoneLiveStatusFilter{
		JobId:       c.Params("id"),
		ProductSlug: constant.SlugPhoneLiveStatus,
		StartDate:   c.Query(constant.StartDate, ""),
		EndDate:     c.Query(constant.EndDate, ""),
		MemberId:    authCtx.UserIdStr(),
		CompanyId:   authCtx.CompanyIdStr(),
		TierLevel:   authCtx.RoleIdStr(),
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
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	filter := &phoneLiveStatusFilter{
		ProductSlug: constant.SlugPhoneLiveStatus,
		StartDate:   c.Query(constant.StartDate, ""),
		EndDate:     c.Query(constant.EndDate, ""),
		MemberId:    authCtx.UserIdStr(),
		CompanyId:   authCtx.CompanyIdStr(),
		TierLevel:   authCtx.RoleIdStr(),
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
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	masked, _ := strconv.ParseBool(c.Query("masked"))
	filter := &phoneLiveStatusFilter{
		ProductSlug: constant.SlugPhoneLiveStatus,
		StartDate:   c.Query(constant.StartDate, ""),
		EndDate:     c.Query(constant.EndDate, ""),
		MemberId:    authCtx.UserIdStr(),
		CompanyId:   authCtx.CompanyIdStr(),
		TierLevel:   authCtx.RoleIdStr(),
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
