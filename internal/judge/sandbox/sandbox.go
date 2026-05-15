package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"gojo/internal/judge/docker"
	"gojo/internal/judge/model"

	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

// ========= 【新增：定义判题机状态码和战报】 =========

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
	resp, err := docker.DockerClient.ContainerCreate(ctx,
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
	defer docker.DockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})

	// 启动编译容器
	docker.DockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{})

	statusCh, errCh := docker.DockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

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
		out, _ := docker.DockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
		var stdoutBuf, stderrBuf bytes.Buffer
		stdcopy.StdCopy(&stdoutBuf, &stderrBuf, out)

		if status.StatusCode != 0 {
			// 编译失败（语法错误等）
			return false, stderrBuf.String(), nil
		}

	case <-timeoutCh:

		// 呼叫 Docker 底层 API，发送极其残暴的 SIGKILL 信号，物理拔电
		if killErr := docker.DockerClient.ContainerKill(ctx, resp.ID, "SIGKILL"); killErr != nil {
			fmt.Println("❌ 强杀容器失败，可能已经变成僵尸进程:", killErr)
		} else {
			fmt.Println("💀 失控容器已被成功销毁！")
		}
		return false, "编译超时 (Compile Time Limit Exceeded)", nil
	}

	// 编译成功！宿主机的 workDir 下现在已经有了一个名为 "solution" 的二进制文件
	return true, "", nil
}

// StartPersistentSandbox 启动一个常驻的测试沙箱
// 返回容器 ID 和 error
func StartPersistentSandbox(ctx context.Context, workDir string) (string, error) {
	// 假设你的 client 叫 dockerClient
	// 1. 创建容器：让它执行 sleep 3600，保证它不会立刻退出
	resp, err := docker.DockerClient.ContainerCreate(ctx, &container.Config{
		Image:      "golang:alpine", // 继续用你之前拉下来的镜像
		Cmd:        []string{"sleep", "3600"},
		WorkingDir: "/app", // 容器内的工作目录
	}, &container.HostConfig{
		NetworkMode: "none",                      // 👈 绝杀！拔掉网线，让代码在一座真正的孤岛上运行！
		Binds:       []string{workDir + ":/app"}, // 把宿主机的代码和编译好的 exe 挂载进去
		// 这里可以加上内存和 CPU 的终极限制，防止死循环炸服
		Resources: container.Resources{
			Memory:    256 * 1024 * 1024, // 限制 256MB 内存
			NanoCPUs:  1 * 1e9,           // 限制 1 个 CPU 核心
			PidsLimit: &[]int64{64}[0],   // 👈 绝杀！最多只允许 64 个进程存活
		},
	}, nil, nil, "")

	if err != nil {
		return "", fmt.Errorf("创建常驻沙箱失败: %v", err)
	}

	// 2. 启动容器
	if err := docker.DockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("启动常驻沙箱失败: %v", err)
	}

	return resp.ID, nil
}

// ExecTestCase 在正在运行的沙箱中执行单次测试
// ExecTestCase 在正在运行的沙箱中执行单次测试
func ExecTestCase(ctx context.Context, containerID string, input string) model.JudgeResult {
	// 1. 创建 Exec 任务（派小弟进去）
	execCreate, err := docker.DockerClient.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          []string{"./solution"},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return model.JudgeResult{Status: model.StatusSystemError, Error: fmt.Errorf("无法创建 Exec 任务: %v", err)}
	}

	// 2. 挂载 I/O 流（拉一根窃听线）
	hijackedResp, err := docker.DockerClient.ContainerExecAttach(ctx, execCreate.ID, container.ExecStartOptions{})
	if err != nil {
		return model.JudgeResult{Status: model.StatusSystemError, Error: fmt.Errorf("无法执行测试: %v", err)}
	}
	defer hijackedResp.Close()

	// 3. 【极速喂饭】把测试输入喂给标准输入流，然后立刻掐断！
	_, _ = hijackedResp.Conn.Write([]byte(input + "\n"))
	hijackedResp.CloseWrite() // 极其关键：防止 scanf 死等

	// 4. 【核心魔法：同步阻塞读取】
	var stdoutBuf, stderrBuf bytes.Buffer

	// 这里的 StdCopy 会一直阻塞，直到小弟把代码跑完，或者外部的 ctx 超时！
	// 根本不需要你手动写 select 和 time.After，极其优雅！
	_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, hijackedResp.Reader)

	// 5. 检查是不是 TLE (超时被强杀)
	// 如果外面传进来的 ctx 带有 WithTimeout，到时间后 StdCopy 会强行中断并返回 DeadlineExceeded
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("⏰ 警报：单次用例执行超时 (TLE)！")
			return model.JudgeResult{Status: model.StatusTimeLimitExceeded}
		}
		return model.JudgeResult{Status: model.StatusSystemError, Error: fmt.Errorf("读取数据流异常: %v", err)}
	}

	// 6. 检查是不是 RE (运行报错，比如数组越界、除以零)
	// 我们不查 Container 的状态，我们查这个特定 Exec 任务的退出码！
	inspectResp, err := docker.DockerClient.ContainerExecInspect(ctx, execCreate.ID)
	if err != nil {
		return model.JudgeResult{Status: model.StatusSystemError, Error: fmt.Errorf("无法获取退出码: %v", err)}
	}

	if inspectResp.ExitCode != 0 {
		// 💡 核心逻辑：137 通常代表被信号 9 (SIGKILL) 强杀
		// 如果不是因为超时（TLE）被我们手动杀掉的，那么在内存限制下，它就是 MLE
		if inspectResp.ExitCode == 137 {
			// 为了更严谨，可以进一步检查容器状态
			containerInfo, _ := docker.DockerClient.ContainerInspect(ctx, containerID)
			if containerInfo.State.OOMKilled {
				return model.JudgeResult{Status: model.StatusMemoryLimitExceeded, Output: "Memory Limit Exceeded"}
			}
			// 如果容器没死但进程死了且退出码是 137，在 OJ 场景下基本也可以判定为 MLE
			return model.JudgeResult{Status: model.StatusMemoryLimitExceeded}
		}

		return model.JudgeResult{
			Status: model.StatusRuntimeError,
			Output: stderrBuf.String(),
		}
	}

	// 7. 完美运行，返回标准输出结果
	return model.JudgeResult{
		Status: model.StatusAccepted,
		Output: stdoutBuf.String(),
	}
}

// RemoveSandbox 物理销毁常驻沙箱（由 Service 在 defer 中调用，确保绝对不漏）
func RemoveSandbox(ctx context.Context, containerID string) {
	if containerID == "" {
		return
	}
	// 发送 Force: true 强制拔电源，管你里面有没有死循环，瞬间超度！
	err := docker.DockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	if err != nil {
		fmt.Printf("⚠️ 警告：沙箱容器 %s 销毁失败，可能产生孤儿容器: %v\n", containerID, err)
	} else {
		fmt.Printf("💥 沙箱容器 %s 已被成功物理抹杀！\n", containerID)
	}
}
