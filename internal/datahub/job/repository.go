package job

import (
	"bytes"
	"context"
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
	"time"
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
	CreateJobAPI(payload *CreateJobRequest) (*createJobRespData, error)
	UpdateJobAPI(jobId string, req map[string]interface{}) error
	GetJobsAPI(filter *logFilter) (*model.AifcoreAPIResponse[any], error)
	GetJobDetailAPI(filter *logFilter) (*model.AifcoreAPIResponse[*jobDetailResponse], error)
	GetJobsSummaryAPI(filter *logFilter) (*model.AifcoreAPIResponse[*jobDetailResponse], error)
}

func (repo *repository) CreateJobAPI(payload *CreateJobRequest) (*createJobRespData, error) {
	url := fmt.Sprintf("%s/api/core/product/jobs", repo.cfg.Env.AifcoreHost)

	bodyBytes, err := repo.marshalFn(payload)
	if err != nil {
		return nil, errors.New(constant.ErrInvalidRequestPayload)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XMemberId, payload.MemberId)
	req.Header.Set(constant.XCompanyId, payload.CompanyId)

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[*createJobRespData](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) UpdateJobAPI(jobId string, payload map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/core/product/jobs/%s", repo.cfg.Env.AifcoreHost, jobId)

	bodyBytes, err := repo.marshalFn(payload)
	if err != nil {
		return errors.New(constant.ErrInvalidRequestPayload)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	resp, err := repo.client.Do(req)
	if err != nil {
		return errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	_, err = helper.ParseAifcoreAPIResponse[*createJobRespData](resp)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) GetJobsAPI(filter *logFilter) (*model.AifcoreAPIResponse[any], error) {
	url := fmt.Sprintf("%s/api/core/product/%s/jobs", repo.cfg.Env.AifcoreHost, filter.ProductSlug)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XMemberId, filter.MemberId)
	req.Header.Set(constant.XCompanyId, filter.CompanyId)
	req.Header.Set(constant.XTierLevel, filter.TierLevel)

	q := req.URL.Query()
	q.Add(constant.Page, filter.Page)
	q.Add(constant.Size, filter.Size)
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

func (repo *repository) GetJobDetailAPI(filter *logFilter) (*model.AifcoreAPIResponse[*jobDetailResponse], error) {
	url := fmt.Sprintf("%s/api/core/product/%s/jobs/%s", repo.cfg.Env.AifcoreHost, filter.ProductSlug, filter.JobId)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XMemberId, filter.MemberId)
	req.Header.Set(constant.XCompanyId, filter.CompanyId)

	q := req.URL.Query()
	q.Add(constant.Page, filter.Page)
	q.Add(constant.Size, filter.Size)
	q.Add(constant.Keyword, filter.Keyword)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	return helper.ParseAifcoreAPIResponse[*jobDetailResponse](resp)
}

func (repo *repository) GetJobsSummaryAPI(filter *logFilter) (*model.AifcoreAPIResponse[*jobDetailResponse], error) {
	url := fmt.Sprintf("%s/api/core/product/%s/jobs-summary", repo.cfg.Env.AifcoreHost, filter.ProductSlug)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XMemberId, filter.MemberId)
	req.Header.Set(constant.XCompanyId, filter.CompanyId)

	q := req.URL.Query()
	q.Add(constant.Keyword, filter.Keyword)
	q.Add(constant.Page, filter.Page)
	q.Add(constant.Size, filter.Size)
	q.Add(constant.StartDate, filter.StartDate)
	q.Add(constant.EndDate, filter.EndDate)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	return helper.ParseAifcoreAPIResponse[*jobDetailResponse](resp)
}
