package judge

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
)

// TestSandbox 极速体验：用代码创建一个真实的容器
func TestSandbox() {
	ctx := context.Background()

	fmt.Println("🚀 正在向 Docker 引擎下达创建容器指令...")

	// 1. 画出小黑屋的图纸 (Config)
	// 这相当于你在命令行里敲 docker run golang:alpine echo "Hello Docker"
	resp, err := DockerClient.ContainerCreate(ctx, &container.Config{
		Image: "golang:alpine",                                 // 指定刚才拉取的只读光盘
		Cmd:   []string{"echo", "Hello from Go-OJ Sandbox!!!"}, // 进屋后执行的第一条命令
	}, nil, nil, nil, "")

	//技术含义： 覆盖容器启动后的默认执行命令（Command）。
	//
	//为什么要写成数组 []string？ 你可能会问，为什么不直接写成一个字符串 "go run main.go"？
	//在大厂的安全规范里，写成数组是为了绕过底层操作系统的 Shell 解析。如果写成一整个字符串，系统可能会调起 /bin/sh -c 去执行，
	//这就给了黑客进行“命令注入”的可乘之机（比如在代码里偷偷拼接个 && rm -rf /）。
	//写成数组，Docker 就会极其精确、纯粹地只执行 go 这个可执行文件，把后面的当做纯参数，安全性极高。
	if err != nil {
		fmt.Println("❌ 容器图纸创建失败:", err)
		return
	}

	fmt.Printf("📦 图纸确认完毕！Docker 已分配了容器 ID: %s\n", resp.ID[:12])

	// 2. 真正通电开机！启动容器
	if err := DockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		fmt.Println("❌ 容器启动失败:", err)
		return
	}

	fmt.Println("🎉 容器成功启动！它现在是一个独立的世界了。")
}
