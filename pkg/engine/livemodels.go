package engine

import (
	"container/list"
	"sync"
)

//LiveModelsList //
type LiveModelsList struct {
	sync.Mutex
	list *list.List
}

//Add model to list
func (l *LiveModelsList) Add() {

}

//Remove model from list
func (l *LiveModelsList) Remove() {

}
