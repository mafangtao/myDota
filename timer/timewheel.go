package timer

import (
	"sync"
	"time"
	"myDota/logger"
	"fmt"
	"errors"
)

/*******************************************
*分层时间轮
*******************************************/
type TimeWheel struct{
	name 	string	//TimeWheel的名称
	interval int64 	//刻度间的时间间隔，单位ms
	scales int		//每个时间轮上的刻度数
	curIndex int	//当前的指向
	maxCap int 		//每个刻度的timer的最大容量
	timerQueue map[int]map[uint32] *Timer	//时间轮上所有timer
	nextTimeWheel *TimeWheel //下一层时间轮
	sync.RWMutex
}

func NewTimeWheel(name string, interval int64, scales int, maxCap int) *TimeWheel{
	tw :=  &TimeWheel{
		name : name,
		interval:interval,
		scales:scales,
		maxCap:maxCap,
		timerQueue: make(map[int]map[uint32]*Timer, scales),
	}
	for i:= 0; i < scales; i++{
		tw.timerQueue[i] = make(map[uint32]*Timer, maxCap)
	}

	return tw
}

func (tw *TimeWheel) AddTimer(tid uint32, t *Timer) error{
	tw.Lock()
	defer tw.Unlock()
	return tw.addTimer(tid, t, false)
}
//把Timer添加至分层时间轮
func (tw *TimeWheel) addTimer(tid uint32, t *Timer, forceNext bool) error{
	defer func() error{
		if err := recover(); err != nil{
			errstr := fmt.Sprintf("AddTimer function err: %s", err)
			logger.Error(errstr)
			return errors.New(errstr)
		}
		return nil
	}()

	interval := t.unixts - UnixMilli()
	//当interval 小于一个刻度， 而且没有下一层时间轮时
	if interval < tw.interval && tw.nextTimeWheel == nil{
		if forceNext {
			tw.timerQueue[(tw.curIndex + 1) % tw.scales][tid] = t
		}else{
			tw.timerQueue[tw.curIndex][tid] = t
		}
		return nil
	}

	//当interval 小于一个刻度，但是有下一层时间轮时
	if interval <tw.interval{
		return tw.nextTimeWheel.AddTimer(tid, t)
	}

	//当interval 大于等于一个刻度时
	dn := interval/tw.interval
	tw.timerQueue[(tw.curIndex+int(dn))%tw.scales][tid] = t
	return nil
}


func (tw *TimeWheel)RemoveTimer (tid uint32){
	tw.Lock()
	defer tw.Unlock()

	for i:= 0; i < tw.scales; i++{
		if _, ok := tw.timerQueue[i][tid]; ok{
			delete(tw.timerQueue[i], tid)
			return
		}
	}

	//如果没找到，去下一层找
	if tw.nextTimeWheel != nil{
		tw.nextTimeWheel.RemoveTimer(tid)
	}
}


func (tw *TimeWheel) AddTimeWheel(next *TimeWheel){
	tw.nextTimeWheel = next
}
//非阻塞的方式让时间轮转起来
func (tw *TimeWheel)Run(){
	go tw.run()
}

func (tw *TimeWheel) run(){
	for{
		time.Sleep(time.Duration(tw.interval) * time.Millisecond)

		tw.Lock()
		curTimers := tw.timerQueue[tw.curIndex]
		tw.timerQueue[tw.curIndex] = make(map[uint32] *Timer, tw.maxCap)
		for k, v := range curTimers{
			tw.addTimer(k, v, true)
		}
		nextTimers := tw.timerQueue[(tw.curIndex+1)%tw.scales]
		tw.timerQueue[(tw.curIndex+1)%tw.scales] = make(map[uint32] *Timer, tw.maxCap)
		for k, v := range nextTimers{
			tw.addTimer(k, v, true)
		}
		tw.curIndex = (tw.curIndex+1)%tw.scales
		tw.Unlock()
	}
}

//获取Timer
func (tw *TimeWheel)GetTimerWithIn(duration time.Duration) map[uint32]*Timer{
	leaftw := tw
	for leaftw.nextTimeWheel != nil{
		leaftw = leaftw.nextTimeWheel
	}

	leaftw.Lock()
	defer leaftw.Unlock()
	timerList := make(map[uint32]*Timer)
	now := UnixMilli()
	for k, v := range leaftw.timerQueue[leaftw.curIndex]{
		if v.unixts - now < int64(duration/1e6){
			timerList[k] = v
			delete(leaftw.timerQueue[leaftw.curIndex], k)
		}
	}
	return timerList
}