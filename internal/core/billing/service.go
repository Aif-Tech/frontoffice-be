package billing

import (
	"bytes"
	"fmt"
	"front-office/configs/application"
	"front-office/internal/core/log/transaction"
	"front-office/internal/mail"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func NewService(
	cfg *application.Config,
	repo Repository,
	transactionRepo transaction.Repository,
	mailSvc *mail.SendMailService,
) Service {
	return &service{
		cfg,
		repo,
		transactionRepo,
		mailSvc,
	}
}

type service struct {
	cfg             *application.Config
	repo            Repository
	transactionRepo transaction.Repository
	mailSvc         *mail.SendMailService
}

type Service interface {
	SendMonthlyUsageReport() error
	DownloadUsageXlsx(input downloadUsageXlsxInput) (*downloadUsageXlsxResult, error)
	GetUsageReport(companyId uint, pricingStrategy string, month, year int) (*usageSummary, error)
	generateUsageXlsx(input XlsxReportInput) ([]byte, error)
}

func (svc *service) NewProcatFetchFn(pricingStrategy string) FetchFn {
	return func(productId, companyId string) ([]LogRow, error) {
		rows, err := svc.transactionRepo.GetLogTransByJobIdAPI(
			"", productId, companyId, pricingStrategy,
		)
		if err != nil {
			return nil, err
		}

		return WrapProcat(rows), nil
	}
}

func (svc *service) NewScoreezyFetchFn(startDate, endDate string) FetchFn {
	return func(productId, companyId string) ([]LogRow, error) {
		rows, err := svc.transactionRepo.GetLogsScoreezyByDateRangeAPI(
			&transaction.LogFilter{
				CompanyId: companyId,
				StartDate: startDate,
				EndDate:   endDate,
				Size:      constant.SizeUnlimited,
			},
		)
		if err != nil {
			return nil, err
		}

		return WrapScoreezy(rows), nil
	}
}

func (svc *service) DownloadUsageXlsx(input downloadUsageXlsxInput) (*downloadUsageXlsxResult, error) {
	if input.PricingStrategy == "" {
		input.PricingStrategy = constant.PaidStatus
	}

	if len(input.Groups) == 0 {
		input.Groups = []string{"procat", "scoreezy"}
	}

	period := time.Date(input.Year, time.Month(input.Month), 1, 0, 0, 0, 0, time.UTC)
	startDate := fmt.Sprintf("%d-%02d-01", input.Year, input.Month)
	endDate := fmt.Sprintf("%d-%02d-%02d", input.Year, input.Month, lastDayOfMonth(period))

	allGroups, companyName, err := svc.buildProductGroups(input.CompanyId, input.PricingStrategy, startDate, endDate, input.Month, input.Year)
	if err != nil {
		return nil, apperror.Internal("failed to build product groups", err)
	}

	filtered := filterGroups(allGroups, input.Groups)
	if len(filtered) == 0 {
		return nil, apperror.BadRequest(fmt.Sprintf("no valid product groups found for: %v", input.Groups))
	}

	xlsxBytes, err := svc.generateUsageXlsx(XlsxReportInput{
		CompanyId:       input.CompanyId,
		CompanyName:     companyName,
		PeriodYear:      input.Year,
		PeriodMonth:     input.Month,
		PricingStrategy: input.PricingStrategy,
		ProductGroups:   filtered,
		Password:        input.Password,
	})
	if err != nil {
		return nil, apperror.Internal("failed to generate report", err)
	}

	filename := fmt.Sprintf(
		"usage_report_%s_%d_%02d.xlsx",
		companyName, input.Year, input.Month,
	)

	return &downloadUsageXlsxResult{
		Filename:    filename,
		ContentType: constant.MimeXlsx,
		Data:        xlsxBytes,
	}, nil
}

func (svc *service) SendMonthlyUsageReport() error {
	summaries, err := svc.repo.GetUsageReport()
	if err != nil {
		return apperror.MapRepoError(err, "failed to get monthly usage report")
	}

	now := time.Now()
	firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastMonth := firstOfThisMonth.AddDate(0, -1, 0)

	year := lastMonth.Year()
	month := lastMonth.Month()
	startDate := fmt.Sprintf("%d-%02d-01", year, int(month))
	endDate := fmt.Sprintf("%d-%02d-%02d", year, int(month), lastDayOfMonth(lastMonth))

	ccEmails := parseCCEmails(svc.cfg.App.MailInternalCC)

	for _, summary := range summaries {
		admins, err := svc.repo.GetAdminsData(summary.CompanyId)
		if err != nil {
			log.Warn().
				Err(err).
				Uint("company_id", summary.CompanyId).
				Msg("failed to get admin emails")

			continue
		}

		if len(admins) == 0 {
			log.Warn().
				Uint("company_id", summary.CompanyId).
				Msg("no admin emails found, skipping")

			continue
		}

		for _, admin := range admins {
			// xlsxPassword := admin.Key
			xlsxPassword := svc.cfg.Mail.Password // todo: update xlsx password

			xlsxBytes, xlsxErr := svc.generateUsageXlsx(XlsxReportInput{
				CompanyId:   summary.CompanyId,
				CompanyName: summary.CompanyName,
				PeriodYear:  year,
				PeriodMonth: int(month),
				ProductGroups: []ProductGroup{
					{
						GroupName: "Procat",
						Products:  toXlsxProducts(summary.ProcatProducts),
						FetchFn:   svc.NewProcatFetchFn(constant.PaidStatus),
					},
					{
						GroupName: "Scoreezy",
						Products:  toXlsxProducts(summary.ScoreezyProducts),
						FetchFn:   svc.NewScoreezyFetchFn(startDate, endDate),
					},
				},
				PricingStrategy: constant.PaidStatus,
				Password:        xlsxPassword,
			})
			if xlsxErr != nil {
				log.Warn().
					Err(xlsxErr).
					Uint("company_id", summary.CompanyId).
					Msg("sending email without attachment")
			}

			var attachments []mail.MailAttachment
			if xlsxBytes != nil {
				attachments = append(attachments, mail.MailAttachment{
					FileName: fmt.Sprintf("Monthly Usage Report for %s - %s %d.xlxs", summary.CompanyName, month, year),
					Content:  xlsxBytes,
					MimeType: constant.MimeXlsx,
				})
			}

			fmt.Println("admin email: ", admin.Email)
			if err := svc.mailSvc.SendWithTemplate(
				"arief@aiforesee.com", // todo: update to admin mail
				ccEmails,
				fmt.Sprintf("Monthly Usage Report for %s - %s %d", summary.CompanyName, month, year),
				"monthly_usage_report.html",
				map[string]any{
					"Name":             summary.CompanyName,
					"ProcatProducts":   summary.ProcatProducts,
					"ScoreezyProducts": summary.ScoreezyProducts,
					"HasUsage":         len(summary.ProcatProducts) > 0 || len(summary.ScoreezyProducts) > 0,
					"Month":            month.String(),
					"Year":             year,
				},
				attachments...,
			); err != nil {
				log.Warn().
					Err(err).
					Uint("company_id", summary.CompanyId).
					Msg("failed to send monthly usage report")
			}
		}

	}

	return nil
}

func (svc *service) GetUsageReport(companyId uint, pricingStrategy string, month, year int) (*usageSummary, error) {
	companyIdStr := strconv.FormatUint(uint64(companyId), 10)
	monthStr := strconv.FormatUint(uint64(month), 10)
	yearStr := strconv.FormatUint(uint64(year), 10)

	summary, err := svc.repo.GetUsageReportByCompany(companyIdStr, pricingStrategy, monthStr, yearStr)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to get usage report")
	}

	return summary, nil
}

func (svc *service) buildProductGroups(
	companyId uint,
	pricingStrategy, startDate, endDate string,
	month, year int,
) ([]ProductGroup, string, error) {
	companyIdStr := strconv.FormatUint(uint64(companyId), 10)
	monthStr := strconv.FormatUint(uint64(month), 10)
	yearStr := strconv.FormatUint(uint64(year), 10)

	summary, err := svc.repo.GetUsageReportByCompany(companyIdStr, pricingStrategy, monthStr, yearStr)
	if err != nil {
		return nil, "", apperror.MapRepoError(err, "failed to get usage report")
	}

	return []ProductGroup{
			{
				GroupName: "Procat",
				Key:       groupKeyProcat,
				Products:  toXlsxProducts(summary.ProcatProducts),
				FetchFn:   svc.NewProcatFetchFn(pricingStrategy),
			},
			{
				GroupName: "Scoreezy",
				Key:       groupKeyScorezy,
				Products:  toXlsxProducts(summary.ScoreezyProducts),
				FetchFn:   svc.NewScoreezyFetchFn(startDate, endDate),
			},
		},
		summary.CompanyName,
		nil
}

func filterGroups(groups []ProductGroup, allowedKeys []string) []ProductGroup {
	if len(allowedKeys) == 0 {
		return groups
	}

	allowed := make(map[string]struct{}, len(allowedKeys))
	for _, k := range allowedKeys {
		allowed[k] = struct{}{}
	}

	var out []ProductGroup
	for _, g := range groups {
		if _, ok := allowed[g.Key]; ok {
			out = append(out, g)
		}
	}

	return out
}

func lastDayOfMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func (svc *service) generateUsageXlsx(input XlsxReportInput) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	defaultSheet := f.GetSheetName(0)
	builtAny := false

	for _, group := range input.ProductGroups {
		for _, product := range group.Products {
			def, ok := productRegistry[product.ProductSlug]
			if !ok {
				log.Warn().
					Str("group", group.GroupName).
					Str("product_slug", product.ProductSlug).
					Msg("product not found in registry, skipping")

				continue
			}

			productId := strconv.FormatUint(uint64(product.ProductId), 10)
			companyId := strconv.FormatUint(uint64(input.CompanyId), 10)

			rows, err := group.FetchFn(productId, companyId)
			if err != nil {
				log.Warn().
					Err(err).
					Str("group", group.GroupName).
					Str("product_slug", product.ProductSlug).
					Uint("company_id", input.CompanyId).
					Msg("failed to fetch transaction data, skipping sheet")

				continue
			}

			sheetName := def.SheetName
			if sheetName == "" {
				sheetName = def.ProductName
			}

			idx, err := f.NewSheet(sheetName)
			if err != nil {
				return nil, apperror.Internal(fmt.Sprintf("failed to create sheet '%s': %s", sheetName, err), err)
			}
			if !builtAny {
				f.SetActiveSheet(idx)
				builtAny = true
			}

			if err := writeProductSheet(f, sheetName, def, rows); err != nil {
				return nil, apperror.Internal(fmt.Sprintf("failed to write sheet '%s': %s", sheetName, err), err)
			}
		}
	}

	if !builtAny {
		return nil, fmt.Errorf("no sheets generated for company %d", input.CompanyId)
	}

	f.DeleteSheet(defaultSheet)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	// encrypt file
	if input.Password != "" {
		encrypted, err := excelize.Encrypt(buf.Bytes(), &excelize.Options{Password: input.Password})
		if err != nil {
			return nil, apperror.Internal("failed to write encypted xlxs in buffer", err)
		}

		return encrypted, nil
	}

	return buf.Bytes(), nil
}

func writeProductSheet(
	f *excelize.File,
	sheetName string,
	def ProductSheetDef,
	rows []LogRow,
) error {
	cols := def.Columns
	colLetters := makeColLetters(len(cols))
	lastCol := colLetters[len(colLetters)-1]

	// Header
	for i, col := range cols {
		cell := colLetters[i] + "1"
		f.SetCellValue(sheetName, cell, col.Header)

		w := col.Width
		if w == 0 {
			w = 18
		}
		f.SetColWidth(sheetName, colLetters[i], colLetters[i], w)
	}
	f.SetRowHeight(sheetName, 1, 22)

	// Freeze header
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A5",
		ActivePane:  "bottomLeft",
	})

	// Data rows
	dataStart := 2
	for ri, row := range rows {
		r := dataStart + ri
		for ci, col := range cols {
			cell := colLetters[ci] + fmt.Sprintf("%d", r)
			val := col.ExtractFn(row)

			setCellTypedValue(f, sheetName, cell, val, col.Type)
		}
		f.SetRowHeight(sheetName, r, 18)
	}

	// total transaction
	totalRow := dataStart + len(rows)
	f.MergeCell(sheetName,
		"A"+fmt.Sprintf("%d", totalRow),
		lastCol+fmt.Sprintf("%d", totalRow))

	f.SetRowHeight(sheetName, totalRow, 22)

	return nil
}

func setCellTypedValue(f *excelize.File, sheet, cell string, val interface{}, colType ColumnType) {
	if val == nil {
		f.SetCellValue(sheet, cell, "")
		return
	}

	switch colType {
	case ColTypeDateTime:
		switch v := val.(type) {
		case time.Time:
			if v.IsZero() {
				f.SetCellValue(sheet, cell, "")
			} else {
				f.SetCellValue(sheet, cell, v.Format("02/01/2006 15:04:05"))
			}
		default:
			f.SetCellValue(sheet, cell, fmt.Sprintf("%v", v))
		}

	case ColTypeDate:
		switch v := val.(type) {
		case time.Time:
			if v.IsZero() {
				f.SetCellValue(sheet, cell, "")
			} else {
				f.SetCellValue(sheet, cell, v.Format("02/01/2006"))
			}
		default:
			f.SetCellValue(sheet, cell, fmt.Sprintf("%v", v))
		}

	default:
		f.SetCellValue(sheet, cell, val)
	}
}

// makeColLetters menghasilkan slice huruf kolom Excel: ["A","B","C",...]
func makeColLetters(n int) []string {
	letters := make([]string, n)
	for i := 0; i < n; i++ {
		name, _ := excelize.ColumnNumberToName(i + 1)
		letters[i] = name
	}

	return letters
}

func toXlsxProducts(products []usagePerProduct) []XlsxReportProduct {
	result := make([]XlsxReportProduct, 0, len(products))

	for _, p := range products {
		result = append(result, XlsxReportProduct{
			ProductId:    p.ProductId,
			ProductSlug:  p.ProductSlug,
			ProductName:  p.ProductName,
			TotalRequest: p.TotalRequest,
			TotalPay:     p.TotalPay,
		})
	}

	return result
}

func parseCCEmails(raw string) []string {
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

type LogRow interface {
	logRow()
}

type ProcatRow struct {
	*transaction.LogTransProductCatalog
}

func (r ProcatRow) logRow() {}

func WrapProcat(rows []*transaction.LogTransProductCatalog) []LogRow {
	out := make([]LogRow, len(rows))
	for i, r := range rows {
		out[i] = ProcatRow{r}
	}
	return out
}

type ScoreezyRow struct {
	*transaction.LogTransScoreezy
}

func (r ScoreezyRow) logRow() {}

func WrapScoreezy(rows []*transaction.LogTransScoreezy) []LogRow {
	out := make([]LogRow, len(rows))
	for i, r := range rows {
		out[i] = ScoreezyRow{r}
	}
	return out
}

type RowExtractFn func(row LogRow) interface{}

func fromProcat(fn func(*transaction.LogTransProductCatalog) interface{}) RowExtractFn {
	return func(row LogRow) interface{} {
		r, ok := row.(ProcatRow)
		if !ok {
			return ""
		}

		return fn(r.LogTransProductCatalog)
	}
}

func fromScoreezy(fn func(*transaction.LogTransScoreezy) interface{}) RowExtractFn {
	return func(row LogRow) interface{} {
		r, ok := row.(ScoreezyRow)
		if !ok {
			return ""
		}
		return fn(r.LogTransScoreezy)
	}
}

func scoreezyNestedDataStr(keys ...string) RowExtractFn {
	return fromScoreezy(func(r *transaction.LogTransScoreezy) interface{} {
		v := helper.ExtractNestedField(r.Data, keys...)
		if s, ok := v.(string); ok {
			return s
		}
		if v == nil || v == "" {
			return ""
		}
		return fmt.Sprintf("%v", v)
	})
}

var (
	ScoreezyExtractTrxID = fromScoreezy(func(r *transaction.LogTransScoreezy) interface{} {
		return r.TrxId
	})
	ScoreezyExtractCreatedAt = fromScoreezy(func(r *transaction.LogTransScoreezy) interface{} {
		return r.CreatedAt
	})
)

func procatExtractData(log *transaction.LogTransProductCatalog, key string) interface{} {
	return helper.LookupKey(helper.ParseJSON(log.Data), key)
}

func procatDataStr(key string) RowExtractFn {
	return fromProcat(func(r *transaction.LogTransProductCatalog) interface{} {
		v := procatExtractData(r, key)
		if s, ok := v.(string); ok {
			return s
		}

		return ""
	})
}

func procatExtractRespInput(log *transaction.LogTransProductCatalog, key string) interface{} {
	body := helper.ParseJSON(log.ResponseBody)
	if body == nil {
		return ""
	}

	inputSection, ok := body["input"].(map[string]interface{})
	if !ok {
		return ""
	}

	return helper.LookupKey(inputSection, key)
}

func procatRespInputStr(key string) RowExtractFn {
	return fromProcat(func(log *transaction.LogTransProductCatalog) interface{} {
		v := procatExtractRespInput(log, key)
		if s, ok := v.(string); ok {
			return s
		}

		return fmt.Sprintf("%v", v)
	})
}

func splitIndex(sep string, n int) func(string) string {
	return func(s string) string {
		if s == "" {
			return ""
		}

		parts := strings.Split(s, sep)
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}

		idx := n
		if idx < 0 {
			idx = len(parts) + n
		}
		if idx < 0 || idx >= len(parts) {
			return ""
		}

		return parts[idx]
	}
}

func procatDataTransform(key string, fn func(string) string) RowExtractFn {
	return fromProcat(func(log *transaction.LogTransProductCatalog) interface{} {
		v := procatExtractData(log, key)
		s, ok := v.(string)
		if !ok {
			s = fmt.Sprintf("%v", v)
		}

		return fn(s)
	})
}

func staticVal(val interface{}) RowExtractFn {
	return func(_ LogRow) interface{} { return val }
}

var (
	ProcatExtractTransactionID = fromProcat(func(log *transaction.LogTransProductCatalog) interface{} {
		return log.TransactionID
	})

	ProcatExtractCreatedAt = fromProcat(func(log *transaction.LogTransProductCatalog) interface{} {
		return log.CreatedAt
	})
)
