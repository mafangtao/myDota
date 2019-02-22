package timer

import (
	"sync"
	"time"
	"math"
	"myDota/logger"
)

const (
	HOUR_NAME = "HOUR"
	HOUR_INTERVAL = 60 * 60 * 1e3
	HOUR_SCALES = 12

	MINUTE_NAME = "MINUTE"
	MINUTE_INTERVAL = 60 * 1e3
	MINUTE_SCALES = 60

	SECOND_NAME = "SECOND"
	SECOND_INTERVAL = 1e3
	SECOND_SCALES = 60

	TIMERS_PER_SCALE = 2048

	//默认最大误差100ms
	MAX_TIME_DELAY = 100

	//默认缓冲队列大小
	MAX_CHAN_BUFF = 2048
)

type TimerSchedule struct{
	timeWheel *TimeWheel
	idGen uint32
	triggerChan chan *DelayFunc
	sync.RWMutex
}

func NewTimerSchedule()* TimerSchedule{
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

	return &TimerSchedule{
		timeWheel:hourWheel,
		triggerChan:make(chan *DelayFunc, MAX_CHAN_BUFF),
	}
}

//创建timer，然后把timer放入分层时间轮，返回timer的tid和可能遇到的错误
func (ts *TimerSchedule) CreateTimerAt(f func(args ...interface{}), args []interface{}, unixNano int64)(uint32, error){
	ts.Lock()
	defer ts.Unlock()

	ts.idGen++
	return ts.idGen, ts.timeWheel.AddTimer(ts.idGen, NewTimerAt(f, args, unixNano))
}

//创建timer，然后把timer放入分层时间轮，返回timer的tid和可能遇到的错误
func (ts *TimerSchedule)CreateTimerAfter(f func(args ...interface{}), args []interface{}, duration time.Duration)(uint32, error){
	ts.Lock()
	defer ts.Unlock()

	ts.idGen++
	return ts.idGen, ts.timeWheel.AddTimer(ts.idGen, NewTimerAfeter(f, args, duration))
}

//删除timer
func (ts *TimerSchedule) CancelTimer(tid uint32){
	ts.Lock()
	defer ts.Unlock()

	ts.timeWheel.RemoveTimer(tid)
}

//获取计时结束的延迟执行函数的通道
func (ts *TimerSchedule) GetTriggerChan() chan *DelayFunc{
	return ts.triggerChan
}

//非阻塞方式启动timerSchedule
func (ts *TimerSchedule) Start(){
	go func(){
		for{
			now := UnixMilli()
			triggerList := ts.timeWheel.GetTimerWithIn(MAX_TIME_DELAY * time.Millisecond)
			for _, timer := range triggerList{
				if math.Abs(float64(now-timer.unixts)) > MAX_TIME_DELAY{
					logger.Error("want call at ", timer.unixts, " ; real call at ", now, " ; delay ", now-timer.unixts)
				}
				//fmt.Println("delayFunc 被写入管道")
				ts.triggerChan <- timer.delayFunc
			}

			time.Sleep(MAX_TIME_DELAY/2 * time.Millisecond)
		}
	}()
}

func NewAutoExecTimerSchedule() *TimerSchedule{
	autoExecSchedule := NewTimerSchedule()
	autoExecSchedule.Start()

	go func(){
		dfsChan := autoExecSchedule.GetTriggerChan()
		for df := range dfsChan{
			go df.Call()
		}
	}()

	return autoExecSchedule
}






























