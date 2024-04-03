package phase

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/piaobeizu/flow/analytics"
	"github.com/piaobeizu/flow/apis"
)

// GenericPhase is a basic phase which gets a config via prepare, sets it into p.Config
type GenericPhase struct {
	analytics.Phase
	ctx     *Context
	manager *Manager
}

// GetConfig is an accessor to phase Config
func (p *GenericPhase) GetConfig() *apis.FlowConfig {
	return p.manager.Config
}

// Prepare the phase
func (p *GenericPhase) Prepare(c *apis.FlowConfig) error {
	p.manager.Config = c
	return nil
}

// Wet is a shorthand for manager.Wet
func (p *GenericPhase) Wet(msg string, funcs ...errorfunc) error {
	return p.manager.Wet(msg, funcs...)
}

// IsWet returns true if manager is in dry-run mode
func (p *GenericPhase) IsWet() bool {
	return !p.manager.DryRun
}

// DryMsg is a shorthand for manager.DryMsg
func (p *GenericPhase) DryMsg(host fmt.Stringer, msg string) {
	p.manager.DryMsg(msg)
}

// DryMsgf is a shorthand for manager.DryMsg + fmt.Sprintf
func (p *GenericPhase) DryMsgf(host fmt.Stringer, msg string, args ...any) {
	p.manager.DryMsg(fmt.Sprintf(msg, args...))
}

// SetManager adds a reference to the phase manager
func (p *GenericPhase) SetManager(m *Manager) {
	p.manager = m
}

func (p *GenericPhase) parallelDo(funcs ...func(c *apis.FlowConfig) error) error {
	// if p.manager.Concurrency == 0 {
	// 	return p.ParallelEach(funcs...)
	// }
	// return hosts.BatchedParallelEach(p.manager.Concurrency, funcs...)
	return p.ParallelEach(funcs...)
}

func (p *GenericPhase) parallelDoUpload(funcs ...func(c *apis.FlowConfig) error) error {
	// if p.manager.Concurrency == 0 {
	// 	return hosts.ParallelEach(funcs...)
	// }
	// return hosts.BatchedParallelEach(p.manager.ConcurrentUploads, funcs...)

	return p.ParallelEach(funcs...)
}

// ParallelEach runs a function (or multiple functions chained) on every Host parallelly.
// Any errors will be concatenated and returned.
func (p *GenericPhase) ParallelEach(filters ...func(c *apis.FlowConfig) error) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []string

	for _, filter := range filters {
		wg.Add(1)
		go func(c *apis.FlowConfig) {
			defer wg.Done()
			if err := filter(c); err != nil {
				mu.Lock()
				fn := runtime.FuncForPC(reflect.ValueOf(filter).Pointer()).Name()
				errors = append(errors, fmt.Sprintf("%s: %s", fn, err.Error()))
				mu.Unlock()
			}
		}(p.manager.Config)

	}
	wg.Wait()
	if len(errors) > 0 {
		return fmt.Errorf("failed on %d hosts:\n - %s", len(errors), strings.Join(errors, "\n - "))
	}

	return nil
}

// BatchedParallelEach runs a function (or multiple functions chained) on every Host parallelly in groups of batchSize hosts.
// func (p *GenericPhase) BatchedParallelEach(batchSize int, filter ...func(c *apis.FlowConfig) error) error {
// 	for i := 0; i < len(hosts); i += batchSize {
// 		end := i + batchSize
// 		if end > len(hosts) {
// 			end = len(hosts)
// 		}
// 		if err := hosts[i:end].ParallelEach(filter...); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
