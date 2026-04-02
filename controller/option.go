package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common/config"
	"one-api/common/logger"
	"one-api/common/utils"
	"one-api/model"
	"one-api/safty"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetOptions(c *gin.Context) {
	var options []*model.Option
	for k, v := range config.GlobalOption.GetAll() {
		if strings.HasSuffix(k, "Token") || strings.HasSuffix(k, "Secret") {
			continue
		}
		options = append(options, &model.Option{
			Key:   k,
			Value: utils.Interface2String(v),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    options,
	})
	return
}

func GetSafeTools(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    safty.GetAllSafeToolsName(),
	})
	return
}

func UpdateOption(c *gin.Context) {
	requestID := c.GetString(logger.RequestIdKey)
	requestStart := time.Now()

	var option model.Option
	err := json.NewDecoder(c.Request.Body).Decode(&option)
	if err != nil {
		logger.SysDebug(fmt.Sprintf("[option-debug] request_id=%s decode_failed err=%v", requestID, err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	logger.SysDebug(fmt.Sprintf("[option-debug] request_id=%s begin key=%s value_len=%d", requestID, option.Key, len(option.Value)))

	switch option.Key {
	case "GitHubOAuthEnabled":
		if option.Value == "true" && config.GitHubClientId == "" {
			logger.SysDebug(fmt.Sprintf("[option-debug] request_id=%s validation_failed key=%s reason=GitHubClientId empty", requestID, option.Key))
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 GitHub OAuth，请先填入 GitHub Client Id 以及 GitHub Client Secret！",
			})
			return
		}
	case "OIDCAuthEnabled":
		if option.Value == "true" && (config.OIDCClientId == "" || config.OIDCClientSecret == "" || config.OIDCIssuer == "" || config.OIDCScopes == "" || config.OIDCUsernameClaims == "") {
			logger.SysDebug(fmt.Sprintf("[option-debug] request_id=%s validation_failed key=%s reason=OIDC config incomplete", requestID, option.Key))
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 OIDC，请先填入OIDC信息！",
			})
			return
		}
	case "EmailDomainRestrictionEnabled":
		if option.Value == "true" && len(config.EmailDomainWhitelist) == 0 {
			logger.SysDebug(fmt.Sprintf("[option-debug] request_id=%s validation_failed key=%s reason=EmailDomainWhitelist empty", requestID, option.Key))
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用邮箱域名限制，请先填入限制的邮箱域名！",
			})
			return
		}
	case "WeChatAuthEnabled":
		if option.Value == "true" && config.WeChatServerAddress == "" {
			logger.SysDebug(fmt.Sprintf("[option-debug] request_id=%s validation_failed key=%s reason=WeChatServerAddress empty", requestID, option.Key))
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用微信登录，请先填入微信登录相关配置信息！",
			})
			return
		}
	case "TurnstileCheckEnabled":
		if option.Value == "true" && config.TurnstileSiteKey == "" {
			logger.SysDebug(fmt.Sprintf("[option-debug] request_id=%s validation_failed key=%s reason=TurnstileSiteKey empty", requestID, option.Key))
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 Turnstile 校验，请先填入 Turnstile 校验相关配置信息！",
			})
			return
		}
	}

	dbStart := time.Now()
	err = model.UpdateOption(option.Key, option.Value)
	dbCost := time.Since(dbStart)
	if err != nil {
		logger.SysError(fmt.Sprintf("[option-debug] request_id=%s update_failed key=%s db_cost=%s total_cost=%s err=%v", requestID, option.Key, dbCost, time.Since(requestStart), err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	logger.SysDebug(fmt.Sprintf("[option-debug] request_id=%s update_success key=%s db_cost=%s total_cost=%s", requestID, option.Key, dbCost, time.Since(requestStart)))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}
