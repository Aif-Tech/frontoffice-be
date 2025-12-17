package constant

const (
	// auth
	EventSignIn  = "sign in"
	EventSignOut = "sign out"

	// user
	EventChangePassword       = "change password"
	EventRequestPasswordReset = "request password reset"
	EventPasswordReset        = "password reset"
	EventRegisterMember       = "register member"
	EventUpdateProfile        = "update profile"
	EventUpdateUserData       = "update user data"
	EventActivateUser         = "activate user"
	EventInactivateUser       = "inactivate user"

	// billing
	EventChangeBillingInformation  = "update billing information"
	EventTopupBalance              = "topup balance"
	EventSubmitPaymentConfirmation = "submit payment confirmation"

	// scoreezy
	EventScoreezySingleReq       = "scoreezy single request"
	EventScoreezyBulkReq         = "scoreezy bulk request"
	EventScoreezySingleDownload  = "scoreezy single download result"
	EventScoreezyBulkDownload    = "scoreezy bulk download result"
	EventScoreezyDownloadSummary = "scoreezy download result summary"

	// loan record
	EventLoanRecordSingleReq       = "loan record single request"
	EventLoanRecordBulkReq         = "loan record bulk request"
	EventLoanRecordDownload        = "loan record download result"
	EventLoanRecordDownloadSummary = "loan record download result summary"

	// 7 days multiple loan
	Event7DMLSingleReq       = "7d multiple loan single request"
	Event7DMLBulkReq         = "7d multiple loan bulk request"
	Event7DMLDownload        = "7d multiple loan download result"
	Event7DMLDownloadSummary = "7d multiple loan download result summary"

	// 30 days multiple loan
	Event30DMLSingleReq       = "30d multiple loan single request"
	Event30DMLBulkReq         = "30d multiple loan bulk request"
	Event30DMLDownload        = "30d multiple loan download result"
	Event30DMLDownloadSummary = "30d multiple loan download result summary"

	// 90 days multiple loan
	Event90DMLSingleReq       = "90d multiple loan single request"
	Event90DMLBulkReq         = "90d multiple loan bulk request"
	Event90DMLDownload        = "90d multiple loan download result"
	Event90DMLDownloadSummary = "90d multiple loan download result summary"

	// npwp verification
	EventNPWPVerificationSingleReq       = "npwp verification single request"
	EventNPWPVerificationBulkReq         = "npwp verification bulk request"
	EventNPWPVerificationDownload        = "npwp verification download result"
	EventNPWPVerificationDownloadSummary = "npwp verification download result summary"

	// phone live status
	EventPhoneLiveSingleReq       = "phone live status single request"
	EventPhoneLiveBulkReq         = "phone live status bulk request"
	EventPhoneLiveDownload        = "phone live status download result"
	EventPhoneLiveDownloadSummary = "phone live status download result summary"

	// recycle number
	EventRecycleNumberSingleReq       = "recycle number single request"
	EventRecycleNumberBulkReq         = "recycle number bulk request"
	EventRecycleNumberDownload        = "recycle number download result"
	EventRecycleNumberDownloadSummary = "recycle number download result summary"

	// tax compliance status
	EventTaxComplianceSingleReq       = "tax compliance single request"
	EventTaxComplianceBulkReq         = "tax compliance bulk request"
	EventTaxComplianceDownload        = "tax compliance download result"
	EventTaxComplianceDownloadSummary = "tax compliance download result summary"

	// tax score
	EventTaxScoreSingleReq       = "tax score single request"
	EventTaxScoreBulkReq         = "tax score bulk request"
	EventTaxScoreDownload        = "tax score download result"
	EventTaxScoreDownloadSummary = "tax score download result summary"

	// tax verification detail
	EventTaxVerificationSingleReq       = "tax verification single request"
	EventTaxVerificationBulkReq         = "tax verification bulk request"
	EventTaxVerificationDownload        = "tax verification download result"
	EventTaxVerificationDownloadSummary = "tax verification download result summary"
)
