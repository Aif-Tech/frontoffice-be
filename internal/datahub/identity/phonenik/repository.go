package phonenik

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
	PhoneToNIKAPI(apiKey, trxId string, payload *phoneNIKRequest) (*model.ProCatAPIResponse[dataPhoneNIKAPI], error)
}

func (repo *repository) PhoneToNIKAPI(apiKey, trxId string, payload *phoneNIKRequest) (*model.ProCatAPIResponse[dataPhoneNIKAPI], error) {
	status := "not match"
	if payload.Phone == "08111111110" && payload.NIK == "3576014403910003" {
		status = "match"
	}

	return &model.ProCatAPIResponse[dataPhoneNIKAPI]{
		Success: true,
		Data: dataPhoneNIKAPI{
			Status: status,
		},
		Input: phoneNIKRequest{
			Phone:  payload.Phone,
			NIK:    payload.NIK,
			LoanNo: payload.LoanNo,
		},
		Message:         "Succeed to Request Data",
		StatusCode:      http.StatusOK,
		PricingStrategy: "FREE",
		TransactionId:   trxId,
		Date:            time.Now().Format(constant.FormatYYYYMMDD),
	}, nil
}
