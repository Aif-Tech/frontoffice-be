package genretail

import (
	"time"
)

type genRetailRequest struct {
	Name     string `json:"name" validate:"required~Name cannot be empty."`
	IdCardNo string `json:"id_card_no" validate:"required~ID Card No cannot be empty., numeric~ID Card No is only number, length(16)~ID Card No must be 16 digit number."`
	PhoneNo  string `json:"phone_no" validate:"required~Phone number cannot be empty, phone~Phone No only allow number and (+)plus sign with minimum 10 maximum 15 digit.,bytelength(10|15)~Phone No only allow number and (+)plus sign with minimum 10 maximum 15 digit., indophone~invalid mobile phone number"`
	LoanNo   string `json:"loan_no" validate:"required~Loan No cannot be empty."`
}

type dataGenRetailV3 struct {
	TransactionId        string  `json:"transaction_id"`
	Name                 string  `json:"name"`
	IdCardNo             string  `json:"id_card_no"`
	PhoneNo              string  `json:"phone_no"`
	LoanNo               string  `json:"loan_no"`
	ProbabilityToDefault float64 `json:"probability_to_default"` //5 digit dibelakang koma
	Grade                string  `json:"grade"`
	Identity             string  `json:"identity"`
	Behavior             string  `json:"behavior"`
	Date                 string  `json:"date"` // 2022-03-22 12:30:22
}

type gradesResponseData struct {
	Logs []*logTransScoreezy `json:"logs"`
}

type filterLogs struct {
	TrxId       string
	StartDate   string
	EndDate     string
	JobId       string
	CompanyId   string
	Name        string
	Grade       string
	ProductType string
	Page        string
	Size        string
}

type genRetailV3ClientReturnSuccess struct {
	Message string           `json:"message"`
	Success bool             `json:"success"`
	Data    *dataGenRetailV3 `json:"data"`
}

type logTransScoreezy struct {
	LogTrxId  uint                  `json:"log_trx_id"`
	CompanyId uint                  `json:"company_id"`
	Data      *dataLogTransScoreezy `json:"data" swaggertype:"object"`
	CreatedAt time.Time             `json:"created_at" format:"date-time"`
}

type dataLogTransScoreezy struct {
	TrxId                string `json:"trx_id"`
	Type                 string `json:"type"`
	Data                 *data  `json:"data"`
	ProbabilityToDefault string `json:"probability_to_default"`
	Grade                string `json:"grade"`
	Behavior             string `json:"behavior"`
	Identity             string `json:"identity"`
	Message              string `json:"message"`
	Status               string `json:"status"`  // Free or Pay
	Success              bool   `json:"success"` // true or false
}

type data struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_no"`
	IdCardNo    string `json:"id_card_no"`
	LoanNo      string `json:"loan_no"`
}

type genRetailContext struct {
	MemberId  uint              `json:"member_id"`
	CompanyId uint              `json:"company_id"`
	ProductId uint              `json:"product_id"`
	JobId     uint              `json:"job_id"`
	Request   *genRetailRequest `json:"request"`
}
