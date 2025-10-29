package models

type GetDishPageParams struct {
	Page       int    `form:"page" binding:"required"`
	Size       int    `form:"pageSize" binding:"required"`
	Name       string `form:"name"`
	Status     string `form:"status"`
	CategoryId string `form:"categoryId"`
}