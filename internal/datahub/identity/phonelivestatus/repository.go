package phonelivestatus

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
	PhoneLiveStatusAPI(apiKey, jobId string, payload *phoneLiveStatusRequest) (*model.ProCatAPIResponse[phoneLiveStatusRespData], error)
	GetPhoneLiveStatusJobAPI(filter *phoneLiveStatusFilter) (*jobListRespData, error)
	GetJobDetailsAPI(filter *phoneLiveStatusFilter) (*jobDetailRaw, error)
	GetJobsSummaryAPI(filter *phoneLiveStatusFilter) (*jobDetailRaw, error)
	GetJobMetricsAPI(filter *phoneLiveStatusFilter) (*jobMetrics, error)
}

func (repo *repository) PhoneLiveStatusAPI(apiKey, jobId string, payload *phoneLiveStatusRequest) (*model.ProCatAPIResponse[phoneLiveStatusRespData], error) {
	url := fmt.Sprintf("%s/product/identity/phone-live-status", repo.cfg.Env.ProductCatalogHost)

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

	q := req.URL.Query()
	q.Add("job_id", jobId)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	return helper.ParseProCatAPIResponse[phoneLiveStatusRespData](resp)
}

func (repo *repository) GetPhoneLiveStatusJobAPI(filter *phoneLiveStatusFilter) (*jobListRespData, error) {
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

	apiResp, err := helper.ParseAifcoreAPIResponse[*jobListRespData](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, err
}

func (repo *repository) GetJobDetailsAPI(filter *phoneLiveStatusFilter) (*jobDetailRaw, error) {
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
	req.Header.Set(constant.XTierLevel, filter.TierLevel)

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

	apiResp, err := helper.ParseAifcoreAPIResponse[*jobDetailRaw](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) GetJobsSummaryAPI(filter *phoneLiveStatusFilter) (*jobDetailRaw, error) {
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

	apiResp, err := helper.ParseAifcoreAPIResponse[*jobDetailRaw](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, err
}

func (repo *repository) GetJobMetricsAPI(filter *phoneLiveStatusFilter) (*jobMetrics, error) {
	url := fmt.Sprintf("%s/api/core/phone-live-status/job-metrics", repo.cfg.Env.AifcoreHost)

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
	q.Add(constant.JobId, filter.JobId)
	q.Add(constant.StartDate, filter.StartDate)
	q.Add(constant.EndDate, filter.EndDate)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[*jobMetrics](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, err
}
