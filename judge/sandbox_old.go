package judge

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

// ========= 【新增：定义判题机状态码和战报】 =========
type JudgeStatus string

const (
	StatusAccepted            JudgeStatus = "AC"  // 答案正确
	StatusWrongAnswer         JudgeStatus = "WA"  // 答案错误
	StatusTimeLimitExceeded   JudgeStatus = "TLE" // 运行超时
	StatusMemoryLimitExceeded JudgeStatus = "MLE" // 内存超限
	StatusRuntimeError        JudgeStatus = "RE"  // 运行错误
	StatusSystemError         JudgeStatus = "SE"  // 系统内部错误
	StatusCompileError        JudgeStatus = "CE"  // 👈 新增：编译错误
)

type JudgeResult struct {
	Status JudgeStatus
	Output string
	Error  error
}

// ===============================================
// CompileCode 负责将玩家代码编译成二进制文件
// 返回值：是否编译成功，以及可能产生的编译报错日志
func CompileCode(ctx context.Context, code string, workDir string) (bool, string, error) {
	// 1. 把代码写入宿主机的临时目录
	codePath := filepath.Join(workDir, "main.go")
	if err := os.WriteFile(codePath, []byte(code), 0644); err != nil {
		return false, "", fmt.Errorf("写入代码文件失败: %v", err)
	}

	// 2. 召唤 Docker 进行【纯编译】
	resp, err := DockerClient.ContainerCreate(ctx,
		&container.Config{
			Image:      "golang:alpine",
			WorkingDir: "/app",
			// 核心指令：go build -o solution main.go (编译并输出为 solution 文件)
			Cmd: []string{"sh", "-c", "GO111MODULE=off go build -o solution main.go"},
		},
		&container.HostConfig{
			Binds: []string{workDir + ":/app"},
		}, nil, nil, "")

	if err != nil {
		return false, "", err
	}
	defer DockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})

	// 启动编译容器
	DockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{})

	statusCh, errCh := DockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	// 祭出第三台对讲机：一个 10 秒的定时器通道
	timeoutCh := time.After(10 * time.Second)
	// 【新增】：定一个变量，专门用来标记是不是运行时错误 (RE)
	// 死死盯住这三个通道，谁先响就听谁的！
	select {
	case err := <-errCh:
		if err != nil {
			return false, "", fmt.Errorf("Docker系统异常: %v", err)
		}
	case status := <-statusCh:
		// 容器正常结束，准备去拿日志
		out, _ := DockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
		var stdoutBuf, stderrBuf bytes.Buffer
		stdcopy.StdCopy(&stdoutBuf, &stderrBuf, out)

		if status.StatusCode != 0 {
			// 编译失败（语法错误等）
			return false, stderrBuf.String(), nil
		}

	case <-timeoutCh:

		// 呼叫 Docker 底层 API，发送极其残暴的 SIGKILL 信号，物理拔电
		if killErr := DockerClient.ContainerKill(ctx, resp.ID, "SIGKILL"); killErr != nil {
			fmt.Println("❌ 强杀容器失败，可能已经变成僵尸进程:", killErr)
		} else {
			fmt.Println("💀 失控容器已被成功销毁！")
		}
		return false, "编译超时 (Compile Time Limit Exceeded)", nil
	}

	// 编译成功！宿主机的 workDir 下现在已经有了一个名为 "solution" 的二进制文件
	return true, "", nil
}

// RunTestCase 负责运行【已经编译好的二进制文件】，并做超时控制
func RunTestCase(ctx context.Context, workDir string, input string) JudgeResult {
	// 1. 写入当前测试用例的输入文件
	inputPath := filepath.Join(workDir, "input.txt")
	os.WriteFile(inputPath, []byte(input), 0644)

	// 2. 召唤 Docker 执行二进制文件
	resp, err := DockerClient.ContainerCreate(ctx,
		&container.Config{
			Image:      "golang:alpine",
			WorkingDir: "/app",
			// 🚀 核心性能跨越：不再 go run！直接运行二进制程序 ./solution
			Cmd: []string{"sh", "-c", "./solution < input.txt"},
		},
		&container.HostConfig{
			Binds: []string{workDir + ":/app"},
			Resources: container.Resources{
				Memory: 256 * 1024 * 1024,
			},
		}, nil, nil, "")

	if err != nil {
		return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("启动容器失败: %v", err)}
	}

	// 同样的保底机制：跑完之后，记得把容器也删掉，别占着硬盘
	defer DockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})

	// ==========================================
	// 3. 通电执行！
	// ==========================================
	fmt.Println("🚀 正在启动容器运行代码...")
	if err := DockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return JudgeResult{
			Status: StatusSystemError,
			Error:  fmt.Errorf("运行容器失败: %v", err),
		}
	}

	fmt.Println("🎉 容器成功启动，代码正在小黑屋里疯狂运算！")

	// ==========================================
	// 4. 等待容器运行完毕 (带超时强杀的阻塞监听)
	// ==========================================
	statusCh, errCh := DockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	// 祭出第三台对讲机：一个 2 秒的定时器通道
	timeoutCh := time.After(15 * time.Second)
	// 【新增】：定一个变量，专门用来标记是不是运行时错误 (RE)
	var isRuntimeError bool
	// 死死盯住这三个通道，谁先响就听谁的！
	select {
	case err := <-errCh:
		if err != nil {
			return JudgeResult{Status: StatusSystemError, Error: err}
		}
	case status := <-statusCh:
		fmt.Printf("🛑 容器停止运行，退出状态码: %d\n", status.StatusCode)
		if status.StatusCode != 0 {
			// 代码出错了！我们只做个记号，不在这里提日志
			isRuntimeError = true
			fmt.Println("⚠️ 发现非零退出码，标记为运行错误(RE)")
		} else {
			fmt.Println("✅ 容器正常结束，准备提取结果...")
		}

	case <-timeoutCh:
		// 闹钟对讲机响了：说明 2 秒过去了，容器还没跑完！
		// 【新增】：死前抢救一下录像带，看看它到底卡在哪了！
		out, _ := DockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		})
		var stdoutBuf, stderrBuf bytes.Buffer
		stdcopy.StdCopy(&stdoutBuf, &stderrBuf, out)

		fmt.Println("⏰ 警报：代码执行超时 (Time Limit Exceeded)！")
		fmt.Println("🔫 正在强行击毙失控容器...")
		fmt.Printf("💀 临死前抓取到的画面(Stdout): \n%s\n", stdoutBuf.String())
		fmt.Printf("💀 临死前抓取到的报错(Stderr): \n%s\n", stderrBuf.String())

		// 呼叫 Docker 底层 API，发送极其残暴的 SIGKILL 信号，物理拔电
		if killErr := DockerClient.ContainerKill(ctx, resp.ID, "SIGKILL"); killErr != nil {
			fmt.Println("❌ 强杀容器失败，可能已经变成僵尸进程:", killErr)
		} else {
			fmt.Println("💀 失控容器已被成功销毁！")
		}

		// 既然超时了，后面的读取日志也就没必要了，直接中止判题
		return JudgeResult{Status: StatusTimeLimitExceeded}
	}

	// ==========================================
	// 5. 窃听小黑屋的输出 (提取日志)
	// ==========================================
	// 告诉 Docker：把标准输出（正确结果）和标准错误（报错信息）都给我拿出来
	out, err := DockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("获取容器日志失败: %v", err)}
	}
	defer out.Close() // 养成好习惯，用完的流一定要关掉

	// 【核心魔法】破解 Docker 的混合数据流
	var stdoutBuf, stderrBuf bytes.Buffer
	// Docker 为了节省通道，会把标准输出和标准错误揉在一个流里，每一行前面加一个 8 字节的包头。
	// 如果你直接读 string，前面会有乱码。必须用官方的 stdcopy.StdCopy 把它们极其干净地分离开来。
	_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, out)
	if err != nil {
		return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("解析日志数据流失败: %v", err)}
	}

	// ==========================================
	// 6. 宣布最终战果
	// ==========================================
	if isRuntimeError {
		return JudgeResult{
			Status: StatusRuntimeError,
			Output: stderrBuf.String(),
		}
	}
	return JudgeResult{
		Status: StatusAccepted,
		Output: stdoutBuf.String(),
	}
}

/*
// RunSandbox 接收玩家的代码字符串，放入容器中运行
func RunSandbox(code string, input string) JudgeResult {
	ctx := context.Background()

	// ==========================================
	// 1. 在宿主机上准备“秘密文件”
	// ==========================================
	// 在系统的临时目录里建一个随机文件夹 (比如 /tmp/judge_12345)
	workDir, err := os.MkdirTemp("", "judge_*")
	if err != nil {
		return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("创建临时目录失败: %v", err)}
	}
	// 极其关键的保底机制：函数执行完后，自动把这个临时文件夹连同里面的代码一起销毁！(毁尸灭迹)
	defer os.RemoveAll(workDir)

	flag, info, err := CompileCode(ctx, workDir, code)
	if err != nil {
		return JudgeResult{
			Status: StatusSystemError,
			Output: err.Error(),
			Error:  err,
		}
	}
	if flag == false {
		return JudgeResult{
			Status: StatusCompileError,
			Output: info,
		}
	}
	return RunTestCase(ctx, workDir, input)

		// 把玩家的字符串代码，写成一个真正的 main.go 文件
		codePath := filepath.Join(workDir, "main.go")
		if err := os.WriteFile(codePath, []byte(code), 0644); err != nil {
			return JudgeResult{
				Status: StatusSystemError,
				Error:  fmt.Errorf("写入代码文件失败: %v", err),
			}
		}
		//filepath.Join 就是 Go 语言官方提供的“跨平台路径拼接神器”。它会极其智能地判断你当前的代码跑在什么系统上，然后自动填入正确的斜杠。

		为什么这个 0644 对我们的 Docker 沙箱极其重要？
		因为一会儿我们要把这个文件挂载（Bind Mount）到 Docker 容器里去。容器里的那个虚拟环境（Alpine Linux）对于宿主机来说，
		就是一个“其他用户”。如果你在这里写了 0600（不给别人读），那容器里的 Go 编译器一启动就会疯狂报错：
		“Permission denied（没有权限读取 main.go）”！0644 完美保证了宿主机能写，而容器能在里面畅通无阻地读取。


		// =================【新增：写入测试用例文件】=================
		inputPath := filepath.Join(workDir, "input.txt")
		if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
			return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("写入输入文件失败: %v", err)}
		}
		// ==========================================================
		fmt.Printf("📝 代码已写入宿主机临时文件: %s\n", codePath)

		// ==========================================
		// 2. 召唤 Docker，配置“时空虫洞”
		// ==========================================
		resp, err := DockerClient.ContainerCreate(ctx,
			&container.Config{
				Image:      "golang:alpine",
				WorkingDir: "/app", // 让特种兵一落地，就站在 /app 这个目录下
				//Cmd:        []string{"go", "run", "main.go"}, // 编译并运行
				// 【核心修改】：利用 sh -c 和 < 符号，完成输入重定向
				// 加上 GO111MODULE=off 强制单机秒级编译
				Cmd: []string{"sh", "-c", "GO111MODULE=off go run main.go < input.txt"},
			},
			&container.HostConfig{
				// 【核心魔法：挂载卷】
				// 格式是 "宿主机绝对路径:容器内绝对路径"
				// 我们把外面刚才建的 workDir，映射成小黑屋里的 /app 目录
				Binds: []string{workDir + ":/app"},

				// 防御机制：限制内存 256MB
				Resources: container.Resources{
					Memory: 256 * 1024 * 1024,
				},
			}, nil, nil, "")

		if err != nil {
			return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("启动容器失败: %v", err)}
		}

		// 同样的保底机制：跑完之后，记得把容器也删掉，别占着硬盘
		defer DockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})

		// ==========================================
		// 3. 通电执行！
		// ==========================================
		fmt.Println("🚀 正在启动容器运行代码...")
		if err := DockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return JudgeResult{
				Status: StatusSystemError,
				Error:  fmt.Errorf("运行容器失败: %v", err),
			}
		}

		fmt.Println("🎉 容器成功启动，代码正在小黑屋里疯狂运算！")

		// ==========================================
		// 4. 等待容器运行完毕 (带超时强杀的阻塞监听)
		// ==========================================
		statusCh, errCh := DockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

		// 祭出第三台对讲机：一个 2 秒的定时器通道
		timeoutCh := time.After(15 * time.Second)
		// 【新增】：定一个变量，专门用来标记是不是运行时错误 (RE)
		var isRuntimeError bool
		// 死死盯住这三个通道，谁先响就听谁的！
		select {
		case err := <-errCh:
			if err != nil {
				return JudgeResult{Status: StatusSystemError, Error: err}
			}
		case status := <-statusCh:
			fmt.Printf("🛑 容器停止运行，退出状态码: %d\n", status.StatusCode)
			if status.StatusCode != 0 {
				// 代码出错了！我们只做个记号，不在这里提日志
				isRuntimeError = true
				fmt.Println("⚠️ 发现非零退出码，标记为运行错误(RE)")
			} else {
				fmt.Println("✅ 容器正常结束，准备提取结果...")
			}

		case <-timeoutCh:
			// 闹钟对讲机响了：说明 2 秒过去了，容器还没跑完！
			// 【新增】：死前抢救一下录像带，看看它到底卡在哪了！
			out, _ := DockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{
				ShowStdout: true,
				ShowStderr: true,
			})
			var stdoutBuf, stderrBuf bytes.Buffer
			stdcopy.StdCopy(&stdoutBuf, &stderrBuf, out)

			fmt.Println("⏰ 警报：代码执行超时 (Time Limit Exceeded)！")
			fmt.Println("🔫 正在强行击毙失控容器...")
			fmt.Printf("💀 临死前抓取到的画面(Stdout): \n%s\n", stdoutBuf.String())
			fmt.Printf("💀 临死前抓取到的报错(Stderr): \n%s\n", stderrBuf.String())

			// 呼叫 Docker 底层 API，发送极其残暴的 SIGKILL 信号，物理拔电
			if killErr := DockerClient.ContainerKill(ctx, resp.ID, "SIGKILL"); killErr != nil {
				fmt.Println("❌ 强杀容器失败，可能已经变成僵尸进程:", killErr)
			} else {
				fmt.Println("💀 失控容器已被成功销毁！")
			}

			// 既然超时了，后面的读取日志也就没必要了，直接中止判题
			return JudgeResult{Status: StatusTimeLimitExceeded}
		}

		// ==========================================
		// 5. 窃听小黑屋的输出 (提取日志)
		// ==========================================
		// 告诉 Docker：把标准输出（正确结果）和标准错误（报错信息）都给我拿出来
		out, err := DockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		})
		if err != nil {
			return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("获取容器日志失败: %v", err)}
		}
		defer out.Close() // 养成好习惯，用完的流一定要关掉

		// 【核心魔法】破解 Docker 的混合数据流
		var stdoutBuf, stderrBuf bytes.Buffer
		// Docker 为了节省通道，会把标准输出和标准错误揉在一个流里，每一行前面加一个 8 字节的包头。
		// 如果你直接读 string，前面会有乱码。必须用官方的 stdcopy.StdCopy 把它们极其干净地分离开来。
		_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, out)
		if err != nil {
			return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("解析日志数据流失败: %v", err)}
		}

		// ==========================================
		// 6. 宣布最终战果
		// ==========================================
		if isRuntimeError {
			return JudgeResult{
				Status: StatusRuntimeError,
				Output: stderrBuf.String(),
			}
		}
		return JudgeResult{
			Status: StatusAccepted,
			Output: stdoutBuf.String(),
		}

}
*/
