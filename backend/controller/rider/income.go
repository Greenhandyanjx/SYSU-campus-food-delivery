package rider

import (
	"backend/global"
	"backend/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetIncome(c *gin.Context) {
	riderID, okk := getRiderIDByBase(c)
	if !okk {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var list []models.RiderIncomeRecord
	global.Db.Where("rider_id = ?", riderID).
		Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&list)

	ok(c, list)
}
