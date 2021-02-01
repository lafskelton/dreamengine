package engine

import (
	"container/list"
	"sync"

	pb "github.com/lafskelton/dreamEngine/pkg/engine/proto"
)

//queue ..
type queue struct {
	//Sync
	sync.Mutex

	//Queue
	list *list.List
}

//PriorityPush task to front of queue
func (q *queue) PriorityPush(task *pb.ServerToClient) {
	q.list.PushFront(task)
	return
}

//Push task to queue
func (q *queue) Push(task *pb.ServerToClient) {
	q.list.PushBack(task)
	return
}

//Pop Task from queue
func (q *queue) Pop() *pb.ServerToClient {
	//
	//get first item from list, nil if type assertion fails
	//could switch errors here
	elem := q.list.Front()
	if task, ok := elem.Value.(*pb.ServerToClient); ok {
		defer q.list.Remove(elem)
		return task
	}
	q.list.Remove(elem)

	return nil
}
