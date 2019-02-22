package timer

import (
	"testing"
	"time"
	"fmt"
)

func TestTimeWheel(t *testing.T){
	//创建秒级时间轮
	secondWheel := NewTimeWheel(SECOND_NAME, SECOND_INTERVAL, SECOND_SCALES, TIMERS_PER_SCALE)
	secondWheel.Run()
	//创建分钟级时间轮
	minuteWheel := NewTimeWheel(MINUTE_NAME, MINUTE_INTERVAL, MINUTE_SCALES, TIMERS_PER_SCALE)
	minuteWheel.AddTimeWheel(secondWheel)
	minuteWheel.Run()
	//创建小时级时间轮
	hourWheel := NewTimeWheel(HOUR_NAME, HOUR_INTERVAL, HOUR_SCALES, TIMERS_PER_SCALE)
	hourWheel.AddTimeWheel(minuteWheel)
	hourWheel.Run()

	hourWheel.Run()
	timer1 := NewTimerAfeter(foo, []interface{}{1, 3000}, 3*time.Second)
	hourWheel.AddTimer(1, timer1)

	timer2 := NewTimerAfeter(foo, []interface{}{2, 1000}, time.Second)
	hourWheel.AddTimer(2, timer2)

	timer3 := NewTimerAfeter(foo, []interface{}{3, 5000}, 5*time.Second)
	hourWheel.AddTimer(3, timer3)

	timer4 := NewTimerAfeter(foo, []interface{}{4, 2000}, 2*time.Second)
	hourWheel.AddTimer(4, timer4)



	go func(){
		for{
			timers := hourWheel.GetTimerWithIn(500*time.Millisecond)
			for _, timer := range timers{
				timer.delayFunc.Call()
			}

			time.Sleep(500*time.Millisecond)
		}

	}()

	k := 0
	for{
		k++
		fmt.Println("/////////////////////////", k)
		time.Sleep(time.Second)
	}
}
