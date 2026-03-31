package billing

import (
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
	GetUsageReport() ([]usageSummary, error)
	GetUsageReportByCompany(companyId, pricingStrategy, month, year string) (*usageSummary, error)
	GetAdminsData(companyId uint) ([]adminEmail, error)
}

func (repo *repository) GetUsageReport() ([]usageSummary, error) {
	url := fmt.Sprintf(`%v/api/core/billing/summaries`, repo.cfg.App.AifcoreHost)

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

	apiResp, err := helper.ParseAifcoreAPIResponse[[]usageSummary](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) GetUsageReportByCompany(companyId, pricingStrategy, month, year string) (*usageSummary, error) {
	url := fmt.Sprintf(`%v/api/core/billing/summary`, repo.cfg.App.AifcoreHost)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	q := req.URL.Query()
	q.Add("company_id", companyId)
	q.Add("pricing_strategy", pricingStrategy)
	q.Add("month", month)
	q.Add("year", year)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[*usageSummary](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) GetAdminsData(companyId uint) ([]adminEmail, error) {
	url := fmt.Sprintf(`%v/api/core/billing/admins/%d`, repo.cfg.App.AifcoreHost, companyId)

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

	apiResp, err := helper.ParseAifcoreAPIResponse[[]adminEmail](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}
