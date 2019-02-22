package iface

type IMsgHandle interface{
	DeliverToMsgQueue(request IRequest)		//把消息发送至消息队列
	DoMsgFromGoRoutine(request IRequest)		//马上以非阻塞方式处理消息
	AddRouter(name string, router IRouter)	//为消息添加具体的处理逻辑
	StartWorker(poolSize int)				//开启worker，循环处理消息
}
