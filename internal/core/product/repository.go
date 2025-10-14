package product

import (
	"context"
	"errors"
	"fmt"
	"front-office/configs/application"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"front-office/pkg/httpclient"
	"net/http"
	"time"
)

func NewRepository(cfg *application.Config, client httpclient.HTTPClient) Repository {
	return &repository{
		cfg:    cfg,
		client: client,
	}
}

type repository struct {
	cfg    *application.Config
	client httpclient.HTTPClient
}

type Repository interface {
	GetProductAPI(slug string) (*productResponseData, error)
}

func (repo *repository) GetProductAPI(slug string) (*productResponseData, error) {
	url := fmt.Sprintf("%s/api/core/product/slug/%s", repo.cfg.Env.AifcoreHost, slug)

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

	apiResp, err := helper.ParseAifcoreAPIResponse[*productResponseData](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}
