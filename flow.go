/*
 @Version : 1.0
 @Author  : steven.wong
 @Email   : 'wangxk1991@gamil.com'
 @Time    : 2024/03/30 14:37:01
 Desc     :
*/

package flow

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/piaobeizu/flow/analytics"
	"github.com/piaobeizu/flow/phase"
	"github.com/sirupsen/logrus"
)

func handlepanic() {
	if err := recover(); err != nil {
		buf := make([]byte, 1<<16)
		ss := runtime.Stack(buf, true)
		msg := string(buf[:ss])
		var bt []string
		for _, row := range strings.Split(msg, "\n") {
			if !strings.HasPrefix(row, "\t") {
				continue
			}
			if strings.Contains(row, "main.") {
				continue
			}
			if strings.Contains(row, "panic") {
				continue
			}
			bt = append(bt, strings.TrimSpace(row))
		}

		analytics.Client.Publish("panic", map[string]interface{}{"backtrace": strings.Join(bt, "\n")})
		log.Fatalf("PANIC: %v\n", err)
	}
}

// func initManager(ctx *phase.Context) error {
// 	c, ok := ctx.Context.Value(ctxConfigKey{}).(*v1beta1.Apollo)
// 	if c == nil || !ok {
// 		return fmt.Errorf("cluster config not available in context")
// 	}

// 	manager, err := phase.NewManager(c)
// 	if err != nil {
// 		return fmt.Errorf("failed to initialize phase manager: %w", err)
// 	}

// 	manager.Concurrency = ctx.Int("concurrency")
// 	manager.ConcurrentUploads = ctx.Int("concurrent-uploads")
// 	manager.DryRun = ctx.Bool("dry-run")

// 	ctx.Context = context.WithValue(ctx.Context, ctxManagerKey{}, manager)

// 	return nil
// }

// func initAnalytics(ctx *phase.Context) error {
// 	client, err := NewClient(segmentWriteKey)
// 	if err != nil {
// 		return err
// 	}
// 	analytics.Client = client

// 	return nil
// }

// func closeAnalytics(_ *phase.Context) error {
// 	analytics.Client.Close()
// 	return nil
// }

type Flow struct {
	// Context is a type that is passed through to
	ctx     *phase.Context
	manager *phase.Manager
	segment bool
}

// NewFlow creates a new context. For use in when invoking an App or Command action.
func NewFlow(ctx *phase.Context, m *phase.Manager) *Flow {
	if ctx == nil {
		ctx = &phase.Context{}
	}
	if m == nil {
		m = &phase.Manager{}
	}
	return &Flow{ctx: ctx, manager: m, segment: false}
}
func (f *Flow) SetSegment(segment bool) *Flow {
	f.segment = segment
	return f
}

// Set sets a context flag to a value.
func (f *Flow) Run() error {
	defer handlepanic()
	analytics.InitAnalytics(f.segment)
	var (
		result error
		start  = time.Now()
	)
	// initLogging(f.debug)
	if result = f.manager.Run(); result != nil {
		analytics.Client.Publish("apply-failure", map[string]interface{}{"flowID": f.manager.Config.Metadata.Name})
		logrus.Info(phase.Colorize.Red("==> Apply failed").String())
		return result
	}
	// analytics.Client.Publish("apply-success",
	// 	map[string]any{
	// 		"duration": time.Since(start),
	// 		"flowID":   f.manager.Config.Metadata.Name,
	// 	},
	// )
	duration := time.Since(start).Truncate(time.Second)

	text := fmt.Sprintf("%s:%s is now finished in %ds.", f.manager.Config.Metadata.Name, f.manager.Config.Metadata.Version, duration)
	logrus.Infof(Colorize.Green(text).String())
	return nil
}
