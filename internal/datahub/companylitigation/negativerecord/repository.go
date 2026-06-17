package negativerecord

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
	NegativeRecordAPI(apiKey, trxId string, payload *negativeRecordRequest) (*model.ProCatAPIResponse[dataNegativeRecord], error)
}

func (repo *repository) NegativeRecordAPI(apiKey, trxId string, payload *negativeRecordRequest) (*model.ProCatAPIResponse[dataNegativeRecord], error) {
	result := []dataNegativeRecordAPI{}
	if payload.CompanyName == "artha" {
		result = []dataNegativeRecordAPI{
			{
				CompanyName:      "KOPERASI SIMPAN PINJAM ARTHA MULIA",
				Status:           "Penyerahan Memori Kasasi",
				Court:            "PENGADILAN NEGERI JAKARTA SELATAN",
				Province:         "DKI Jakarta",
				CaseNumber:       "411/Pdt.G/2025/PN JKT.SEL",
				CaseType:         "Wanprestasi",
				RegistrationDate: "2025-04-28",
				ProcessDuration:  "165 Hari",
				LastUpdated:      time.Now().Format("02 January 2006 15:04") + " WIB",
				SimilarityScore:  "100.0",
			},
			{
				CompanyName:      "ARTHA PRIMA FINANCE",
				Status:           "Minutasi",
				Court:            "PENGADILAN NEGERI BANDUNG",
				Province:         "Jawa Barat",
				CaseNumber:       "203/Pdt.G/2024/PN BDG",
				CaseType:         "Perbuatan Melawan Hukum",
				RegistrationDate: "2024-11-15",
				ProcessDuration:  "210 Hari",
				LastUpdated:      time.Now().Format("02 January 2006 15:04") + " WIB",
				SimilarityScore:  "90.5",
			},
		}
	}

	return &model.ProCatAPIResponse[dataNegativeRecord]{
		Success: true,
		Data: dataNegativeRecord{
			Result: result,
		},
		Input: negativeRecordRequest{
			CompanyName: payload.CompanyName,
			LoanNo:      payload.LoanNo,
		},
		Message:         "Succeed to Request Data",
		StatusCode:      http.StatusOK,
		PricingStrategy: "FREE",
		TransactionId:   trxId,
		Date:            time.Now().Format(constant.FormatYYYYMMDD),
	}, nil
}
