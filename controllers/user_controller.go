package controllers

import (
	"gojo/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// controllers/user_controller.go

// Register 处理用户注册
func Register(c *gin.Context) {
	var req services.UserAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "账号或密码不能为空，且密码至少6位"})
		return
	}

	if err := services.RegisterUser(c.Request.Context(), req); err != nil {
		if err.Error() == "username_exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "注册失败，用户名已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，注册失败"})
		return
	}

	// ✅ 安全返回：绝对不要把带密码或哈希的对象返回给前端
	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功！欢迎来到 OJ 平台",
	})
}

// Login 处理用户登录
func Login(c *gin.Context) {
	var req services.UserAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	token, err := services.LoginUser(c.Request.Context(), req)
	if err != nil {
		// 模糊报错保护：不管账号错还是密码错，都提示“账号或密码错误”，防黑客试探
		if err.Error() == "user_not_found" || err.Error() == "wrong_password" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，登录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功！",
		"token":   token,
	})
}

// GetProfile 获取当前登录用户的个人主页
func GetProfile(c *gin.Context) {
	// 1. 🛡️ 极其关键的安全获取身份与断言 (修复宕机雷)
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	userID, ok := userIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户身份解析异常"})
		return
	}

	// 2. 获取战绩
	profile, err := services.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取战绩大屏成功",
		"data":    profile,
	})
}
