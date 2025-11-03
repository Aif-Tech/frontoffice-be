package npwpverification

type npwpVerificationRequest struct {
	NPWP string `json:"npwp"`
}

type npwpVerificationRespData struct {
	Name string `json:"nama"`
}
