## crontable

This is a simple library to handle scheduled tasks. Tasks can be run in a minimum delay of once a second--for which Cron isn't actually designed. Comparisons are fast and efficient and take place in a goroutine; matched jobs are also executed in goroutines.

## how to use

### new crontable
```golang
    cronTable := NewCron()
```

### put in a task and start now
```golang
	cronTable.CronIn(time.Second*1, &CallbackInfoStruct{
		CallbackFunc: testCallBack,
		TaskId:       "taskId",
	})
```
### stop a task 
```golang
	cronTable.CronOut("taskId")
```

###  put in a task and start after wait
```golang
	cronTable.CronInAfterWait(time.Second*1, time.Second*7, &CallbackInfoStruct{
		CallbackFunc: testCallBack,
		TaskId:       "taskId",
	})
```