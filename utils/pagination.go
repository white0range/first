package utils

import (
	"gorm.io/gorm"
)

// Paginate 是一个 GORM 的 Scope 插件！
// 它会自动从 Gin 的上下文里提取 page 和 limit 参数，并转化成 MySQL 的 Offset 和 Limit
//
//	func Paginate(c *gin.Context) func(db *gorm.DB) *gorm.DB {
//		return func(db *gorm.DB) *gorm.DB {
//			// 1. 提取参数并给默认值
//			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
//			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20")) // 默认每页 20 条
//
//			// 2. 绝对防御：防止前端瞎传参数搞崩数据库
//			if page <= 0 {
//				page = 1
//			}
//			switch {
//			case limit > 100:
//				limit = 100 // 限制单次最大查询量，防止恶意扒库
//			case limit <= 0:
//				limit = 20
//			}
//
//			// 3. 计算偏移量
//			offset := (page - 1) * limit
//
//			// 4. 返回拼接好分页指令的 DB 对象
//			return db.Offset(offset).Limit(limit)
//		}
//	}
//
// ✅ 重构后的神器：纯粹的领域层分页工具
func Paginate(page, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if limit <= 0 || limit > 100 {
			limit = 10
		}
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}
