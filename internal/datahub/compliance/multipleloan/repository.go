package multipleloan

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
		marshalFn = json.Marshal // default behavior
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
	CallMultipleLoan7Days(apiKey, jobId, memberId, companyId string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error)
	CallMultipleLoan30Days(apiKey, jobId, memberId, companyId string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error)
	CallMultipleLoan90Days(apiKey, jobId, memberId, companyId string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error)
}

func (repo *repository) CallMultipleLoan7Days(apiKey, jobId, memberId, companyId string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error) {
	url := fmt.Sprintf("%s/product/compliance/multiple-loan/7-days", repo.cfg.App.ProductCatalogHost)

	bodyBytes, err := repo.marshalFn(reqBody)
	if err != nil {
		return nil, errors.New(constant.ErrInvalidRequestPayload)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XAPIKey, apiKey)
	req.Header.Set(constant.XMemberId, memberId)
	req.Header.Set(constant.XCompanyId, companyId)

	q := req.URL.Query()
	q.Add("job_id", jobId)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	return helper.ParseProCatAPIResponse[dataMultipleLoanResponse](resp)
}

func (repo *repository) CallMultipleLoan30Days(apiKey, jobId, memberId, companyId string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error) {
	url := fmt.Sprintf("%s/product/compliance/multiple-loan/30-days", repo.cfg.App.ProductCatalogHost)

	bodyBytes, err := repo.marshalFn(reqBody)
	if err != nil {
		return nil, errors.New(constant.ErrInvalidRequestPayload)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XAPIKey, apiKey)
	req.Header.Set(constant.XMemberId, memberId)
	req.Header.Set(constant.XCompanyId, companyId)

	q := req.URL.Query()
	q.Add("job_id", jobId)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	return helper.ParseProCatAPIResponse[dataMultipleLoanResponse](resp)
}

func (repo *repository) CallMultipleLoan90Days(apiKey, jobId, memberId, companyId string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error) {
	url := fmt.Sprintf("%s/product/compliance/multiple-loan/90-days", repo.cfg.App.ProductCatalogHost)

	bodyBytes, err := repo.marshalFn(reqBody)
	if err != nil {
		return nil, errors.New(constant.ErrInvalidRequestPayload)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XAPIKey, apiKey)
	req.Header.Set(constant.XMemberId, memberId)
	req.Header.Set(constant.XCompanyId, companyId)

	q := req.URL.Query()
	q.Add("job_id", jobId)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	return helper.ParseProCatAPIResponse[dataMultipleLoanResponse](resp)
}
