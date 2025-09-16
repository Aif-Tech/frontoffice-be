package job

import (
	"time"
)

type logTransProductCatalog struct {
	MemberID               uint           `json:"member_id"`
	CompanyID              uint           `json:"company_id"`
	JobID                  uint           `json:"job_id"`
	ProductID              uint           `json:"product_id"`
	Status                 string         `json:"status"`
	Message                *string        `json:"message"`
	Input                  *logTransInput `json:"input"`
	Data                   *logTransData  `json:"data"`
	PricingStrategy        string         `json:"pricing_strategy"`
	TransactionId          string         `json:"transaction_id"`
	DateTime               string         `json:"datetime"`
	RefTransProductCatalog any            `json:"ref_trans_product_catalog"`
}

type refTransProductCatalog struct {
	Data struct {
		NPWP string `json:"npwp"`
	}
	Input struct {
		Name        string `json:"name"`
		NIK         string `json:"nik"`
		PhoneNumber string `json:"phone_number"`
		NPWP        string `json:"npwp"`
		NPWPOrNIK   string `json:"npwp_or_nik"`
	} `json:"input"`
}

type logTransData struct {
	Remarks          *string `json:"remarks,omitempty"`
	Status           *string `json:"status,omitempty"`
	QueryCount       *int    `json:"query_count,omitempty"`
	Nama             *string `json:"nama,omitempty"`
	Score            *string `json:"score,omitempty"`
	Alamat           *string `json:"alamat,omitempty"`
	NPWP             *string `json:"npwp,omitempty"`
	NPWPVerification *string `json:"npwp_verification,omitempty"`
	TaxCompliance    *string `json:"tax_compliance,omitempty"`
}

type logTransInput struct {
	Name        *string `json:"name,omitempty"`
	NIK         *string `json:"nik,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	NPWP        *string `json:"npwp,omitempty"`
	NPWPOrNIK   *string `json:"npwp_or_nik,omitempty"`
}

type jobDetailResponse struct {
	TotalData                  int64                     `json:"total_data"`
	TotalDataPercentageSuccess int64                     `json:"total_data_percentage_success"`
	TotalDataPercentageFail    int64                     `json:"total_data_percentage_fail"`
	TotalDataPercentageError   int64                     `json:"total_data_percentage_error"`
	JobDetails                 []*logTransProductCatalog `json:"job_details"`
}

type CreateJobRequest struct {
	ProductId uint   `json:"product_id" validate:"required~Field product id is required"`
	MemberId  string `json:"member_id" validate:"required~Field member id is required"`
	CompanyId string `json:"company_id" validate:"required~Field company id is required"`
	Total     int    `json:"total" validate:"required~Field total is required"`
}

type UpdateJobRequest struct {
	SuccessCount *uint      `json:"success_count"`
	Status       *string    `json:"status"`
	EndAt        *time.Time `json:"end_at"`
}

type createJobRespData struct {
	JobId     uint `json:"id"`
	MemberId  uint `json:"member_id"`
	CompanyId uint `json:"company_id"`
}

type logFilter struct {
	Page        string
	Size        string
	Offset      string
	StartDate   string
	EndDate     string
	JobId       string
	ProductSlug string
	MemberId    string
	CompanyId   string
	TierLevel   string
	IsMasked    bool
	Keyword     string
}
