package operation

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
	GetLogsOperationAPI(filter *logOperationFilter) (*model.AifcoreAPIResponse[any], error)
	GetLogsByRangeAPI(filter *logRangeFilter) (*model.AifcoreAPIResponse[any], error)
	AddLogOperation(req *AddLogRequest) error
}

func (repo *repository) GetLogsOperationAPI(filter *logOperationFilter) (*model.AifcoreAPIResponse[any], error) {
	url := fmt.Sprintf("%s/api/core/logging/operation/list/%s", repo.cfg.Env.AifcoreHost, filter.CompanyId)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	q := req.URL.Query()
	q.Add(constant.Page, filter.Page)
	q.Add(constant.Size, filter.Size)
	q.Add("name", filter.Name)
	q.Add("role", filter.Role)
	q.Add("event", filter.Event)
	q.Add(constant.StartDate, filter.StartDate)
	q.Add(constant.EndDate, filter.EndDate)
	req.URL.RawQuery = q.Encode()

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

func (repo *repository) GetLogsByRangeAPI(filter *logRangeFilter) (*model.AifcoreAPIResponse[any], error) {
	url := fmt.Sprintf("%s/api/core/logging/operation/range", repo.cfg.Env.AifcoreHost)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	q := req.URL.Query()
	q.Add(constant.Page, filter.Page)
	q.Add(constant.Size, filter.Size)
	q.Add("company_id", filter.CompanyId)
	q.Add(constant.StartDate, filter.StartDate)
	q.Add(constant.EndDate, filter.EndDate)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	return helper.ParseAifcoreAPIResponse[any](resp)
}

func (repo *repository) AddLogOperation(payload *AddLogRequest) error {
	url := fmt.Sprintf("%s/api/core/logging/operation", repo.cfg.Env.AifcoreHost)

	bodyBytes, err := repo.marshalFn(payload)
	if err != nil {
		return errors.New(constant.ErrInvalidRequestPayload)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
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
