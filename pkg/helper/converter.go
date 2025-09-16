package helper

import (
	"fmt"
	"strconv"
	"time"
)

func ConvertUintToString(arg uint) string {
	return strconv.Itoa(int(arg))
}

func InterfaceToUint(input interface{}) (uint, error) {
	if val, ok := input.(uint); ok {
		return val, nil
	}

	return 0, fmt.Errorf("cannot convert %T to uint", input)
}

func BoolPtr(b bool) *bool           { return &b }
func IntPtr(i int) *int              { return &i }
func UintPtr(u uint) *uint           { return &u }
func StringPtr(s string) *string     { return &s }
func TimePtr(t time.Time) *time.Time { return &t }
