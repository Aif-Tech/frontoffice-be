package recyclenumber

import (
	"encoding/json"
	"front-office/configs/application"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
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
	RecycleNumberAPI(apiKey, jobId, memberId, companyId string, payload *recycleNumberRequest) (*model.ProCatAPIResponse[dataRecycleNumberAPI], error)
}

func (repo *repository) RecycleNumberAPI(apiKey, jobId, memberId, companyId string, payload *recycleNumberRequest) (*model.ProCatAPIResponse[dataRecycleNumberAPI], error) {
	status := "phone number never happens recycled"
	if payload.Phone == "08111111110" {
		status = "phone number has never been recycled"
	}

	return &model.ProCatAPIResponse[dataRecycleNumberAPI]{
		Success: true,
		Data: dataRecycleNumberAPI{
			Status: status,
		},
		Input: recycleNumberRequest{
			Phone:     payload.Phone,
			LoanNo:    payload.LoanNo,
			Timestamp: time.Now().Format(constant.FormatYYYYMMDD),
			Period:    payload.Period,
		},
		Message:         "Succeed to Request Data",
		StatusCode:      http.StatusOK,
		PricingStrategy: "FREE",
		TransactionId:   "dummy-trx-id",
		Date:            time.Now().Format(constant.FormatYYYYMMDD),
	}, nil
}
