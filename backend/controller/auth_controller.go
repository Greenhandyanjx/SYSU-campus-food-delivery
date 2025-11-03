package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Register_Base_User_temp(ctx *gin.Context){
	var base_user models.BaseUser
	body, _ := io.ReadAll(ctx.Request.Body)
    fmt.Println("Request Body:", string(body)) // 打印请求体内容
    ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // 重置请求体
	if err := ctx.ShouldBind(&base_user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "bind error",
		})
		return	
	}

	//重置请求体
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	hspd, err := utils.Hpwd(base_user.Password)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "HSPD ERROR",
		})
		return 
	}
	// 使用 utils 包中的 Hpwd 函数对用户密码进行哈希处理，并将结果存储在 hspd 变量中，同时捕获可能的错误信息存储在 err 变量中

	base_user.Password = hspd
	token, err := utils.GenerateJWT(base_user.Username)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "token generate error",
		})
		return 
	}
	if err := global.Db.Table("base_users").AutoMigrate(&base_user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg": "table create error",
		})
		return 
	}
	if err := global.Db.Create(&base_user).Error; err != nil {
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
	fmt.Println("baseuser successs")
	ctx.JSON(http.StatusOK,gin.H{
           "code":"1",
		   "msg":"success",
		   "token":token,
	})
}

func Register_Base_User(ctx *gin.Context)(uint, string,string, error) {
	var base_user models.BaseUser
	body, _ := io.ReadAll(ctx.Request.Body)
    fmt.Println("Request Body:", string(body)) // 打印请求体内容
    ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // 重置请求体
	if err := ctx.ShouldBind(&base_user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "bind error",
		})
		return 0, "", "",err
	}
	//重置请求体
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	hspd, err := utils.Hpwd(base_user.Password)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "HSPD ERROR",
		})
		return 0, "", "",err
	}
	// 使用 utils 包中的 Hpwd 函数对用户密码进行哈希处理，并将结果存储在 hspd 变量中，同时捕获可能的错误信息存储在 err 变量中

	base_user.Password = hspd
	token, err := utils.GenerateJWT(base_user.Username)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "token generate error",
		})
		return 0, "", "",err
	}
	if err := global.Db.Table("base_users").AutoMigrate(&base_user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg": "table create error",
		})
		return 0, "","",err
	}
	if err := global.Db.Create(&base_user).Error; err != nil {
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
    return 0, "", "",err
    }
    // 以下为正常情况
	fmt.Println("baseuser successs")
	return base_user.ID, token, base_user.Password,nil
}

func Register_User(ctx *gin.Context) {
	// 调用基础用户注册函数获取基础用户信息和JWT
	baseid, token,password, err := Register_Base_User(ctx)
	if err != nil {
		return
	}
	// 再次绑定 user 特定的属性，不包括 role 字段
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		log.Printf("绑定错误: %v", err) // 打印具体的绑定错误
		
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg": "绑定错误",
		})
		return
	}
	// 设置基础用户信息，忽略 role 字段
	// 设置 base_id 字段
	user.BaseID = baseid
	user.Password=password
	// 创
	// 创建 user 记录
	if err := global.Db.Table("users").AutoMigrate(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg": "table create error",
		})
		return
	}
	if err := global.Db.Create(&user).Error; err != nil {
		log.Printf("数据库创建错误: %v", err) // 记录详细错误日志
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg":  "服务器内部错误，请稍后再试",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": "200",
		"msg":  "注册成功",
		"token": token,
	})
}


func Register_Rider(ctx *gin.Context) {
	// 调用基础用户注册函数获取基础用户信息和JWT
	baseid, token, password,err := Register_Base_User(ctx)
	if err != nil {
		return
	}
	// 再次绑定 rider 特定的属性，不包括 role 字段
	var rider models.Rider
	if err := ctx.ShouldBindJSON(&rider); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg": "绑定错误",
		})
		return
	}
	// 设置 base_id 字段
	rider.BaseID = baseid
	rider.Password=password
	// 创建 rider 记录
	if err := global.Db.Table("riders").AutoMigrate(&rider); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg": "table user create error",
		})
		return
	}
	if err := global.Db.Create(&rider).Error; err != nil {
		log.Printf("数据库创建错误: %v", err) // 记录详细错误日志
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg":  "服务器内部错误，请稍后再试",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": "200",
		"msg":  "注册成功",
		"token": token,
	})
}

func Register_Merchant(ctx *gin.Context) {
	// 调用基础用户注册函数获取基础用户信息和JWT
	baseid, token, password,err := Register_Base_User(ctx)
	if err != nil {
		return
	}
	// 再次绑定 merchant 特定的属性，不包括 role 字段
	var merchant models.Merchant
	if err := ctx.ShouldBindJSON(&merchant); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg": "绑定错误",
		})
		return
	}
	// 设置 base_id 字段
	merchant.BaseID = baseid
	merchant.Password = password
	// 创建 merchant 记录
	if err := global.Db.Table("merchants").AutoMigrate(&merchant); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg": "table create error",
		})
		return
	}
	if err := global.Db.Create(&merchant).Error; err != nil {
		log.Printf("数据库创建错误: %v", err) // 记录详细错误日志
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg":  "服务器内部错误，请稍后再试",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": "200",
		"msg":  "注册成功",
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

func ChangePassword(c *gin.Context) {
		tokenUsername := c.MustGet("username").(string)
	// 解析请求体
	var request struct {
		Username    string `json:"username"`
		OldPassword string `json:"oldpassword"`
		NewPassword string `json:"newpassword"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":"400",
			"msg": "Invalid request body",
		})
		return
	}
	// 检查请求体中的 username 是否与解析出的用户名一致
	if request.Username != "" && request.Username != tokenUsername {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":"401",
			"msg": "Username mismatch",
			})
		return
	}
	// 获取用户的真实用户名
	//获取用户的密码哈希
	hash, err := utils.GetUserHashByUsernameuser(tokenUsername)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":"500",
			"msg": "Failed to get user hash",
		})
		return
	}
	// 验证旧密码
	if !utils.CheckPassword(request.OldPassword, hash) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"cdoe":"401",
			"msg": "Old password is incorrect",
		})
		return
	}
	// 更新新密码
	newHash, err := utils.Hpwd(request.NewPassword)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":"500",
			"msg": "Failed to hash new password",
		})
		return
	}
	// 假设这里有一个函数 `updateUserPasswordHash` 来更新用户的密码哈希
	if err := utils.UpdateUserPasswordHash(tokenUsername, newHash); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":"500",
			"msg": "Failed to update password",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":"0",
		"msg": "Password updated successfully",
	})
}