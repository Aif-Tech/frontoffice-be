package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	t.Setenv("FO_APP_ENV", "test")
	t.Setenv("FO_APP_PORT", "1234")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "dbname")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_HOST", "localhost")

	env := GetConfig()

	assert.Equal(t, "1234", env.App.Port)
	assert.Equal(t, "user", env.Db.User)
	assert.Equal(t, "pass", env.Db.Password)
	assert.Equal(t, "dbname", env.Db.Name)
	assert.Equal(t, "5432", env.Db.Port)
	assert.Equal(t, "localhost", env.Db.Host)
	assert.Equal(t, "disable", env.Db.SSLMode)
	assert.Equal(t, "Asia/Jakarta", env.Db.TimeZone)

	assert.Equal(t, "test", env.Env.Env)
}
