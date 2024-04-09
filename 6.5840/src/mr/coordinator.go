package mr

import "log"
import "net"
import "sync"
import "os"
import "fmt"
import "time"
import "net/rpc"
import "net/http"

type Coordinator struct {
	mu                       sync.Mutex
	nReduce                  int // how many tasks need to done
	files                    []string
	finishedMapfiles         []bool
	TimesfinishedMapfiles    []time.Time
	finishedReducefiles      []bool
	TimesfinishedReducefiles []time.Time
	Isdone                   bool
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (c *Coordinator) GetTask(args *TaskArgs, reply *Task) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for {
		var brk bool
		for i, v := range c.finishedMapfiles {
			if !v {
				if c.TimesfinishedMapfiles[i].IsZero() || time.Since(c.TimesfinishedMapfiles[i]).Seconds() > 10 {
					reply.TaskType = MapTask
					reply.TaskNum = i
					reply.ReduceSum = c.nReduce
					reply.Filename = c.files[i]
					c.TimesfinishedMapfiles[i] = time.Now()
					fmt.Printf("GetTask:%d %d %v %v\n", reply.TaskNum, reply.TaskType, time.Now(), v)
					return nil
				}
			} else {
				brk = true
			}
		}

		if brk && c.CheckIsFinish(MapTask) {
			break
		}
	}

	for {
		var brk bool
		for i, v := range c.finishedReducefiles {
			if !v {
				if c.TimesfinishedReducefiles[i].IsZero() || time.Since(c.TimesfinishedReducefiles[i]).Seconds() > 10 {
					reply.TaskType = ReduceTask
					reply.TaskNum = i
					reply.ReduceSum = c.nReduce
					reply.MapSum = len(c.files)
					c.TimesfinishedReducefiles[i] = time.Now()
					return nil
				}
			} else {
				brk = true
			}
		}
		if brk && c.CheckIsFinish(ReduceTask) {
			break
		}
	}
	reply.TaskType = CompleteTask
	c.Isdone = true

	return nil
}

func (c *Coordinator) CheckIsFinish(tp TaskType) bool {

	switch tp {

	case MapTask:
		count := 0
		for _, v := range c.finishedMapfiles {
			if v {
				count++
			}
		}
		if count == len(c.finishedMapfiles) {
			return true
		}
		return false

	case ReduceTask:
		count := 0
		for _, v := range c.finishedReducefiles {
			if v {
				count++
			}
		}
		if count == len(c.finishedReducefiles) {
			return true
		}
		return false

	}

	return false

}

func (c *Coordinator) FinishedTask(args *FinishArgs, reply *Task) error {
	switch args.Ts.TaskType {
	case MapTask:
		c.finishedMapfiles[args.Ts.TaskNum] = true
		fmt.Println(args.Ts.TaskNum)

	case ReduceTask:
		c.finishedReducefiles[args.Ts.TaskNum] = true
	}
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()

	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	ret := false
	// Your code here.
	c.mu.Lock()
	ret = c.Isdone
	c.mu.Unlock()
	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	c.nReduce = nReduce
	c.files = files
	c.finishedMapfiles = make([]bool, len(files))
	c.TimesfinishedMapfiles = make([]time.Time, len(files))

	c.finishedReducefiles = make([]bool, nReduce)
	c.TimesfinishedReducefiles = make([]time.Time, nReduce)

	c.server()
	return &c
}
