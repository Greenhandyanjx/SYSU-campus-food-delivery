package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// commonDownload 处理通用下载接口
func CommonDownload(c *gin.Context) {
	// 获取请求参数
	downloadType := c.Query("type")
	format := c.Query("format")
	// 根据 type 和 format 进行不同的处理
	switch downloadType {
	case "dishes":
		switch format {
		case "csv":
			DownloadDishesCSV(c)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "不支持的格式", "data": nil})
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "不支持的类型", "data": nil})
	}
}

// downloadDishesCSV 将菜品信息导出为 CSV 文件
func DownloadDishesCSV(c *gin.Context) {
	// 查询所有菜品
	var dishes []models.Dish
	if err := global.Db.Find(&dishes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询菜品失败", "data": nil})
		return
	}
	// 创建 CSV 编写器
	w := csv.NewWriter(c.Writer)
	defer w.Flush()
	// 设置响应头
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=dishes.csv")
	// 写入 CSV 头
	headers := []string{"ID", "Name", "Price", "Images", "Description", "CategoryID", "Stock"}
	if err := w.Write(headers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "写入 CSV 头失败", "data": nil})
		return
	}
	// 写入 CSV 数据
	for _, dish := range dishes {
		imagePath_split := strings.Split(dish.ImagePath, ",")
		record := []string{
			fmt.Sprintf("%d", dish.ID),
			dish.DishName,
			dish.Price,
			strings.Join(imagePath_split, ","),
			dish.Description,
			fmt.Sprintf("%d", dish.Category),
		}
		if err := w.Write(record); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "写入 CSV 数据失败", "data": nil})
			return
		}
	}
}

// GetUserProfile 返回当前登录用户的基本信息（username, points, orderCount, couponCount）
func GetUserProfile(c *gin.Context) {
	baseUserIDIface, exists := c.Get("baseUserID")
	if !exists {
		// 应该由中间件保证，这里兜底返回 demo
		utils.Success(c, gin.H{
			"username":    "游客",
			"vipLevel":    "普通用户",
			"points":      120,
			"orderCount":  0,
			"couponCount": 0,
		})
		return
	}
	baseUserID := baseUserIDIface.(uint)

	// 尝试读取 BaseUser 以获取 username
	var base models.BaseUser
	username := "游客"
	if err := global.Db.Where("id = ?", baseUserID).First(&base).Error; err == nil {
		if base.Username != "" {
			username = base.Username
		}
	}

	// 查找关联的用户表（可选），以获取昵称、手机号、头像等信息
	var user models.User
	var phoneStr string
	var avatarUrl string
	if err := global.Db.Where("base_id = ?", baseUserID).First(&user).Error; err == nil {
		if user.Nickname != "" {
			username = user.Nickname
		}
		phoneStr = user.Phone
		avatarUrl = user.AvatarURL
	}

	// 统计该用户的订单数量（userid 字段在 Order 表中对应 baseUserID）
	var orderCount int64
	global.Db.Model(&models.Order{}).Where("userid = ?", baseUserID).Count(&orderCount)

	// couponCount 暂时返回 0（或你可以接入 coupons 表）
	couponCount := 0

	// points 固定为 120（满足当前需求）
	points := 120

	utils.Success(c, gin.H{
		"username":    username,
		"vipLevel":    "普通用户",
		"points":      points,
		"orderCount":  orderCount,
		"couponCount": couponCount,
		"phone":       phoneStr,
		"avatar_url":  avatarUrl,
	})
}

// UpdateUserProfile 更新当前用户的昵称、手机号和头像 URL
func UpdateUserProfile(c *gin.Context) {
	baseUserIDIface, exists := c.Get("baseUserID")
	if !exists {
		utils.Fail(c, "unauthenticated")
		return
	}
	baseUserID := baseUserIDIface.(uint)

	var req struct {
		Nickname  string `json:"nickname"`
		Phone     string `json:"phone"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "invalid request")
		return
	}

	// 更新或创建 models.User （以 base_id 关联）
	var user models.User
	if err := global.Db.Where("base_id = ?", baseUserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 仅当提供了至少一个可保存的字段时才创建新行，避免产生空记录
			if req.Nickname != "" || req.Phone != "" || req.AvatarURL != "" {
				user = models.User{
					BaseID:    baseUserID,
					Nickname:  req.Nickname,
					Phone:     req.Phone,
					AvatarURL: req.AvatarURL,
				}
				if err := global.Db.Create(&user).Error; err != nil {
					utils.Error(c, err)
					return
				}
			} else {
				// 没有可保存的数据，直接返回成功（无操作）
				utils.Success(c, gin.H{"ok": true})
				return
			}
		} else {
			utils.Error(c, err)
			return
		}
	} else {
		// 已存在，执行更新
		updates := map[string]interface{}{}
		if req.Nickname != "" {
			updates["nickname"] = req.Nickname
		}
		if req.Phone != "" {
			updates["phone"] = req.Phone
		}
		if req.AvatarURL != "" {
			updates["avatar_url"] = req.AvatarURL
		}
		if len(updates) > 0 {
			if err := global.Db.Model(&user).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
				utils.Error(c, err)
				return
			}
		}
	}

	// Optionally update BaseUser.Username if nickname provided and desired
	if req.Nickname != "" {
		var base models.BaseUser
		if err := global.Db.Where("id = ?", baseUserID).First(&base).Error; err == nil {
			if base.Username == "" || base.Username != req.Nickname {
				// do not overwrite intentionally; only update if empty
				// (keep conservative behavior)
			}
		}
	}

	utils.Success(c, gin.H{"ok": true})
}
