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

	page := c.Query(constant.Page, "1")
	size := c.Query(constant.Size, "10")
	role := strings.ToLower(c.Query("role"))
	eventQuery := c.Query("event")
	name := strings.ToLower(c.Query("name", ""))
	startDate := c.Query(constant.StartDate)
	endDate := c.Query(constant.EndDate)

	mappedEvent, valid := mapEventKeyword(eventQuery)
	if eventQuery != "" && !valid {
		return apperror.BadRequest("invalid event type")
	}

	filter := &logOperationFilter{
		CompanyId: authCtx.CompanyIdStr(),
		Page:      page,
		Size:      size,
		Role:      role,
		Event:     mappedEvent,
		Name:      name,
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

	page := c.Query(constant.Page, "1")
	size := c.Query(constant.Size, "10")
	startDate := c.Query(constant.StartDate)
	endDate := c.Query((constant.EndDate))

	if startDate == "" || endDate == "" {
		return apperror.BadRequest("start_date and end_date are required")
	}

	filter := &logRangeFilter{
		Page:      page,
		Size:      size,
		CompanyId: authCtx.CompanyIdStr(),
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
		"sign_in":                     constant.EventSignIn,
		"sign_out":                    constant.EventSignOut,
		"change_password":             constant.EventChangePassword,
		"add_new_user":                constant.EventRegisterMember,
		"request_password_reset":      constant.EventRequestPasswordReset,
		"password_reset":              constant.EventPasswordReset,
		"update_profile_account":      constant.EventUpdateProfile,
		"updates_user_data":           constant.EventUpdateUserData,
		"activate_user":               constant.EventActivateUser,
		"inactivate_user":             constant.EventInactivateUser,
		"calculate_score":             constant.EventCalculateScore,
		"download_history_hit":        constant.EventDownloadScoreHistory,
		"change_billing_information":  constant.EventChangeBillingInformation,
		"topup_balance":               constant.EventTopupBalance,
		"submit_payment_confirmation": constant.EventSubmitPaymentConfirmation,
	}

	normalized := strings.ToLower(strings.ReplaceAll(input, " ", "_"))
	event, ok := eventMap[normalized]
	return event, ok
}
