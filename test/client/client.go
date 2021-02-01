package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"time"

	pb "github.com/lafskelton/dreamEngine/pkg/engine/proto"
	"github.com/lafskelton/dreamEngine/pkg/typeslib"
	"google.golang.org/grpc"
	"gopkg.in/mgo.v2/bson"
)

//GenerateEncodedModel ..
func GenerateEncodedModel() []byte {
	model := typeslib.ModelDefinition{}
	// 3 Layers, 3 Neurons per layer

	//
	//Generate test model
	//
	//Params
	model.ModelLibraryID = "43738095743985"
	model.NumLayers = 10
	model.InputVecLen = 3 //Input vector length
	model.Layers = make([]typeslib.LayerDefinition, model.NumLayers)
	//
	for i, layerDef := range model.Layers {
		//
		rand.Seed(time.Now().Unix())
		//Index
		layerDef.LayerIndex = i
		//Neurons
		layerDef.LayerNeurons = rand.Intn(4) + 1
		//Activation
		layerDef.Activation = "ReLU"
		//
		//Inputs to this layer equals previous outputs
		if i != 0 {
			layerDef.PrevLayerNeurons = model.Layers[i-1].LayerNeurons
		} else {
			layerDef.PrevLayerNeurons = model.InputVecLen
		}
		//
		//Alloc slices
		layerDef.Weights = make([]float32, layerDef.PrevLayerNeurons*layerDef.LayerNeurons) //Matrix of prevN*N
		layerDef.Biases = make([]float32, layerDef.LayerNeurons)                            //Bias for each neuron
		//
		//Populate slices with random data
		//
		//Weights
		rand.Seed(time.Now().Unix())
		for i := range layerDef.Weights {
			layerDef.Weights[i] = rand.Float32()
		}
		// Biases
		rand.Seed(time.Now().Unix())
		for i := range layerDef.Biases {
			layerDef.Biases[i] = rand.Float32()
		}
		model.Layers[i] = layerDef
	}
	fmt.Println(model)
	data, err := bson.Marshal(&model)
	if err != nil {
		panic(err)
	}
	return data
}

func main() {

	flag.Parse()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(":8080", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewDreamdatastreamClient(conn)

	stream, err := client.Data(context.Background())

	stream.Send(&pb.ClientToServer{
		Data: &pb.ClientToServer_Handshake{
			Handshake: &pb.HandshakeManifest{},
		},
	})

	stream.Send(&pb.ClientToServer{
		Data: &pb.ClientToServer_Task{
			Task: &pb.TaskManifest{
				Data: &pb.TaskManifest_Load{
					Load: &pb.LoadModelManifest{
						BinaryModelDefinition: GenerateEncodedModel(),
					},
				},
			},
		},
	})

	stream.Send(&pb.ClientToServer{
		Data: &pb.ClientToServer_Task{
			Task: &pb.TaskManifest{
				Data: &pb.TaskManifest_Exec{
					Exec: &pb.ExecuteDataManifest{},
				},
			},
		},
	})

	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			fmt.Println(in)

		}
	}()

	stream.CloseSend()
	<-waitc
}
