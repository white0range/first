package controllers

import (
	"gojo/utils"
	"net/http"

	"gojo/models" // 呼叫你自己的模型部门 (注意 go-oj 要和你的模块名一致)

	"github.com/gin-gonic/gin"
)

// Register 是处理用户注册的开放服务（首字母大写 R）
func Register(c *gin.Context) {
	// 1. 准备一个结构体来接收前端传来的 JSON
	// 这里咱们直接借用 models 部门画好的图纸
	var req models.UserRequest

	// 2. 尝试把前端传来的 JSON 绑定到 req 变量上
	// ShouldBindJSON 会自动检查 binding:"required"。如果没传账号密码，直接报错，咱们连 if 校验都省了！
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "账号或密码不能为空，且格式必须正确"})
		return
	}

	//// 3. 简单的业务校验：密码不能为空
	//if req.Username == "" || req.Password == "" {
	//	fmt.Println(req.Username, " ", req.Password)
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "账号或密码不能为空"})
	//	return
	//}

	//hashed
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加密失败"})
		return
	}
	req.Password = hash

	user := models.User{
		Username: req.Username,
		Password: req.Password,
	}
	// 4. 【核心联动】呼叫 models 部门的万能钥匙（DB），把数据写进 MySQL
	// 相当于执行：INSERT INTO users (username, password) VALUES (...)
	result := models.DB.Create(&user)

	if result.Error != nil {
		// 如果数据库报错（比如用户名被别人抢注了，触发了 unique 约束）
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败，用户名可能已存在"})
		return
	}

	// 5. 注册成功，返回前端
	// 注意：因为 models.User 里的 Password 字段加了 json:"-" 标签，所以密码不会被打印出来，非常安全！
	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功！欢迎来到 OJ 平台",
		"data":    req,
	})
}

// Login 是处理用户登录的接口
func Login(c *gin.Context) {
	// 1. 接收前端传来的账号密码
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误或者账号密码为空"})
		return
	}

	// 2. 去数据库里找这个人
	var user models.User
	// 注意 GORM 的语法：去数据库查找 Username 等于 req.Username 的那行数据，存进 user 变量里
	result := models.DB.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}

	// 3. 核心校验：比对密码
	// req.Password 是用户刚才输入的明文，user.Password 是数据库里取出来的乱码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 4. 密码正确！呼叫 utils 部门，给他打印专属手环
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "手环生成失败"})
		return
	}

	// 5. 双手奉上手环给前端
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功！",
		"token":   token, // 前端拿到这个 token 后，会把它存进浏览器的 LocalStorage 里
	})
}
