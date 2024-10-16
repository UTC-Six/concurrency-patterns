package main

import (
	"fmt"
	"sync"
)

// 主函数
func main() {
	wg := sync.WaitGroup{}
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})
	defer close(ch1)
	defer close(ch2)
	defer close(ch3)

	wg.Add(3)

	// 打印 A 的 goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			<-ch1
			fmt.Print("A")
			ch2 <- struct{}{}
		}
	}()

	// 打印 B 的 goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			<-ch2
			fmt.Print("B")
			ch3 <- struct{}{}
		}
	}()

	// 打印 C 的 goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			<-ch3
			fmt.Print("C")
			if i < 9 {
				ch1 <- struct{}{}
			}
		}
	}()

	ch1 <- struct{}{}
	wg.Wait()
}

// 流量消峰 kafka
// 扣费（非实时）保证高性能

//1. 收到请求就把消息 push 到 kafka 中
//2. 在消费端去扣款
//3. 扣款时，如果余额小于 0，修改 redis 用户的状态，拒绝新请求
//
//4. 另外也可以同通过一个 redis 计数一个大于  余额请求次数的值，如果周期内大于这个次数限制，就再次触发同步校验余额，小于 0 立即修改 redis 状态，拒绝新请求

// 简化的生产者-消费者模式示例
var ch2 = make(chan struct{})

func pub() {
	go func() { // 生产者
		for i := 0; i < 10; i++ {
			ch2 <- struct{}{}
		}
	}()
}

func cus() {
	go func() { // 消费者
		for i := 0; i < 10; i++ {
			<-ch2
			fmt.Print("B")
		}
	}()
}
