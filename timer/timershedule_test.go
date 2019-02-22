package timer

import (
	"testing"
	"fmt"
	"time"
)

func TestNewTimerSchedule(t *testing.T) {
	timerSchedule := NewTimerSchedule()
	timerSchedule.Start()

	//在schedule中添加timer
	for i := 1; i < 2000; i++{
		tid, err := timerSchedule.CreateTimerAfter(foo, []interface{}{i, 3*i}, time.Duration(3*i) * time.Millisecond)
		if err != nil{
			fmt.Println("createtimerafter err", tid, err)
		}
		//fmt.Println("tid:", tid)
	}

	//执行到时的延迟函数
	go func(){
		dfsChan := timerSchedule.GetTriggerChan()
		for df := range dfsChan{
			df.Call()
		}
	}()

	//读秒等待
	k := 0
	for{
		k++
		fmt.Println("//////////////////////////////////////////////////", k)
		time.Sleep(time.Second)
	}
}


func foo (args ...interface{}){
	fmt.Printf("I am No.%d functiong, delay %dms\n", args[0].(int), args[1].(int))
}

func TestNewAutoExecTimerSchedule(t *testing.T) {
	tsAuto := NewAutoExecTimerSchedule()
	//在schedule中添加timer
	for i := 1; i < 2000; i++{
		tid, err := tsAuto.CreateTimerAfter(foo, []interface{}{i, 3*i}, time.Duration(3*i) * time.Millisecond)
		if err != nil{
			fmt.Println("createtimerafter err", tid, err)
		}
	}

	//读秒等待
	k := 0
	for{
		k++
		fmt.Println("//////////////////////////////////////////////////", k)
		time.Sleep(time.Second)
	}
}