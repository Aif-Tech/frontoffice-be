package taxscore

type taxScoreRequest struct {
	Npwp   string `json:"npwp" validate:"required~NPWP tidak boleh kosong., numeric~NPWP hanya berupa angka., length(16)~NPWP 15 digit tidak berlaku. Untuk pribadi gunakan NIK. Bila badan atau perusahaan tambahkan angka 0 di depan."`
	LoanNo string `json:"loan_no" validate:"required~Loan No cannot be empty."`
}

type taxScoreRespData struct {
	Nama   string `json:"nama"`
	Alamat string `json:"alamat"`
	Score  string `json:"score"`
	Status string `json:"status"`
}

type taxScoreContext struct {
	APIKey         string           `json:"api_key"`
	JobIdStr       string           `json:"job_id_str"`
	MemberIdStr    string           `json:"member_id_str"`
	CompanyIdStr   string           `json:"company_id_str"`
	MemberId       uint             `json:"member_id"`
	CompanyId      uint             `json:"company_id"`
	ProductId      uint             `json:"product_id"`
	ProductGroupId uint             `json:"product_group_id"`
	JobId          uint             `json:"job_id"`
	Request        *taxScoreRequest `json:"request"`
}
