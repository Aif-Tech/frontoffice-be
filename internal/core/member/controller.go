package member

import (
	"front-office/internal/core/log/operation"
	"front-office/internal/core/role"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"

	"github.com/gofiber/fiber/v2"
)

func NewController(
	service Service,
	roleService role.Service,
	logOperationService operation.Service) Controller {
	return &controller{
		svc:             service,
		roleSvc:         roleService,
		logOperationSvc: logOperationService,
	}
}

type controller struct {
	svc             Service
	roleSvc         role.Service
	logOperationSvc operation.Service
}

type Controller interface {
	GetBy(c *fiber.Ctx) error
	GetById(c *fiber.Ctx) error
	GetList(c *fiber.Ctx) error
	UpdateProfile(c *fiber.Ctx) error
	UploadProfileImage(c *fiber.Ctx) error
	UpdateMemberById(c *fiber.Ctx) error
	DeleteById(c *fiber.Ctx) error
}

func (ctrl *controller) GetBy(c *fiber.Ctx) error {
	member, err := ctrl.svc.GetMemberBy(&MemberParams{
		Email:    c.Query("email"),
		Username: c.Query("username"),
		Key:      c.Query("key"),
	})
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		"succeed to get a user",
		member,
	))
}

func (ctrl *controller) GetById(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apperror.BadRequest(constant.MissingUserId)
	}

	member, err := ctrl.svc.GetMemberBy(&MemberParams{
		Id: id,
	})
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		"succeed to get a user",
		member,
	))
}

func (ctrl *controller) GetList(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	filter := &MemberParams{
		CompanyId: authCtx.CompanyIdStr(),
		Page:      c.Query(constant.Page, "1"),
		Limit:     c.Query("limit", "10"),
		Keyword:   c.Query("keyword", ""),
		StartDate: c.Query("startDate", ""),
		EndDate:   c.Query("endDate", ""),
	}

	users, meta, err := ctrl.svc.GetMemberList(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		"succeed to get member list",
		&memberListResponse{
			Data:      users,
			TotalData: meta.Total,
		},
	))
}

func (ctrl *controller) UpdateProfile(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*updateProfileRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	updateResp, err := ctrl.svc.UpdateProfile(authCtx.UserIdStr(), authCtx.RoleId, reqBody)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		"succeed to update profile",
		updateResp,
	))
}

func (ctrl *controller) UploadProfileImage(c *fiber.Ctx) error {
	filename, err := helper.GetStringLocal(c, "filename")
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	resp, err := ctrl.svc.UploadProfileImage(authCtx.UserIdStr(), &filename)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		"success to upload profile image",
		resp,
	))
}

func (ctrl *controller) UpdateMemberById(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*updateUserRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	memberId := c.Params("id")
	if memberId == "" {
		return apperror.BadRequest(constant.MissingUserId)
	}

	if err := ctrl.svc.UpdateMemberById(authCtx, memberId, reqBody); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"success to update user",
		nil,
	))
}

func (ctrl *controller) DeleteById(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	id := c.Params("id")
	if id == "" {
		return apperror.BadRequest(constant.MissingUserId)
	}

	if err := ctrl.svc.DeleteMemberById(id, authCtx.CompanyIdStr()); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"succeed to delete member",
		nil,
	))
}
