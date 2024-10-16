package workerpool

import (
	"fmt"
	"sync"
)

// Job 代表一个需要处理的任务
// 这里使用 int 类型简化了任务的表示
type Job int

// Result 代表任务处理的结果
// 同样使用 int 类型简化了结果的表示
type Result int

// Worker 函数处理任务并产生结果
// id: worker 的唯一标识符
// jobs: 用于接收任务的只读通道
// results: 用于发送结果的只写通道
// wg: 用于同步的 WaitGroup，确保所有 worker 完成工作
func Worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	// 确保在函数返回时调用 wg.Done()，表示这个 worker 已完成工作
	defer wg.Done()
	// 持续从 jobs 通道中接收任务
	for job := range jobs {
		// 打印 worker 正在处理的任务信息
		fmt.Printf("worker %d processing job %d\n", id, job)
		// 将任务结果（这里简单地将任务值乘以2）发送到 results 通道
		results <- Result(job * 2) // 模拟工作处理过程
	}
}

// RunWorkerPool 启动工作池并处理任务
func RunWorkerPool() {
	// 设置任务数量和工作者数量
	const numJobs = 9
	const numWorkers = 3

	// 创建用于发送任务和接收结果的缓冲通道
	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	// 创建 WaitGroup 来同步所有 worker
	var wg sync.WaitGroup
	// 添加要等待的 worker 数量
	wg.Add(numWorkers)

	// 创建 worker
	for w := 1; w <= numWorkers; w++ {
		// 启动 worker goroutine
		go Worker(w, jobs, results, &wg)
	}

	// 发送任务
	for j := 1; j <= numJobs; j++ {
		jobs <- Job(j)
	}
	// 关闭 jobs 通道，表示不再有新任务
	close(jobs)

	// 等待所有 worker 完成
	go func() {
		// 等待所有 worker 完成工作
		wg.Wait()
		// 关闭 results 通道，表示所有结果都已处理完毕
		close(results)
	}()

	// 收集并打印结果
	for result := range results {
		fmt.Printf("Received result: %d\n", result)
	}
}
