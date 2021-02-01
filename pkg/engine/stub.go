package engine

import (
	"container/list"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"

	pb "github.com/lafskelton/dreamEngine/pkg/engine/proto"
	"github.com/lafskelton/dreamEngine/pkg/neural"
	"google.golang.org/grpc"
	"gopkg.in/mgo.v2/bson"
)

//DreamEngineInstance ..
type DreamEngineInstance struct {
	pb.DreamdatastreamServer
	serverOK   bool
	instanceID string
	liveModels *LiveModelsList
	//*[]typeslib.LiveModel
}

func parseLNWBA(query string) *neural.ModelObject {

	return nil
}

type replyQ struct {
	//
	cond sync.Cond
	q    queue
}

//PushReplytoQueue and send to client
func (rQ *replyQ) pushReplytoQueue(data *pb.ServerToClient) {
	//
	rQ.q.Lock()
	defer rQ.q.Unlock()
	rQ.q.Push(data)
	rQ.cond.Signal()
	return
}

//Data ..
func (d *DreamEngineInstance) Data(stream pb.Dreamdatastream_DataServer) error {

	//Exiting
	exiting := false
	userID := uuid.NewString()
	_ = userID

	rQ := new(replyQ)
	rQ.cond.L = &rQ.q.Mutex
	rQ.q.list = list.New()
	//Init queue

	//GPU test

	//Ready healthcheck
	// Two go routines, one for receiving
	// one that watches (sync.cond) the Q for replies
	// and sends them - NON-BLOCKING

	//RECEIVER
	wait1 := make(chan struct{})
	go func(rQ *replyQ) {
		//
		for {
			//
			in, err := stream.Recv()
			if err == io.EOF {
				close(wait1)
				return
			}
			if err != nil {
				fmt.Println(err)
				close(wait1)
				return
			}
			//
			//Switch data from oneof field
			switch data := in.Data.(type) {
			//Perform task respective to received data
			case *pb.ClientToServer_Handshake:
				fmt.Println("Received handshake")

				//Reply
				rQ.pushReplytoQueue(&pb.ServerToClient{
					MsgID: uuid.NewString(),
				})
				continue
				//
			case *pb.ClientToServer_Task:
				//Switch task type... LoadModelFromLNWBA, ExecuteInputVector...
				taskManifest := data.Task
				switch task := taskManifest.Data.(type) {
				case *pb.TaskManifest_Load:
					//Load new model on GPU
					model := new(neural.ModelDefinition)
					modelDef := task.Load.BinaryModelDefinition
					err := bson.Unmarshal(modelDef, model)
					if err != nil {
						panic(err)
					}
					fmt.Println(model)
					//Queue reply
					rQ.pushReplytoQueue(&pb.ServerToClient{
						MsgID: uuid.NewString(),
						Data: &pb.ServerToClient_Task{
							Task: &pb.TaskReceipt{
								Data: &pb.TaskReceipt_Load{
									Load: &pb.LoadModelReceipt{},
								},
							},
						},
					})
					//
					//Send to GPU
					// neural.LoadModelToGPU()
					//
					continue
				case *pb.TaskManifest_Exec:
					//Exec data on loaded model
				}
				//Execute with cGO

				continue
				//
			case *pb.ClientToServer_Healthcheck:
				continue
			// case *pb.InstanceToAPIDataStream_Shutdown:
			// 	fmt.Println(data)
			// 	go handleShutdown(data)
			default:
				_ = data
				//nil
			}
		}
	}(rQ)

	//Responce routine
	wait2 := make(chan struct{})
	//
	go func(rQ *replyQ) {
		for {
			//
			rQ.q.Lock()
			//
			//Block this routine if nothing in queue
			if rQ.q.list.Len() == 0 {
				//
				//exiting is true and queue is empty
				if exiting {
					close(wait2)
					return
				}
				//
				rQ.cond.Wait()
			}
			//
			//Send reply!
			rply := rQ.q.Pop()
			stream.Send(rply)
			//
			rQ.q.Unlock()
		}
	}(rQ)

	//Wait for receiver to exit
	<-wait1 //Responder empties queue
	//Set connected instance to not OK
	exiting = true
	<-wait2
	fmt.Println("Exited!")
	//Reciever closed, ensure queue empty before exit
	fmt.Println("Bye")
	return nil

}

//RunDreamEngine ..
func RunDreamEngine() {
	//
	//

	//
	//Declare server
	server := DreamEngineInstance{}
	server.liveModels = new(LiveModelsList)
	server.liveModels.list = list.New()
	//
	//gRPC
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{grpc.ConnectionTimeout(time.Second * 3)}
	s := grpc.NewServer(opts...)
	pb.RegisterDreamdatastreamServer(s, &server)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	return

}
