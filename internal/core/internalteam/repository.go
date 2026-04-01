package internalteam

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"front-office/configs/application"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
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
	AddMemberAPI(req *registerRequest) (*model.AifcoreAPIResponse[any], error)
	GetMemberAPI() (*model.AifcoreAPIResponse[[]MstInternalTeam], error)
}

func (repo *repository) AddMemberAPI(payload *registerRequest) (*model.AifcoreAPIResponse[any], error) {
	url := fmt.Sprintf("%s/api/core/aifuser/internal", repo.cfg.App.AifcoreHost)

	bodyBytes, err := repo.marshalFn(payload)
	if err != nil {
		return nil, errors.New(constant.ErrInvalidRequestPayload)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[any](resp)
	if err != nil {
		return nil, err
	}

	return apiResp, nil
}

func (repo *repository) GetMemberAPI() (*model.AifcoreAPIResponse[[]MstInternalTeam], error) {
	url := fmt.Sprintf("%s/api/core/aifuser/internal/list", repo.cfg.App.AifcoreHost)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[[]MstInternalTeam](resp)
	if err != nil {
		return nil, err
	}

	return apiResp, nil
}
