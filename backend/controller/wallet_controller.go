package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserWallet 获取当前用户的钱包余额和最近交易流水
func GetUserWallet(c *gin.Context) {
	baseUserID := getBaseUserID(c)
	if baseUserID == 0 {
		return
	}

	wallet, _ := ensureUserWallet(baseUserID)

	// 取最近 50 条流水
	var txns []models.WalletTransaction
	global.Db.Where("user_id = ?", baseUserID).
		Order("created_at desc").
		Limit(50).
		Find(&txns)
	if txns == nil {
		txns = []models.WalletTransaction{}
	}

	utils.Success(c, gin.H{
		"balance":      wallet.Balance,
		"transactions": txns,
	})
}

// RechargeWallet 充值（模拟 QR 码支付，自动成功）
type RechargeRequest struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description,omitempty"`
}

func RechargeWallet(c *gin.Context) {
	baseUserID := getBaseUserID(c)
	if baseUserID == 0 {
		return
	}

	var req RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "无效的请求参数")
		return
	}
	if req.Amount <= 0 {
		utils.Fail(c, "充值金额必须大于 0")
		return
	}
	if req.Amount > 99999 {
		utils.Fail(c, "单次充值金额不能超过 99999")
		return
	}

	// 模拟 QR 码 —— 返回一个假的 code_url
	codeURL := fmt.Sprintf("https://qr.example.com/fake?t=%d&amt=%.2f", time.Now().UnixNano(), req.Amount)

	utils.Success(c, gin.H{
		"code_url": codeURL,
		"amount":   req.Amount,
		"message":  "请使用微信/支付宝扫码支付",
	})
}

// ConfirmRecharge 确认充值（前端模拟支付成功后调用，类似 PayOrder）
func ConfirmRecharge(c *gin.Context) {
	baseUserID := getBaseUserID(c)
	if baseUserID == 0 {
		return
	}

	var req struct {
		Amount      float64 `json:"amount"`
		Description string  `json:"description,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "无效的请求参数")
		return
	}
	if req.Amount <= 0 {
		utils.Fail(c, "金额无效")
		return
	}

	desc := req.Description
	if desc == "" {
		desc = fmt.Sprintf("余额充值 ¥%.2f", req.Amount)
	}

	wallet, _ := ensureUserWallet(baseUserID)

	before := wallet.Balance
	after := before + req.Amount

	tx := global.Db.Begin()
	if err := tx.Model(&models.UserWallet{}).Where("user_id = ?", baseUserID).
		Update("balance", gorm.Expr("balance + ?", req.Amount)).Error; err != nil {
		tx.Rollback()
		utils.Fail(c, "充值失败")
		return
	}

	// 流水记录
	if err := tx.Create(&models.WalletTransaction{
		UserID:        baseUserID,
		Amount:        req.Amount,
		Type:          "recharge",
		BalanceBefore: before,
		BalanceAfter:  after,
		Description:   desc,
	}).Error; err != nil {
		tx.Rollback()
		utils.Fail(c, "记录流水失败")
		return
	}

	tx.Commit()

	// 生成一个假的 6 位交易号供显示
	refNo := fmt.Sprintf("R%06d", rand.Intn(1000000))

	utils.Success(c, gin.H{
		"balance":          after,
		"amount":           req.Amount,
		"reference_no":     refNo,
		"transaction_type": "recharge",
	})
}

// PayWithWallet 使用钱包余额支付订单（在 PayOrder 中被调用，也可独立使用）
func PayWithWallet(c *gin.Context) {
	baseUserID := getBaseUserID(c)
	if baseUserID == 0 {
		return
	}

	var req struct {
		OrderID int     `json:"order_id"`
		Amount  float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "无效的请求参数")
		return
	}

	wallet, _ := ensureUserWallet(baseUserID)
	if wallet.Balance < req.Amount {
		utils.Fail(c, fmt.Sprintf("余额不足，当前余额 ¥%.2f，需要 ¥%.2f", wallet.Balance, req.Amount))
		return
	}

	before := wallet.Balance
	after := before - req.Amount

	tx := global.Db.Begin()
	if err := tx.Model(&models.UserWallet{}).Where("user_id = ?", baseUserID).
		Update("balance", gorm.Expr("balance - ?", req.Amount)).Error; err != nil {
		tx.Rollback()
		utils.Fail(c, "扣款失败")
		return
	}

	// 流水记录
	desc := fmt.Sprintf("支付订单 #%d", req.OrderID)
	if err := tx.Create(&models.WalletTransaction{
		UserID:        baseUserID,
		Amount:        req.Amount,
		Type:          "payment",
		BalanceBefore: before,
		BalanceAfter:  after,
		Description:   desc,
		OrderID:       req.OrderID,
	}).Error; err != nil {
		tx.Rollback()
		utils.Fail(c, "记录流水失败")
		return
	}

	tx.Commit()

	utils.Success(c, gin.H{
		"balance": after,
		"success": true,
	})
}

// ensureUserWallet 确保用户有钱包记录，不存在则创建
func ensureUserWallet(userID uint) (*models.UserWallet, bool) {
	var wallet models.UserWallet
	if err := global.Db.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			wallet = models.UserWallet{UserID: userID, Balance: 100} // 新用户送 100 元
			if createErr := global.Db.Create(&wallet).Error; createErr != nil {
				return &wallet, false
			}

			// 记录一条初始流水
			global.Db.Create(&models.WalletTransaction{
				UserID:        userID,
				Amount:        100,
				Type:          "recharge",
				BalanceBefore: 0,
				BalanceAfter:  100,
				Description:   "新用户注册赠送余额",
			})
		}
	}
	return &wallet, true
}

// getBaseUserID 从上下文中提取 baseUserID（uint），出错时自动返回 0 并写响应
func getBaseUserID(c *gin.Context) uint {
	baseIf, exists := c.Get("baseUserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "message": "未登录"})
		return 0
	}
	switch v := baseIf.(type) {
	case uint:
		return v
	case int:
		return uint(v)
	case int64:
		return uint(v)
	case float64:
		return uint(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "无效的用户 ID 类型"})
		return 0
	}
}

// init 中注册钱包相关路由（在 router.go 调用）
