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
	RecycleNumberAPI(apiKey string, jobId string, memberId string, companyId string, payload *recycleNumberRequest) (*model.ProCatAPIResponse[dataRecycleNumberAPI], error)
}

func (repo *repository) RecycleNumberAPI(apiKey string, jobId string, memberId string, companyId string, payload *recycleNumberRequest) (*model.ProCatAPIResponse[dataRecycleNumberAPI], error) {
	return &model.ProCatAPIResponse[dataRecycleNumberAPI]{
		Success: true,
		Data: dataRecycleNumberAPI{
			Status: "phone number has never been recycled",
		},
		Input: recycleNumberRequest{
			Phone:  payload.Phone,
			LoanNo: payload.LoanNo,
		},
		Message:         "Succeed to Request Data",
		StatusCode:      http.StatusOK,
		PricingStrategy: "FREE",
		TransactionId:   "dummy-trx-id",
		Date:            time.Now().Format(constant.FormatYYYYMMDD),
	}, nil
}
