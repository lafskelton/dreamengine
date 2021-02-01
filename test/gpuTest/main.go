package main

import "C"
import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/lafskelton/dreamEngine/pkg/neural"
)

//
func generateTestLayer(neurons int, previousNeurons int) *neural.LayerDefinition {
	//Generate layers
	//
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	layerDef := new(neural.LayerDefinition)
	//
	//Structure
	layerDef.LayerNeurons = neurons
	layerDef.PrevLayerNeurons = previousNeurons
	//
	//
	//Biases
	layerDef.Biases = make([]float32, neurons, neurons)
	for i := 0; i < neurons; i++ {
		layerDef.Biases[i] = float32(r.NormFloat64()*0.1)
	}
	//

	//
	//Weights
	layerDef.Weights = make([]float32, neurons*previousNeurons, neurons*previousNeurons)
	for i := 0; i < neurons*previousNeurons; i++ {
		layerDef.Weights[i] = float32(r.NormFloat64()*0.1)
	}
	return layerDef
}

func runGPUTest1000() {
	//
	fmt.Println("- ---- GPU TEST ---- -")
	// fmt.Println("Using first available GPU!! ")
	// fmt.Println("")
	// //
	// neural.PrintGPUInfo()
	// //
	// fmt.Println("")
	// fmt.Println("- Loading test model -")
	// fmt.Println("")
	//
	// new test model
	model := new(neural.ModelObject)
	_ = model
	modelDef := new(neural.ModelDefinition)
	//
	modelLibraryID := 123456789
	numLayers := 200
	inputVenLen := 224*224
	N := 500
	// inputVenLen := 64
	// N := 64
	//
	//Initialize definition
	modelDef.Init(modelLibraryID, numLayers, inputVenLen)
	//
	// Generate layers
	//
	for i := 0; i < numLayers; i++ {
		//
		if i == 0 {
			err := modelDef.AppendLayer(generateTestLayer(N, inputVenLen))
			if err != nil {
				log.Fatal(err)
			}
			continue
		}
		//
		//	Increment layer size to ensure we never see a square model in testing
		//
		err := modelDef.AppendLayer(generateTestLayer(int(math.Max(float64(N-(i*2)), 500)), modelDef.Layers[i-1].LayerNeurons)) //int(65535/(i*2))
		if err != nil {
			log.Fatal(err)
		}
	}
	//
	//
	//Initialize ModelObject
	err := model.Init(modelDef)
	if err != nil {
		log.Fatal(err)
	}
	modelDef = nil;
	//
	// Generate inputs
	inputs := make([]float32, inputVenLen)
	for i := 0; i < inputVenLen; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		inputs[i] = float32(r.Float64())
	}
	//
	// model.FireModel(inputs)
	cstart := time.Now();
	runFor := time.Minute * 120
	totalCount := 0;
	totalAvg := []float32{}
	for {
		if time.Since(cstart) > runFor {
			break
		}
		intStart := time.Now()
		avgMs := []float32{}
		for{
			if time.Since(intStart) > time.Second/4{
				break;
			}
			var wg sync.WaitGroup
			//
			//launch 24 kernels per second
			for g := 0; g < 16; g++{
				wg.Add(1)
				go func(avg *[]float32, wg *sync.WaitGroup){
					defer wg.Done()
					op, _ := model.FireModel(inputs)
					go func(avg *[]float32, op2 float32){
						avgMs = append(avgMs, op2)
					}(&avgMs, op)
					return;
				}(&avgMs, &wg)
			}
			wg.Wait()
		}
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
			fmt.Print("\n\n\n\n\nDREAMENGINE V0.0.1\nCUDA kernel: 4d_flash_forward\n\nResNet-200 concurrency test...\n\nCUDA 11.2 \nDevice[0]: ", neural.DeviceName(),"\n-------------\n\nbatch size: 16 \n\n")
			ops := float64(len(avgMs))/time.Since(intStart).Seconds()
			fmt.Println(float32(ops), " infr/s")
			var avg float32 = 0;
			for i := 0; i < len(avgMs)-1; i++ {
				avg += avgMs[i];
			}
			avg = avg/float32(len(avgMs))
			totalCount += len(avgMs)
			totalAvg = append(totalAvg, avg)
			fmt.Print("\navg exec time: ", avg, "ms \n\n")
		
	}
	var tAvg float32 = 0;
	for i := 0; i < len(totalAvg)-1; i++ {
		tAvg += totalAvg[i];
	}
	tAvg = tAvg / float32(len(totalAvg))
	cmd := exec.Command("clear") //Linux example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Print("\n\n\n\n\nDREAMENGINE 0.0.1\n\n\nRESNET1000 CONCURRENCY TEST\n\n-------------\n\n")
	fmt.Print("TEST COMPLETED:\n\n\n")
	fmt.Print("run time: ", runFor.Seconds(), "s\n")
	fmt.Print(float32(totalCount)/float32(runFor.Seconds()), " ops/s avg\n")
	fmt.Print("total ops: ", totalCount, "\n")
	fmt.Print("avg op time: ", tAvg, "ms \n")

	
	// fmt.Println(model.FireModel(inputs));
	//
	model.StopAndDispose()
	//
	return;
}
func runGPUTest30() {
	//
	fmt.Println("- ---- GPU TEST ---- -")
	// fmt.Println("Using first available GPU!! ")
	// fmt.Println("")
	// //
	// neural.PrintGPUInfo()
	// //
	// fmt.Println("")
	// fmt.Println("- Loading test model -")
	// fmt.Println("")
	//
	// new test model
	model := new(neural.ModelObject)
	_ = model
	modelDef := new(neural.ModelDefinition)
	//
	modelLibraryID := 123456789
	numLayers := 32
	inputVenLen := 233*3
	N := 233*3
	// inputVenLen := 64
	// N := 64
	//
	//Initialize definition
	modelDef.Init(modelLibraryID, numLayers, inputVenLen)
	//
	// Generate layers
	//
	for i := 0; i < numLayers; i++ {
		//
		if i == 0 {
			err := modelDef.AppendLayer(generateTestLayer(N, inputVenLen))
			if err != nil {
				log.Fatal(err)
			}
			continue
		}
		//
		//	Increment layer size to ensure we never see a square model in testing
		//
		err := modelDef.AppendLayer(generateTestLayer(N-(i*4), modelDef.Layers[i-1].LayerNeurons)) //int(65535/(i*2))
		if err != nil {
			log.Fatal(err)
		}
	}
	//
	//
	//Initialize ModelObject
	err := model.Init(modelDef)
	if err != nil {
		log.Fatal(err)
	}
	modelDef = nil;
	//
	// Generate inputs
	inputs := make([]float32, inputVenLen)
	for i := 0; i < inputVenLen; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		inputs[i] = float32(r.Float64())
	}
	//
	// model.FireModel(inputs)
	cstart := time.Now();
	runFor := time.Minute * 120
	totalCount := 0;
	totalAvg := []float32{}
	for {
		if time.Since(cstart) > runFor {
			break
		}
		intStart := time.Now()
		avgMs := []float32{}
		for{
			if time.Since(intStart) > time.Second/2{
				break;
			}
			var wg sync.WaitGroup
			//
			//launch 24 kernels per second
			for g := 0; g < 4; g++{
				wg.Add(1)
				go func(avg []float32, wg *sync.WaitGroup){
					defer wg.Done()
					op, _ := model.FireModel(inputs)
					avgMs = append(avgMs, op)
					return;
				}(avgMs, &wg)
			}
			wg.Wait()
		}
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
			fmt.Print("\n\n\n\n\nDREAMENGINE 0.0.1\n\n\nRESNET30 CONCURRENCY TEST\n\n-------------\n\n")
			ops := float64(len(avgMs))/time.Since(intStart).Seconds()
			fmt.Println(float32(ops), " op/s")
			var avg float32 = 0;
			for i := 0; i < len(avgMs)-1; i++ {
				avg += avgMs[i];
			}
			avg = avg/float32(len(avgMs))
			totalCount += len(avgMs)
			totalAvg = append(totalAvg, avg)
			fmt.Print("\navg exec time: ", avg, "ms \n\n")
		
	}
	var tAvg float32 = 0;
	for i := 0; i < len(totalAvg)-1; i++ {
		tAvg += totalAvg[i];
	}
	tAvg = tAvg / float32(len(totalAvg))
	cmd := exec.Command("clear") //Linux example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Print("\n\n\n\n\nDREAMENGINE 0.0.1\n\n\nRESNET1000 CONCURRENCY TEST\n\n-------------\n\n")
	fmt.Print("TEST COMPLETED:\n\n\n")
	fmt.Print("run time: ", runFor.Seconds(), "s\n")
	fmt.Print(float32(totalCount)/float32(runFor.Seconds()), " ops/s avg\n")
	fmt.Print("total ops: ", totalCount, "\n")
	fmt.Print("avg op time: ", tAvg, "ms \n")
	//
	// fmt.Println(model.FireModel(inputs));
	//
	model.StopAndDispose()
	//
	return;
}

func main(){
	runGPUTest1000()
	// runGPUTest30()
}