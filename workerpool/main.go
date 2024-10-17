package main

import (
	"context" // 导入context包,用于处理超时和取消操作
	"fmt"     // 导入fmt包,用于打印输出
	"sync"    // 导入sync包,用于同步goroutine
	"time"    // 导入time包,用于处理时间相关操作
)

// Job 代表一个需要处理的任务
// 这里使用int类型简化任务的表示
type Job int

// Result 代表任务处理的结果
// 同样使用int类型简化结果的表示
type Result int

// Worker 函数处理任务并产生结果
// ctx: 用于控制goroutine的生命周期
// id: worker的唯一标识符
// jobs: 用于接收任务的只读通道
// results: 用于发送结果的只写通道
// wg: 用于同步的WaitGroup,确保所有worker完成工作
func Worker(ctx context.Context, id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done() // 确保在函数返回时调用wg.Done(),表示这个worker已完成工作
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				// 如果jobs通道已关闭,worker退出
				return
			}
			fmt.Printf("worker %d processing job %d\n", id, job)
			results <- Result(job * 2) // 将任务结果(这里简单地将任务值乘以2)发送到results通道
		case <-ctx.Done():
			// 如果context被取消,打印错误信息并退出
			fmt.Printf("worker %d stopping due to context cancellation: %v\n", id, ctx.Err())
			return
		}
	}
}

func main() {
	// 创建一个3秒后超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // 确保在main函数结束时调用cancel(),释放资源

	// 创建任务和结果通道,缓冲区大小为100
	jobs := make(chan Job, 100)
	results := make(chan Result, 100)

	// 创建WaitGroup来同步worker
	var wg sync.WaitGroup

	// 启动工作池,创建3个worker
	for i := 0; i < 3; i++ {
		wg.Add(1) // 每启动一个worker,将WaitGroup计数器加1
		go Worker(ctx, i, jobs, results, &wg)
	}

	// 发送任务
	go func() {
		for i := 0; i < 5; i++ {
			jobs <- Job(i) // 发送5个任务到jobs通道
		}
		close(jobs) // 所有任务发送完毕后关闭jobs通道
	}()

	// 使用 for + select 接收结果
	resultCount := 0
	for {
		select {
		case result := <-results:
			fmt.Printf("Result: %d\n", result)
			resultCount++
			if resultCount == 5 {
				// 如果已接收5个结果,程序结束
				return
			}
		case <-ctx.Done():
			// 如果context被取消(超时或手动取消),打印错误信息
			fmt.Printf("Main: context cancelled: %v\n", ctx.Err())
			wg.Wait() // 等待所有worker结束
			return
		default:
			// 检查context是否已经结束
			if err := ctx.Err(); err != nil {
				fmt.Printf("Main: context error: %v\n", err)
				wg.Wait() // 等待所有worker结束
				return
			}
			// 如果context没有结束,短暂休眠后继续循环
			time.Sleep(100 * time.Millisecond)
		}
	}
}
