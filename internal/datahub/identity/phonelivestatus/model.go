package phonelivestatus

import (
	"front-office/internal/core/company"
	"front-office/internal/core/member"
)

type mstPhoneLiveStatusJob struct {
	Id           uint               `json:"id"`
	Total        int                `json:"total"`
	SuccessCount int                `json:"success_count"`
	Status       string             `json:"status"`
	MemberId     uint               `json:"member_id"`
	Member       member.MstMember   `json:"-"`
	CompanyId    uint               `json:"company_id"`
	Company      company.MstCompany `json:"-"`
	CreatedAt    string             `json:"start_time"`
	EndAt        string             `json:"end_time"`
}

type mstPhoneLiveStatusJobDetail struct {
	MemberId         uint      `json:"member_id"`
	CompanyId        uint      `json:"company_id"`
	JobId            uint      `json:"job_id"`
	PhoneNumber      string    `json:"phone_number" validate:"required~phone number is required, min(10)~phone number must be at least 10 characters, indophone~invalid number"`
	InProgress       bool      `json:"in_progess"`
	Status           string    `json:"status"`
	Message          *string   `json:"message"`
	SubscriberStatus string    `json:"subscriber_status"`
	DeviceStatus     string    `json:"device_status"`
	PhoneType        string    `json:"phone_type"`
	Operator         string    `json:"operator"`
	PricingStrategy  string    `json:"pricing_strategy"`
	TransactionId    string    `json:"transaction_id"`
	CreatedAt        string    `json:"created_at"`
	RefLogTrx        RefLogTrx `json:"ref_log_trx"`
}

type RefLogTrx struct {
	PhoneNumber string `json:"phone_number"`
}

type phoneLiveStatusRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required~phone number is required, min(10)~phone number must be at least 10 characters, indophone~invalid number"`
	TrxId       string `json:"trx_id"`
	LoanNo      string `json:"loan_no" validate:"required~Loan No cannot be empty."`
}

type phoneLiveStatusRespData struct {
	LiveStatus string      `json:"live_status"`
	PhoneType  string      `json:"phone_type"`
	Operator   string      `json:"operator"`
	Errors     []errorData `json:"errors"`
}

type errorData struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

type phoneLiveStatusFilter struct {
	Page        string
	Size        string
	Offset      string
	StartDate   string
	EndDate     string
	ProductSlug string
	JobId       string
	MemberId    string
	CompanyId   string
	TierLevel   string
	Keyword     string
	Masked      bool
}

type jobListRespData struct {
	Jobs      []*mstPhoneLiveStatusJob `json:"jobs"`
	TotalData int                      `json:"total_data"`
}

type jobDetailsDTO struct {
	TotalData                  int64                          `json:"total_data"`
	TotalDataPercentageSuccess int64                          `json:"total_data_percentage_success"`
	TotalDataPercentageFail    int64                          `json:"total_data_percentage_fail"`
	TotalDataPercentageError   int64                          `json:"total_data_percentage_error"`
	SubsActive                 int64                          `json:"subs_active"`
	SubsDisconnected           int64                          `json:"subs_disconnected"`
	DevReachable               int64                          `json:"dev_reachable"`
	DevUnreachable             int64                          `json:"dev_unreachable"`
	DevUnavailable             int64                          `json:"dev_unavailable"`
	JobDetails                 []*mstPhoneLiveStatusJobDetail `json:"job_details"`
}

type jobsSummaryDTO struct {
	TotalData                  int64 `json:"total_data"`
	TotalDataPercentageSuccess int64 `json:"total_data_percentage_success"`
	TotalDataPercentageFail    int64 `json:"total_data_percentage_fail"`
	TotalDataPercentageError   int64 `json:"total_data_percentage_error"`
	SubsActive                 int64 `json:"subs_active"`
	SubsDisconnected           int64 `json:"subs_disconnected"`
	DevReachable               int64 `json:"dev_reachable"`
	DevUnreachable             int64 `json:"dev_unreachable"`
	DevUnavailable             int64 `json:"dev_unavailable"`
	Mobile                     int64 `json:"mobile"`
	FixedLine                  int64 `json:"fixed_line"`
}

type jobDetailRaw struct {
	TotalData                  int64                     `json:"total_data"`
	TotalDataPercentageSuccess int64                     `json:"total_data_percentage_success"`
	TotalDataPercentageFail    int64                     `json:"total_data_percentage_fail"`
	TotalDataPercentageError   int64                     `json:"total_data_percentage_error"`
	JobDetails                 []*logTransProductCatalog `json:"job_details"`
}

type jobMetrics struct {
	SubsActive       int64 `json:"subs_active"`
	SubsDisconnected int64 `json:"subs_disconnected"`
	DevReachable     int64 `json:"dev_reachable"`
	DevUnreachable   int64 `json:"dev_unreachable"`
	DevUnavailable   int64 `json:"dev_unavailable"`
	Mobile           int64 `json:"mobile"`
	FixedLine        int64 `json:"fixed_line"`
}

type logTransProductCatalog struct {
	MemberID               uint                   `json:"member_id"`
	CompanyID              uint                   `json:"company_id"`
	JobID                  uint                   `json:"job_id"`
	ProductID              uint                   `json:"product_id"`
	Status                 string                 `json:"status"`
	Message                *string                `json:"message"`
	Input                  *logTransInput         `json:"input"`
	Data                   *logTransData          `json:"data"`
	PricingStrategy        string                 `json:"pricing_strategy"`
	TransactionId          string                 `json:"transaction_id"`
	DateTime               string                 `json:"datetime"`
	RefTransProductCatalog RefTransProductCatalog `json:"ref_trans_product_catalog"`
}

type RefTransProductCatalog struct {
	Input logTransInput `json:"input"`
}

type logTransData struct {
	Operator   string `json:"operator"`
	PhoneType  string `json:"phone_type"`
	LiveStatus string `json:"live_status"`
}

type logTransInput struct {
	PhoneNumber string `json:"phone_number,omitempty"`
}

type phoneLiveStatusContext struct {
	APIKey         string                  `json:"api_key"`
	JobIdStr       string                  `json:"job_id_str"`
	MemberId       uint                    `json:"member_id"`
	CompanyId      uint                    `json:"company_id"`
	ProductId      uint                    `json:"product_id"`
	ProductGroupId uint                    `json:"product_group_id"`
	JobId          uint                    `json:"job_id"`
	Request        *phoneLiveStatusRequest `json:"request"`
}
