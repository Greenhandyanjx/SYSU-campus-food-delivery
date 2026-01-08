package rider

import "github.com/gin-gonic/gin"

// 统一成功返回：code=1
func ok(c *gin.Context, data any) {
	c.JSON(200, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": data,
	})
}

// 统一失败返回：code=0
func fail(c *gin.Context, msg string) {
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  msg,
	})
	c.Abort() // 🚨 关键修复：终止后续中间件和处理函数的执行
}
