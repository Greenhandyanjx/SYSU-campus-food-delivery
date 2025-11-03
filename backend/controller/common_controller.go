package controller

import (
	"backend/global"
	"backend/models"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

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