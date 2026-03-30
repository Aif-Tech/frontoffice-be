package billing

import (
	"front-office/internal/core/log/transaction"
	"front-office/pkg/common/constant"
)

type downloadUsageXlsxRequest struct {
	CompanyId       uint
	Year            int
	Month           int
	Groups          []string // opsional, example: "procat,scoreezy"
	PricingStrategy string   // opsional, default: "PAY"
	Password        string
}

type downloadUsageXlsxInput struct {
	CompanyId       uint
	Year            int
	Month           int
	Groups          []string
	PricingStrategy string // kosong = default constant.PaidStatus
	Password        string
}

type downloadUsageXlsxResult struct {
	Filename    string
	ContentType string
	Data        []byte
}

const (
	groupKeyProcat  = "procat"
	groupKeyScorezy = "scoreezy"
)

type usagePerProduct struct {
	ProductId    uint   `json:"product_id"`
	ProductSlug  string `json:"product_slug"`
	ProductName  string `json:"product_name"`
	TotalRequest int    `json:"total_request"`
	TotalPay     int    `json:"total_pay"`
}

type usageSummary struct {
	CompanyId          uint                `json:"company_id"`
	CompanyName        string              `json:"company_name"`
	PeriodYear         int                 `json:"period_year"`
	PeriodMonth        int                 `json:"period_month"`
	SubscribedProducts []subscribedProduct `json:"subscribed_products"`
	ProcatProducts     []usagePerProduct   `json:"procat_products"`
	ScoreezyProducts   []usagePerProduct   `json:"scoreezy_products"`
	// GrandTotalRequest int64             `json:"grand_total_request"`
	// GrandTotalPay     int64             `json:"grand_total_pay"`
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
	TotalRequest int
	TotalPay     int
}

type FetchFn func(productId, companyId string) ([]LogRow, error)

type ProductGroup struct {
	GroupName string
	Key       string
	Products  []XlsxReportProduct
	FetchFn   FetchFn
}

type XlsxReportInput struct {
	CompanyId       uint
	CompanyName     string
	PeriodYear      int
	PeriodMonth     int
	ProductGroups   []ProductGroup
	PricingStrategy string
	Password        string
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
	ExtractFn RowExtractFn
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

type TemplateProduct struct {
	ProductName string
	TotalPay    int
}

type MonthlyUsageTemplateData struct {
	Subject  string
	Name     string
	Month    string
	Year     int
	Products []TemplateProduct
}

type subscribedProduct struct {
	ProductId   uint   `json:"product_id"`
	ProductName string `json:"product_name"`
}

var productRegistry = map[string]ProductSheetDef{
	constant.SlugGenRetailV3:           productGenRetail,
	constant.SlugLoanRecordChecker:     productLoanRecord,
	constant.Slug7DaysMultipleLoan:     product7DMultipleLoan,
	constant.Slug30DaysMultipleLoan:    product30DMultipleLoan,
	constant.Slug90DaysMultipleLoan:    product90DMultipleLoan,
	constant.SlugNPWPVerification:      productNPWPVerification,
	constant.SlugPhoneLiveStatus:       productPhoneLive,
	constant.SlugRecycleNumber:         productRecycleNumber,
	constant.SlugTaxComplianceStatus:   productTaxCompliance,
	constant.SlugTaxScore:              productTaxScore,
	constant.SlugTaxVerificationDetail: productTaxVerification,
}

var productGenRetail = ProductSheetDef{
	ProductName: constant.GenRetail,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ScoreezyExtractTrxID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.GenRetail)},
		{Header: constant.CSVHeaderIDCard, Type: ColTypeText, Width: 20, ExtractFn: scoreezyNestedDataStr("data", "data", "id_card_no")},
		{Header: constant.CSVHeaderIdentity, Type: ColTypeText, Width: 24, ExtractFn: scoreezyNestedDataStr("data", "behavior")},
		{Header: constant.CSVHeaderBehavior, Type: ColTypeText, Width: 36, ExtractFn: scoreezyNestedDataStr("data", "identity")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ScoreezyExtractCreatedAt},
	},
}

var productLoanRecord = ProductSheetDef{
	ProductName: constant.LoanRecordChecker,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ProcatExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.LoanRecordChecker)},
		{Header: constant.CSVHeaderIDCard, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("nik")},
		{Header: constant.CSVHeaderResult, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("status")},
		{Header: constant.CSVHeaderRemarks, Type: ColTypeText, Width: 36, ExtractFn: procatDataStr("remarks")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ProcatExtractCreatedAt},
	},
}

var product7DMultipleLoan = ProductSheetDef{
	ProductName: constant.MultipleLoan7D,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ProcatExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.MultipleLoan7D)},
		{Header: constant.CSVHeaderIDCard, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("nik")},
		{Header: constant.CSVHeaderPhone, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("phone_number")},
		{Header: constant.CSVHeaderQueryCount, Type: ColTypeText, Width: 36, ExtractFn: procatDataStr("query_count")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ProcatExtractCreatedAt},
	},
}

var product30DMultipleLoan = ProductSheetDef{
	ProductName: constant.MultipleLoan30D,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ProcatExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.MultipleLoan30D)},
		{Header: constant.CSVHeaderIDCard, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("nik")},
		{Header: constant.CSVHeaderPhone, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("phone_number")},
		{Header: constant.CSVHeaderQueryCount, Type: ColTypeText, Width: 36, ExtractFn: procatDataStr("query_count")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ProcatExtractCreatedAt},
	},
}

var product90DMultipleLoan = ProductSheetDef{
	ProductName: constant.MultipleLoan90D,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ProcatExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.MultipleLoan30D)},
		{Header: constant.CSVHeaderIDCard, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("nik")},
		{Header: constant.CSVHeaderPhone, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("phone_number")},
		{Header: constant.CSVHeaderQueryCount, Type: ColTypeText, Width: 36, ExtractFn: procatDataStr("query_count")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ProcatExtractCreatedAt},
	},
}

var productNPWPVerification = ProductSheetDef{
	ProductName: constant.NPWPVerification,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ProcatExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.NPWPVerification)},
		{Header: constant.CSVHeaderNPWP, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("npwp")},
		{Header: constant.CSVHeaderName, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("nama")},
		{Header: constant.CSVHeaderAddress, Type: ColTypeText, Width: 36, ExtractFn: procatDataStr("alamat")},
		{Header: constant.CSVHeaderResult, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("status")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ProcatExtractCreatedAt},
	},
}

var productPhoneLive = ProductSheetDef{
	ProductName: constant.PhoneLiveStatus,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ProcatExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.PhoneLiveStatus)},
		{Header: constant.CSVHeaderPhone, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("phone_number")},
		{Header: constant.CSVHeaderSubscriberStatus, Type: ColTypeText, Width: 24, ExtractFn: procatDataTransform("live_status", splitIndex(",", 0))},
		{Header: constant.CSVHeaderDeviceStatus, Type: ColTypeText, Width: 24, ExtractFn: procatDataTransform("live_status", splitIndex(",", 1))},
		{Header: constant.CSVHeaderOperator, Type: ColTypeText, Width: 36, ExtractFn: procatDataStr("operator")},
		{Header: constant.CSVHeaderPhoneType, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("phone_type")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ProcatExtractCreatedAt},
	},
}

var productRecycleNumber = ProductSheetDef{
	ProductName: constant.RecycleNumber,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ProcatExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.RecycleNumber)},
		{Header: constant.CSVHeaderPhone, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("phone_number")},
		{Header: constant.CSVHeaderStatus, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("status")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ProcatExtractCreatedAt},
	},
}

var productTaxCompliance = ProductSheetDef{
	ProductName: constant.TaxCompliance,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ProcatExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.TaxCompliance)},
		{Header: constant.CSVHeaderNPWP, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("npwp")},
		{Header: constant.CSVHeaderName, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("nama")},
		{Header: constant.CSVHeaderAddress, Type: ColTypeText, Width: 36, ExtractFn: procatDataStr("alamat")},
		{Header: constant.CSVHeaderResult, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("status")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ProcatExtractCreatedAt},
	},
}

var productTaxScore = ProductSheetDef{
	ProductName: constant.TaxScore,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ProcatExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.TaxScore)},
		{Header: constant.CSVHeaderNPWP, Type: ColTypeText, Width: 20, ExtractFn: procatRespInputStr("npwp")},
		{Header: constant.CSVHeaderName, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("nama")},
		{Header: constant.CSVHeaderAddress, Type: ColTypeText, Width: 36, ExtractFn: procatDataStr("alamat")},
		{Header: constant.CSVHeaderResult, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("status")},
		{Header: constant.CSVHeaderScore, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("score")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ProcatExtractCreatedAt},
	},
}

var productTaxVerification = ProductSheetDef{
	ProductName: constant.TaxVerification,
	Columns: []ColumnDef{
		{Header: constant.CSVHeaderTransactionID, Type: ColTypeText, Width: 32, ExtractFn: ProcatExtractTransactionID},
		{Header: constant.CSVHeaderProductName, Type: ColTypeText, Width: 18, ExtractFn: staticVal(constant.TaxVerification)},
		{Header: constant.CSVHeaderNPWP, Type: ColTypeText, Width: 20, ExtractFn: procatDataStr("npwp")},
		{Header: constant.CSVHeaderName, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("nama")},
		{Header: constant.CSVHeaderAddress, Type: ColTypeText, Width: 36, ExtractFn: procatDataStr("alamat")},
		{Header: constant.CSVHeaderNPWPVerification, Type: ColTypeText, Width: 24, ExtractFn: procatDataStr("npwp_verification")},
		{Header: constant.CSVHeaderDateCreated, Type: ColTypeDateTime, Width: 22, ExtractFn: ProcatExtractCreatedAt},
	},
}
