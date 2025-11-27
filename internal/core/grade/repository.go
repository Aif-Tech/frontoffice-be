package grade

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"front-office/configs/application"
	"front-office/pkg/common/constant"
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
	SaveGradingAPI(payload *createGradePayload) error
	GetGradesAPI(productSlug, companyId string) (*gradesResponseData, error)
}

func (repo *repository) SaveGradingAPI(payload *createGradePayload) error {
	url := fmt.Sprintf("%s/api/core/product/%s/grades", repo.cfg.App.AifcoreHost, payload.ProductSlug)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bodyBytes, err := repo.marshalFn(payload.Request)
	if err != nil {
		return errors.New(constant.ErrInvalidRequestPayload)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XCompanyId, payload.CompanyId)

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

func (repo *repository) GetGradesAPI(productSlug, companyId string) (*gradesResponseData, error) {
	url := fmt.Sprintf("%s/api/core/product/%s/grades", repo.cfg.App.AifcoreHost, productSlug)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XCompanyId, companyId)

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrUpstreamUnavailable)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[*gradesResponseData](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}
