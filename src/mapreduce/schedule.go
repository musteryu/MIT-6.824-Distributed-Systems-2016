package mapreduce

import (
	"fmt"
	"sync/atomic"
)

// schedule starts and waits for all tasks in the given phase (Map or Reduce).
func (mr *Master) schedule(phase jobPhase) {
	var ntasks int
	var nios int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mr.files)
		nios = mr.nReduce
	case reducePhase:
		ntasks = mr.nReduce
		nios = len(mr.files)
	}

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, nios)

	// All ntasks tasks have to be scheduled on workers, and only once all of
	// them have been completed successfully should the function return.
	// Remember that workers may fail, and that any given worker may finish
	// multiple tasks.
	//
	// TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO
	//
	taskChan := make(chan DoTaskArgs)
	var count int32 = 0
	doChan := make(chan bool)
	go func() {
		for doTaskArgs := range taskChan {
			worker := <- mr.registerChannel
			go func(worker string, doTaskArgs DoTaskArgs) {
				if call(worker, "Worker.DoTask", doTaskArgs, new(struct{})) {
					if atomic.AddInt32(&count, int32(1)) == int32(ntasks) {
						doChan <- true
					}
				} else {
					go func() {
						taskChan <- doTaskArgs
					}()
				}
				//go func() {
					mr.registerChannel <- worker
				//}()
			}(worker, doTaskArgs)
		}
	}()

	for i := 0; i < ntasks; i++ {
		doTaskArgs := DoTaskArgs{
			JobName:       mr.jobName,
			Phase:         phase,
			TaskNumber:    i,
			NumOtherPhase: nios}
		if phase == mapPhase {
			doTaskArgs.File = mr.files[i]
		}
		taskChan <- doTaskArgs
	}
	fmt.Printf("Schedule: %v phase done\n", phase)
	<- doChan
	close(taskChan)
}
