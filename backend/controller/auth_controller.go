package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Register 统一注册接口：根据 role 创建 base_user 和角色表记录（事务）
func Register(ctx *gin.Context) {
	body, _ := io.ReadAll(ctx.Request.Body)
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "0", "msg": "invalid json"})
		return
	}

	getString := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := payload[k]; ok && v != nil {
				switch t := v.(type) {
				case string:
					return t
				case float64:
					return strconv.FormatFloat(t, 'f', -1, 64)
				case int:
					return strconv.Itoa(t)
				}
			}
		}
		return ""
	}

	username := getString("username")
	password := getString("password")
	role := getString("role")
	if username == "" || password == "" || role == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "0", "msg": "username/password/role required"})
		return
	}
	if role != "user" && role != "rider" && role != "merchant" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "0", "msg": "invalid role"})
		return
	}

	hpwd, err := utils.Hpwd(password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": "0", "msg": "hash error"})
		return
	}

	tx := global.Db.Begin()
	base := models.BaseUser{Username: username, Password: hpwd, Role: role}
	if err := tx.Create(&base).Error; err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "Error 1062") {
			ctx.JSON(http.StatusConflict, gin.H{"code": "409", "msg": "用户名已被注册"})
			return
		}
		log.Printf("create base_user error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": "500", "msg": "db error"})
		return
	}

	// 根据 role 创建对应记录
	switch role {
	case "user":
		u := models.User{
			BaseID:   base.ID,
			Nickname: getString("nickname", "nick"),
			Phone:    getString("phone"),
			Address:  getString("address"),
		}
		if err := tx.Create(&u).Error; err != nil {
			tx.Rollback()
			log.Printf("create user error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": "500", "msg": "create user error"})
			return
		}
	case "rider":
		r := models.Rider{
			BaseID:   base.ID,
			RealName: getString("realname", "real_name"),
			IDNumber: getString("idNumber", "id_number"),
			// IDPhoto:  getString("idPhoto", "id_photo"),
			Phone: getString("phone"),
		}
		if err := tx.Create(&r).Error; err != nil {
			tx.Rollback()
			log.Printf("create rider error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": "500", "msg": "create rider error"})
			return
		}
		// ==== 新增：创建 rider_profile（骑手资料）====
		rp := models.RiderProfile{
			UserID:  base.ID,
			RiderID: r.ID,
			Name:    "李骑手", // 默认名
			Phone:   r.Phone,
			Avatar:  "",
		}
		if err := tx.Create(&rp).Error; err != nil {
			tx.Rollback()
			log.Printf("create rider_profile error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": "500", "msg": "create rider_profile error"})
			return
		}
	case "merchant":
		m := models.Merchant{
			BaseID:       base.ID,
			ShopName:     getString("shopName", "shop_name"),
			ShopLocation: getString("shopLocation", "shop_location"),
			Owner:        getString("owner"),
			Phone:        getString("phone"),
			Logo:         getString("logo", "logoUrl"),
			License:      getString("license", "licenseUrl"),
			Status:       "opening",
		}
		if err := tx.Create(&m).Error; err != nil {
			tx.Rollback()
			log.Printf("create merchant error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": "500", "msg": "create merchant error"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": "500", "msg": "commit error"})
		return
	}

	token, err := utils.GenerateJWTWithRole(base.Username, base.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": "0", "msg": "token error"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": "1", "msg": "注册成功", "token": token, "role": base.Role})
}

func Login(ctx *gin.Context) {
	var input models.LoginInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		fmt.Println("Login bind error:", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code": "0",
			"msg":  "server or input error",
		})
		return
	}
	// 根据角色查询不同表
	user := &models.BaseUser{}
	//查询用户是否存在
	if err := global.Db.Table("base_users").Where("username = ?", input.Username).First(user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code": "0",
			"msg":  "用户名不存在",
		})
		return
	}
	//验证密码
	if !utils.CheckPassword(input.Password, user.Password) {
		// ...密码错误处理...
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code": "0",
			"msg":  "密码错误",
		})
		return
	}

	// 在查询用户后添加角色检查
	if user.Role != input.Role { // 假设 BaseUser 有 Role 字段
		ctx.JSON(http.StatusForbidden, gin.H{
			"code": "0",
			"msg":  "角色不匹配",
		})
		return
	}

	token, err := utils.GenerateJWTWithRole(user.Username, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": "0",
			"msg":  "JWT 生成错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    "1",
		"msg":     "登录成功",
		"token":   token,
		"role":    user.Role,
		"base_id": user.ID,
	})
}

func ChangePassword(c *gin.Context) {
	// 从中间件获取 baseUserID（中间件会把 base_users 的 ID 存入上下文）
	baseIDIf, ok := c.Get("baseUserID")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "401", "msg": "no user in context"})
		return
	}

	// 处理可能的底层类型（uint 或数值类型）
	var baseID uint
	switch v := baseIDIf.(type) {
	case uint:
		baseID = v
	case int:
		baseID = uint(v)
	case int64:
		baseID = uint(v)
	case float64:
		baseID = uint(v)
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "500", "msg": "invalid user id type"})
		return
	}

	// 解析请求体（只需要旧密码和新密码）
	var request struct {
		OldPassword string `json:"oldpassword"`
		NewPassword string `json:"newpassword"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": "400", "msg": "Invalid request body"})
		return
	}

	// 从数据库读取 base_users 当前哈希
	var baseUser models.BaseUser
	if err := global.Db.First(&baseUser, baseID).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "500", "msg": "Failed to load user"})
		return
	}

	// 验证旧密码
	if !utils.CheckPassword(request.OldPassword, baseUser.Password) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "401", "msg": "Old password is incorrect"})
		return
	}

	// 计算新密码哈希
	newHash, err := utils.Hpwd(request.NewPassword)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "500", "msg": "Failed to hash new password"})
		return
	}

	// 更新数据库
	if err := global.Db.Model(&models.BaseUser{}).Where("id = ?", baseID).Update("password", newHash).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "500", "msg": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "0", "msg": "Password updated successfully"})
}
