package crontable

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestCron(t *testing.T) {
	cronTable := NewCron()
	cronTable.CronInAfterWait(time.Second*1, time.Second*7, &CallbackInfoStruct{
		CallbackFunc: testCallBack,
		TaskId:       "4",
	})
	cronTable.CronIn(time.Second*1, &CallbackInfoStruct{
		CallbackFunc: testCallBack,
		TaskId:       "1",
	})
	cronTable.CronIn(time.Second*1, &CallbackInfoStruct{
		CallbackFunc: testCallBack,
		TaskId:       "2",
	})

	cronTable.CronIn(time.Second*1, &CallbackInfoStruct{
		CallbackFunc: testCallBack,
		TaskId:       "3",
	})
	time.Sleep(time.Second * 10)
	cronTable.CronOut("2")
	time.Sleep(time.Second * 100)
}
func testCallBack(context context.Context, tid string) error {
	fmt.Println(time.Now().Format(time.Stamp), "do:", tid)
	return nil
}
