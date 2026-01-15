package constant

const (
	CSVHeaderName                 = "Name"
	CSVHeaderDateCreated          = "Date Created"
	CSVHeaderIDCard               = "ID Card Number"
	CSVHeaderPhone                = "Phone Number"
	CSVHeaderLoanNumber           = "Loan Number"
	CSVHeaderNPWP                 = "NPWP"
	CSVHeaderRemarks              = "Remarks"
	CSVHeaderQueryCount           = "Query Count"
	CSVHeaderDataStatus           = "Data Status"
	CSVHeaderAddress              = "Address"
	CSVHeaderScore                = "Score"
	CSVHeaderNPWPVerification     = "NPWP Verification"
	CSVHeaderTaxCompliance        = "Tax Compliance"
	CSVHeaderSubscriberStatus     = "Subscriber Status"
	CSVHeaderDeviceStatus         = "Device Status"
	CSVHeaderOperator             = "Operator"
	CSVHeaderPhoneType            = "Phone Type"
	CSVHeaderProbabilityToDefault = "Probability To Default"
	CSVHeaderGrade                = "Grade"
	CSVHeaderBehavior             = "Behavior"
	CSVHeaderIdentity             = "Identity"
	CSVHeaderPeriod               = "Period"
	CSVHeaderStatus               = "Status"
	CSVHeaderDescription          = "Description"
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
	// CSVHeaderLoanNumber,
}

var CSVTemplateHeaderTaxScore = []string{
	CSVHeaderNPWP,
	CSVHeaderLoanNumber,
}

var CSVTemplateHeaderTaxVerification = []string{
	CSVHeaderIDCard,
	CSVHeaderLoanNumber,
}

var CSVTemplateHeaderNPWPVerification = []string{
	CSVHeaderNPWP,
	CSVHeaderLoanNumber,
}

var CSVExportHeaderPhoneLive = []string{
	CSVHeaderLoanNumber,
	CSVHeaderPhone,
	CSVHeaderSubscriberStatus,
	CSVHeaderDeviceStatus,
	CSVHeaderOperator,
	CSVHeaderPhoneType,
	CSVHeaderStatus,
	CSVHeaderDescription,
}

var CSVTemplateHeaderRecycleNumber = []string{
	CSVHeaderPhone,
	CSVHeaderLoanNumber,
	CSVHeaderPeriod,
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
	// CSVHeaderLoanNumber,
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

var CSVExportHeaderNPWPVerification = []string{
	CSVHeaderLoanNumber,
	CSVHeaderNPWP,
	CSVHeaderName,
	CSVHeaderAddress,
	CSVHeaderDataStatus,
	CSVHeaderStatus,
	CSVHeaderDescription,
}

var CSVExportHeaderRecycleNumber = []string{
	CSVHeaderLoanNumber,
	CSVHeaderPhone,
	CSVHeaderDataStatus,
	CSVHeaderStatus,
	CSVHeaderDescription,
}

var CSVExportHeaderGenRetail = []string{
	// CSVHeaderDateCreated,
	CSVHeaderLoanNumber,
	CSVHeaderName,
	CSVHeaderIDCard,
	CSVHeaderPhone,
	CSVHeaderProbabilityToDefault,
	CSVHeaderGrade,
	CSVHeaderBehavior,
	CSVHeaderIdentity,
	CSVHeaderDescription,
}
