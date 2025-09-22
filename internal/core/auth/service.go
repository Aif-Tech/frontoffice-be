package auth

import (
	"errors"
	"fmt"
	"front-office/configs/application"
	"front-office/internal/core/activation"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/member"
	"front-office/internal/core/passwordreset"
	"front-office/internal/core/role"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"front-office/pkg/utility/mailjet"

	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

func NewService(
	cfg *application.Config,
	repo Repository,
	memberRepo member.Repository,
	roleRepo role.Repository,
	operationRepo operation.Repository,
	activationRepo activation.Repository,
	passwordResetRepo passwordreset.Repository,
) Service {
	return &service{
		cfg,
		repo,
		memberRepo,
		roleRepo,
		operationRepo,
		activationRepo,
		passwordResetRepo,
	}
}

type service struct {
	cfg               *application.Config
	repo              Repository
	memberRepo        member.Repository
	roleRepo          role.Repository
	operationRepo     operation.Repository
	activationRepo    activation.Repository
	passwordResetRepo passwordreset.Repository
}

type Service interface {
	// RegisterAdminSvc(req *RegisterAdminRequest) (*user.User, string, error)
	LoginMember(loginReq *userLoginRequest) (accessToken, refreshToken string, loginResp *loginResponse, err error)
	RefreshAccessToken(userId, companyId, tierLevel uint, apiKey string) (string, error)
	Logout(userId, companyId uint) error
	AddMember(currentUserId uint, req *member.RegisterMemberRequest) error
	RequestActivation(email string) error
	RequestPasswordReset(email string) error
	PasswordReset(token string, req *PasswordResetRequest) error
	VerifyMember(token string, req *PasswordResetRequest) error
	ChangePassword(userId string, req *ChangePasswordRequest) error
}

// func (svc *service) RegisterAdminSvc(req *RegisterAdminRequest) (*user.User, string, error) {
// 	secret := svc.Cfg.Env.JwtSecretKey
// 	minutesToExpired, _ := strconv.Atoi(svc.Cfg.Env.JwtVerificationExpiresMinutes)

// 	isPasswordStrength := helper.ValidatePasswordStrength(req.Password)
// 	if !isPasswordStrength {
// 		return nil, "", errors.New(constant.InvalidPassword)
// 	}

// 	var tierLevel uint
// 	if req.RoleId != "" {
// 		result, err := svc.RepoRole.FindOneById(req.RoleId)
// 		if result == nil {
// 			return nil, "", errors.New(constant.DataNotFound)
// 		} else if err != nil {
// 			return nil, "", err
// 		} else {
// 			tierLevel = result.TierLevel
// 		}
// 	}

// 	companyId := uuid.NewString()
// 	dataCompany := &company.Company{
// 		Id:              companyId,
// 		CompanyName:     req.CompanyName,
// 		CompanyAddress:  req.CompanyAddress,
// 		CompanyPhone:    req.CompanyPhone,
// 		AgreementNumber: req.AgreementNumber,
// 		PaymentScheme:   req.PaymentScheme,
// 		IndustryId:      req.IndustryId,
// 	}

// 	userId := uuid.NewString()
// 	dataUser := &user.User{
// 		Id:     userId,
// 		Name:   req.Name,
// 		Email:  req.Email,
// 		Phone:  req.Phone,
// 		Key:    helper.GenerateAPIKey(),
// 		RoleId: req.RoleId,
// 	}

// 	dataUser.Password = user.SetPassword(req.Password)

// 	token, err := helper.GenerateToken(secret, minutesToExpired, 1, 1, tierLevel)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	tokenId := uuid.NewString()
// 	dataActivationToken := &activationtoken.MstActivationToken{
// 		Id:       tokenId,
// 		Token:    token,
// 		MemberId: userId,
// 	}

// 	user, err := svc.Repo.CreateAdmin(dataCompany, dataUser, dataActivationToken)
// 	if err != nil {
// 		return user, "", err
// 	}

// 	return user, token, nil
// }

func (svc *service) VerifyMember(token string, req *PasswordResetRequest) error {
	activationData, err := svc.activationRepo.GetActivationTokenAPI(token)
	if err != nil {
		return apperror.MapRepoError(err, "failed to retrieve activation token")
	}

	userId := fmt.Sprintf("%d", activationData.MemberId)

	user, err := svc.memberRepo.GetMemberAPI(&member.MemberParams{
		Id: userId,
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchMember)
	}
	if user.MemberId == 0 {
		return apperror.NotFound(constant.UserNotFound)
	}
	if user.IsVerified && user.Active {
		return apperror.BadRequest(constant.AlreadyVerified)
	}

	minutesToExpired, err := strconv.Atoi(svc.cfg.Env.JwtActivationExpiresMinutes)
	if err != nil {
		return apperror.Internal("invalid activation expiry config", err)
	}

	elapsedMinutes := time.Since(activationData.CreatedAt).Minutes()
	if elapsedMinutes > float64(minutesToExpired) {
		updateFields := map[string]interface{}{
			"mail_status": mailStatusResend,
			"updated_at":  time.Now(),
		}

		err := svc.memberRepo.UpdateMemberAPI(userId, updateFields)
		if err != nil {
			return apperror.MapRepoError(err, "failed to update member after token expired")
		}

		return apperror.Forbidden(constant.ActivationTokenExpired)
	}

	if !helper.ValidatePasswordStrength(req.Password) {
		return apperror.BadRequest(constant.InvalidPassword)
	}

	if req.Password != req.ConfirmPassword {
		return apperror.BadRequest(constant.ConfirmPasswordMismatch)
	}

	if err := svc.repo.VerifyMemberAPI(userId, req); err != nil {
		return apperror.MapRepoError(err, "failed to verify member")
	}

	return nil
}

func (svc *service) PasswordReset(token string, req *PasswordResetRequest) error {
	data, err := svc.passwordResetRepo.GetPasswordResetTokenAPI(token)
	if err != nil {
		return apperror.Forbidden(constant.InvalidPasswordResetLink)
	}

	idStr := strconv.Itoa(int(data.Id))

	minutesToExpired, err := strconv.Atoi(svc.cfg.Env.JwtResetPasswordExpiresMinutes)
	if err != nil {
		return apperror.Internal("invalid password reset expiry config", err)
	}

	elapsedMinutes := time.Since(data.CreatedAt).Minutes()
	if elapsedMinutes > float64(minutesToExpired) {
		if err := svc.passwordResetRepo.DeletePasswordResetTokenAPI(idStr); err != nil {
			return apperror.MapRepoError(err, "failed to delete password reset token")
		}

		return apperror.Forbidden(constant.InvalidPasswordResetLink)
	}

	if !helper.ValidatePasswordStrength(req.Password) {
		return apperror.BadRequest(constant.InvalidPassword)
	}

	if req.Password != req.ConfirmPassword {
		return apperror.BadRequest(constant.ConfirmPasswordMismatch)
	}

	err = svc.repo.PasswordResetAPI(strconv.Itoa(int(data.MemberId)), token, req)
	if err != nil {
		return apperror.MapRepoError(err, "failed to password reset")
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  data.MemberId,
		CompanyId: data.Member.CompanyId,
		Action:    constant.EventPasswordReset,
	}); err != nil {
		log.Warn().Err(err).Msg("failed to log password reset event")
	}

	return nil
}

func (svc *service) AddMember(currentUserId uint, req *member.RegisterMemberRequest) error {
	user, err := svc.memberRepo.AddMemberAPI(req)
	if err != nil {
		return apperror.MapRepoError(err, "failed to register member")
	}

	tokenPayload := &tokenPayload{
		MemberId:  user.MemberId,
		CompanyId: req.CompanyId,
		RoleId:    req.RoleId,
	}
	activationToken, err := svc.generateToken(tokenPayload, svc.cfg.Env.JwtSecretKey, svc.cfg.Env.JwtActivationExpiresMinutes)
	if err != nil {
		return apperror.Internal("generate activation token failed", err)
	}

	userIdStr := helper.ConvertUintToString(user.MemberId)

	err = svc.activationRepo.CreateActivationTokenAPI(userIdStr, &activation.CreateActivationTokenRequest{
		Token: activationToken,
	})
	if err != nil {
		return apperror.MapRepoError(err, "failed to create activation")
	}

	err = mailjet.SendEmailActivation(req.Email, activationToken)
	if err != nil {
		updateFields := map[string]interface{}{
			"mail_status": mailStatusResend,
			"updated_at":  time.Now(),
		}

		err := svc.memberRepo.UpdateMemberAPI(userIdStr, updateFields)
		if err != nil {
			return apperror.MapRepoError(err, "failed to update member after email failure")
		}

		return apperror.Internal("failed to send activation email", err)
	}

	err = svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  currentUserId,
		CompanyId: req.CompanyId,
		Action:    constant.EventRegisterMember,
	})
	if err != nil {
		log.Warn().Err(err).Msg("failed to log register member event")
	}

	return nil
}

func (svc *service) RequestActivation(email string) error {
	user, err := svc.memberRepo.GetMemberAPI(&member.MemberParams{
		Email: email,
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchMember)
	}

	if user.MemberId == 0 {
		return apperror.NotFound(constant.UserNotFound)
	}

	if user.IsVerified {
		return apperror.Conflict(constant.AlreadyVerified)
	}

	tokenPayload := &tokenPayload{
		MemberId:  user.MemberId,
		CompanyId: user.CompanyId,
		RoleId:    user.RoleId,
	}
	token, err := svc.generateToken(tokenPayload, svc.cfg.Env.JwtSecretKey, svc.cfg.Env.JwtActivationExpiresMinutes)
	if err != nil {
		return apperror.Internal("generate activation token failed", err)
	}

	userIdStr := helper.ConvertUintToString(user.MemberId)

	if err := svc.activationRepo.CreateActivationTokenAPI(userIdStr, &activation.CreateActivationTokenRequest{
		Token: token,
	}); err != nil {
		return apperror.MapRepoError(err, "failed to create activation")
	}

	if err := mailjet.SendEmailActivation(email, token); err != nil {
		return apperror.Internal("failed to send activation email", err)
	}

	updateFields := map[string]interface{}{
		"mail_status": mailStatusPending,
		"updated_at":  time.Now(),
	}

	if err := svc.memberRepo.UpdateMemberAPI(userIdStr, updateFields); err != nil {
		return apperror.MapRepoError(err, constant.FailedUpdateMember)
	}

	return nil
}

func (svc *service) RequestPasswordReset(email string) error {
	user, err := svc.memberRepo.GetMemberAPI(&member.MemberParams{
		Email: email,
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchMember)
	}
	if user.MemberId == 0 {
		return apperror.NotFound(constant.UserNotFoundForgotEmail)
	}

	if !user.IsVerified {
		return apperror.Unauthorized(constant.UnverifiedUser)
	}

	tokenPayload := &tokenPayload{
		MemberId:  user.MemberId,
		CompanyId: user.CompanyId,
		RoleId:    user.RoleId,
	}
	token, err := svc.generateToken(tokenPayload, svc.cfg.Env.JwtSecretKey, svc.cfg.Env.JwtActivationExpiresMinutes)
	if err != nil {
		return apperror.Internal("generate password reset token failed", err)
	}

	userIdStr := helper.ConvertUintToString(user.MemberId)

	if err := svc.passwordResetRepo.CreatePasswordResetTokenAPI(userIdStr, &passwordreset.CreatePasswordResetTokenRequest{
		Token: token,
	}); err != nil {
		return apperror.MapRepoError(err, "failed to create password reset token")
	}

	if err := mailjet.SendEmailPasswordReset(email, user.Name, token); err != nil {
		return apperror.Internal("failed to send password reset email email", err)
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  user.MemberId,
		CompanyId: user.CompanyId,
		Action:    constant.EventRequestPasswordReset,
	}); err != nil {
		log.Warn().Err(err).Msg("failed to log request password reset event")
	}

	return nil
}

func (svc *service) LoginMember(req *userLoginRequest) (accessToken, refreshToken string, loginResp *loginResponse, err error) {
	user, err := svc.repo.AuthMemberAPI(req)
	if err != nil {
		var apiErr *apperror.ExternalAPIError
		if errors.As(err, &apiErr) {
			return "", "", nil, apperror.MapAuthError(apiErr)
		}

		return "", "", nil, apperror.Internal("auth failed", err)
	}

	tokenPayload := &tokenPayload{
		MemberId:  user.MemberId,
		CompanyId: user.CompanyId,
		RoleId:    user.RoleId,
		QuotaType: user.QuotaType,
		ApiKey:    user.ApiKey,
	}
	accessToken, err = svc.generateToken(tokenPayload, svc.cfg.Env.JwtSecretKey, svc.cfg.Env.JwtExpiresMinutes)
	if err != nil {
		return "", "", nil, apperror.Internal("generate access token failed", err)
	}

	refreshToken, err = svc.generateToken(tokenPayload, svc.cfg.Env.JwtSecretKey, svc.cfg.Env.JwtRefreshTokenExpiresMinutes)
	if err != nil {
		return "", "", nil, apperror.Internal("generate refresh token failed", err)
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  user.MemberId,
		CompanyId: user.CompanyId,
		Action:    constant.EventSignIn,
	}); err != nil {
		log.Warn().Err(err).Msg("failed to log sign-in event")
	}

	loginResp = &loginResponse{
		Id:                 user.MemberId,
		Name:               user.Name,
		Email:              user.Email,
		CompanyId:          user.CompanyId,
		CompanyName:        user.CompanyName,
		QuotaType:          user.QuotaType,
		TierLevel:          user.RoleId,
		Image:              user.Image,
		SubscriberProducts: user.SubscriberProducts,
	}

	return accessToken, refreshToken, loginResp, nil
}

func (svc *service) RefreshAccessToken(userId, companyId, roleId uint, apiKey string) (string, error) {
	tokenPayload := &tokenPayload{
		MemberId:  userId,
		CompanyId: companyId,
		RoleId:    roleId,
		ApiKey:    apiKey,
	}

	accessToken, err := svc.generateToken(tokenPayload, svc.cfg.Env.JwtSecretKey, svc.cfg.Env.JwtExpiresMinutes)
	if err != nil {
		return "", apperror.Internal("generate access token failed", err)
	}

	return accessToken, nil
}

func (svc *service) Logout(userId, companyId uint) error {
	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  userId,
		CompanyId: companyId,
		Action:    constant.EventSignOut,
	}); err != nil {
		log.Warn().Err(err).Msg("failed to log sign-out event")
	}

	return nil
}

func (svc *service) ChangePassword(userId string, reqBody *ChangePasswordRequest) error {
	user, err := svc.memberRepo.GetMemberAPI(&member.MemberParams{
		Id: userId,
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchMember)
	}

	if !helper.ValidatePasswordStrength(reqBody.NewPassword) {
		return apperror.BadRequest(constant.InvalidPassword)
	}

	if reqBody.NewPassword != reqBody.ConfirmNewPassword {
		return apperror.BadRequest(constant.ConfirmPasswordMismatch)
	}

	if err := svc.repo.ChangePasswordAPI(userId, reqBody); err != nil {
		var apiErr *apperror.ExternalAPIError
		if errors.As(err, &apiErr) {
			return apperror.MapChangePasswordError(apiErr)
		}

		return apperror.Internal("failed to change password", err)
	}

	if err := mailjet.SendConfirmationEmailPasswordChangeSuccess(user.Name, user.Email); err != nil {
		return apperror.Internal("failed to send confirmation password change", err)
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  user.MemberId,
		CompanyId: user.CompanyId,
		Action:    constant.EventRequestPasswordReset,
	}); err != nil {
		log.Warn().Err(err).Msg("failed to log change password event")
	}

	return nil
}

func (svc *service) generateToken(payload *tokenPayload, secret, minutesStr string) (string, error) {
	minutes, err := strconv.Atoi(minutesStr)
	if err != nil {
		return "", fmt.Errorf("invalid duration: %w", err)
	}

	return helper.GenerateToken(secret, minutes, payload.MemberId, payload.CompanyId, payload.RoleId, payload.QuotaType, payload.ApiKey)
}
