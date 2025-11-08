package controller

import (
	"backend/global"
	"backend/models"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Meal_add(ctx *gin.Context) {
    var meal models.Meal
    // 中间结构体用于绑定
    type SetmealDish struct {
        Copies int    `json:"copies"`
        Name   string `json:"name"`
        Price  string `json:"price"`
    }
    type mealRequest struct {
        DishIDs []SetmealDish `json:"setmealDishes"`
    }
    var request mealRequest
    //
    baseUserID := ctx.MustGet("baseUserID").(uint)
    // 将用户ID赋给套餐的MerchantID字段
	meal.MerchantID = baseUserID
    //第一次绑定
    body, _ := io.ReadAll(ctx.Request.Body)
    fmt.Println("Request Body:", string(body)) // 打印请求体内容
    ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // 重置请求体
    if err := ctx.ShouldBindJSON(&meal); err != nil {
        // 打印绑定错误的详细信息
		log.Printf("绑定错误: %v", err.Error())
        ctx.JSON(http.StatusBadRequest, gin.H{
            "code": "400",
            "msg":  "binding error",
        })
        return
    }

    fmt.Println("Request Body:", string(body)) // 打印请求体内容
   	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
    // 获取 dish_ids，第二次绑定
    if err := ctx.ShouldBindJSON(&request); err != nil {
        // 打印绑定错误的详细信息
		log.Printf("绑定错误: %v", err.Error())
        ctx.JSON(http.StatusBadRequest, gin.H{
            "code": "400",
            "msg":  "binding dish_ids error",
        })
        return
    }

    if err := global.Db.Create(&meal).Error; err != nil {
        fmt.Println(err.Error())
        ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "code": "500",
            "msg":  "服务器内部错误，请稍后再试",
        })
        return
    }

    // 创建 meal_dish 关联关系
    for _, setmealDish := range request.DishIDs {
        mealDish := models.MealDish{
            MealID: meal.ID,
            DishID: 0, // 假设 setmealDish 中没有 DishID，你需要从 DishList 或其他地方获取
            Num:    setmealDish.Copies,
        }
        // 如果 setmealDish 中有 DishID，可以直接使用
        // 如果没有，你需要根据 setmealDish.Name 或其他信息查询 DishID
        // 这里假设你需要通过名称查询 DishID
        var dish models.Dish
        if err := global.Db.Where("dish_name = ?", setmealDish.Name).First(&dish).Error; err != nil {
            log.Printf("查询 DishID 错误: %v", err)
            ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                "code": "500",
                "msg":  "查询 DishID 错误",
            })
            return
        }
        mealDish.DishID = dish.ID
        if err := global.Db.Create(&mealDish).Error; err != nil {
            log.Printf("创建关联关系错误: %v", err)
            ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                "code": "500",
                "msg":  "服务器内部错误，请稍后再试",
            })
            return
        }
    }

    ctx.JSON(http.StatusOK, gin.H{
        "code": "200",
        "msg":  "meal 创建成功",
    })
}

func QueryDishById(ctx *gin.Context) {
    id := ctx.Query("id")
    var dish models.Dish
    if err := global.Db.Model(&models.Dish{}).Where("ID = ?", id).First(&dish).Error; err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "code": "500",
            "msg":  "查询菜品失败",
        })
        return
    }
    ctx.JSON(http.StatusOK, gin.H{
        "code": "1",
        "msg":  "success",
        "data": dish,
    })
}


func Meal_Edit(c *gin.Context) {
    var meal models.Meal
    if err := c.ShouldBindJSON(&meal); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求体格式错误", "data": nil})
        return
    }

    // 根据 ID 查找套餐
    var existingMeal models.Meal
    if err := global.Db.First(&existingMeal, meal.ID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "套餐未找到", "data": nil})
        return
    }

    // 更新套餐信息
    if err := global.Db.Model(&existingMeal).Updates(models.Meal{
        Mealname:    meal.Mealname,
        Price:       meal.Price,
        Description: meal.Description,
        MerchantID:  meal.MerchantID,
        Status:      meal.Status,
        ImagePath:   meal.ImagePath,
        Category:    meal.Category,
    }).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "更新套餐信息失败", "data": nil})
        return
    }

    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true, "mealId": strconv.Itoa(meal.ID)}})
}

func Meal_Delete(c*gin.Context){
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
        if err := global.Db.Table("meals").Delete(&models.Meal{}, ids).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "删除菜品失败", "data": nil})
            return
        }
    case []interface{}:
        // 批量删除菜品
        for _, id := range ids {
            if idStr, ok := id.(string); ok {
                removedIDs = append(removedIDs, idStr)
                if err := global.Db.Table("meals").Delete(&models.Meal{}, idStr).Error; err != nil {
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


func Edit_Meal_Status(c *gin.Context) {
    // 绑定请求体到结构体
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
    var existingMeal models.Dish
    if err := global.Db.Table("meals").First(&existingMeal, request.ID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "套餐未找到", "data": nil})
        return
    }
    // 更新菜品状态
    if err :=global.Db.Table("meals").Model(&existingMeal).Updates(models.Meal{
        Status: status,
    }).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "更新套餐状态失败", "data": nil})
        return
    }
    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true}})
}


// 分页获取套餐信息
func GetMealsPage(c *gin.Context) {
    // 获取请求参数
    page, err1 := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, err2 := strconv.Atoi(c.DefaultQuery("size", "20"))
    name := c.Query("name")

    // 检查参数是否合法
    if err1 != nil || err2 != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "无效的分页参数", "data": nil})
        return
    }

    // 计算分页的偏移量
    offset := (page - 1) * size

    // 构建查询条件
    var meals []models.Meal
    var total int64

    query := global.Db.Model(&models.Meal{})

    if name != "" {
        query = query.Where("mealname LIKE ?", "%"+name+"%")
    }

    // 获取总记录数
    if err := query.Count(&total).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "获取套餐总数失败", "data": nil})
        return
    }

    // 获取分页数据
    if err := query.Offset(offset).Limit(size).Find(&meals).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "获取套餐列表失败", "data": nil})
        return
    }

    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"items": meals, "total": total}})
}


func Get_Meal_ById(c *gin.Context) {
    // 获取请求参数中的 id
    id := c.Query("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数中缺少 id 字段", "data": nil})
        return
    }
    // 构建查询条件
    var meal models.Meal
    if err := global.Db.First(&meal, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "套餐不存在", "data": nil})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询套餐失败", "data": nil})
        }
        return
    }
    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{"code": 1, "data": meal})
}