package transaction

import (
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"

	"github.com/gofiber/fiber/v2"
)

func (ctrl *controller) GetLogScoreezy(c *fiber.Ctx) error {
	logs, err := ctrl.svc.GetScoreezyLogs()
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		constant.SucceedGetLogTrans,
		logs,
	))
}

func (ctrl *controller) GetLogScoreezyByDate(c *fiber.Ctx) error {
	date := c.Query("date")
	companyId := c.Query("company_id")

	if date == "" {
		return apperror.BadRequest("date are required")
	}

	logs, err := ctrl.svc.GetScoreezyLogsByDate(companyId, date)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		constant.SucceedGetLogTrans,
		logs,
	))
}

func (ctrl *controller) GetLogScoreezyByDateRange(c *fiber.Ctx) error {
	page := c.Query(constant.Page, "1")
	startDate := c.Query(constant.StartDate)
	endDate := c.Query(constant.EndDate)
	companyId := c.Query("company_id")

	if startDate == "" || endDate == "" {
		return apperror.BadRequest("start_date and end_date  are required")
	}

	logs, err := ctrl.svc.GetScoreezyLogsByDateRange(startDate, endDate, companyId, page)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		constant.SucceedGetLogTrans,
		logs,
	))
}

func (ctrl *controller) GetLogScoreezyByMonth(c *fiber.Ctx) error {
	companyId := c.Query("company_id")
	month := c.Query("month")

	if month == "" {
		return apperror.BadRequest("month are required")
	}

	logs, err := ctrl.svc.GetScoreezyLogsByMonth(companyId, month)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		constant.SucceedGetLogTrans,
		logs,
	))
}
