package genretail

import (
	"front-office/internal/core/company"
	"front-office/internal/core/member"
	"time"

	"gorm.io/gorm"
)

type genRetailRequest struct {
	Name     string `json:"name" validate:"required~Name cannot be empty."`
	IdCardNo string `json:"id_card_no" validate:"required~ID Card No cannot be empty., numeric~ID Card No is only number, length(16)~ID Card No must be 16 digit number."`
	PhoneNo  string `json:"phone_no" validate:"required~Phone number cannot be empty, phone~Phone No only allow number and (+)plus sign with minimum 10 maximum 15 digit.,bytelength(10|15)~Phone No only allow number and (+)plus sign with minimum 10 maximum 15 digit."`
	LoanNo   string `json:"loan_no" validate:"required~Loan No cannot be empty."`
}

type GenRetailV3ModelResponse struct {
	Message      string           `json:"message"`
	ErrorMessage string           `json:"error_message"`
	Success      bool             `json:"success"`
	Data         *dataGenRetailV3 `json:"data"`
	StatusCode   int              `json:"status_code"`
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
	TrxId     string
	StartDate string
	EndDate   string
	CompanyId string
	Grade     string
	Size      string
}

type GenRetailV3ClientReturnSuccess struct {
	Message string           `json:"message"`
	Success bool             `json:"success"`
	Data    *dataGenRetailV3 `json:"data"`
}

type logTransScoreezy struct {
	LogTrxId             uint                  `json:"log_trx_id" gorm:"primaryKey;autoIncrement"`
	TrxId                string                `json:"trx_id"`
	CompanyId            uint                  `json:"company_id"`
	Status               string                `json:"status"`  // Free or Pay
	Success              bool                  `json:"success"` // true or false
	Message              string                `json:"message"`
	ProbabilityToDefault string                `json:"probability_to_default"`
	Grade                string                `json:"grade"`
	LoanNo               string                `json:"loan_no"`
	Data                 *dataLogTransScoreezy `json:"data" swaggertype:"object"`
	CreatedAt            time.Time             `json:"created_at" format:"date-time"`
}

type dataLogTransScoreezy struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_no"`
	IdCardNo    string `json:"id_card_no"`
	LoanNo      string `json:"loan_no"`
}

type genRetailContext struct {
	MemberId  uint              `json:"member_id"`
	CompanyId uint              `json:"company_id"`
	ProductId uint              `json:"product_id"`
	Request   *genRetailRequest `json:"request"`
}

type UploadScoringRequest struct {
	Files []byte `json:"files"`
}

type UploadScoringReturnError struct {
	Message string `json:"message"`
}

type BulkSearch struct {
	Id                   uint               `gorm:"primarykey;autoIncrement" json:"id"`
	UploadId             string             `gorm:"not null" json:"upload_id"`
	TransactionId        string             `gorm:"not null" json:"transaction_id"`
	Name                 string             `gorm:"not null" json:"name"`
	IdCardNo             string             `gorm:"not null" json:"id_card_no"`
	PhoneNo              string             `gorm:"not null" json:"phone_no"`
	LoanNo               string             `gorm:"not null" json:"loan_no"`
	ProbabilityToDefault float64            `gorm:"not null" json:"probability_to_default"`
	Grade                string             `gorm:"not null" json:"grade"`
	Date                 string             `gorm:"not null" json:"date"`
	Type                 string             `gorm:"not null" json:"type"`
	UserId               string             `gorm:"not null" json:"user_id"`
	User                 member.MstMember   `gorm:"foreignKey:UserId" json:"user"`
	CompanyId            string             `json:"company_id"`
	Company              company.MstCompany `gorm:"foreignKey:CompanyId" json:"company"`
	CreatedAt            time.Time          `json:"-"`
	UpdatedAt            time.Time          `json:"-"`
	DeletedAt            gorm.DeletedAt     `gorm:"index" json:"-"`
}

type BulkSearchRequest struct {
	LoanNo      string `json:"loan_no"`
	Name        string `json:"name"`
	NIK         string `json:"nik"`
	PhoneNumber string `json:"phone_number"`
}

type BulkSearchResponse struct {
	TransactionId        string  `json:"transaction_id"`
	Name                 string  `json:"name"`
	PIC                  string  `json:"pic"`
	IdCardNo             string  `json:"id_card_no"`
	PhoneNo              string  `json:"phone_no"`
	LoanNo               string  `json:"loan_no"`
	ProbabilityToDefault float64 `json:"probability_to_default"`
	Grade                string  `json:"grade"`
	Type                 string  `json:"type"`
	Date                 string  `json:"date"`
}
