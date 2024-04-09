package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

type Task struct {
	TaskType  TaskType
	TaskNum   int
	ReduceSum int // sum of tasks needed to reduce
	MapSum    int // sum of tasks needed to map
	Filename  string
}

type TaskArgs struct {
	TaskType TaskType
	TaskNum  int
}

type FinishArgs struct {
	Ts Task
}

type TaskType int

type Phase int

type State int

const (
	MapTask TaskType = iota
	ReduceTask
	CompleteTask
)

const (
	MapPhase Phase = iota
	ReducePhase
	AllDone
)

const (
	Working State = iota
	Waiting
	Done
)

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
