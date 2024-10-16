package fan_in

import (
	"fmt"
	"sync"
)

// Generator 生成整数序列
// done: 用于发送停止信号的通道
// nums: 要生成的整数序列
// 返回一个只读的整数通道
func Generator(done <-chan struct{}, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // 确保在函数结束时关闭通道
		for _, n := range nums {
			select {
			case out <- n: // 尝试发送数字到 out 通道
			case <-done: // 如果收到结束信号，则退出
				return
			}
		}
	}()
	return out
}

// FanIn 将多个输入通道合并到一个输出通道
// done: 用于发送停止信号的通道
// channels: 要合并的输入通道列表
// 返回一个只读的整数通道，包含所有输入通道的数据
func FanIn(done <-chan struct{}, channels ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	multiplexedStream := make(chan int)

	// multiplex 函数用于从一个输入通道读取数据并发送到 multiplexedStream
	multiplex := func(c <-chan int) {
		defer wg.Done()
		for i := range c {
			select {
			case multiplexedStream <- i:
			case <-done:
				return
			}
		}
	}

	// 为每个输入通道启动一个 goroutine
	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	// 等待所有输入通道处理完毕，然后关闭输出通道
	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	return multiplexedStream
}

// RunFanIn 运行 Fan-In 模式示例
func RunFanIn() {
	done := make(chan struct{})
	defer close(done) // 确保在函数结束时发送结束信号

	// 创建三个生成器，每个生成器产生不同的整数序列
	in1 := Generator(done, 1, 2, 3)
	in2 := Generator(done, 4, 5, 6)
	in3 := Generator(done, 7, 8, 9)

	// 使用 FanIn 函数合并三个输入通道
	merged := FanIn(done, in1, in2, in3)

	// 从合并后的通道中读取并打印所有值
	for v := range merged {
		fmt.Println(v)
	}
}
