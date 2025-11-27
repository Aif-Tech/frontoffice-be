package constant

const (
	// auth
	EventSignIn               = "sign in"
	EventSignOut              = "sign out"
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
	EventScoreezySingleReq      = "scoreezy single request"
	EventScoreezyBulkReq        = "scoreezy bulk request"
	EventScoreezySingleDownload = "scoreezy single download result"
	EventScoreezyBulkDownload   = "scoreezy bulk download result"

	// loan record
	EventLoanRecordSingleReq      = "loan record single request"
	EventLoanRecordBulkReq        = "loan record bulk request"
	EventLoanRecordDownloadResult = "loan record download result"

	Event7DMLSingleReq      = "7d multiple loan single request"
	Event7DMLBulkReq        = "7d multiple loan bulk request"
	Event7DMLDownloadResult = "7d multiple loan download result"

	Event30DMLSingleReq      = "30d multiple loan single request"
	Event30DMLBulkReq        = "30d multiple loan bulk request"
	Event30DMLDownloadResult = "30d multiple loan download result"

	Event90DMLSingleReq      = "90d multiple loan single request"
	Event90DMLBulkReq        = "90d multiple loan bulk request"
	Event90DMLDownloadResult = "90d multiple loan download result"

	EventNPWPVerificationSingleReq = "npwp verification single request"
	EventPhoneLiveSingleReq        = "phone live status single request"
	EventTaxComplianceSingleReq    = "tax compliance status single request"
	EventTaxScoreSingleReq         = "tax score single request"
	EventTaxVerificationSingleReq  = "tax verification single request"
)
