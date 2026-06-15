package negativerecord

type negativeRecordRequest struct {
	LoanNo      string `json:"loan_no" validate:"required~loan number cannot be empty."`
	CompanyName string `json:"company_name" validate:"required~company name cannot be empty"`
}

type dataNegativeRecord struct {
	Result []dataNegativeRecordAPI `json:"result"`
}

type dataNegativeRecordAPI struct {
	CompanyName         string `json:"company_name"`
	CaseStatus          string `json:"case_status"`
	Court               string `json:"court"`
	CaseNumber          string `json:"case_number"`
	CaseCodeDescription string `json:"case_code_description"`
	PartyStatus         string `json:"party_status"`
	CaseClassification  string `json:"case_classification"`
	RegistrationDate    string `json:"registration_date"`
	CaseDuration        string `json:"case_duration"`
	SimilarityScore     string `json:"similarity_score"`
}

type negativeRecordContext struct {
	APIKey         string                 `json:"api_key"`
	JobIdStr       string                 `json:"job_id_str"`
	MemberIdStr    string                 `json:"member_id_str"`
	CompanyIdStr   string                 `json:"company_id_str"`
	MemberId       uint                   `json:"member_id"`
	CompanyId      uint                   `json:"company_id"`
	ProductId      uint                   `json:"product_id"`
	ProductGroupId uint                   `json:"product_group_id"`
	JobId          uint                   `json:"job_id"`
	Request        *negativeRecordRequest `json:"request"`
}
