package judge

import (
	"bytes"
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

// StartPersistentSandbox 启动一个常驻的测试沙箱
// 返回容器 ID 和 error
func StartPersistentSandbox(ctx context.Context, workDir string) (string, error) {
	// 假设你的 client 叫 dockerClient
	// 1. 创建容器：让它执行 sleep 3600，保证它不会立刻退出
	resp, err := DockerClient.ContainerCreate(ctx, &container.Config{
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
	if err := DockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("启动常驻沙箱失败: %v", err)
	}

	return resp.ID, nil
}

// ExecTestCase 在正在运行的沙箱中执行单次测试
// ExecTestCase 在正在运行的沙箱中执行单次测试
func ExecTestCase(ctx context.Context, containerID string, input string) JudgeResult {
	// 1. 创建 Exec 任务（派小弟进去）
	execCreate, err := DockerClient.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          []string{"./solution"},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("无法创建 Exec 任务: %v", err)}
	}

	// 2. 挂载 I/O 流（拉一根窃听线）
	hijackedResp, err := DockerClient.ContainerExecAttach(ctx, execCreate.ID, container.ExecStartOptions{})
	if err != nil {
		return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("无法执行测试: %v", err)}
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
			return JudgeResult{Status: StatusTimeLimitExceeded}
		}
		return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("读取数据流异常: %v", err)}
	}

	// 6. 检查是不是 RE (运行报错，比如数组越界、除以零)
	// 我们不查 Container 的状态，我们查这个特定 Exec 任务的退出码！
	inspectResp, err := DockerClient.ContainerExecInspect(ctx, execCreate.ID)
	if err != nil {
		return JudgeResult{Status: StatusSystemError, Error: fmt.Errorf("无法获取退出码: %v", err)}
	}

	if inspectResp.ExitCode != 0 {
		fmt.Printf("⚠️ 发现非零退出码 %d，标记为运行错误(RE)\n", inspectResp.ExitCode)
		return JudgeResult{
			Status: StatusRuntimeError,
			Output: stderrBuf.String(), // 提取报错日志
		}
	}

	// 7. 完美运行，返回标准输出结果
	return JudgeResult{
		Status: StatusAccepted,
		Output: stdoutBuf.String(),
	}
}
