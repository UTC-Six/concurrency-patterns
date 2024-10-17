package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Generator 生成有限的整数序列
// ctx: 用于控制 goroutine 生命周期的 context
// id: 生成器的唯一标识符
// nums: 要生成的整数序列
// 返回一个只读的整数通道
func Generator(ctx context.Context, id int, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // 确保在函数结束时关闭通道
		for _, n := range nums {
			select {
			case out <- n:
				// 模拟数据生成的时间，添加随机延迟
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			case <-ctx.Done():
				// 如果收到取消信号，打印消息并退出
				fmt.Printf("Generator %d: 收到 done 信号\n", id)
				return
			}
		}
	}()
	return out
}

// FanIn 将多个输入通道合并到一个输出通道
// ctx: 用于控制 goroutine 生命周期的 context
// channels: 要合并的输入通道列表
// 返回一个只读的整数通道，包含所有输入通道的数据
func FanIn(ctx context.Context, channels ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	multiplexedStream := make(chan int)

	// multiplex 函数用于从一个输入通道读取数据并发送到 multiplexedStream
	multiplex := func(c <-chan int) {
		defer wg.Done()
		for i := range c {
			select {
			case multiplexedStream <- i:
			case <-ctx.Done():
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

func main() {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 减少超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// 减少生成的数字数量
	in1 := Generator(ctx, 1, 1, 2, 3)
	in2 := Generator(ctx, 2, 4, 5, 6)
	in3 := Generator(ctx, 3, 7, 8, 9)

	merged := FanIn(ctx, in1, in2, in3)

	isDone := false
	results := make(chan string, 100)

	go func() {
		for v := range merged {
			if !isDone {
				select {
				case <-ctx.Done():
					isDone = true
					fmt.Println("主 goroutine: 收到 done 信号")
				default:
				}
			}

			// 增加处理时间
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

			if isDone {
				results <- fmt.Sprintf("%d (done 后的输出)", v)
			} else {
				results <- fmt.Sprintf("%d", v)
			}
		}
		close(results)
	}()

	<-ctx.Done()
	fmt.Println("Context 已结束，但仍在处理剩余数据...")

	// 打印所有结果
	for result := range results {
		fmt.Println(result)
	}

	fmt.Println("所有工作已完成，程序退出")
}
