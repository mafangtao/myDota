package timer

import (
	"testing"
	"fmt"
	"time"
)

func TestTimer(t *testing.T) {
	for i:= 0; i < 5; i++{
		NewTimerAfeter(f, []interface{}{i, 2*i}, time.Duration(2*i) * time.Second).Run()
		fmt.Printf("第%d个timer 定时完毕\n", i)
	}

	time.Sleep(time.Minute)
}

func f(v ...interface{}){
	fmt.Printf("No.%d function, delay %d second(s)\n", v[0].(int), v[1].(int))
}
