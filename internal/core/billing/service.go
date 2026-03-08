package billing

import (
	"fmt"
	"front-office/internal/mail"
	"front-office/pkg/apperror"
	"time"

	"github.com/rs/zerolog/log"
)

func NewService(
	repo Repository,
	mailSvc *mail.SendMailService,
) Service {
	return &service{
		repo,
		mailSvc,
	}
}

type service struct {
	repo    Repository
	mailSvc *mail.SendMailService
}

type Service interface {
	SendMonthlyUsageReport() error
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

	for _, summary := range summaries {
		adminEmails, err := svc.repo.GetAdminEmails(summary.CompanyId)
		if err != nil {
			log.Warn().
				Err(err).
				Uint("company_id", summary.CompanyId).
				Msg("failed to get admin emails")
			continue
		}

		if len(adminEmails) == 0 {
			log.Warn().
				Uint("company_id", summary.CompanyId).
				Msg("no admin emails found, skipping")
			continue
		}

		toEmails := make([]string, 0, len(adminEmails))
		for _, admin := range adminEmails {
			toEmails = append(toEmails, admin.Email)
		}

		if err := svc.mailSvc.SendWithTemplate(
			toEmails,
			nil,
			fmt.Sprintf("Monthly Usage Report for %s - %s %d", summary.CompanyName, month, year),
			"monthly_usage_report.html",
			map[string]any{
				"Name":     summary.CompanyName,
				"Products": summary.Products,
				"Month":    month.String(),
				"Year":     year,
			},
			nil,
		); err != nil {
			fmt.Println("errrrrrr ", err)
			log.Warn().
				Err(err).
				Uint("company_id", summary.CompanyId).
				Msg("failed to send monthly usage report")
		}
	}

	return nil
}
