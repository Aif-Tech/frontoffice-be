package genretail

import (
	"bytes"
	"fmt"
	"front-office/pkg/apperror"
	"front-office/pkg/helper"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"front-office/pkg/common/constant"
)

func NewController(
	service Service,
) Controller {
	return &controller{
		service,
	}
}

type controller struct {
	service Service
}

type Controller interface {
	DummyRequestScore(c *fiber.Ctx) error
	SingleRequest(c *fiber.Ctx) error
	BulkRequest(c *fiber.Ctx) error
	GetLogsScoreezy(c *fiber.Ctx) error
	GetLogScoreezy(c *fiber.Ctx) error
	ExportJobDetails(c *fiber.Ctx) error
	// DownloadCSV(c *fiber.Ctx) error
	// UploadCSV(c *fiber.Ctx) error
	// GetBulkSearch(c *fiber.Ctx) error
}

func (ctrl *controller) DummyRequestScore(c *fiber.Ctx) error {
	response := genRetailV3ClientReturnSuccess{
		Message: "Succeed to Request Scores",
		Success: true,
		Data: &dataGenRetailV3{
			TransactionId:        "TRX123456789",
			Name:                 "John Doe",
			IdCardNo:             "1234567890123456",
			PhoneNo:              "081234567890",
			LoanNo:               "LN987654321",
			ProbabilityToDefault: 0.12345,
			Grade:                "A",
			Identity:             "Verified in more than 50% social media platform and registered on one of the telecommunication platforms",
			Behavior:             "This individual is not identified to have a history of loan applications and is not indicated to have defaulted on payments.",
			Date:                 time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	return c.Status(200).JSON(response)
}

func (ctrl *controller) SingleRequest(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*genRetailRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	result, err := ctrl.service.GenRetailV3(authCtx.UserId, authCtx.CompanyId, reqBody)
	if err != nil {
		return err
	}

	return c.Status(result.StatusCode).JSON(result)
}

func (ctrl *controller) BulkRequest(c *fiber.Ctx) error {
	file, ok := c.Locals(constant.ValidatedFile).(*multipart.FileHeader)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	jobId, err := ctrl.service.BulkGenRetailV3(authCtx.UserId, authCtx.CompanyId, authCtx.QuotaTypeStr(), file)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		constant.Success,
		fiber.Map{"job_id": jobId},
	))
}

func (ctrl *controller) GetLogsScoreezy(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	filter := &filterLogs{
		CompanyId:   authCtx.CompanyIdStr(),
		JobId:       c.Query(constant.JobId),
		StartDate:   c.Query(constant.StartDate),
		EndDate:     c.Query(constant.EndDate),
		Name:        strings.ToLower(strings.TrimSpace(c.Query("name"))),
		ProductType: strings.ToLower(strings.TrimSpace(c.Query("product_type"))),
		Grade:       strings.ToLower(c.Query("grade")),
		Page:        c.Query(constant.Page),
		Size:        c.Query(constant.Size),
	}

	result, err := ctrl.service.GetLogsScoreezy(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		constant.Success,
		gradesResponseData{Logs: result.Data},
		result.Meta,
	))
}

func (ctrl *controller) GetLogScoreezy(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	filter := &filterLogs{
		CompanyId: authCtx.CompanyIdStr(),
		TrxId:     c.Params("trx_id"),
	}

	result, err := ctrl.service.GetLogScoreezy(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(constant.Success, result))
}

func (ctrl *controller) ExportJobDetails(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	masked, _ := strconv.ParseBool(c.Query(constant.Masked))

	filter := &filterLogs{
		Masked:    masked,
		CompanyId: authCtx.CompanyIdStr(),
		JobId:     c.Query(constant.JobId),
		StartDate: c.Query(constant.StartDate),
		EndDate:   c.Query(constant.EndDate),
		Size:      constant.SizeUnlimited,
	}

	var buf bytes.Buffer
	filename, err := ctrl.service.ExportJobDetails(filter, &buf)
	if err != nil {
		return err
	}

	c.Set(constant.HeaderContentType, constant.TextOrCSVContentType)
	c.Set(constant.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))

	return c.SendStream(bytes.NewReader(buf.Bytes()))
}
