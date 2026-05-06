package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// 定义配置结构体，与 yaml 文件的层级一一对应
type Config struct {
	SQL   SQLConfig
	Redis RedisConfig
	JWT   JWTConfig
	AI    AIConfig
}
type SQLConfig struct {
	Dsn string
}
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
}

type AIConfig struct {
	APIKey string `mapstructure:"api_key"`
}

// AppConfig 是全局可用的配置实例
var AppConfig Config

// InitConfig 初始化加载配置
func InitConfig() {
	viper.SetConfigName("config")   // 配置文件名称(无扩展名)
	viper.SetConfigType("yaml")     // 文件类型
	viper.AddConfigPath("./config") // 告诉 viper 去哪里找这个文件

	// 1. 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("❌ 致命错误：读取配置文件失败: %v", err)
	}

	// 2. 把读取到的配置反序列化到结构体里
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("❌ 致命错误：解析配置文件失败: %v", err)
	}

	fmt.Println("⚙️  系统配置加载成功！")
}
