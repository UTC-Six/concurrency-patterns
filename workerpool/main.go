package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Job 代表一个需要处理的任务
type Job int

// Result 代表任务处理的结果
type Result int

// Worker 函数处理任务并产生结果
func Worker(ctx context.Context, id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			fmt.Printf("worker %d processing job %d\n", id, job)
			results <- Result(job * 2)
		case <-ctx.Done():
			fmt.Printf("worker %d stopping due to context cancellation\n", id)
			return
		}
	}
}

func main() {
	// 创建一个3秒后超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 创建任务和结果通道
	jobs := make(chan Job, 100)
	results := make(chan Result, 100)

	// 创建WaitGroup来同步worker
	var wg sync.WaitGroup

	// 启动工作池
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go Worker(ctx, i, jobs, results, &wg)
	}

	// 发送任务
	go func() {
		for i := 0; i < 5; i++ {
			jobs <- Job(i)
		}
		close(jobs)
	}()

	// 使用 for + select 接收结果
	resultCount := 0
	for {
		select {
		case result := <-results:
			fmt.Printf("Result: %d\n", result)
			resultCount++
			if resultCount == 5 {
				return
			}
		case <-ctx.Done():
			fmt.Println("Main: context deadline exceeded")
			return
		}
	}
}
