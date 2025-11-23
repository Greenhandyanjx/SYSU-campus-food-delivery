package controller

import (
	"backend/global"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ==================== 查询商家信息 ====================
// /api/store/query?name=xxx 或 ?id=123
func StoreQuery(ctx *gin.Context) {
	name := ctx.Query("name")
	id := ctx.Query("id")
	var merchant models.Merchant

	// 按 ID 查询
	if id != "" {
		if err := global.Db.Where("base_id = ?", id).First(&merchant).Error; err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "merchant not found",
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"code": 1,
			"data": merchant,
		})
		return
	}

	// 按名称查询
	if name != "" {
		if err := global.Db.Where("shop_name LIKE ?", "%"+name+"%").First(&merchant).Error; err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "merchant not found",
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"code": 1,
			"data": merchant,
		})
		return
	}

	ctx.JSON(http.StatusBadRequest, gin.H{
		"code": 0,
		"msg":  "missing query parameters",
	})
}

// ==================== 查询商家下的菜品 ====================
// /api/store/dishes?storeId=233
func StoreDishes(ctx *gin.Context) {
	storeId := ctx.Query("storeId")
	if storeId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 0,
			"msg":  "storeId is required",
		})
		return
	}

	var dishes []models.Dish
	if err := global.Db.Where("merchant_id = ?", storeId).Find(&dishes).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 0,
			"msg":  "failed to query dishes",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": dishes,
	})
}
