package controller

import (
	"backend/global"
	"backend/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Dish_add(ctx*gin.Context){
    var dish models.Dish
    if err := ctx.ShouldBind(&dish); err!= nil {
		ctx.JSON(http.StatusBadRequest,gin.H{
			"code":"400",
			"msg":"binding error",
		})
		return
	}
	if err := global.Db.Create(&dish).Error; err != nil {
		log.Printf("数据库插入错误: %v", err) // 记录详细错误日志
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg":  "服务器内部错误，请稍后再试",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":"1",
		"msg":"dish added successfully",
	})
}

func Get_dishes(ctx *gin.Context) {
	var params models.GetDishPageParams
    if err := ctx.ShouldBindQuery(&params); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "code": "400",
            "msg":  "请求参数错误",
        })
        return
    }
    // 计算分页参数
    offset := (params.Page - 1) * params.Size
    limit := params.Size
    // 构建查询条件
    var query = global.Db.Model(&models.Dish{}).Preload("Flavors")
    if params.Name != "" {
        query = query.Where("dish_name LIKE ?", "%"+params.Name+"%")
    }
    if params.CategoryId != "" {
        categoryId, err := strconv.Atoi(params.CategoryId)
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "code": "400",
                "msg":  "分类ID格式错误",
            })
            return
        }
        query = query.Where("category = ?", categoryId)
    }
	if params.Status != "" {
		status, err := strconv.Atoi(params.Status)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": "400",
				"msg":  "状态格式错误",	
		})
			return
		}
		query = query.Where("status = ?", status)
	}
    // 查询菜品列表
    var dishes []models.Dish
    if err := query.Offset(offset).Limit(limit).Find(&dishes).Error; err != nil {
        log.Printf("数据库查询错误: %v", err)
        ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "code": "500",
            "msg":  "数据库查询错误",
        })
        return
    }
    // 查询总记录数
    var total int64
    if err := query.Count(&total).Error; err != nil {
        log.Printf("数据库计数错误: %v", err)
        ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "code": "500",
            "msg":  "数据库计数错误",
        })
        return
    }
    // 准备返回数据
    items := make([]gin.H, len(dishes))
    for i, dish := range dishes {
        items[i] = gin.H{
            "id":          dish.ID,
            "name":        dish.DishName,
            "price":       dish.Price,
            "status":      dish.Status,
            "categoryId":  dish.Category,
            "stock":       0, // 假设 stock 字段在 Dish 结构体中不存在，这里返回 0
        }
    }
    // 返回结果
    ctx.JSON(http.StatusOK, gin.H{
        "code": "1",
		"msg":  "获取菜品列表成功",
        "data": gin.H{
            "items": items,
            "total": total,
        },
    })
}

func Edit_dish(c*gin.Context){
	 var dish models.Dish
    if err := c.ShouldBindJSON(&dish); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求体格式错误", "data": nil})
        return
    }
    // 根据 ID 查找菜品
    var existingDish models.Dish
    if err := global.Db.First(&existingDish, dish.ID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "菜品未找到", "data": nil})
        return
    }
    // 更新菜品信息
    if err := global.Db.Model(&existingDish).Updates(models.Dish{
        DishName:    dish.DishName,
        Price:       dish.Price,
        Description: dish.Description,
        MerchantID:  dish.MerchantID,
        Tastes:      dish.Tastes,
        ImagePath:   dish.ImagePath,
        Category:    dish.Category,
        Status:      dish.Status,
    }).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "更新菜品信息失败", "data": nil})
        return
    }
    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true, "dishId": strconv.Itoa(dish.ID)}})
}

func Delete_dish(c*gin.Context){
	 // 绑定请求体到 map 结构体
    var request map[string]interface{}
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求体格式错误", "data": nil})
        return
    }
    // 获取请求中的 id 或 id 列表
    idOrList, ok := request["id"]
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求体中缺少 id 字段", "data": nil})
        return
    }
    var removedIDs []string
    switch ids := idOrList.(type) {
    case string:
        // 单个菜品删除
        removedIDs = append(removedIDs, ids)
        if err := global.Db	.Delete(&models.Dish{}, ids).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "删除菜品失败", "data": nil})
            return
        }
    case []interface{}:
        // 批量删除菜品
        for _, id := range ids {
            if idStr, ok := id.(string); ok {
                removedIDs = append(removedIDs, idStr)
                if err := global.Db.Delete(&models.Dish{}, idStr).Error; err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "删除菜品失败", "data": nil})
                    return
                }
            } else {
                c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "id 列表中包含非字符串类型", "data": nil})
                return
            }
        }
    default:
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "id 字段类型错误", "data": nil})
        return
    }
    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true, "removed": removedIDs}})
}

func Edit_DishStatus_By_Status(c *gin.Context) {
    // 绑定请求体到 Dish 结构体
    var request struct {
        ID     string `json:"id" form:"id"`
        Status string `json:"status" form:"status"`
    }
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求体格式错误", "data": nil})
        return
    }
    // 根据前端传递的 status 字段转换为数据库中的整数值
    var status int
    switch request.Status {
    case "on":
        status = 1
    case "off":
        status = 0
    case "recommended":
        status = 1 // 假设推荐状态对应的值为 2
    default:
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "状态值无效", "data": nil})
        return
    }
    // 根据 ID 查找菜品
    var existingDish models.Dish
    if err := global.Db.First(&existingDish, request.ID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "菜品未找到", "data": nil})
        return
    }
    // 更新菜品状态
    if err :=global.Db.Model(&existingDish).Updates(models.Dish{
        Status: status,
    }).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "更新菜品状态失败", "data": nil})
        return
    }
    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true}})
}

//GET /merchant/dish/list
// queryDishList 处理基于条件查询菜品的请求
func QueryDishList(c *gin.Context) {
    // 获取请求参数
    categoryId := c.Query("categoryId")
    name := c.Query("name")
    // 构建查询条件
    var dishes []models.Dish
    db := global.Db
    if categoryId != "" && name != "" {
        db = db.Where("category_id = ? AND name LIKE ?", categoryId, "%"+name+"%")
    } else if categoryId != "" {
        db = db.Where("category_id = ?", categoryId)
    } else if name != "" {
        db = db.Where("name LIKE ?", "%"+name+"%")
    }
    // 执行查询
    if err := db.Find(&dishes).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询菜品失败", "data": nil})
        return
    }
    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{"code": 1, "data": dishes})
}

// queryDishById 处理查询单个菜品详情的请求
func Get_Dish_ById(c *gin.Context) {
    // 获取请求参数中的 id
    id := c.Query("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数中缺少 id 字段", "data": nil})
        return
    }
    // 构建查询条件
    var dish models.Dish
    if err := global.Db.First(&dish, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "菜品不存在", "data": nil})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询菜品失败", "data": nil})
        }
        return
    }
    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{"code": 1, "data": dish})
}