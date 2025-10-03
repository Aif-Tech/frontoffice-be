package loanrecordchecker

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
	LoanRecordCheckerAPI(apiKey, jobId, memberId, companyId string, payload *loanRecordCheckerRequest) (*model.ProCatAPIResponse[dataLoanRecord], error)
}

func (repo *repository) LoanRecordCheckerAPI(apiKey, jobId, memberId, companyId string, payload *loanRecordCheckerRequest) (*model.ProCatAPIResponse[dataLoanRecord], error) {
	url := fmt.Sprintf("%s/product/compliance/loan-record-checker", repo.cfg.Env.ProductCatalogHost)

	bodyBytes, err := repo.marshalFn(payload)
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

	return helper.ParseProCatAPIResponse[dataLoanRecord](resp)
}
