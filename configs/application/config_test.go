package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	t.Setenv("FO_APP_ENV", "test")
	t.Setenv("FO_APP_PORT", "1234")

	env := GetConfig()

	assert.Equal(t, "1234", env.App.Port)
	assert.Equal(t, "test", env.App.AppEnv)
}
