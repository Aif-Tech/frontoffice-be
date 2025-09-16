package helper

import (
	"front-office/pkg/common/constant"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertUintToString(t *testing.T) {
	assert.Equal(t, "123", ConvertUintToString(123))
	assert.Equal(t, "0", ConvertUintToString(0))
}

func TestInterfaceToUint(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		val, err := InterfaceToUint(uint(42))
		assert.NoError(t, err)
		assert.Equal(t, uint(42), val)
	})

	t.Run("Invalid Type", func(t *testing.T) {
		val, err := InterfaceToUint("not a uint")
		assert.Error(t, err)
		assert.Equal(t, uint(0), val)
		assert.Contains(t, err.Error(), "cannot convert string to uint")
	})
}

func TestBoolPtr(t *testing.T) {
	b := true
	assert.Equal(t, &b, BoolPtr(true))
}

func TestIntPtr(t *testing.T) {
	i := 42
	assert.Equal(t, &i, IntPtr(42))
}

func TestUintPtr(t *testing.T) {
	u := uint(123)
	assert.Equal(t, &u, UintPtr(123))
}

func TestStringPtr(t *testing.T) {
	s := "hello"
	assert.Equal(t, &s, StringPtr("hello"))
}

func TestTimePtr(t *testing.T) {
	now := time.Now()
	assert.Equal(t, &now, TimePtr(now))
}
