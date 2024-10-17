package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Generator 生成有限的整数序列
func Generator(ctx context.Context, limit int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; i <= limit; i++ {
			select {
			case out <- i:
				time.Sleep(10 * time.Millisecond) // 减缓生成速度
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// FanOut 将输入通道的数据分发到多个输出通道
func FanOut(ctx context.Context, in <-chan int, n int) []<-chan int {
	outputs := make([]<-chan int, n)
	for i := 0; i < n; i++ {
		outputs[i] = Worker(ctx, in, i)
	}
	return outputs
}

// Worker 处理输入通道的数据
func Worker(ctx context.Context, in <-chan int, id int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		isDone := false
		for n := range in {
			if !isDone {
				select {
				case <-ctx.Done():
					isDone = true
					fmt.Printf("Worker %d: 收到 done 信号\n", id)
				default:
				}
			}
			// 模拟处理时间
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			result := n * n
			if isDone {
				fmt.Printf("Worker %d: %d squared is %d (done 后的输出)\n", id, n, result)
			} else {
				fmt.Printf("Worker %d: %d squared is %d\n", id, n, result)
			}
			out <- result
		}
	}()
	return out
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	in := Generator(ctx, 20) // 限制生成 20 个数字
	outputs := FanOut(ctx, in, 3)

	var wg sync.WaitGroup
	wg.Add(len(outputs))

	results := make(chan string, 100) // 用于收集所有输出

	for i, c := range outputs {
		go func(i int, c <-chan int) {
			defer wg.Done()
			isDone := false
			for n := range c {
				if !isDone {
					select {
					case <-ctx.Done():
						isDone = true
						fmt.Printf("Output %d: 收到 done 信号\n", i)
					default:
					}
				}
				if isDone {
					results <- fmt.Sprintf("Output %d: %d (done 后的输出)", i, n)
				} else {
					results <- fmt.Sprintf("Output %d: %d", i, n)
				}
			}
		}(i, c)
	}

	// 等待 context 结束或所有数据处理完毕
	go func() {
		wg.Wait()
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
