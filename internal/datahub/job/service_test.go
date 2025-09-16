package job

import (
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapLoanRecordCheckerRow(t *testing.T) {
	t.Run("should map all fields correctly when isMasked is true", func(t *testing.T) {
		message := "Succeed"
		result := mapLoanRecordCheckerRow(true, &logTransProductCatalog{
			Input: &logTransInput{
				Name:        helper.StringPtr(constant.DummyName),
				NIK:         helper.StringPtr("unmasked-nik"),
				PhoneNumber: helper.StringPtr("unmasked-phone"),
			},
			Data: &logTransData{
				Remarks: helper.StringPtr("-"),
				Status:  helper.StringPtr(""),
			},
			Message: &message,
			RefTransProductCatalog: map[string]interface{}{
				"input": map[string]interface{}{
					"name":         constant.DummyName,
					"nik":          constant.DummyNIK,
					"phone_number": constant.DummyPhoneNumber,
				},
			},
		})

		expected := []string{
			constant.DummyName,
			constant.DummyNIK,
			constant.DummyPhoneNumber,
			"-",
			"",
			"",
			"Succeed",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("should map all fields correctly when isMasked is false", func(t *testing.T) {
		message := "Succeed"
		result := mapLoanRecordCheckerRow(false, &logTransProductCatalog{
			Input: &logTransInput{
				Name:        helper.StringPtr(constant.DummyName),
				NIK:         helper.StringPtr(constant.DummyNIK),
				PhoneNumber: helper.StringPtr(constant.DummyPhoneNumber),
			},
			Data: &logTransData{
				Remarks: helper.StringPtr("-"),
				Status:  helper.StringPtr(""),
			},
			Message: &message,
		})

		expected := []string{
			constant.DummyName,
			constant.DummyNIK,
			constant.DummyPhoneNumber,
			"-",
			"",
			"",
			"Succeed",
		}
		assert.Equal(t, expected, result)
	})
}
