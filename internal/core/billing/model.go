package billing

import "front-office/internal/core/log/transaction"

type usagePerProduct struct {
	ProductId    uint   `json:"product_id"`
	ProductSlug  string `json:"product_slug"`
	ProductName  string `json:"product_name"`
	TotalRequest int64  `json:"total_request"`
	TotalPay     int64  `json:"total_pay"`
}

type companyUsageSummary struct {
	CompanyId         uint              `json:"company_id"`
	CompanyName       string            `json:"company_name"`
	PeriodYear        int               `json:"period_year"`
	PeriodMonth       int               `json:"period_month"`
	Products          []usagePerProduct `json:"products"`
	GrandTotalRequest int64             `json:"grand_total_request"`
	GrandTotalPay     int64             `json:"grand_total_pay"`
}

type adminEmail struct {
	MemberId  uint   `json:"member_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CompanyId uint   `json:"company_id"`
	Key       string `json:"key"`
}

type XlsxReportProduct struct {
	ProductId    uint
	ProductName  string
	ProductSlug  string
	TotalRequest int64
	TotalPay     int64
}

type XlsxReportInput struct {
	CompanyId   uint
	CompanyName string
	PeriodYear  int
	PeriodMonth int
	Products    []XlsxReportProduct
	Password    string
}

type ColumnType int

const (
	ColTypeText ColumnType = iota
	ColTypeNumber
	ColTypeDate
	ColTypeDateTime
)

type ColumnDef struct {
	Header    string
	Type      ColumnType
	Width     float64
	ExtractFn ExtractFn
}

type ExtractFn func(log *transaction.LogTransProductCatalog) interface{}

type ProductSheetDef struct {
	ProductName string
	SheetName   string
	Columns     []ColumnDef
}

type RowData map[string]interface{}

type ProductData struct {
	Def  ProductSheetDef
	Rows []RowData
}

var ProductTaxVerification = ProductSheetDef{
	ProductName: "Tax Verification Detail",
	Columns: []ColumnDef{
		{Header: "Transaction ID", Type: ColTypeText, Width: 32, ExtractFn: ExtractTransactionID},
		{Header: "Product Name", Type: ColTypeText, Width: 18, ExtractFn: staticVal("Tax Verification Detail")},
		{Header: "NPWP", Type: ColTypeText, Width: 20, ExtractFn: dataStr("npwp")},
		{Header: "Name", Type: ColTypeText, Width: 24, ExtractFn: dataStr("nama")},
		{Header: "Address", Type: ColTypeText, Width: 36, ExtractFn: dataStr("alamat")},
		{Header: "NPWP Verification", Type: ColTypeText, Width: 24, ExtractFn: dataStr("npwp_verification")},
		{Header: "Date Time", Type: ColTypeDateTime, Width: 22, ExtractFn: ExtractRequestTime},
	},
}

var ProductRegistry = map[string]ProductSheetDef{
	"INCOMETAX_tax_verification_detail": ProductTaxVerification,
}
