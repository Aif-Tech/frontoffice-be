package taxcompliancestatus

import (
	"bytes"
	"encoding/json"
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
	TaxComplianceStatusAPI(apiKey, jobId string, request *taxComplianceStatusRequest) (*model.ProCatAPIResponse[taxComplianceRespData], error)
}

func (repo *repository) TaxComplianceStatusAPI(apiKey, jobId string, reqBody *taxComplianceStatusRequest) (*model.ProCatAPIResponse[taxComplianceRespData], error) {
	url := fmt.Sprintf("%s/product/incometax/tax-compliance-status", repo.cfg.Env.ProductCatalogHost)

	bodyBytes, err := repo.marshalFn(reqBody)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgMarshalReqBody, err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XAPIKey, apiKey)

	q := req.URL.Query()
	q.Add("job_id", jobId)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	return helper.ParseProCatAPIResponse[taxComplianceRespData](resp)
}
