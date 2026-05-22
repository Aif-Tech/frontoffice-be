package billing

import (
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"front-office/pkg/helper"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func NewController(service Service) Controller {
	return &controller{
		svc: service,
	}
}

type controller struct {
	svc Service
}

type Controller interface {
	ExportUsage(c *fiber.Ctx) error
	SendMonthlyUsageReport(c *fiber.Ctx) error
	GetUsageReport(c *fiber.Ctx) error
}

func (ctrl *controller) ExportUsage(c *fiber.Ctx) error {
	var err error
	req, err := parseDownloadRequest(c)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	result, err := ctrl.svc.ExportUsageXlsx(downloadUsageXlsxInput{
		CompanyId:       req.CompanyId,
		Year:            req.Year,
		Month:           req.Month,
		Groups:          req.Groups,
		PricingStrategy: req.PricingStrategy,
		Password:        req.Password,
	})
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to generate report")

		return err
	}

	c.Set(constant.HeaderContentType, result.ContentType)
	c.Set(constant.HeaderContentDisposition, `attachment; filename="`+result.Filename+`"`)
	c.Set("Content-Length", strconv.Itoa(len(result.Data)))

	return c.Send(result.Data)
}

func (ctrl *controller) SendMonthlyUsageReport(c *fiber.Ctx) error {
	if err := ctrl.svc.SendMonthlyUsageReport(); err != nil {
		log.Error().
			Err(err).
			Msg("failed to send monthly usage report")
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"succeed to send monthly usage report",
		nil,
	))
}

func (ctrl *controller) GetUsageReport(c *fiber.Ctx) error {
	var err error
	req, err := parseDownloadRequest(c)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	result, err := ctrl.svc.GetUsageReport(req.CompanyId, req.PricingStrategy, req.Month, req.Year)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"succeed to get monthly usage report",
		result,
	))
}

func parseDownloadRequest(c *fiber.Ctx) (*downloadUsageXlsxRequest, error) {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return nil, apperror.Unauthorized(err.Error())
	}

	companyIdUint, err := parseAndValidateCompanyId(c, authCtx)
	if err != nil {
		return nil, err
	}

	year, month, err := parseYearMonth(c)
	if err != nil {
		return nil, err
	}

	groups := parseGroups(c.Query("groups"))
	pricingStrategy := strings.ToUpper(c.Query("pricing_strategy"))

	return &downloadUsageXlsxRequest{
		CompanyId:       uint(companyIdUint),
		Year:            year,
		Month:           month,
		Groups:          groups,
		PricingStrategy: pricingStrategy,
		Password:        authCtx.APIKey,
	}, nil
}

func parseAndValidateCompanyId(c *fiber.Ctx, authCtx *model.AuthContext) (uint64, error) {
	companyId := c.Query("company_id")
	if companyId == "0" {
		return 0, apperror.BadRequest("company_id is required")
	}

	if authCtx.RoleId != 0 && authCtx.CompanyIdStr() != companyId {
		return 0, apperror.Unauthorized(constant.RequestProhibited)
	}

	companyIdUint, err := strconv.ParseUint(companyId, 10, 64)
	if err != nil {
		return 0, apperror.BadRequest("company_id must be a valid number")
	}

	return companyIdUint, nil
}

func parseYearMonth(c *fiber.Ctx) (int, int, error) {
	now := time.Now()
	lastMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)

	year, month := lastMonth.Year(), int(lastMonth.Month())

	if y := c.Query("year"); y != "" {
		parsed, err := strconv.Atoi(y)
		if err != nil || parsed < 2000 || parsed > now.Year() {
			return 0, 0, apperror.BadRequest("year must be a valid 4-digit year")
		}
		year = parsed
	}

	if m := c.Query("month"); m != "" {
		parsed, err := strconv.Atoi(m)
		if err != nil || parsed < 1 || parsed > 12 {
			return 0, 0, apperror.BadRequest("month must be between 1 and 12")
		}
		month = parsed
	}

	return year, month, nil
}

func parseGroups(raw string) []string {
	if raw == "" {
		return nil
	}

	var groups []string
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(strings.ToLower(part)); trimmed != "" {
			groups = append(groups, trimmed)
		}
	}

	return groups
}
