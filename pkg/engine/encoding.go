package engine

// import (
// 	"fmt"
// 	"strconv"
// 	"sync"

// 	"github.com/lafskelton/dreamEngine/pkg/typeslib"
// )

// //EncodeModelFromString ./
// func EncodeModelFromString(in string, model *typeslib.ModelDefinition) (*[]byte, error) {
// 	//
// 	//Parse layer count
// 	layerNum := ""
// 	for i := 0; i < len(in); i++ {
// 		if string(in[i]) == "{" {
// 			break
// 		}
// 		layerNum += string(in[i])
// 	}
// 	//
// 	layerN, err := strconv.Atoi(layerNum)
// 	if err != nil {
// 		return nil, err
// 	}
// 	fmt.Println(layerN, "layers ")
// 	//
// 	//
// 	//Set layer count
// 	model.LayerN = layerN
// 	model.Layers = make([]typeslib.LayerDefinition, layerN)
// 	//
// 	//Parse Layers
// 	parsedLayers := make([]string, layerN)
// 	parsedLayersCurrent := 0
// 	//
// 	for i := len(layerNum); i <= len(in)-1; i++ {
// 		char := string(in[i])
// 		//Record char
// 		if char != "" {
// 			parsedLayers[parsedLayersCurrent] += char
// 		}
// 		//Jump to next layer
// 		if char == "}" {
// 			parsedLayersCurrent++
// 		}
// 		//
// 		//If the layer count doesn't match the definitions, panic condition solution
// 		if parsedLayersCurrent >= len(parsedLayers) {
// 			break
// 		}
// 	}
// 	//
// 	//	Parse layer strings concurrently
// 	//
// 	var wg sync.WaitGroup
// 	//
// 	for i, layerStr := range parsedLayers {
// 		model.Layers[i] = typeslib.LayerDefinition{}
// 		//
// 		// parse weights
// 		//
// 		wg.Add(1)
// 		go func(layerDef *typeslib.LayerDefinition, str string, wg *sync.WaitGroup) {
// 			defer wg.Done()
// 			//
// 			//
// 			for _, char := range str {
// 				if string(char) == "(" {
// 					//Start of weights
// 				}
// 			}
// 			//
// 		}(&model.Layers[i], layerStr, &wg)
// 		//
// 		// parse biases
// 		//
// 		wg.Add(1)
// 		go func(layerDef *typeslib.LayerDefinition, str string, wg *sync.WaitGroup) {
// 			defer wg.Done()
// 			//
// 			//
// 			bIndex := 0
// 			for j := range str {
// 				if j > 2 {
// 					if string(str[j-2]) == ")" && string(str[j-1]) == "(" {
// 						//Start of biases i:
// 						bIndex = j
// 						break
// 					}
// 				}
// 			}
// 			fmt.Println(str)
// 			endstring := str[bIndex:]
// 			tmp := ""
// 			biases := []float32{}
// 			for j := range endstring {
// 				if string(endstring[j]) == "," {
// 					//Next case
// 					parsedFloat, _ := strconv.ParseFloat(tmp, 64)
// 					tmp = ""
// 					biases = append(biases, float32(parsedFloat))
// 					continue
// 				}
// 				//Record case
// 				tmp += string(endstring[j])
// 			}
// 			fmt.Println(biases)
// 			//
// 		}(&model.Layers[i], layerStr, &wg)
// 		//
// 		wg.Wait()
// 	}
// 	// fmt.Println(model.Layers)

// 	return nil, nil
// }

// func decodeModel(in *[]byte) (*typeslib.ModelDefinition, error) {

// 	return nil, nil
// }
