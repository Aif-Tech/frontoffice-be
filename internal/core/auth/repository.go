package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"front-office/configs/application"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"front-office/pkg/httpclient"
	"front-office/pkg/jsonutil"

	"net/http"
)

func NewRepository(cfg *application.Config, client httpclient.HTTPClient, marshalFn jsonutil.Marshaller) Repository {
	if marshalFn == nil {
		marshalFn = json.Marshal
	}

	return &repository{
		cfg:       cfg,
		client:    client,
		marshalFn: marshalFn,
	}
}

type repository struct {
	cfg       *application.Config
	client    httpclient.HTTPClient
	marshalFn jsonutil.Marshaller
}

type Repository interface {
	// CreateAdmin(company *company.MstCompany, user *member.MstMember, activationToken *activationtoken.MstActivationToken) (*member.MstMember, error)
	// CreateMember(user *member.MstMember, activationToken *activationtoken.MstActivationToken) (*member.MstMember, error)
	VerifyMemberAPI(userId string, req *passwordResetRequest) error
	ChangePasswordAPI(userId string, req *changePasswordRequest) error
	PasswordResetAPI(userId, token string, req *passwordResetRequest) error
	AuthMemberAPI(req *userLoginRequest) (*loginResponseData, error)
}

// func (repo *repository) CreateAdmin(company *company.MstCompany, user *member.MstMember, activationToken *activationtoken.MstActivationToken) (*member.MstMember, error) {
// 	errTx := repo.DB.Transaction(func(tx *gorm.DB) error {
// 		if err := tx.Create(&company).Error; err != nil {
// 			return err
// 		}

// 		user.CompanyId = company.CompanyId
// 		if err := tx.Create(&user).Error; err != nil {
// 			return err
// 		}

// 		if err := tx.Create(&activationToken).Error; err != nil {
// 			return err
// 		}

// 		return nil
// 	})

// 	if errTx != nil {
// 		return user, errTx
// 	}

// 	repo.DB.Preload("Company").Preload("Company.Industry").Preload("Role").First(&user)

// 	return user, errTx
// }

// func (repo *repository) CreateMember(user *member.MstMember, activationToken *activationtoken.MstActivationToken) (*member.MstMember, error) {
// 	errTx := repo.DB.Transaction(func(tx *gorm.DB) error {
// 		if err := tx.Create(&user).Error; err != nil {
// 			return err
// 		}

// 		if err := tx.Create(&activationToken).Error; err != nil {
// 			return err
// 		}

// 		return nil
// 	})

// 	if errTx != nil {
// 		return nil, errTx
// 	}

// 	return user, nil
// }

func (repo *repository) VerifyMemberAPI(userId string, payload *passwordResetRequest) error {
	url := fmt.Sprintf(`%v/api/core/member/%v/activation-tokens`, repo.cfg.Env.AifcoreHost, userId)

	bodyBytes, err := repo.marshalFn(payload)
	if err != nil {
		return errors.New(constant.ErrInvalidRequestPayload)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	resp, err := repo.client.Do(req)
	if err != nil {
		return errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	_, err = helper.ParseAifcoreAPIResponse[any](resp)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) PasswordResetAPI(userId, token string, payload *passwordResetRequest) error {
	url := fmt.Sprintf(`%v/api/core/member/%v/password-reset-tokens/%v`, repo.cfg.Env.AifcoreHost, userId, token)

	bodyBytes, err := repo.marshalFn(payload)
	if err != nil {
		return errors.New(constant.ErrInvalidRequestPayload)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	resp, err := repo.client.Do(req)
	if err != nil {
		return errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	_, err = helper.ParseAifcoreAPIResponse[any](resp)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) ChangePasswordAPI(userId string, payload *changePasswordRequest) error {
	url := fmt.Sprintf(`%v/api/core/member/%v/change-password`, repo.cfg.Env.AifcoreHost, userId)

	bodyBytes, err := repo.marshalFn(payload)
	if err != nil {
		return errors.New(constant.ErrInvalidRequestPayload)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	resp, err := repo.client.Do(req)
	if err != nil {
		return errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	_, err = helper.ParseAifcoreAPIResponse[any](resp)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) AuthMemberAPI(payload *userLoginRequest) (*loginResponseData, error) {
	url := fmt.Sprintf("%s/api/middleware/auth-member-login", repo.cfg.Env.AifcoreHost)

	bodyBytes, err := repo.marshalFn(payload)
	if err != nil {
		return nil, errors.New(constant.ErrInvalidRequestPayload)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	// send http request
	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	// parse structured response
	apiResp, err := helper.ParseAifcoreAPIResponse[*loginResponseData](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}
