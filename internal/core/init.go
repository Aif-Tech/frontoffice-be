package core

import (
	"front-office/configs/application"
	"front-office/internal/core/auth"
	"front-office/internal/core/grade"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/member"
	"front-office/internal/core/role"
	"front-office/internal/core/template"
	"front-office/internal/datahub"
	"front-office/internal/mail"
	"front-office/internal/scoreezy"
	"front-office/pkg/httpclient"

	"time"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(routeGroup fiber.Router, cfg *application.Config) {
	client := httpclient.NewDefaultClient(10 * time.Second)

	mailModule := mail.Init(cfg)

	userGroup := routeGroup.Group("users")
	auth.SetupInit(userGroup, cfg, client, mailModule.SendMail)
	member.SetupInit(userGroup, cfg, client)

	roleGroup := routeGroup.Group("roles")
	role.SetupInit(roleGroup, cfg, client)

	gradeGroup := routeGroup.Group("grades")
	grade.SetupInit(gradeGroup, cfg, client)

	logGroup := routeGroup.Group("logs")
	transaction.SetupInit(logGroup, cfg, client)
	operation.SetupInit(logGroup, cfg, client)

	productGroup := routeGroup.Group("products")
	datahub.SetupInit(productGroup, cfg)
	scoreezy.SetupInit(productGroup, cfg, client)

	templateGroup := routeGroup.Group("templates")
	template.SetupInit(templateGroup)
}
