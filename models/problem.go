package models

import (
	"gorm.io/gorm"
)

// Problem 题目的数据库模型图纸
type Problem struct {
	// gorm.Model 是官方提供的一个超级贴心的小工具
	// 只要嵌入它，你的表里就会自动多出 4 个字段：
	// ID (主键), CreatedAt (创建时间), UpdatedAt (更新时间), DeletedAt (软删除标志)
	gorm.Model

	// 题目的核心信息
	Title       string `gorm:"type:varchar(100);not null" json:"title"` // 题目标题，比如 "A+B Problem"
	Description string `gorm:"type:text;not null" json:"description"`   // 题目描述，通常是 Markdown 格式的长文本

	// 判题核心限制（OJ 系统的命脉）
	TimeLimit   int `gorm:"type:int;default:1000" json:"time_limit"`  // 时间限制，单位是毫秒 (ms)，默认 1000ms = 1秒
	MemoryLimit int `gorm:"type:int;default:256" json:"memory_limit"` // 内存限制，单位是兆 (MB)，默认 256MB

	// 统计信息（未来大屏幕上展示的炫酷数据）
	SubmitCount   int `gorm:"type:int;default:0" json:"submit_count"`   // 别人总共提交了多少次
	AcceptedCount int `gorm:"type:int;default:0" json:"accepted_count"` // 有多少次是 AC (完全正确) 的

	// 【GORM魔法口袋】：逻辑外键，告诉 GORM 怎么拼装数据，绝对不会在 MySQL 里生成这一列！
	TestCases []TestCase `gorm:"foreignKey:ProblemID" json:"-"`
	// 💥 新增这个字段！
	// gorm:"-" 的意思是：包工头GORM你不要管它，数据库里没有这列！
	// json:"is_ac" 的意思是：转成 JSON 发给前端时，叫这个名字。
	IsAC bool `gorm:"-" json:"is_ac"`
	// 这行代码会让 GORM 自动在底层帮你建一张叫 `problem_tags` 的中间表！
	// 里面只有两个字段：problem_id 和 tag_id
	Tags []Tag `gorm:"many2many:problem_tags;" json:"tags"`
}

// 顺便在这里准备一个接客用的 DTO 表单
// 因为我们不希望管理员在发布题目时，还要手填 SubmitCount 这种应该由系统管理的字段
type ProblemRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	TimeLimit   int    `json:"time_limit"` // 如果前端不传，在 Controller 里我们会给它赋默认值
	MemoryLimit int    `json:"memory_limit"`

	// 直接复用 TestCaseRequest！
	TestCases []TestCaseRequest `json:"test_cases"`
	TagIDs    []uint            `json:"tag_ids"`
}
