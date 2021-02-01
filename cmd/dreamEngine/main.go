package main

import (
	"github.com/lafskelton/dreamEngine/pkg/engine"
	"github.com/lafskelton/dreamEngine/pkg/neural"
	"github.com/lafskelton/dreamEngine/pkg/typeslib"
)

//DreamEngine ..
type DreamEngine struct {
	engine       *engine.DreamEngineInstance
	loadedModels []*typeslib.LiveModel
}

//AddLiveModel ..
func (d *DreamEngine) AddLiveModel() {

}

//

func main() {
	//
	//GPU Test will panic if failed
	neural.GPUTest()
	//
	// engine.RunDreamEngine()
}
