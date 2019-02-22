package fnet

import (
	"myDota/iface"
	"myDota/utils"
	"strconv"
	"myDota/logger"
	"fmt"
	"runtime/debug"
	"time"
)

type MsgHandle struct{
	PoolSize int32
	TaskQueue []chan iface.IRequest
	Apis map[uint32]iface.IRouter
}

func NewMsgHandle() *MsgHandle{
	return &MsgHandle{
		PoolSize:utils.GlobalObject.PoolSize,
		TaskQueue:make([]chan iface.IRequest, utils.GlobalObject.PoolSize),
		Apis:make(map[uint32]iface.IRouter),
	}
}

//一致性路由，把统一连接的数据转发给同一个goroutine
func (mh *MsgHandle) DeliverToMsgQueue(request iface.IRequest){
	index := request.GetConnection().GetSessionId()% uint32(mh.PoolSize)
	logger.Debug(fmt.Sprintf("add to pool: %d", index))
	mh.TaskQueue[index] <- request
}

func (mh *MsgHandle)DoMsgFromGoRoutine(request iface.IRequest){
	logger.Debug("do from goroutine:", request.GetData())
	go mh.doApi(request)
}

func (mh *MsgHandle )AddRouter(name string, router iface.IRouter){
	index, err := strconv.Atoi(name)
	if err != nil{
		panic("error api:" + name)
	}

	//api已经存在时
	if _, ok := mh.Apis[uint32(index)]; ok{
		panic("repeated api" + name)
	}

	mh.Apis[uint32(index)] = router
	logger.Info("add api" + name)
}
func (mh *MsgHandle)StartWorker(poolSize int){
	for i:= 0; i < int(mh.PoolSize); i++{
		mh.TaskQueue[i] = make(chan iface.IRequest, utils.GlobalObject.MaxWorkerLen)
		go func(index int, taskQueue chan iface.IRequest){
			logger.Info("init goroutine %d.", index)
			delayTaskCh := utils.GlobalObject.GetTimerSchedule().GetTriggerChan()
			for{
				select {
				case request := <- taskQueue:
					mh.doApi(request)
				case delayTask := <- delayTaskCh:
					delayTask.Call()
				}
			}
		}(i, mh.TaskQueue[i])
	}
}

func (mh *MsgHandle)doApi(request iface.IRequest){
	defer func(){
		if err := recover(); err != nil{
			logger.Info("============do api panic recover============")
			logger.Error(err)
			debug.PrintStack()
		}
	}()

	f, ok := mh.Apis[request.GetMsgId()]
	if !ok{
		logger.Error(fmt.Sprintf("api: %d not found", request.GetMsgId()))
	}

	ts := time.Now()
	f.PreHandle(request)
	f.Handle(request)
	f.PostHandle(request)
	logger.Debug(fmt.Sprintf("Api_%d cost total time: %f ms", request.GetMsgId(), float64(time.Since(ts).Nanoseconds())/1e6))
}