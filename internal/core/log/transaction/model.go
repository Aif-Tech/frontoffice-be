package transaction

import (
	"front-office/internal/core/member"
	"time"

	"gorm.io/datatypes"
)

type LogTransScoreezy struct {
	LogTrxId             uint             `json:"log_trx_id" gorm:"primaryKey;autoIncrement"`
	TrxId                string           `json:"trx_id"`
	MemberId             uint             `json:"member_id"`
	Member               member.MstMember `json:"-"`
	CompanyId            uint             `json:"company_id"`
	IpClient             string           `json:"ip_client"`
	ProductId            uint             `json:"product_id"`
	JobId                uint             `json:"job_id"`
	Status               string           `json:"status"`  // Free or Pay
	Success              bool             `json:"success"` // true or false
	Message              string           `json:"message"`
	ProbabilityToDefault string           `json:"probability_to_default"`
	Grade                string           `json:"grade"`
	LoanNo               string           `json:"loan_no"`
	Data                 datatypes.JSON   `json:"data" swaggertype:"object"`
	Duration             time.Duration    `json:"duration" format:"duration" example:"2h30m"`
	CreatedAt            time.Time        `json:"created_at" format:"date-time"`
}

type LogTransProductCatalog struct {
	LogTrxID        uint           `json:"log_trx_id" gorm:"primaryKey;autoIncrement"`
	TransactionID   string         `json:"transaction_id"`
	MemberID        uint           `json:"member_id"`
	CompanyID       uint           `json:"company_id"`
	IpClient        string         `json:"ip_client"`
	JobID           *uint          `json:"job_id"`
	ProductID       uint           `json:"product_id"`
	ProductGroupID  uint           `json:"product_group_id"`
	RequestBody     datatypes.JSON `json:"request_body" swaggertype:"object"`
	ResponseBody    datatypes.JSON `json:"response_body" swaggertype:"object"`
	Data            datatypes.JSON `json:"data" swaggertype:"object"`
	Status          int            `json:"status"`
	Success         bool           `json:"success"` // true or false
	Message         string         `json:"message"`
	PricingStrategy string         `json:"pricing_strategy"` //PAY , FREE
	Notes           string         `json:"notes"`
	Duration        time.Duration  `json:"duration" format:"duration" example:"2h30m"`
	RequestTime     time.Time      `json:"request_time" format:"date-time"`
	ResponseTime    time.Time      `json:"response_time" format:"date-time"`
	CreatedAt       time.Time      `json:"created_at" format:"date-time"`
}

type LogTransProCatRequest struct {
	TransactionID   string        `json:"transaction_id"`
	MemberID        uint          `json:"member_id"`
	CompanyID       uint          `json:"company_id"`
	IpClient        string        `json:"ip_client"`
	JobID           uint          `json:"job_id"`
	ProductID       uint          `json:"product_id"`
	ProductGroupID  uint          `json:"product_group_id"`
	RequestBody     any           `json:"request_body" swaggertype:"object"`
	ResponseBody    *ResponseBody `json:"response_body" swaggertype:"object"`
	Data            any           `json:"data" swaggertype:"object"`
	Status          int           `json:"status"`
	Success         bool          `json:"success"` // true or false
	Message         string        `json:"message"`
	PricingStrategy string        `json:"pricing_strategy"` //PAY , FREE
	Notes           string        `json:"notes"`
	Duration        time.Duration `json:"duration" format:"duration" example:"2h30m"`
	RequestTime     time.Time     `json:"request_time" format:"date-time"`
	ResponseTime    time.Time     `json:"response_time" format:"date-time"`
}

type ResponseBody struct {
	Data            any    `json:"data"`
	Input           any    `json:"input"`
	TransactionId   string `json:"transaction_id"`
	PricingStrategy string `json:"pricing_strategy"`
	DateTime        string `json:"datetime"`
}

type scoreezyLogResponse struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Grade     string    `json:"grade"`
	CreatedAt time.Time `json:"created_at"`
}

type getProcessedCountResp struct {
	ProcessedCount uint `json:"processed_count"`
}

type UpdateTransRequest struct {
	Success *bool `json:"success"`
}
