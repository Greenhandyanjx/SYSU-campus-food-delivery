package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GET /user/addresses
func GetUserAddresses(c *gin.Context) {
	userID := c.MustGet("baseUserID").(uint)
	var consignees []models.Consignee
	if err := global.Db.Where("userid = ?", userID).Find(&consignees).Error; err != nil {
		utils.Error(c, err)
		return
	}

	// Load address details for each consignee
	out := make([]gin.H, 0, len(consignees))
	for _, con := range consignees {
		var addr models.Address
		if err := global.Db.First(&addr, con.Addressid).Error; err != nil {
			// ignore missing address, still return consignee base
			out = append(out, gin.H{
				"id":        con.ID,
				"name":      con.Name,
				"phone":     con.Phone,
				"addressId": con.Addressid,
				"tag":       con.Tag,
				"isDefault": con.IsDefault,
			})
			continue
		}
		out = append(out, gin.H{
			"id":        con.ID,
			"name":      con.Name,
			"phone":     con.Phone,
			"addressId": con.Addressid,
			"province":  addr.Province,
			"city":      addr.City,
			"district":  addr.District,
			"street":    addr.Street,
			"detail":    addr.Detail,
			"tag":       con.Tag,
			"isDefault": con.IsDefault,
		})
	}

	// 返回地址数组（前端期望直接是 list）
	utils.Success(c, out)
}

// POST /user/address
func AddUserAddress(c *gin.Context) {
	userID := c.MustGet("baseUserID").(uint)
	var req struct {
		Name      string `json:"name" binding:"required"`
		Phone     string `json:"phone" binding:"required"`
		Province  string `json:"province"`
		City      string `json:"city"`
		District  string `json:"district"`
		Street    string `json:"street"`
		Detail    string `json:"detail" binding:"required"`
		Tag       string `json:"tag"`
		IsDefault bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数错误："+err.Error())
		return
	}

	// If province/city/district not provided, try a simple split of detail by spaces
	prov := req.Province
	city := req.City
	dist := req.District
	street := req.Street
	if prov == "" && req.Detail != "" {
		parts := strings.Fields(req.Detail)
		if len(parts) > 0 {
			prov = parts[0]
		}
		if len(parts) > 1 {
			city = parts[1]
		}
		if len(parts) > 2 {
			dist = parts[2]
		}
		if len(parts) > 3 {
			street = strings.Join(parts[3:], " ")
		}
	}

	// create Address (set UserID so address row is owned by the user)
	addr := models.Address{
		UserID:   int(userID),
		Province: prov,
		City:     city,
		District: dist,
		Street:   street,
		Detail:   req.Detail,
	}
	if err := global.Db.Create(&addr).Error; err != nil {
		utils.Error(c, err)
		return
	}

	// if IsDefault true, unset other default for user
	if req.IsDefault {
		if err := global.Db.Model(&models.Consignee{}).Where("userid = ?", userID).Update("is_default", false).Error; err != nil {
			// log but continue
			fmt.Println("failed to unset other defaults:", err)
		}
	}

	con := models.Consignee{
		Userid:    userID,
		Name:      req.Name,
		Phone:     req.Phone,
		Addressid: addr.ID,
		Tag:       req.Tag,
		IsDefault: req.IsDefault,
	}
	if err := global.Db.Create(&con).Error; err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{"consignee": con})
}

// PUT /user/address/:id
func EditUserAddress(c *gin.Context) {
	userID := c.MustGet("baseUserID").(uint)
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Fail(c, "无效地址 id")
		return
	}

	var req struct {
		Name      string `json:"name" binding:"required"`
		Phone     string `json:"phone" binding:"required"`
		Province  string `json:"province"`
		City      string `json:"city"`
		District  string `json:"district"`
		Street    string `json:"street"`
		Detail    string `json:"detail" binding:"required"`
		Tag       string `json:"tag"`
		IsDefault bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数错误："+err.Error())
		return
	}

	var con models.Consignee
	if err := global.Db.First(&con, id).Error; err != nil {
		utils.Fail(c, "地址联系人不存在")
		return
	}
	if con.Userid != userID {
		utils.Fail(c, "没有权限编辑该地址")
		return
	}

	// update address
	var addr models.Address
	if err := global.Db.First(&addr, con.Addressid).Error; err != nil {
		utils.Fail(c, "关联地址未找到")
		return
	}
	// If province/city/district not provided in edit, try to split detail
	prov := req.Province
	city := req.City
	dist := req.District
	street := req.Street
	if prov == "" && req.Detail != "" {
		parts := strings.Fields(req.Detail)
		if len(parts) > 0 {
			prov = parts[0]
		}
		if len(parts) > 1 {
			city = parts[1]
		}
		if len(parts) > 2 {
			dist = parts[2]
		}
		if len(parts) > 3 {
			street = strings.Join(parts[3:], " ")
		}
	}

	addr.Province = prov
	addr.City = city
	addr.District = dist
	addr.Street = street
	addr.Detail = req.Detail
	if err := global.Db.Save(&addr).Error; err != nil {
		utils.Error(c, err)
		return
	}

	// handle default flag
	if req.IsDefault {
		if err := global.Db.Model(&models.Consignee{}).Where("userid = ?", userID).Update("is_default", false).Error; err != nil {
			fmt.Println("failed to unset other defaults:", err)
		}
	}

	con.Name = req.Name
	con.Phone = req.Phone
	con.Tag = req.Tag
	con.IsDefault = req.IsDefault
	if err := global.Db.Save(&con).Error; err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{"consignee": con})
}

// POST /user/address/:id/default
func SetDefaultAddress(c *gin.Context) {
	userID := c.MustGet("baseUserID").(uint)
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Fail(c, "无效地址 id")
		return
	}

	var con models.Consignee
	if err := global.Db.First(&con, id).Error; err != nil {
		utils.Fail(c, "地址联系人不存在")
		return
	}
	if con.Userid != userID {
		utils.Fail(c, "没有权限操作该地址")
		return
	}

	// 事务：先清空其它默认，然后设置当前
	tx := global.Db.Begin()
	if err := tx.Model(&models.Consignee{}).Where("userid = ?", userID).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		utils.Error(c, err)
		return
	}
	if err := tx.Model(&models.Consignee{}).Where("id = ?", id).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		utils.Error(c, err)
		return
	}
	if err := tx.Commit().Error; err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{"success": true})
}

// DELETE /user/address/:id
func DeleteUserAddress(c *gin.Context) {
	userID := c.MustGet("baseUserID").(uint)
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Fail(c, "无效地址 id")
		return
	}

	var con models.Consignee
	if err := global.Db.First(&con, id).Error; err != nil {
		utils.Fail(c, "地址联系人不存在")
		return
	}
	if con.Userid != userID {
		utils.Fail(c, "没有权限删除该地址")
		return
	}

	// 删除 consignee 与其 address（可选）
	tx := global.Db.Begin()
	if err := tx.Delete(&models.Consignee{}, id).Error; err != nil {
		tx.Rollback()
		utils.Error(c, err)
		return
	}
	if err := tx.Delete(&models.Address{}, con.Addressid).Error; err != nil {
		// non-fatal: rollback address deletion only
		fmt.Println("warning: failed to delete address row:", err)
	}
	if err := tx.Commit().Error; err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{"success": true})
}
