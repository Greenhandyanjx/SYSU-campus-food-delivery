package controller

import (
	"backend/global"
	"backend/utils"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadImage(ctx *gin.Context){
    // 处理图片上传
    file, err := ctx.FormFile("file")
    if err != nil {
       ctx.JSON(http.StatusInternalServerError, gin.H{
		"code":"500",
		"msg": "无法获取上传的图片",
	})
    }

    // 生成唯一文件名
    ext := filepath.Ext(file.Filename)
    timestamp := time.Now().Unix()
    filename := "meal_" + strconv.FormatInt(timestamp, 10) + ext
    baseUploadPath := global.Meal_image_path // 绝对路径
    relativePath := filepath.Join(baseUploadPath, filename)
    fullPath := relativePath // 直接使用相对路径作为完整路径

    // 确保上传目录存在
    if err := utils.EnsureDir(fullPath); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
		"code":"500",
		"msg": "无法创建上传目录",
	})
    }

    // 保存文件
    if err := ctx.SaveUploadedFile(file, fullPath); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
		"code":"500",
		"msg": "无法保存图片",
		})
    }

    ctx.JSON(http.StatusOK, gin.H{
		"code": "200",
        "msg":  "图片上传成功",
        "data": gin.H{
            "url": relativePath, // 或者返回完整的 URL，例如 base URL + relativePath
        },
	})
}
