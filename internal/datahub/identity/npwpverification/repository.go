package npwpverification

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
		cfg,
		client,
		marshalFn,
	}
}

type repository struct {
	cfg       *application.Config
	client    httpclient.HTTPClient
	marshalFn jsonutil.Marshaller
}

type Repository interface {
	NPWPVerificationAPI(apiKey, jobId string, payload *npwpVerificationRequest) (*model.ProCatAPIResponse[npwpVerificationRespData], error)
}

func (repo *repository) NPWPVerificationAPI(apiKey, jobId string, payload *npwpVerificationRequest) (*model.ProCatAPIResponse[npwpVerificationRespData], error) {
	url := fmt.Sprintf("%s/product/identity/npwp-verification", repo.cfg.App.ProductCatalogHost)

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

	return helper.ParseProCatAPIResponse[npwpVerificationRespData](resp)
}
