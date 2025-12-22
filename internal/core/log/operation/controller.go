package operation

import (
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func NewController(service Service) Controller {
	return &controller{svc: service}
}

type controller struct {
	svc Service
}

type Controller interface {
	GetList(c *fiber.Ctx) error
	GetListByRange(c *fiber.Ctx) error
}

func (ctrl *controller) GetList(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	eventQuery := c.Query("event")
	startDate := c.Query(constant.StartDate)
	endDate := c.Query(constant.EndDate)

	mappedEvent, valid := mapEventKeyword(eventQuery)
	if eventQuery != "" && !valid {
		return apperror.BadRequest("invalid event type")
	}

	filter := &logOperationFilter{
		CompanyId: authCtx.CompanyIdStr(),
		Page:      c.Query(constant.Page, "1"),
		Size:      c.Query(constant.Size, "10"),
		Role:      c.Query("role"),
		Event:     mappedEvent,
		Name:      c.Query("name", ""),
		SortBy:    strings.ToLower(c.Query("sort_by", "created_at")),
		Order:     strings.ToLower(c.Query("order", "desc")),
		StartDate: startDate,
		EndDate:   endDate,
	}

	result, err := ctrl.svc.GetLogsOperation(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (ctrl *controller) GetListByRange(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	startDate := c.Query(constant.StartDate)
	endDate := c.Query((constant.EndDate))

	if startDate == "" || endDate == "" {
		return apperror.BadRequest("start_date and end_date are required")
	}

	filter := &logRangeFilter{
		Page:      c.Query(constant.Page, "1"),
		Size:      c.Query(constant.Size, "10"),
		CompanyId: authCtx.CompanyIdStr(),
		SortBy:    strings.ToLower(c.Query("sort_by", "created_at")),
		Order:     strings.ToLower(c.Query("order", "desc")),
		StartDate: startDate,
		EndDate:   endDate,
	}

	result, err := ctrl.svc.GetLogsByRange(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func mapEventKeyword(input string) (string, bool) {
	eventMap := map[string]string{
		// auth
		"sign-in":  constant.EventSignIn,
		"sign-out": constant.EventSignOut,

		// user
		"change-password":        constant.EventChangePassword,
		"register-member":        constant.EventRegisterMember,
		"request-password-reset": constant.EventRequestPasswordReset,
		"password-reset":         constant.EventPasswordReset,
		"update-profile":         constant.EventUpdateProfile,
		"update-user-data":       constant.EventUpdateUserData,
		"activate-user":          constant.EventActivateUser,
		"inactivate-user":        constant.EventInactivateUser,

		// balance
		"update-billing-information":  constant.EventChangeBillingInformation,
		"topup-balance":               constant.EventTopupBalance,
		"submit-payment-confirmation": constant.EventSubmitPaymentConfirmation,

		// scoreezy
		"scoreezy-single-request":          constant.EventScoreezySingleReq,
		"scoreezy-bulk-request":            constant.EventScoreezyBulkReq,
		"scoreezy-single-download-result":  constant.EventScoreezySingleDownload,
		"scoreezy-bulk-download-result'":   constant.EventScoreezyBulkDownload,
		"scoreezy-download-result-summary": constant.EventScoreezyDownloadSummary,

		// loan record checker
		"loan-record-single-request":          constant.EventLoanRecordSingleReq,
		"loan-record-bulk-request":            constant.EventLoanRecordBulkReq,
		"loan-record-download-result":         constant.EventLoanRecordDownload,
		"loan-record-download-result-summary": constant.EventLoanRecordDownloadSummary,

		// 7 days multiple loan
		"7dml-single-request":          constant.Event7DMLSingleReq,
		"7dml-bulk-request":            constant.Event7DMLBulkReq,
		"7dml-download-result":         constant.Event7DMLDownload,
		"7dml-download-result-summary": constant.Event7DMLDownloadSummary,

		// 30 days multiple loan
		"30dml-single-request":          constant.Event30DMLSingleReq,
		"30dml-bulk-request":            constant.Event30DMLBulkReq,
		"30dml-download-result":         constant.Event30DMLDownload,
		"30dml-download-result-summary": constant.Event30DMLDownloadSummary,

		// 90 days multiple loan
		"90dml-single-request":          constant.Event90DMLSingleReq,
		"90dml-bulk-request":            constant.Event90DMLBulkReq,
		"90dml-download-result":         constant.Event90DMLDownload,
		"90dml-download-result-summary": constant.Event90DMLDownloadSummary,

		// npwp verification
		"npwp-verification-single-request":          constant.EventNPWPVerificationSingleReq,
		"npwp-verification-bulk-request":            constant.EventNPWPVerificationBulkReq,
		"npwp-verification-download-result":         constant.EventNPWPVerificationDownload,
		"npwp-verification-download-result-summary": constant.EventNPWPVerificationDownloadSummary,

		// phone live status
		"phone-live-single-request":          constant.EventPhoneLiveSingleReq,
		"phone-live-bulk-request":            constant.EventPhoneLiveBulkReq,
		"phone-live-download-result":         constant.EventPhoneLiveDownload,
		"phone-live-download-result-summary": constant.EventPhoneLiveDownloadSummary,

		// tax compliance status
		"tax-compliance-single-request":          constant.EventTaxComplianceSingleReq,
		"tax-compliance-bulk-request":            constant.EventTaxComplianceBulkReq,
		"tax-compliance-download-result":         constant.EventTaxComplianceDownload,
		"tax-compliance-download-result-summary": constant.EventTaxComplianceDownloadSummary,

		// tax score
		"tax-score-single-request":          constant.EventTaxScoreSingleReq,
		"tax-score-bulk-request":            constant.EventTaxScoreBulkReq,
		"tax-score-download-result":         constant.EventTaxScoreDownload,
		"tax-score-download-result-summary": constant.EventTaxScoreDownloadSummary,

		// tax verification detail
		"tax-verification-single-request":          constant.EventTaxVerificationSingleReq,
		"tax-verification-bulk-request":            constant.EventTaxVerificationBulkReq,
		"tax-verification-download-result":         constant.EventTaxVerificationDownload,
		"tax-verification-download-result-summary": constant.EventTaxVerificationDownloadSummary,
	}

	normalized := strings.ToLower(input)
	event, ok := eventMap[normalized]
	return event, ok
}
