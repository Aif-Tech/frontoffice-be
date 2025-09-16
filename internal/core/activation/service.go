package activation

import (
	"errors"
	"strconv"

	"front-office/configs/application"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"strings"
)

func NewService(repo Repository, cfg *application.Config) Service {
	return &service{repo, cfg}
}

type service struct {
	repo Repository
	cfg  *application.Config
}

type Service interface {
	CreateActivationToken(memberId, companyId uint, roleId uint) (string, error)
	ValidateActivationToken(authHeader string) (string, uint, error)
	GetActivationToken(token string) (*MstActivationToken, error)
}

func (svc *service) CreateActivationToken(memberId, companyId, roleId uint) (string, error) {
	secret := svc.cfg.Env.JwtSecretKey
	minutesToExpired, err := strconv.Atoi(svc.cfg.Env.JwtActivationExpiresMinutes)
	if err != nil {
		return "", apperror.Internal("invalid activation expiry config", err)
	}

	token, err := helper.GenerateToken(secret, minutesToExpired, memberId, companyId, roleId, "")
	if err != nil {
		return "", apperror.Internal("generate activation token failed", err)
	}

	req := &CreateActivationTokenRequest{
		Token: token,
	}

	memberIdStr := helper.ConvertUintToString(memberId)
	err = svc.repo.CreateActivationTokenAPI(memberIdStr, req)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (svc *service) ValidateActivationToken(authHeader string) (string, uint, error) {
	secret := svc.cfg.Env.JwtSecretKey

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 {
		return "", 0, errors.New(constant.InvalidActivationLink)
	}

	token := bearerToken[1]

	claims, err := helper.ExtractClaimsFromJWT(token, secret)
	if err != nil {
		return "", 0, errors.New(constant.InvalidActivationLink)
	}

	userId, err := helper.ExtractUserIdFromClaims(claims)
	if err != nil {
		return "", 0, errors.New(constant.InvalidActivationLink)
	}

	return token, userId, nil
}

func (svc *service) GetActivationToken(token string) (*MstActivationToken, error) {
	activationToken, err := svc.repo.GetActivationTokenAPI(token)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to get activation token")
	}

	return activationToken, nil
}
