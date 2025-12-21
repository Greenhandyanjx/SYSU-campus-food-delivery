package rider

import (
	"backend/global"
	"backend/models"

	"github.com/gin-gonic/gin"
)

func getRiderIDByBase(c *gin.Context) (uint, bool) {
	baseID := c.GetUint("baseUserID")
	var r models.Rider
	if err := global.Db.Where("base_id = ?", baseID).First(&r).Error; err != nil {
		fail(c, "未找到骑手实名信息")
		return 0, false
	}
	return r.ID, true
}
