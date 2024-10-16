package fan_out

import (
	"fmt"
	"sync"
	"time"
)

// Generator 生成整数序列
// done: 用于发送停止信号的通道
// 返回一个只读的整数通道
func Generator(done <-chan struct{}) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)    // 确保在函数结束时关闭通道
		for i := 1; ; i++ { // 无限循环生成整数
			select {
			case out <- i: // 尝试发送数字到 out 通道
			case <-done: // 如果收到结束信号，则退出
				return
			}
		}
	}()
	return out
}

// FanOut 将输入通道的数据分发到多个输出通道
// done: 用于发送停止信号的通道
// in: 输入的整数通道
// n: 要创建的 worker 数量
// 返回一个只读的整数通道切片
func FanOut(done <-chan struct{}, in <-chan int, n int) []<-chan int {
	outputs := make([]<-chan int, n)
	for i := 0; i < n; i++ {
		outputs[i] = Worker(done, in, i)
	}
	return outputs
}

// Worker 处理输入通道的数据
// done: 用于发送停止信号的通道
// in: 输入的整数通道
// id: worker 的唯一标识符
// 返回一个只读的整数通道
func Worker(done <-chan struct{}, in <-chan int, id int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // 确保在函数结束时关闭通道
		for n := range in {
			select {
			case out <- n * n: // 计算平方并发送结果
				fmt.Printf("Worker %d: %d squared is %d\n", id, n, n*n)
			case <-done: // 如果收到结束信号，则退出
				return
			}
		}
	}()
	return out
}

// RunFanOut 运行 Fan-Out 模式示例
func RunFanOut() {
	done := make(chan struct{})
	defer close(done) // 确保在函数结束时发送结束信号

	in := Generator(done)          // 创建一个生成器
	outputs := FanOut(done, in, 3) // 创建 3 个 worker

	var wg sync.WaitGroup
	wg.Add(len(outputs))

	// 为每个输出通道启动一个 goroutine 来处理结果
	for i, c := range outputs {
		go func(i int, c <-chan int) {
			defer wg.Done()
			for n := range c {
				fmt.Printf("Output %d: %d\n", i, n)
			}
		}(i, c)
	}

	// 让程序运行一段时间后停止
	go func() {
		<-time.After(2 * time.Second)
		close(done)
	}()

	wg.Wait() // 等待所有 goroutine 完成
}
