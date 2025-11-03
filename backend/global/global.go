package global

import (
	"gorm.io/gorm"
)

var(
  Db *gorm.DB
  Meal_image_path string ="E:\\campus_food\\images\\meals"
)