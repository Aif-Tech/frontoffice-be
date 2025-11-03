package npwpverification

type npwpVerificationRequest struct {
	Npwp   string `json:"npwp" validate:"required~NPWP tidak boleh kosong., numeric~NPWP hanya berupa angka., length(16)~NPWP 15 digit tidak berlaku. Untuk pribadi gunakan NIK. Bila badan atau perusahaan tambahkan angka 0 di depan."`
	LoanNo string `json:"loan_no" validate:"required~Loan No cannot be empty."`
}

type npwpVerificationRespData struct {
	Name string `json:"nama"`
}
