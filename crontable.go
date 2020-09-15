package crontable

import (
	"sync"
	"time"

	"github.com/rs/xid"
	"golang.org/x/net/context"
)

func NewCron() *CronServer {
	cronTable := &CronServer{}
	cronTable.CronMap = make(map[string]*cronInfo)
	cronTable.myKeepFreshS_C.Channel = time.NewTimer(time.Second * 10000)
	cronTable.CronInChannel = make(chan string, 2)
	go cronTable.start()
	return cronTable
}

type CronServer struct {
	myKeepFreshS_C cronOutS_C
	CronMap        map[string]*cronInfo
	CronInChannel  chan string
	lock           sync.Mutex
}
type cronOutS_C struct {
	Channel *time.Timer
	Tid     []string
}

type cronInfo struct {
	longestFreshTime  time.Duration
	nextFreshTime     time.Time
	lastFreshTime     time.Time
	HowOftenKeepFresh time.Duration
	callbackInfo      *CallbackInfoStruct
}
type CallbackInfoStruct struct {
	CallbackFunc func(ctx context.Context, tid string) error
	Context      context.Context
	TaskId       string
}

func (this *CronServer) keepFreshButNotUpdateAll(vv *cronInfo) {
	now := time.Now()
	vv.lastFreshTime = now
	vv.nextFreshTime = now.Add(vv.HowOftenKeepFresh)
	go vv.callbackInfo.CallbackFunc(vv.callbackInfo.Context, vv.callbackInfo.TaskId)

}
func (this *CronServer) keepFresh(cronList []*cronInfo) {
	now := time.Now()
	for index := range cronList {
		cronList[index].lastFreshTime = now
		a := now.Add(cronList[index].HowOftenKeepFresh)
		cronList[index].nextFreshTime = a
	}
	this.updateNextFreshAndCronOutTime()
	for index := range cronList {
		go cronList[index].callbackInfo.CallbackFunc(cronList[index].callbackInfo.Context, cronList[index].callbackInfo.TaskId)
	}
}
func (this *CronServer) updateNextFreshAndCronOutTime() {
	this.lock.Lock()
	var myNextFreshDuration time.Duration = -1
	var myNextFreshTid []string = []string{}
	for tid, vv := range this.CronMap {
		if nil == vv {
			continue
		}
		tempNextFreshDuration := vv.nextFreshTime.Sub(time.Now())
		if tempNextFreshDuration <= 0 {
			this.keepFreshButNotUpdateAll(vv)
			tempNextFreshDuration = vv.HowOftenKeepFresh
		}
		if myNextFreshDuration < 0 {
			myNextFreshDuration = tempNextFreshDuration
		}
		if tempNextFreshDuration.Seconds() < myNextFreshDuration.Seconds() {
			myNextFreshDuration = tempNextFreshDuration
			myNextFreshTid = []string{tid}
		} else if tempNextFreshDuration.Seconds() == myNextFreshDuration.Seconds() {
			myNextFreshTid = append(myNextFreshTid, tid)
		}
	}
	this.lock.Unlock()
	if myNextFreshDuration >= 0 {
		this.myKeepFreshS_C.Tid = myNextFreshTid
		this.myKeepFreshS_C.Channel.Reset(myNextFreshDuration)
	} else {
		this.myKeepFreshS_C.Tid = []string{}
		this.myKeepFreshS_C.Channel.Reset(time.Second * 19999)
	}
}
func (this *CronServer) start() {
	for {
		select {
		case <-this.myKeepFreshS_C.Channel.C:
			existList := []*cronInfo{}
			this.lock.Lock()
			for index := range this.myKeepFreshS_C.Tid {
				v, ok := this.CronMap[this.myKeepFreshS_C.Tid[index]]
				if ok {
					existList = append(existList, v)
				}
			}
			this.lock.Unlock()
			this.keepFresh(existList)
		case <-this.CronInChannel:
			this.updateNextFreshAndCronOutTime()
		}
	}
}

// func (this *CronServer) Start() {
// 	go this.start()
// }
//从取消某个任务的定时执行
func (this *CronServer) CronOut(tid string) {
	this.lock.Lock()
	delete(this.CronMap, tid)
	this.lock.Unlock()
}

//从现在开始定时执行某个任务
func (this *CronServer) CronIn(howOftenKeepFresh time.Duration, callBack *CallbackInfoStruct) {
	now := time.Now()
	a := now.Add(howOftenKeepFresh)
	info := &cronInfo{
		nextFreshTime:     a,
		HowOftenKeepFresh: howOftenKeepFresh,
		callbackInfo:      callBack,
	}
	this.lock.Lock()
	this.CronMap[callBack.TaskId] = info
	this.lock.Unlock()
	this.CronInChannel <- callBack.TaskId
}

//过一阵子再开始定时执行某个任务
func (this *CronServer) CronInAfterWait(howOftenKeepFresh time.Duration, afterWait time.Duration, callBack *CallbackInfoStruct) {
	taskTemp := xid.New().String()
	outAfterWait := func(ctx context.Context, tid string) error {
		this.CronOut(tid)
		this.CronIn(howOftenKeepFresh, callBack)
		return nil
	}
	this.CronIn(afterWait, &CallbackInfoStruct{
		CallbackFunc: outAfterWait,
		Context:      context.Background(),
		TaskId:       taskTemp,
	})

}
