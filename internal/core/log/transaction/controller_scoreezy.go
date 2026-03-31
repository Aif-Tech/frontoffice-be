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
	startDate := c.Query(constant.StartDate)
	endDate := c.Query(constant.EndDate)

	if startDate == "" || endDate == "" {
		return apperror.BadRequest("start_date and end_date  are required")
	}

	filter := &LogFilter{
		Page:      c.Query(constant.Page, "1"),
		Size:      c.Query(constant.Size, "10"),
		CompanyId: c.Query("company_id"),
		StartDate: startDate,
		EndDate:   endDate,
	}

	logs, err := ctrl.svc.GetScoreezyLogsByDateRange(filter)
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
