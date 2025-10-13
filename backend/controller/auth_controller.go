package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Register(ctx *gin.Context) {
	var user models.BaseUser
	if err := ctx.ShouldBind(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "bind error",
		})
		return
	}
	hspd, err := utils.Hpwd(user.Password)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "HSPD ERROR",
		})
		return
	}
	// 使用 utils 包中的 Hpwd 函数对用户密码进行哈希处理，并将结果存储在 hspd 变量中，同时捕获可能的错误信息存储在 err 变量中

	user.Password = hspd
	token, err := utils.GenerateJWT(user.Username)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "token generate error",
		})
		return
	}
	if err := global.Db.Table("base_users").AutoMigrate(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg": "table create error",
		})
		return
	}
	if err := global.Db.Create(&user).Error; err != nil {
    // 检查是否为MySQL唯一键冲突错误
    if strings.Contains(err.Error(), "Error 1062") || 
       strings.Contains(err.Error(), "Duplicate entry") {
        ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{ // 409 Conflict
            "code": "409",
            "msg":  "用户名已被注册",
        })
    } else {
        // 其他数据库错误
        log.Printf("数据库创建错误: %v", err) // 记录详细错误日志
        ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "code": "500",
            "msg":  "服务器内部错误，请稍后再试",
        })
    }
    return
    }
    // 以下为正常情况

	ctx.JSON(http.StatusOK, gin.H{
		"msg":   "register success",
		"token": token,
	})
}

func Login(ctx *gin.Context) {
	var input models.LoginInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code": "0",
			"msg":"server or input error",
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
   if !utils.CheckPassword(input.Password,user.Password) {
    // ...密码错误处理...
	ctx.JSON(http.StatusUnauthorized, gin.H{
			"code": "0",
			"msg":  "密码错误",
		})
		return
   } 
   
	// 在查询用户后添加角色检查
if user.Role != input.Role {  // 假设 BaseUser 有 Role 字段
    ctx.JSON(http.StatusForbidden, gin.H{
        "code": "0",
        "msg":  "角色不匹配",
    })
    return
}

	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": "0",
			"msg":  "JWT 生成错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":  "1",
		"msg":   "登录成功",
		"token": token,
		"role":  input.Role, // 返回角色供前端使用
	})
}