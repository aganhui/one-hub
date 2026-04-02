package model

import (
	"fmt"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/logger"
	"strings"
	"time"

	"gorm.io/gorm/clause"
)

type Option struct {
	Key   string `json:"key" gorm:"primaryKey"`
	Value string `json:"value"`
}

type DBStats struct {
	OpenConnections int
	InUse           int
	Idle            int
	WaitCount       int64
	WaitDuration    time.Duration
}

func DBStatsSnapshot() DBStats {
	sqlDB, err := DB.DB()
	if err != nil {
		return DBStats{}
	}

	stats := sqlDB.Stats()
	return DBStats{
		OpenConnections: stats.OpenConnections,
		InUse:           stats.InUse,
		Idle:            stats.Idle,
		WaitCount:       stats.WaitCount,
		WaitDuration:    stats.WaitDuration,
	}
}

func AllOption() ([]*Option, error) {
	var options []*Option
	err := DB.Find(&options).Error
	return options, err
}

func GetOption(key string) (option Option, err error) {
	err = DB.First(&option, Option{Key: key}).Error
	return
}

func InitOptionMap() {

	config.GlobalOption.RegisterBool("PasswordLoginEnabled", &config.PasswordLoginEnabled)
	config.GlobalOption.RegisterBool("PasswordRegisterEnabled", &config.PasswordRegisterEnabled)
	config.GlobalOption.RegisterBool("EmailVerificationEnabled", &config.EmailVerificationEnabled)
	config.GlobalOption.RegisterBool("GitHubOAuthEnabled", &config.GitHubOAuthEnabled)
	config.GlobalOption.RegisterBool("WeChatAuthEnabled", &config.WeChatAuthEnabled)
	config.GlobalOption.RegisterBool("LarkAuthEnabled", &config.LarkAuthEnabled)
	config.GlobalOption.RegisterBool("OIDCAuthEnabled", &config.OIDCAuthEnabled)
	config.GlobalOption.RegisterBool("TurnstileCheckEnabled", &config.TurnstileCheckEnabled)
	config.GlobalOption.RegisterBool("RegisterEnabled", &config.RegisterEnabled)
	config.GlobalOption.RegisterBool("AutomaticDisableChannelEnabled", &config.AutomaticDisableChannelEnabled)
	config.GlobalOption.RegisterBool("AutomaticEnableChannelEnabled", &config.AutomaticEnableChannelEnabled)
	config.GlobalOption.RegisterBool("ApproximateTokenEnabled", &config.ApproximateTokenEnabled)
	config.GlobalOption.RegisterBool("LogConsumeEnabled", &config.LogConsumeEnabled)
	config.GlobalOption.RegisterBool("DisplayInCurrencyEnabled", &config.DisplayInCurrencyEnabled)
	config.GlobalOption.RegisterFloat("ChannelDisableThreshold", &config.ChannelDisableThreshold)
	config.GlobalOption.RegisterBool("EmailDomainRestrictionEnabled", &config.EmailDomainRestrictionEnabled)

	config.GlobalOption.RegisterCustom("EmailDomainWhitelist", func() string {
		return strings.Join(config.EmailDomainWhitelist, ",")
	}, func(value string) error {
		config.EmailDomainWhitelist = strings.Split(value, ",")
		return nil
	}, "")

	config.GlobalOption.RegisterString("SMTPServer", &config.SMTPServer)
	config.GlobalOption.RegisterString("SMTPFrom", &config.SMTPFrom)
	config.GlobalOption.RegisterInt("SMTPPort", &config.SMTPPort)
	config.GlobalOption.RegisterString("SMTPAccount", &config.SMTPAccount)
	config.GlobalOption.RegisterString("SMTPToken", &config.SMTPToken)
	config.GlobalOption.RegisterValue("Notice")
	config.GlobalOption.RegisterValue("About")
	config.GlobalOption.RegisterValue("HomePageContent")
	config.GlobalOption.RegisterString("Footer", &config.Footer)
	config.GlobalOption.RegisterString("SystemName", &config.SystemName)
	config.GlobalOption.RegisterString("Logo", &config.Logo)
	config.GlobalOption.RegisterString("AnalyticsCode", &config.AnalyticsCode)
	config.GlobalOption.RegisterString("ServerAddress", &config.ServerAddress)
	config.GlobalOption.RegisterString("GitHubClientId", &config.GitHubClientId)
	config.GlobalOption.RegisterString("GitHubClientSecret", &config.GitHubClientSecret)

	config.GlobalOption.RegisterString("OIDCClientId", &config.OIDCClientId)
	config.GlobalOption.RegisterString("OIDCClientSecret", &config.OIDCClientSecret)
	config.GlobalOption.RegisterString("OIDCIssuer", &config.OIDCIssuer)
	config.GlobalOption.RegisterString("OIDCScopes", &config.OIDCScopes)
	config.GlobalOption.RegisterString("OIDCUsernameClaims", &config.OIDCUsernameClaims)

	config.GlobalOption.RegisterString("WeChatServerAddress", &config.WeChatServerAddress)
	config.GlobalOption.RegisterString("WeChatServerToken", &config.WeChatServerToken)
	config.GlobalOption.RegisterString("WeChatAccountQRCodeImageURL", &config.WeChatAccountQRCodeImageURL)
	config.GlobalOption.RegisterString("TurnstileSiteKey", &config.TurnstileSiteKey)
	config.GlobalOption.RegisterString("TurnstileSecretKey", &config.TurnstileSecretKey)
	config.GlobalOption.RegisterInt("QuotaForNewUser", &config.QuotaForNewUser)
	config.GlobalOption.RegisterInt("QuotaForInviter", &config.QuotaForInviter)
	config.GlobalOption.RegisterInt("QuotaForInvitee", &config.QuotaForInvitee)
	config.GlobalOption.RegisterInt("QuotaRemindThreshold", &config.QuotaRemindThreshold)
	config.GlobalOption.RegisterInt("PreConsumedQuota", &config.PreConsumedQuota)

	config.GlobalOption.RegisterString("TopUpLink", &config.TopUpLink)
	config.GlobalOption.RegisterString("ChatLink", &config.ChatLink)
	config.GlobalOption.RegisterString("ChatLinks", &config.ChatLinks)
	config.GlobalOption.RegisterFloat("QuotaPerUnit", &config.QuotaPerUnit)
	config.GlobalOption.RegisterInt("RetryTimes", &config.RetryTimes)
	config.GlobalOption.RegisterInt("RetryCooldownSeconds", &config.RetryCooldownSeconds)

	config.GlobalOption.RegisterBool("MjNotifyEnabled", &config.MjNotifyEnabled)
	config.GlobalOption.RegisterString("ChatImageRequestProxy", &config.ChatImageRequestProxy)
	config.GlobalOption.RegisterFloat("PaymentUSDRate", &config.PaymentUSDRate)
	config.GlobalOption.RegisterInt("PaymentMinAmount", &config.PaymentMinAmount)

	config.GlobalOption.RegisterCustom("RechargeDiscount", func() string {
		return common.RechargeDiscount2JSONString()
	}, func(value string) error {
		config.RechargeDiscount = value
		common.UpdateRechargeDiscountByJSONString(value)
		return nil
	}, "")

	config.GlobalOption.RegisterString("CFWorkerImageUrl", &config.CFWorkerImageUrl)
	config.GlobalOption.RegisterString("CFWorkerImageKey", &config.CFWorkerImageKey)
	config.GlobalOption.RegisterInt("OldTokenMaxId", &config.OldTokenMaxId)
	config.GlobalOption.RegisterBool("GitHubOldIdCloseEnabled", &config.GitHubOldIdCloseEnabled)

	config.GlobalOption.RegisterBool("GeminiAPIEnabled", &config.GeminiAPIEnabled)
	config.GlobalOption.RegisterBool("ClaudeAPIEnabled", &config.ClaudeAPIEnabled)

	config.GlobalOption.RegisterCustom("DisableChannelKeywords", func() string {
		return common.DisableChannelKeywordsInstance.GetKeywords()
	}, func(value string) error {
		common.DisableChannelKeywordsInstance.Load(value)
		return nil
	}, common.GetDefaultDisableChannelKeywords())

	config.GlobalOption.RegisterInt("RetryTimeOut", &config.RetryTimeOut)

	config.GlobalOption.RegisterBool("EnableSafe", &config.EnableSafe)
	config.GlobalOption.RegisterString("SafeToolName", &config.SafeToolName)
	config.GlobalOption.RegisterCustom("SafeKeyWords", func() string {
		return strings.Join(config.SafeKeyWords, "\n")
	}, func(value string) error {
		config.SafeKeyWords = strings.Split(value, "\n")
		return nil
	}, "")

	loadOptionsFromDatabase()
}

func loadOptionsFromDatabase() {
	options, _ := AllOption()
	for _, option := range options {
		err := config.GlobalOption.Set(option.Key, option.Value)
		if err != nil {
			logger.SysError("failed to update option map: " + err.Error())
		}
	}
}

func SyncOptions(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		logger.SysLog("syncing options from database")
		loadOptionsFromDatabase()
	}
}

func UpdateOption(key string, value string) error {
	start := time.Now()

	statsBefore := DBStatsSnapshot()
	logger.SysDebug(fmt.Sprintf("[option-debug] db_start key=%s open=%d in_use=%d idle=%d wait_count=%d wait_duration=%s", key, statsBefore.OpenConnections, statsBefore.InUse, statsBefore.Idle, statsBefore.WaitCount, statsBefore.WaitDuration))

	option := Option{
		Key:   key,
		Value: value,
	}

	err := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&option).Error

	statsAfter := DBStatsSnapshot()
	if err != nil {
		logger.SysError(fmt.Sprintf("[option-debug] db_failed key=%s cost=%s open=%d in_use=%d idle=%d wait_count=%d wait_duration=%s err=%v", key, time.Since(start), statsAfter.OpenConnections, statsAfter.InUse, statsAfter.Idle, statsAfter.WaitCount, statsAfter.WaitDuration, err))
		return err
	}

	logger.SysDebug(fmt.Sprintf("[option-debug] db_success key=%s cost=%s open=%d in_use=%d idle=%d wait_count=%d wait_duration=%s", key, time.Since(start), statsAfter.OpenConnections, statsAfter.InUse, statsAfter.Idle, statsAfter.WaitCount, statsAfter.WaitDuration))

	// Update OptionMap
	return config.GlobalOption.Set(key, value)
}
