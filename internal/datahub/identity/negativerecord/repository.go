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
				CompanyName:         "KOPERASI SIMPAN PINJAM ARTHA MULIA",
				CaseStatus:          "Penyerahan Memori Kasasi",
				Court:               "PENGADILAN NEGERI JAKARTA SELATAN",
				CaseNumber:          "411/Pdt.G/2025/PN JKT.SEL",
				CaseCodeDescription: "Perkara Perdata Gugatan",
				PartyStatus:         "Turut Tergugat",
				CaseClassification:  "Wanprestasi",
				RegistrationDate:    "2025-04-28",
				CaseDuration:        "165 Hari",
				SimilarityScore:     "100.0",
			},
			{
				CompanyName:         "ARTHA PRIMA FINANCE",
				CaseStatus:          "Minutasi",
				Court:               "PENGADILAN NEGERI BANDUNG",
				CaseNumber:          "203/Pdt.G/2024/PN BDG",
				CaseCodeDescription: "Perkara Perdata Gugatan",
				PartyStatus:         "Tergugat",
				CaseClassification:  "Perbuatan Melawan Hukum",
				RegistrationDate:    "2024-11-15",
				CaseDuration:        "210 Hari",
				SimilarityScore:     "87.5",
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
