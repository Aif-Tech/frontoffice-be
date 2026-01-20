package member

import (
	"front-office/internal/core/log/operation"
	"front-office/internal/core/role"
	"front-office/internal/mail"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"front-office/pkg/helper"
	"time"

	"github.com/rs/zerolog/log"
)

func NewService(
	repo Repository,
	roleRepo role.Repository,
	operationRepo operation.Repository,
	mailSvc *mail.SendMailService,
) Service {
	return &service{
		repo,
		roleRepo,
		operationRepo,
		mailSvc,
	}
}

type service struct {
	repo          Repository
	roleRepo      role.Repository
	operationRepo operation.Repository
	mailSvc       *mail.SendMailService
}

type Service interface {
	GetMemberBy(query *MemberParams) (*MstMember, error)
	GetMemberList(filter *MemberParams) ([]*MstMember, *model.Meta, error)
	UpdateProfile(userId string, currentUserRoleId uint, req *updateProfileRequest) (*userUpdateResponse, error)
	UploadProfileImage(id string, filename *string) (*userUpdateResponse, error)
	UpdateMemberById(authCtx *model.AuthContext, memberId string, req *updateUserRequest) error
	UpdateExpiredTokens() error
	DeleteMemberById(memberId, companyId string) error
}

func (svc *service) GetMemberBy(query *MemberParams) (*MstMember, error) {
	member, err := svc.repo.GetMemberAPI(query)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to get member")
	}
	if member.MemberId == 0 {
		return nil, apperror.NotFound(constant.UserNotFound)
	}

	return member, nil
}

func (svc *service) GetMemberList(filter *MemberParams) ([]*MstMember, *model.Meta, error) {
	users, meta, err := svc.repo.GetMemberListAPI(filter)
	if err != nil {
		return nil, nil, err
	}

	return users, meta, nil
}

func (svc *service) UpdateProfile(userId string, currentUserRoleId uint, req *updateProfileRequest) (*userUpdateResponse, error) {
	user, err := svc.repo.GetMemberAPI(&MemberParams{Id: userId})
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedFetchMember)
	}
	if user.MemberId == 0 {
		return nil, apperror.NotFound(constant.UserNotFound)
	}

	updateFields := make(map[string]interface{})
	shouldSendEmailConfirmation := false
	var newEmail string

	if req.Name != nil {
		updateFields["name"] = *req.Name
		user.Name = *req.Name
	}

	if req.Email != nil {
		if currentUserRoleId == uint(memberRoleID) {
			return nil, apperror.Unauthorized("you are not allowed to update email")
		}

		existing, err := svc.repo.GetMemberAPI(&MemberParams{Email: *req.Email})
		if err != nil {
			return nil, apperror.MapRepoError(err, "failed to check existing email")
		}
		if existing.MemberId != 0 {
			return nil, apperror.Conflict(constant.EmailAlreadyExists)
		}

		updateFields["email"] = *req.Email
		shouldSendEmailConfirmation = true
		newEmail = *req.Email
	}

	updateFields["updated_at"] = time.Now()

	if err := svc.repo.UpdateMemberAPI(userId, updateFields); err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedUpdateMember)
	}

	if shouldSendEmailConfirmation {
		if err := svc.mailSvc.SendWithTemplate(
			newEmail,
			"Scoreezy Account Email Updated",
			"email_changed.html",
			map[string]any{
				"Name":         user.Name,
				"OldEmail":     user.Email,
				"NewEmail":     newEmail,
				"DateOfChange": helper.FormatWIB(time.Now()),
				"Year":         time.Now().Year(),
			},
		); err != nil {
			log.Warn().
				Err(err).
				Str("member_id", userId).
				Msg("failed to send email change confirmation")
		}

		user.Email = newEmail
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  user.MemberId,
		CompanyId: user.CompanyId,
		Action:    constant.EventUpdateProfile,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventUpdateProfile).
			Msg("failed to add operation log")
	}

	return &userUpdateResponse{
		Id:        user.MemberId,
		Name:      user.Name,
		Email:     user.Email,
		Active:    user.Active,
		CompanyId: user.CompanyId,
		RoleId:    user.RoleId,
	}, nil
}

func (svc *service) UploadProfileImage(userId string, filename *string) (*userUpdateResponse, error) {
	user, err := svc.repo.GetMemberAPI(&MemberParams{Id: userId})
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedFetchMember)
	}
	if user.MemberId == 0 {
		return nil, apperror.NotFound(constant.UserNotFound)
	}

	updateFields := make(map[string]interface{})

	if filename != nil {
		updateFields["image"] = *filename
	}

	updateFields["updated_at"] = time.Now()

	if err := svc.repo.UpdateMemberAPI(userId, updateFields); err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedUpdateMember)
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  user.MemberId,
		CompanyId: user.CompanyId,
		Action:    constant.EventUpdateProfile,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventUpdateProfile).
			Msg("failed to add operation log")
	}

	return &userUpdateResponse{
		Id:        user.MemberId,
		Name:      user.Name,
		Email:     user.Email,
		Active:    user.Active,
		CompanyId: user.CompanyId,
		RoleId:    user.RoleId,
	}, nil
}

func (svc *service) UpdateMemberById(authCtx *model.AuthContext, memberId string, req *updateUserRequest) error {
	member, err := svc.repo.GetMemberAPI(&MemberParams{Id: memberId, CompanyId: authCtx.CompanyIdStr()})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchMember)
	}
	if member.MemberId == 0 {
		return apperror.NotFound(constant.UserNotFound)
	}

	updateFields := make(map[string]interface{})
	var (
		sendEmailConfirmation bool
		newEmail              string
		logEvents             []string
	)

	if req.Name != nil || req.Email != nil || req.RoleId != nil {
		logEvents = append(logEvents, constant.EventUpdateUserData)
	}

	if req.Name != nil {
		updateFields["name"] = *req.Name
	}

	if req.Email != nil {
		if authCtx.RoleId == uint(memberRoleID) {
			return apperror.Unauthorized("you are not allowed to update email")
		}

		existing, err := svc.repo.GetMemberAPI(&MemberParams{Email: *req.Email})
		if err != nil {
			return apperror.MapRepoError(err, "failed to check existing email")
		}
		if existing.MemberId != 0 {
			return apperror.Conflict(constant.EmailAlreadyExists)
		}

		updateFields["email"] = *req.Email
		newEmail = *req.Email
		sendEmailConfirmation = true
	}

	if req.RoleId != nil {
		role, err := svc.roleRepo.GetRoleByIdAPI(*req.RoleId)
		if err != nil {
			return apperror.MapRepoError(err, "failed to fetch role")
		}
		if role.RoleId == 0 {
			return apperror.NotFound("role not found")
		}

		updateFields["role_id"] = *req.RoleId
	}

	if req.Active != nil {
		updateFields["active"] = *req.Active

		if *req.Active {
			updateFields["mail_status"] = "active"
			logEvents = append(logEvents, constant.EventActivateUser)
		} else {
			updateFields["mail_status"] = "inactive"
			logEvents = append(logEvents, constant.EventInactivateUser)
		}
	}

	updateFields["updated_at"] = time.Now()

	if err := svc.repo.UpdateMemberAPI(memberId, updateFields); err != nil {
		return apperror.MapRepoError(err, constant.FailedUpdateMember)
	}

	if sendEmailConfirmation {
		if err := svc.mailSvc.SendWithTemplate(
			member.Email,
			"Scoreezy Account Email Updated",
			"email_changed.html",
			map[string]any{
				"Name":         member.Name,
				"OldEmail":     member.Email,
				"NewEmail":     newEmail,
				"DateOfChange": helper.FormatWIB(time.Now()),
				"Year":         time.Now().Year(),
			},
		); err != nil {
			log.Warn().
				Err(err).
				Str("member_id", memberId).
				Msg("failed to send email change confirmation")
		}
	}

	for _, event := range logEvents {
		if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
			MemberId:  authCtx.UserId,
			CompanyId: member.CompanyId,
			Action:    event,
		}); err != nil {
			log.Warn().
				Err(err).
				Str("action", event).
				Msg("failed to add operation log")
		}
	}

	return nil
}

func (svc *service) UpdateExpiredTokens() error {
	if err := svc.repo.UpdateExpiredTokensAPI(); err != nil {
		return apperror.MapRepoError(err, "failed to update expired mail")
	}

	return nil
}

func (svc *service) DeleteMemberById(memberId, companyId string) error {
	member, err := svc.repo.GetMemberAPI(&MemberParams{Id: memberId, CompanyId: companyId})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchMember)
	}
	if member.MemberId == 0 {
		return apperror.NotFound(constant.UserNotFound)
	}

	if err := svc.repo.DeleteMemberAPI(memberId); err != nil {
		return apperror.MapRepoError(err, "failed to delete member")
	}

	return nil
}
