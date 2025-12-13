package recyclenumber

type recycleNumberRequest struct {
	Phone  string `json:"phone_number" validate:"required~Phone Number cannot be empty, indophone, min(9)"`
	LoanNo string `json:"loan_no" validate:"required~Loan No cannot be empty."`
}

type dataRecycleNumberAPI struct {
	Status string `json:"status"`
}

type recycleNumberContext struct {
	APIKey         string                `json:"api_key"`
	JobIdStr       string                `json:"job_id_str"`
	MemberIdStr    string                `json:"member_id_str"`
	CompanyIdStr   string                `json:"company_id_str"`
	MemberId       uint                  `json:"member_id"`
	CompanyId      uint                  `json:"company_id"`
	ProductId      uint                  `json:"product_id"`
	ProductGroupId uint                  `json:"product_group_id"`
	JobId          uint                  `json:"job_id"`
	Request        *recycleNumberRequest `json:"request"`
}
