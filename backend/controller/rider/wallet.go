package rider

import (
	"backend/global"
	"backend/models"

	"github.com/gin-gonic/gin"
)

func GetWallet(c *gin.Context) {
	riderID, okk := getRiderIDByBase(c)
	if !okk {
		return
	}

	var w models.RiderWallet
	err := global.Db.Where("rider_id = ?", riderID).First(&w).Error
	if err != nil {
		// 没有就创建
		w = models.RiderWallet{RiderID: riderID, Balance: 0, FrozenAmount: 0, TotalIncome: 0}
		if err2 := global.Db.Create(&w).Error; err2 != nil {
			fail(c, "创建钱包失败")
			return
		}
	}
	ok(c, w)
}
