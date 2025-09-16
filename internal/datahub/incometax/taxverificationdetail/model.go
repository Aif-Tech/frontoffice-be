package taxverificationdetail

type taxVerificationRequest struct {
	NpwpOrNik string `json:"npwp_or_nik" validate:"required~NPWP or NIK cannot be empty., numeric~NPWP is only number., length(16)~NPWP 15 digit tidak berlaku. Untuk pribadi gunakan NIK. Bila badan atau perusahaan tambahkan angka 0 di depan."`
}

type taxVerificationRespData struct {
	Nama             string `json:"nama"`
	Alamat           string `json:"alamat"`
	NPWP             string `json:"npwp"`
	NPWPVerification string `json:"npwp_verification"`
	TaxCompliance    string `json:"tax_compliance"`
	Status           string `json:"status"`
}

type taxVerificationContext struct {
	APIKey         string                  `json:"api_key"`
	JobIdStr       string                  `json:"job_id_str"`
	MemberIdStr    string                  `json:"member_id_str"`
	CompanyIdStr   string                  `json:"company_id_str"`
	MemberId       uint                    `json:"member_id"`
	CompanyId      uint                    `json:"company_id"`
	ProductId      uint                    `json:"product_id"`
	ProductGroupId uint                    `json:"product_group_id"`
	JobId          uint                    `json:"job_id"`
	Request        *taxVerificationRequest `json:"request"`
}
