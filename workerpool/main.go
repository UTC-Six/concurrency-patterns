package main

import (
	"context"   // 导入context包,用于处理超时和取消操作
	"fmt"       // 导入fmt包,用于打印输出
	"math/rand" // 导入math/rand包,用于生成随机数
	"sync"      // 导入sync包,用于同步goroutine
	"time"      // 导入time包,用于处理时间相关操作
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
			fmt.Printf("Worker %d started job %d\n", id, job)
			// 模拟一个耗时的任务
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			result := Result(job * 2)
			fmt.Printf("Worker %d completed job %d, result: %d\n", id, job, result)
			results <- result
		case <-ctx.Done():
			// 如果context被取消,打印错误信息并退出
			fmt.Printf("Worker %d stopping due to context cancellation: %v\n", id, ctx.Err())
			return
		}
	}
}

func main() {
	// 创建一个5秒后超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 确保在main函数结束时调用cancel(),释放资源

	// 创建任务和结果通道,缓冲区大小为100
	jobs := make(chan Job, 100)
	results := make(chan Result, 100)

	// 创建WaitGroup来同步worker
	var wg sync.WaitGroup

	// 启动5个worker
	for i := 0; i < 5; i++ {
		wg.Add(1) // 每启动一个worker,将WaitGroup计数器加1
		go Worker(ctx, i, jobs, results, &wg)
	}

	// 发送20个任务
	go func() {
		for i := 0; i < 20; i++ {
			jobs <- Job(i) // 发送20个任务到jobs通道
		}
		close(jobs) // 所有任务发送完毕后关闭jobs通道
	}()

	resultCount := 0
	for {
		select {
		case result := <-results:
			fmt.Printf("Received result: %d\n", result)
			resultCount++
			if resultCount == 20 {
				fmt.Println("All results received")
				return
			}
		case <-ctx.Done():
			// 如果context被取消(超时或手动取消),打印错误信息
			fmt.Printf("Main: context cancelled: %v\n", ctx.Err())
			wg.Wait() // 等待所有worker结束
			return
		}
	}
}
