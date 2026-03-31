package transaction

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"net/http"
	"time"
)

func (repo *repository) CreateLogTransAPI(payload *LogTransProCatRequest) error {
	url := fmt.Sprintf("%s/api/core/logging/transaction/product-catalog", repo.cfg.App.AifcoreHost)

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

func (repo *repository) ProcessedLogCountAPI(jobId string) (*getProcessedCountResp, error) {
	url := fmt.Sprintf("%s/api/core/logging/transaction/product-catalog/%s/processed_count", repo.cfg.App.AifcoreHost, jobId)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[*getProcessedCountResp](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) GetLogTransByCompanyAPI(jobId, productId, companyId, pricingStrategy, productSlug, applyDedup string) ([]*LogTransProductCatalog, error) {
	url := fmt.Sprintf("%s/api/core/logging/transaction/product-catalog/by-company", repo.cfg.App.AifcoreHost)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XCompanyId, companyId)

	q := req.URL.Query()
	q.Add("job_id", jobId)
	q.Add("product_id", productId)
	q.Add("pricing_strategy", pricingStrategy)
	q.Add("product_slug", productSlug)
	q.Add("apply_dedup", applyDedup)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[[]*LogTransProductCatalog](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) UpdateLogTransAPI(transId string, payload map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/core/logging/transaction/product-catalog/%s", repo.cfg.App.AifcoreHost, transId)

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

	_, err = helper.ParseAifcoreAPIResponse[*any](resp)
	if err != nil {
		return err
	}

	return nil
}
