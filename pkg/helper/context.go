package helper

import (
	"fmt"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetUintLocal(c *fiber.Ctx, key string) (uint, error) {
	val, ok := c.Locals(key).(uint)
	if !ok {
		return 0, fmt.Errorf("invalid or missing '%s' in context", key)
	}

	return val, nil
}

func GetStringLocal(c *fiber.Ctx, key string) (string, error) {
	val, ok := c.Locals(key).(string)
	if !ok {
		return "", fmt.Errorf("invalid or missing '%s' in context", key)
	}

	return val, nil
}

func GetUintLocalStr(c *fiber.Ctx, key string) (string, error) {
	val, err := GetUintLocal(c, key)
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(uint64(val), 10), nil
}

func GetAuthContext(c *fiber.Ctx) (*model.AuthContext, error) {
	userId, ok := c.Locals(constant.UserId).(uint)
	if !ok {
		return nil, fmt.Errorf("invalid or missing user id")
	}

	companyId, ok := c.Locals(constant.CompanyId).(uint)
	if !ok {
		return nil, fmt.Errorf("invalid or missing company id")
	}

	roleID, ok := c.Locals(constant.RoleId).(uint)
	if !ok {
		return nil, fmt.Errorf("invalid or missing role id")
	}

	apiKey, ok := c.Locals(constant.APIKey).(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing api key")
	}

	quotaType, ok := c.Locals(constant.QuotaType).(uint)
	if !ok {
		return nil, fmt.Errorf("invalid or missing quota type")
	}

	return &model.AuthContext{
		UserId:    userId,
		CompanyId: companyId,
		RoleId:    roleID,
		APIKey:    apiKey,
		QuotaType: quotaType,
	}, nil
}
