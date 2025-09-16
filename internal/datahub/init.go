package datahub

import (
	"front-office/configs/application"
	"front-office/internal/datahub/compliance/loanrecordchecker"
	"front-office/internal/datahub/compliance/multipleloan"
	"front-office/internal/datahub/identity/phonelivestatus"
	"front-office/internal/datahub/incometax/taxcompliancestatus"
	"front-office/internal/datahub/incometax/taxscore"
	"front-office/internal/datahub/incometax/taxverificationdetail"
	"front-office/internal/datahub/job"
	"front-office/pkg/httpclient"

	"time"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(routeAPI fiber.Router, cfg *application.Config) {
	client := httpclient.NewDefaultClient(10 * time.Second)

	complianceGroupAPI := routeAPI.Group("compliance")
	loanrecordchecker.SetupInit(complianceGroupAPI, cfg, client)
	multipleloan.SetupInit(complianceGroupAPI, cfg, client)
	job.SetupInit(complianceGroupAPI, cfg, client)

	incomeTaxGroupAPI := routeAPI.Group("incometax")
	taxcompliancestatus.SetupInit(incomeTaxGroupAPI, cfg, client)
	taxscore.SetupInit(incomeTaxGroupAPI, cfg, client)
	taxverificationdetail.SetupInit(incomeTaxGroupAPI, cfg, client)
	job.SetupInit(incomeTaxGroupAPI, cfg, client)

	identityGroupAPI := routeAPI.Group("identity")
	phonelivestatus.SetupInit(identityGroupAPI, cfg, client)
}
