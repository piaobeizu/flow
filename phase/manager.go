package phase

import (
	"fmt"
	"sync"

	"github.com/logrusorgru/aurora"
	"github.com/piaobeizu/flow/apis"
	log "github.com/sirupsen/logrus"
)

// NoWait is used by various phases to decide if node ready state should be waited for or not
var NoWait bool

// Force is used by various phases to attempt a forced installation
var Force bool

// Colorize is an instance of "aurora", used to colorize the output
var Colorize = aurora.NewAurora(false)

type phase interface {
	Run() error
	Title() string
}

type withconfig interface {
	Title() string
	Prepare(*apis.FlowConfig) error
}

type conditional interface {
	ShouldRun() bool
}

// beforehook receives the phase title as an argument because of reasons.
type beforehook interface {
	Before(string) error
}

type afterhook interface {
	After(error) error
}

type propsetter interface {
	SetProp(string, interface{})
}

type withcleanup interface {
	CleanUp()
}

type withmanager interface {
	SetManager(*Manager)
}

type withDryRun interface {
	DryRun() error
}

// Manager executes phases to construct the cluster
type Manager struct {
	phases            []phase
	Config            *apis.FlowConfig
	Concurrency       int
	ConcurrentUploads int
	DryRun            bool

	dryMessages map[string][]string
	dryMu       sync.Mutex
}

// NewManager creates a new Manager
func NewManager(config *apis.FlowConfig) (*Manager, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	return &Manager{Config: config}, nil
}

// AddPhase adds a Phase to Manager
func (m *Manager) AddPhase(p ...phase) {
	m.phases = append(m.phases, p...)
}

type errorfunc func() error

// DryMsg prints a message in dry-run mode
func (m *Manager) DryMsg(msg string) {
	m.dryMu.Lock()
	defer m.dryMu.Unlock()
	if m.dryMessages == nil {
		m.dryMessages = make(map[string][]string)
	}
	var key string
	// if host == nil {
	key = "local"
	// } else {
	// key = host.String()
	// }
	m.dryMessages[key] = append(m.dryMessages[key], msg)
}

// Wet runs the first given function when not in dry-run mode. The second function will be
// run when in dry-mode and the message will be displayed. Any error returned from the
// functions will be returned and will halt the operation.
func (m *Manager) Wet(msg string, funcs ...errorfunc) error {
	if !m.DryRun {
		if len(funcs) > 0 && funcs[0] != nil {
			return funcs[0]()
		}
		return nil
	}

	m.DryMsg(msg)

	if m.DryRun && len(funcs) == 2 && funcs[1] != nil {
		return funcs[1]()
	}

	return nil
}

// Run executes all the added Phases in order
func (m *Manager) Run() error {
	var ran []phase
	var result error

	defer func() {
		if m.DryRun {
			if len(m.dryMessages) == 0 {
				fmt.Println(Colorize.BrightGreen("dry-run: no cluster state altering actions would be performed"))
				return
			}

			fmt.Println(Colorize.BrightRed("dry-run: cluster state altering actions would be performed:"))
			for host, msgs := range m.dryMessages {
				fmt.Println(Colorize.BrightRed("dry-run:"), Colorize.Bold(fmt.Sprintf("* %s :", host)))
				for _, msg := range msgs {
					fmt.Println(Colorize.BrightRed("dry-run:"), Colorize.Red(" -"), msg)
				}
			}
			return
		}
		if result != nil {
			for _, p := range ran {
				if c, ok := p.(withcleanup); ok {
					log.Infof(Colorize.Red("* Running clean-up for phase: %s").String(), p.Title())
					c.CleanUp()
				}
			}
		}
	}()

	for _, p := range m.phases {
		title := p.Title()

		text := Colorize.Green("Starting phase: %s").String()
		log.Infof(text, title)
		if p, ok := p.(withmanager); ok {
			p.SetManager(m)
		}

		if p, ok := p.(withconfig); ok {
			log.Debugf("Preparing phase '%s'", p.Title())
			if err := p.Prepare(m.Config); err != nil {
				return err
			}
		}

		if p, ok := p.(conditional); ok {
			if !p.ShouldRun() {
				text = Colorize.Yellow("Skipping phase: %s").String()
				log.Warnf(text, title)
				continue
			}
		}

		if p, ok := p.(beforehook); ok {
			log.Debugf("Before phase '%s'", title)
			if err := p.Before(title); err != nil {
				log.Debugf("before hook failed '%s'", err.Error())
				return err
			}
		}

		if p, ok := p.(propsetter); ok {
			// if m.Config.Spec.K0s == nil {
			// 	m.Config.Spec.K0s = &cluster.K0s{Metadata: cluster.K0sMetadata{}}
			// }
			// if m.Config.Spec.K0s.Metadata.ClusterID != "" {
			// 	p.SetProp("clusterID", m.Config.Spec.K0s.Metadata.ClusterID)
			// }
			p.SetProp("apolloID", m.Config.Metadata.Name)
		}

		if dp, ok := p.(withDryRun); ok && m.DryRun {
			if err := dp.DryRun(); err != nil {
				return err
			}
			continue
		}
		text = Colorize.Green("Running phase: %s").String()
		log.Infof(text, title)
		result = p.Run()
		ran = append(ran, p)

		if p, ok := p.(afterhook); ok {
			if err := p.After(result); err != nil {
				log.Debugf("after hook failed: '%s' (phase result: %s)", err.Error(), result)
				return err
			}
		}

		if result != nil {
			return result
		}
	}

	return nil
}
