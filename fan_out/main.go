package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Generator 生成整数序列
func Generator(ctx context.Context) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; ; i++ {
			select {
			case out <- i:
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
				default:
				}
			}
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
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	in := Generator(ctx)
	outputs := FanOut(ctx, in, 3)

	var wg sync.WaitGroup
	wg.Add(len(outputs))

	for i, c := range outputs {
		go func(i int, c <-chan int) {
			defer wg.Done()
			isDone := false
			for n := range c {
				if !isDone {
					select {
					case <-ctx.Done():
						isDone = true
					default:
					}
				}
				if isDone {
					fmt.Printf("Output %d: %d (done 后的输出)\n", i, n)
				} else {
					fmt.Printf("Output %d: %d\n", i, n)
				}
			}
		}(i, c)
	}

	<-ctx.Done()
	fmt.Println("Context 已结束，但仍在处理剩余数据...")

	wg.Wait()
	fmt.Println("所有工作已完成，程序退出")
}
