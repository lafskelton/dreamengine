package neural

/*
void printCudaInfo(void);
int getModelID(char* ptr);
char* newModel(int ModelLibraryID, int LayerNum);
void disposeModel(char* ptr);
void execute(int max_threads, char* ptr, float* inputs, float *out);
int get_max_threads(int device);
char* get_device_name(int device);
int appendLayer(char* modelPtr, const int N, const int prevN, float* weights, float* biases);
#cgo LDFLAGS: -L./lib/ -lmodeldef
*/
import "C"

import (
	"errors"
	"log"
	"sync"
	"time"
	"unsafe"
)

//ModelObject denotes the go lifecycle and structure of a model allocated and managed by C & CUDA
type ModelObject struct {
	sync.Mutex
	Live    bool
	ModelID int32
	// GPUID        int
	RunningSince time.Time
	CGOPtr       *C.char
	MemSize      int32
	Def          *ModelDefinition //Disposed of after load
}

//Init ..
func (m *ModelObject) Init(def *ModelDefinition) error {
	//Initialize model
	//
	// Live indicates if this definition is avaliable in the GPUs memory
	m.Live = false
	m.Def = def
	//
	m.ModelID = 123456 // This will need to be tracked in a top level state object
	//
	m.CGOPtr = C.newModel((C.int)(m.ModelID), (C.int)(len(m.Def.Layers)))
	//
	// Confirm model initialized
	modelID := C.getModelID(m.CGOPtr)
	if modelID != (C.int)(m.ModelID) {
		return errors.New("C failed to initialize the model correctly")
	}
	//
	//	Validate ModelDefinition
	if m.Def.NumLayers != len(m.Def.Layers) {
		return errors.New("Invalid definition")
	}
	//
	//Alloc layers in C enviroment
	for _, layer := range m.Def.Layers {
		//
		// Create C arrays
		weights := make([]C.float, len(layer.Weights), len(layer.Weights))
		biases := make([]C.float, len(layer.Biases), len(layer.Biases))
		//
		// Copy values
		for i, w := range layer.Weights {
			weights[i] = C.float(w)
		}
		for i, b := range layer.Biases {
			biases[i] = C.float(b)
		}
		//
		//
		code := C.appendLayer(
			//ModelPtr,
			m.CGOPtr,
			//n Neurons
			C.int(layer.LayerNeurons),
			//n Previous Neurons
			C.int(layer.PrevLayerNeurons),
			//Weights
			(*C.float)(unsafe.Pointer(&weights[0])),
			//Biases
			(*C.float)(unsafe.Pointer(&biases[0])),
		)
		if code != 0 {
			log.Fatal("Append failed")
		}
	}
	//
	return nil

}

//StopAndDispose ..
func (m *ModelObject) StopAndDispose() error {
	C.disposeModel(m.CGOPtr)
	return nil
}

//FireModel ..
func (m *ModelObject) FireModel(inputs []float32) (float32, []float32) {
	//
	//
	// serialize inputs
	in := make([]C.float, m.Def.Layers[0].PrevLayerNeurons)
	for i, val := range inputs {
		in[i] = C.float(val)
	}
	//
	//
	out := make([]C.float, m.Def.Layers[len(m.Def.Layers)-1].LayerNeurons)
	//
	// maxT := maxThreads()
	start := time.Now()
	//
	C.execute(1024, m.CGOPtr, (*C.float)(unsafe.Pointer(&inputs[0])), (*C.float)(unsafe.Pointer(&out[0])))
	//
	elapsed := time.Since(start);
	op := float32(elapsed.Milliseconds())
	// fmt.Print("\n- --- TEST EXIT ---- -\n\n op: ", op," ms\n\n")
	outputs := make([]float32, m.Def.Layers[len(m.Def.Layers)-1].LayerNeurons)
	//
	for i, val := range out {
		outputs[i] = float32(val)
	}
	//
	return op, outputs
}

//ModelDefinition ..
type ModelDefinition struct {
	ModelLibraryID int
	NumLayers      int                //Number of layers
	InputVecLen    int                //Input vector size
	Layers         []*LayerDefinition //Ordered list of layers
}

//LayerDefinition ..
type LayerDefinition struct {
	LayerNeurons     int
	PrevLayerNeurons int
	Weights          []float32
	Biases           []float32
	Activation       string
}

//Init new model def
func (m *ModelDefinition) Init(modelLibraryID int, numLayers int, inputVenLen int) error {
	//Init values
	m.ModelLibraryID = modelLibraryID
	m.NumLayers = numLayers
	m.InputVecLen = inputVenLen
	m.Layers = make([]*LayerDefinition, numLayers, numLayers)
	//
	return nil
}

//AppendLayer ...append to last layer, inputs must match previous outputs
func (m *ModelDefinition) AppendLayer(layerDef *LayerDefinition) error {
	for i := range m.Layers {
		if m.Layers[i] == nil {
			if i > 0 {
				if layerDef.PrevLayerNeurons == m.Layers[i-1].LayerNeurons {
					m.Layers[i] = layerDef
				} else {
					return errors.New("invalid layer")
				}
			} else {
				m.Layers[i] = layerDef
			}
			break
		}
	}
	return nil
}

//MaxThreads of GPU 0 
func maxThreads() C.int {	
	return C.get_max_threads(C.int(0))
}
//DeviceName of GPU 0 
func DeviceName() string {	
	str := C.get_device_name(C.int(0))
	return C.GoString(str);
}