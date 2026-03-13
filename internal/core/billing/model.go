package billing

import (
	"front-office/internal/core/log/transaction"
	"front-office/pkg/common/constant"
)

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

var ProductRegistry = map[string]ProductSheetDef{
	constant.SlugPhoneLiveStatus:       ProductPhoneLive,
	constant.SlugRecycleNumber:         ProductRecycleNumber,
	constant.SlugTaxVerificationDetail: ProductTaxVerification,
}

var ProductPhoneLive = ProductSheetDef{
	ProductName: constant.PhoneLiveStatus,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.PhoneLiveStatus)},
		{Header: constant.CSVHeaderPhone, Type: ColTypeText, Width: 20, ExtractFn: respInputStr("phone_number")},
		{Header: constant.CSVHeaderSubscriberStatus, Type: ColTypeText, Width: 24, ExtractFn: dataTransform("live_status", splitIndex(",", 0))},
		{Header: constant.CSVHeaderDeviceStatus, Type: ColTypeText, Width: 24, ExtractFn: dataTransform("live_status", splitIndex(",", 1))},
		{Header: constant.CSVHeaderOperator, Type: ColTypeText, Width: 36, ExtractFn: dataStr("operator")},
		{Header: constant.CSVHeaderPhoneType, Type: ColTypeText, Width: 24, ExtractFn: dataStr("phone_type")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ExtractRequestTime},
	},
}

var ProductRecycleNumber = ProductSheetDef{
	ProductName: constant.RecycleNumber,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.RecycleNumber)},
		{Header: constant.CSVHeaderPhone, Type: ColTypeText, Width: 20, ExtractFn: respInputStr("phone_number")},
		{Header: constant.CSVHeaderStatus, Type: ColTypeText, Width: 24, ExtractFn: dataStr("status")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ExtractRequestTime},
	},
}

var ProductTaxVerification = ProductSheetDef{
	ProductName: constant.TaxVerification,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.TaxVerification)},
		{Header: constant.CSVHeaderNPWP, Type: ColTypeText, Width: 20, ExtractFn: dataStr("npwp")},
		{Header: constant.CSVHeaderName, Type: ColTypeText, Width: 24, ExtractFn: dataStr("nama")},
		{Header: constant.CSVHeaderAddress, Type: ColTypeText, Width: 36, ExtractFn: dataStr("alamat")},
		{Header: constant.CSVHeaderNPWPVerification, Type: ColTypeText, Width: 24, ExtractFn: dataStr("npwp_verification")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ExtractRequestTime},
	},
}
