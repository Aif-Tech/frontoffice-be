package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"front-office/configs/application"
	"front-office/internal/core/log/transaction"
	"front-office/internal/mail"
	"front-office/pkg/apperror"
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
	generateUsageXlsx(input XlsxReportInput) ([]byte, error)
}

func (svc *service) SendMonthlyUsageReport() error {
	summaries, err := svc.repo.GetMonthlyReport()
	if err != nil {
		return apperror.MapRepoError(err, "failed to get monthly usage report")
	}

	now := time.Now()
	firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastMonth := firstOfThisMonth.AddDate(0, -1, 0)
	year := lastMonth.Year()
	month := lastMonth.Month()

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
			xlxsPassword := admin.Key

			xlsxBytes, xlsxErr := svc.generateUsageXlsx(XlsxReportInput{
				CompanyId:   summary.CompanyId,
				CompanyName: summary.CompanyName,
				PeriodYear:  year,
				PeriodMonth: int(month),
				Products:    toXlsxProducts(summary.Products),
				Password:    xlxsPassword,
			})
			if xlsxErr != nil {
				log.Warn().
					Err(xlsxErr).
					Uint("company_id", summary.CompanyId).
					Msg("failed to generate xlsx, sending email without attachment")
			}

			var attachments []mail.MailAttachment
			if xlsxBytes != nil {
				attachments = append(attachments, mail.MailAttachment{
					FileName: fmt.Sprintf("Monthly Usage Report for %s - %s %d>", summary.CompanyName, month, year),
					Content:  xlsxBytes,
					MimeType: mail.MimeXlsx,
				})
			}

			if err := svc.mailSvc.SendWithTemplate(
				admin.Email,
				ccEmails,
				fmt.Sprintf("Monthly Usage Report for %s - %s %d", summary.CompanyName, month, year),
				"monthly_usage_report.html",
				map[string]any{
					"Name":     summary.CompanyName,
					"Products": summary.Products,
					"Month":    month.String(),
					"Year":     year,
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

func (svc *service) generateUsageXlsx(input XlsxReportInput) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	defaultSheet := f.GetSheetName(0)
	builtAny := false

	for _, key := range input.Products {
		def, ok := ProductRegistry[key.ProductSlug]
		if !ok {
			return nil, fmt.Errorf("produk '%s' tidak ditemukan di ProductRegistry", key.ProductName)
		}

		productId := strconv.FormatUint(uint64(key.ProductId), 10)
		companyId := strconv.FormatUint(uint64(input.CompanyId), 10)
		rows, err := svc.transactionRepo.GetLogTransByJobIdAPI("", productId, companyId)
		if err != nil {
			return nil, fmt.Errorf("gagal mengambil data untuk '%s': %w", key.ProductName, err)
		}

		sheetName := def.SheetName
		if sheetName == "" {
			sheetName = def.ProductName
		}

		idx, err := f.NewSheet(sheetName)
		if err != nil {
			return nil, fmt.Errorf("gagal membuat sheet '%s': %w", sheetName, err)
		}
		if !builtAny {
			f.SetActiveSheet(idx)
			builtAny = true
		}

		if err := writeProductSheet(f, sheetName, def, rows); err != nil {
			return nil, fmt.Errorf("gagal menulis sheet '%s': %w", sheetName, err)
		}
	}

	// Hapus sheet default Excel
	f.DeleteSheet(defaultSheet)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	// encrypt file
	if input.Password != "" {
		encrypted, err := excelize.Encrypt(buf.Bytes(), &excelize.Options{Password: input.Password})
		if err != nil {
			return nil, fmt.Errorf("gagal menulis xlsx terenkripsi ke buffer: %w", err)
		}

		return encrypted, nil
	}

	return buf.Bytes(), nil
}

func writeProductSheet(
	f *excelize.File,
	sheetName string,
	def ProductSheetDef,
	rows []*transaction.LogTransProductCatalog,
) error {
	cols := def.Columns
	colLetters := makeColLetters(len(cols))
	lastCol := colLetters[len(colLetters)-1]

	// ── Sub-header tanggal generate ───────────────────────────────────────────
	// f.MergeCell(sheetName, "A2", lastCol+"2")
	// f.SetCellValue(sheetName, "A2",
	// 	fmt.Sprintf("Generated on: %s", time.Now().Format("02 January 2006, 15:04:05")))
	// f.SetRowHeight(sheetName, 2, 18)

	f.SetRowHeight(sheetName, 3, 6) // spacer

	// Header
	for i, col := range cols {
		cell := colLetters[i] + "4"
		f.SetCellValue(sheetName, cell, col.Header)

		w := col.Width
		if w == 0 {
			w = 18
		}
		f.SetColWidth(sheetName, colLetters[i], colLetters[i], w)
	}
	f.SetRowHeight(sheetName, 4, 22)

	// Freeze header
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		YSplit:      4,
		TopLeftCell: "A5",
		ActivePane:  "bottomLeft",
	})

	// Data rows
	dataStart := 5
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
			TotalSuccess: p.TotalSuccess,
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

func extractData(log *transaction.LogTransProductCatalog, key string) interface{} {
	if log.Data == nil {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(log.Data, &m); err != nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return v
}

func dataStr(key string) ExtractFn {
	return func(log *transaction.LogTransProductCatalog) interface{} {
		v := extractData(log, key)
		if s, ok := v.(string); ok {
			return s
		}

		return fmt.Sprintf("%v", v)
	}
}

func dataNum(key string) ExtractFn {
	return func(log *transaction.LogTransProductCatalog) interface{} {
		return extractData(log, key)
	}
}

func staticVal(val interface{}) ExtractFn {
	return func(_ *transaction.LogTransProductCatalog) interface{} {
		return val
	}
}

var (
	ExtractTransactionID = func(log *transaction.LogTransProductCatalog) interface{} {
		return log.TransactionID
	}

	ExtractNPWP = func(log *transaction.LogTransProductCatalog) interface{} {
		return dataStr("npwp")(log)
	}

	ExtractRequestTime = func(log *transaction.LogTransProductCatalog) interface{} {
		return log.RequestTime
	}
)
