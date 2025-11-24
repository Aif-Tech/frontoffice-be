package constant

const (
	EventSignIn                     = "sign in"
	EventSignOut                    = "sign out"
	EventChangePassword             = "change password"
	EventRequestPasswordReset       = "request password reset"
	EventPasswordReset              = "password reset"
	EventRegisterMember             = "add new user"
	EventUpdateProfile              = "update profile account"
	EventUpdateUserData             = "updates user data"
	EventActivateUser               = "activate user"
	EventInactivateUser             = "inactivate user"
	EventCalculateScoreSingle       = "calculate score single"
	EventCalculateScoreBulk         = "calculate score bulk"
	EventDownloadScoreHistorySingle = "download history single"
	EventDownloadScoreHistoryBulk   = "download result bulk"
	EventChangeBillingInformation   = "change billing information"
	EventTopupBalance               = "topup balance"
	EventSubmitPaymentConfirmation  = "submit payment confirmation"

	EventLoanRecordSingleHit      = "loan record checker single hit request"
	EventLoanRecordBulkHit        = "loan record checker bulk hit request"
	EventLoanRecordDownloadResult = "loan record checker download result"
)
