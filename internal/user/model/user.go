// 第一行必须是 package + 文件夹名。这就像挂门牌号，告诉 Go 编译器：“这屋里的代码都属于 models 这个部门”
package model

import "gorm.io/gorm"

// 定义 User 结构体。
// 注意：首字母大写 'U'！因为将来我们在 controllers 部门处理登录时，需要用到这个结构体。
type User struct {
	// 这里直接嵌套咱们学过的 gorm.Model，白嫖 ID、创建时间等字段
	gorm.Model

	// 字段名 Username 首字母必须大写 'U'！
	// 如果你写成 username，GORM 框架在底层会因为“没有权限访问私有字段”而崩溃，数据库里根本建不出这个列！
	Username string `gorm:"type:varchar(50);unique;not null" json:"username"`

	// 密码同样必须大写 'P'。
	// json:"-" 的意思是：以后如果把 User 变成 JSON 发给前端，主动把这个字段隐藏掉，防止密码泄露。
	Password string `gorm:"type:varchar(255);not null" json:"-"`

	// OJ 平台特有字段：解决的题目数量 (大写 'S')
	SolvedCount int `gorm:"default:0" json:"solved_count"`
	Role        int `gorm:"type:tinyint;default:0;comment:0-普通用户, 1-超级管理员"`
}

// UserRequest 是专门用来接收前端登录/注册数据的表单 (dto)
// 注意这里的 json 标签，我们光明正大地接收 password
// binding:"required" 是 Gin 的神器，自动帮你拦截没填字段的人！
type UserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
