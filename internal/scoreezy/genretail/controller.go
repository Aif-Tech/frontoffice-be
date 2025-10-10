package genretail

import (
	"bytes"
	"fmt"
	"front-office/pkg/apperror"
	"front-office/pkg/helper"
	"mime/multipart"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"front-office/pkg/common/constant"
)

func NewController(
	service Service,
) Controller {
	return &controller{
		service,
	}
}

type controller struct {
	service Service
}

type Controller interface {
	DummyRequestScore(c *fiber.Ctx) error
	SingleRequest(c *fiber.Ctx) error
	BulkRequest(c *fiber.Ctx) error
	GetLogsScoreezy(c *fiber.Ctx) error
	GetLogScoreezy(c *fiber.Ctx) error
	ExportJobDetails(c *fiber.Ctx) error
	// DownloadCSV(c *fiber.Ctx) error
	// UploadCSV(c *fiber.Ctx) error
	// GetBulkSearch(c *fiber.Ctx) error
}

func (ctrl *controller) DummyRequestScore(c *fiber.Ctx) error {
	response := GenRetailV3ClientReturnSuccess{
		Message: "Succeed to Request Scores",
		Success: true,
		Data: &dataGenRetailV3{
			TransactionId:        "TRX123456789",
			Name:                 "John Doe",
			IdCardNo:             "1234567890123456",
			PhoneNo:              "081234567890",
			LoanNo:               "LN987654321",
			ProbabilityToDefault: 0.12345,
			Grade:                "A",
			Identity:             "Verified in more than 50% social media platform and registered on one of the telecommunication platforms",
			Behavior:             "This individual is not identified to have a history of loan applications and is not indicated to have defaulted on payments.",
			Date:                 time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	return c.Status(200).JSON(response)
}

func (ctrl *controller) SingleRequest(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*genRetailRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	result, err := ctrl.service.GenRetailV3(authCtx.UserId, authCtx.CompanyId, reqBody)
	if err != nil {
		return err
	}

	return c.Status(result.StatusCode).JSON(result)
}

func (ctrl *controller) BulkRequest(c *fiber.Ctx) error {
	file, ok := c.Locals(constant.ValidatedFile).(*multipart.FileHeader)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	if err := ctrl.service.BulkGenRetailV3(authCtx.UserId, authCtx.CompanyId, authCtx.QuotaTypeStr(), file); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		constant.Success,
		nil,
	))
}

func (ctrl *controller) GetLogsScoreezy(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	filter := &filterLogs{
		CompanyId:   authCtx.CompanyIdStr(),
		JobId:       c.Query("job_id"),
		StartDate:   c.Query(constant.StartDate),
		EndDate:     c.Query(constant.EndDate),
		Name:        strings.ToLower(strings.TrimSpace(c.Query("name"))),
		ProductType: strings.ToLower(strings.TrimSpace(c.Query("product_type"))),
		Grade:       strings.ToLower(c.Query("grade")),
		Page:        c.Query(constant.Page),
		Size:        c.Query(constant.Size),
	}

	result, err := ctrl.service.GetLogsScoreezy(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		constant.Success,
		gradesResponseData{Logs: result.Data},
		result.Meta,
	))
}

func (ctrl *controller) GetLogScoreezy(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	filter := &filterLogs{
		CompanyId: authCtx.CompanyIdStr(),
		TrxId:     c.Params("trx_id"),
	}

	result, err := ctrl.service.GetLogScoreezy(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(constant.Success, result))
}

func (ctrl *controller) ExportJobDetails(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	filter := &filterLogs{
		CompanyId: authCtx.CompanyIdStr(),
		StartDate: c.Query(constant.StartDate),
		EndDate:   c.Query(constant.EndDate),
		Size:      constant.SizeUnlimited,
	}

	if filter.StartDate == "" || filter.EndDate == "" {
		return apperror.BadRequest("start_date and end_date are required")
	}

	var buf bytes.Buffer
	filename, err := ctrl.service.ExportJobDetails(filter, &buf)
	if err != nil {
		return err
	}

	c.Set(constant.HeaderContentType, constant.TextOrCSVContentType)
	c.Set(constant.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))

	return c.SendStream(bytes.NewReader(buf.Bytes()))
}

// func (ctrl *controller) UploadCSV(c *fiber.Ctx) error {
// 	userId := fmt.Sprintf("%v", c.Locals(constant.UserId))
// 	companyId := fmt.Sprintf("%v", c.Locals(constant.CompanyId))
// 	tierLevel, _ := strconv.ParseUint(fmt.Sprintf("%v", c.Locals("tierLevel")), 10, 64)
// 	tempType := fmt.Sprintf("%v", c.Locals("tempType"))
// 	apiKey := c.Get(constant.XAPIKey)

// 	// Get the file from the form data
// 	fileHeader, err := c.FormFile("file")
// 	if err != nil {
// 		statusCode, resp := helper.GetError(constant.ErrorGettingFile)
// 		return c.Status(statusCode).JSON(resp)
// 	}

// 	file, err := fileHeader.Open()
// 	if err != nil {
// 		statusCode, resp := helper.GetError(constant.ErrorOpeningFile)
// 		return c.Status(statusCode).JSON(resp)
// 	}
// 	defer file.Close()

// 	// Create a CSV reader
// 	reader := csv.NewReader(file)

// 	// Read the header row
// 	header, err := reader.Read()
// 	if err != nil {
// 		statusCode, resp := helper.GetError(constant.ErrorReadingCSV)
// 		return c.Status(statusCode).JSON(resp)
// 	}

// 	// Process the header (first line)
// 	var validHeaderTemplate []string
// 	if tempType == "personal" {
// 		validHeaderTemplate = append(validHeaderTemplate, "loan_no", "name", "nik", "phone_number")
// 	} else {
// 		validHeaderTemplate = append(validHeaderTemplate, "company_id", "company_name", "npwp_company", "phone_number")
// 	}

// 	for _, v := range header {
// 		isValidHeader := helper.IsValidTemplateHeader(validHeaderTemplate, v)

// 		if !isValidHeader {
// 			statusCode, resp := helper.GetError(constant.HeaderTemplateNotValid)
// 			return c.Status(statusCode).JSON(resp)
// 		}
// 	}

// 	storeData := []BulkSearchRequest{}
// 	// Iterate over CSV records
// 	for {
// 		record, err := reader.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			statusCode, resp := helper.GetError(constant.ErrorReadingCSVRecords)
// 			return c.Status(statusCode).JSON(resp)
// 		}

// 		// Process the CSV record
// 		insertNew := BulkSearchRequest{}
// 		for _, v := range record {
// 			fmt.Println("v: ", v)
// 			insertNew.LoanNo = record[0]
// 			insertNew.Name = record[1]
// 			insertNew.NIK = record[2]
// 			insertNew.PhoneNumber = record[3]
// 		}
// 		storeData = append(storeData, insertNew)
// 	}

// 	processInsert := ctrl.Svc.BulkSearchUploadSvc(storeData, tempType, apiKey, userId, companyId)

// 	if processInsert != nil {
// 		statusCode, resp := helper.GetError(constant.ErrorUploadDataCSV)
// 		return c.Status(statusCode).JSON(resp)
// 	}

// 	bulkSearch, err := ctrl.Svc.GetBulkSearchSvc(uint(tierLevel), userId, companyId)
// 	if err != nil {
// 		statusCode, resp := helper.GetError(err.Error())
// 		return c.Status(statusCode).JSON(resp)
// 	}

// 	totalData, _ := ctrl.Svc.GetTotalDataBulk(uint(tierLevel), userId, companyId)

// 	fullResponsePage := map[string]interface{}{
// 		"total_data": totalData,
// 		"data":       bulkSearch,
// 	}

// 	resp := helper.SuccessResponse(
// 		"succeed to upload data",
// 		fullResponsePage,
// 	)

// 	return c.Status(fiber.StatusOK).JSON(resp)
// }

// func (ctrl *controller) GetBulkSearch(c *fiber.Ctx) error {
// 	userId := fmt.Sprintf("%v", c.Locals(constant.UserId))
// 	companyId := fmt.Sprintf("%v", c.Locals(constant.CompanyId))
// 	tierLevel, _ := strconv.ParseUint(fmt.Sprintf("%v", c.Locals("tierLevel")), 10, 64)
// 	// find user loggin detail

// 	bulkSearch, err := ctrl.Svc.GetBulkSearchSvc(uint(tierLevel), userId, companyId)
// 	if err != nil {
// 		statusCode, resp := helper.GetError(err.Error())
// 		return c.Status(statusCode).JSON(resp)
// 	}

// 	totalData, _ := ctrl.Svc.GetTotalDataBulk(uint(tierLevel), userId, companyId)

// 	fullResponsePage := map[string]interface{}{
// 		"total_data": totalData,
// 		"data":       bulkSearch,
// 	}

// 	resp := helper.SuccessResponse(
// 		"succeed to get bulk search data",
// 		fullResponsePage,
// 	)

// 	return c.Status(fiber.StatusOK).JSON(resp)
// }
