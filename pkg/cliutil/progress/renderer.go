package progress

import (
	"time"

	"go.uber.org/atomic"
)

type renderer struct {
	isUpdaterRunning atomic.Bool
	stopChan         chan struct{}
	stopFinishedChan chan struct{}
	renderFn         func()
}

func newRenderer() *renderer {
	return &renderer{
		isUpdaterRunning: atomic.Bool{},
		stopChan:         nil,
		stopFinishedChan: nil,
		renderFn:         nil,
	}
}

func (r *renderer) startRenderLoop() {
	if r.renderFn == nil {
		panic("renderFn must be set")
	}
	if !r.isUpdaterRunning.CAS(false, true) {
		return
	}
	r.stopChan = make(chan struct{})
	r.stopFinishedChan = make(chan struct{})
	go r.renderLoopFn()
}

func (r *renderer) stopRenderLoop() {
	if !r.isUpdaterRunning.CAS(true, false) {
		return
	}
	r.stopChan <- struct{}{}
	close(r.stopChan)
	r.stopChan = nil

	<-r.stopFinishedChan
	close(r.stopFinishedChan)
	r.stopFinishedChan = nil
}

func (r *renderer) renderLoopFn() {
	ticker := time.NewTicker(refreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.renderFn()
		case <-r.stopChan:
			r.renderFn()
			r.stopFinishedChan <- struct{}{}
			return
		}
	}
}
