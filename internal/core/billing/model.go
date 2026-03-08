package billing

type usagePerProduct struct {
	ProductId    uint   `json:"product_id"`
	ProductName  string `json:"product_name"`
	TotalRequest int64  `json:"total_request"`
	TotalSuccess int64  `json:"total_success"`
	// TotalError   int64  `json:"total_error"`
	// TotalPay     int64  `json:"total_pay"`
	// TotalFree    int64  `json:"total_free"`
}

type companyUsageSummary struct {
	CompanyId         uint              `json:"company_id"`
	CompanyName       string            `json:"company_name"`
	PeriodYear        int               `json:"period_year"`
	PeriodMonth       int               `json:"period_month"`
	Products          []usagePerProduct `json:"products"`
	GrandTotalRequest int64             `json:"grand_total_request"`
	GrandTotalSuccess int64             `json:"grand_total_success"`
	// GrandTotalError   int64 `json:"grand_total_error"`
	// GrandTotalPay     int64 `json:"grand_total_pay"`
	// GrandTotalFree    int64 `json:"grand_total_free"`
}

type adminEmail struct {
	MemberId  uint   `json:"member_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CompanyId uint   `json:"company_id"`
}
