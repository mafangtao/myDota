package timer

import (
	"fmt"
	"reflect"
	"myDota/logger"
	"time"
)

func UnixMilli() int64 {
	return time.Now().UnixNano()/1e6
}

/**************************************************
*延迟调用的函数类型
***************************************************/
type DelayFunc struct{
	f func( ...interface{})
	args []interface{}
}

func NewDelayFunc(f func(v ...interface{}), args []interface{}) *DelayFunc{
	return &DelayFunc{
		f : f,
		args : args,
	}
}

func (df *DelayFunc) String() string{
	return fmt.Sprintf("{DelayFunc:%s, args:%v}", reflect.TypeOf(df.f).Name(), df.args)
}

func (df *DelayFunc) Call(){
	defer func(){
		if err := recover(); err != nil{
			logger.Error(df.String(), "Call err:", err)
		}
	}()

	df.f(df.args...)
}

/**********************************************
*定时器的实现
**********************************************/
type Timer struct{
	//延迟调用函数
	delayFunc *DelayFunc
	//调用时间（unix时间，单位ms）
	unixts int64
}

func NewTimerAt (f func(v ...interface{}), args []interface{}, unixNano int64) *Timer{
	return &Timer{
		delayFunc : NewDelayFunc(f, args),
		unixts:unixNano/1e6,
	}
}

func NewTimerAfeter(f func(v ...interface{}), args []interface{}, duration time.Duration) *Timer{
	return NewTimerAt(f, args, time.Now().UnixNano() + int64(duration))
}

//非阻塞调用timer的Run方法
func (t *Timer)Run(){
	go func(){
		now := UnixMilli()
		if (t.unixts > now){
			time.Sleep(time.Duration(t.unixts - now) * time.Millisecond)
		}

		t.delayFunc.Call()
	}()
}