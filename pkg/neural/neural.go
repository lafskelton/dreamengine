package neural

/*
void printCudaInfo(void);
#cgo LDFLAGS: -L./lib/ -lneural
*/
import "C"

//PrintGPUInfo ..
func PrintGPUInfo() {
	C.printCudaInfo()
}
