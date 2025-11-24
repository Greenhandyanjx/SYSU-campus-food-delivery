package controller

import (
	"backend/global"
	"backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ReplaceCategories 用于一次性替换 categories 表为 1..15 的固定分类。
// 为避免误操作，必须在请求中传入 ?force=1
func ReplaceCategories(c *gin.Context) {
	force := c.Query("force")
	if force != "1" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "force=1 required"})
		return
	}

	tx := global.Db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "begin tx failed"})
		return
	}

	if err := tx.Exec("DELETE FROM categories").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "delete existing categories failed"})
		return
	}

	fixed := []models.Category{
		{ID: 1, Name: "招牌套餐", Type: "common", Sort: 1},
		{ID: 2, Name: "现煮粉面", Type: "common", Sort: 2},
		{ID: 3, Name: "汉堡炸鸡", Type: "common", Sort: 3},
		{ID: 4, Name: "奶茶咖啡", Type: "common", Sort: 4},
		{ID: 5, Name: "日式便当", Type: "common", Sort: 5},
		{ID: 6, Name: "烧烤烤肉", Type: "common", Sort: 6},
		{ID: 7, Name: "水果拼盘", Type: "common", Sort: 7},
		{ID: 8, Name: "精致甜品", Type: "common", Sort: 8},
		{ID: 9, Name: "家常快炒", Type: "common", Sort: 9},
		{ID: 10, Name: "粥粉面饭", Type: "common", Sort: 10},
		{ID: 11, Name: "极速配送", Type: "common", Sort: 11},
		{ID: 12, Name: "午餐推荐", Type: "common", Sort: 12},
		{ID: 13, Name: "低价满减", Type: "common", Sort: 13},
		{ID: 14, Name: "沙拉轻食", Type: "common", Sort: 14},
		{ID: 15, Name: "精致下午茶", Type: "common", Sort: 15},
	}

	for _, fc := range fixed {
		if err := tx.Exec("INSERT INTO categories (id, name, type, sort) VALUES (?, ?, ?, ?)", fc.ID, fc.Name, fc.Type, fc.Sort).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "insert category failed", "err": err.Error()})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "commit failed", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "categories replaced", "count": strconv.Itoa(len(fixed))})
}
