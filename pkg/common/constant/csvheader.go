package constant

const (
	CSVHeaderName             = "Name"
	CSVHeaderIDCard           = "ID Card Number"
	CSVHeaderPhone            = "Phone Number"
	CSVHeaderLoanNumber       = "Loan Number"
	CSVHeaderNPWP             = "NPWP"
	CSVHeaderRemarks          = "Remarks"
	CSVHeaderQueryCount       = "Query Count"
	CSVHeaderDataStatus       = "Data Status"
	CSVHeaderAddress          = "Address"
	CSVHeaderScore            = "Score"
	CSVHeaderNPWPVerification = "NPWP Verification"
	CSVHeaderTaxCompliance    = "Tax Compliance"
	CSVHeaderStatus           = "Status"
	CSVHeaderDescription      = "Description"
)

var CSVTemplateHeaderLoanRecord = []string{
	CSVHeaderName,
	CSVHeaderIDCard,
	CSVHeaderPhone,
	CSVHeaderLoanNumber,
}

var CSVTemplateHeaderMultipleLoan = []string{
	CSVHeaderIDCard,
	CSVHeaderPhone,
	CSVHeaderLoanNumber,
}

var CSVTemplateHeaderPhoneLive = []string{
	CSVHeaderPhone,
	CSVHeaderLoanNumber,
}

var CSVTemplateHeaderTaxCompliance = []string{
	CSVHeaderNPWP,
	CSVHeaderLoanNumber,
}

var CSVTemplateHeaderTaxScore = []string{
	CSVHeaderNPWP,
	CSVHeaderLoanNumber,
}

var CSVTemplateHeaderTaxVerification = []string{
	CSVHeaderIDCard,
	CSVHeaderLoanNumber,
}

var CSVExportHeaderLoanRecord = []string{
	CSVHeaderLoanNumber,
	CSVHeaderName,
	CSVHeaderIDCard,
	CSVHeaderPhone,
	CSVHeaderRemarks,
	CSVHeaderDataStatus,
	CSVHeaderStatus,
	CSVHeaderDescription,
}

var CSVExportHeaderMultipleLoan = []string{
	CSVHeaderLoanNumber,
	CSVHeaderIDCard,
	CSVHeaderPhone,
	CSVHeaderQueryCount,
	CSVHeaderStatus,
	CSVHeaderDescription,
}

var CSVExportHeaderTaxCompliance = []string{
	CSVHeaderLoanNumber,
	CSVHeaderNPWP,
	CSVHeaderName,
	CSVHeaderAddress,
	CSVHeaderDataStatus,
	CSVHeaderStatus,
	CSVHeaderDescription,
}

var CSVExportHeaderTaxScore = []string{
	CSVHeaderLoanNumber,
	CSVHeaderNPWP,
	CSVHeaderName,
	CSVHeaderAddress,
	CSVHeaderDataStatus,
	CSVHeaderScore,
	CSVHeaderStatus,
	CSVHeaderDescription,
}

var CSVExportHeaderTaxVerification = []string{
	CSVHeaderLoanNumber,
	CSVHeaderName,
	CSVHeaderAddress,
	CSVHeaderNPWP,
	CSVHeaderNPWPVerification,
	CSVHeaderDataStatus,
	CSVHeaderTaxCompliance,
	CSVHeaderStatus,
	CSVHeaderDescription,
}
