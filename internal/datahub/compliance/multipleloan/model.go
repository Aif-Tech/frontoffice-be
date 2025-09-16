package multipleloan

type multipleLoanRequest struct {
	Nik   string `json:"nik" validate:"required~NIK cannot be empty., numeric~ID Card No is only number, length(16)~ID Card No must be 16 digit number."`
	Phone string `json:"phone_number" validate:"required~Phone Number cannot be empty, indophone, min(9)"`
}

type dataMultipleLoanResponse struct {
	QueryCount uint `json:"query_count"`
}

type multipleLoanContext struct {
	APIKey         string               `json:"api_key"`
	JobIdStr       string               `json:"job_id_str"`
	MemberIdStr    string               `json:"member_id_str"`
	CompanyIdStr   string               `json:"company_id_str"`
	ProductSlug    string               `json:"product_slug"`
	MemberId       uint                 `json:"member_id"`
	CompanyId      uint                 `json:"company_id"`
	ProductId      uint                 `json:"product_id"`
	ProductGroupId uint                 `json:"product_group_id"`
	JobId          uint                 `json:"job_id"`
	Request        *multipleLoanRequest `json:"request"`
}
