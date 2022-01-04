package operator

import (
	"sync"

	"github.com/yuyicai/kubei/pkg/ssh"
)

var sharedRunnerFactory = newSharedRunnerFactory

type RunnerFactory interface {
	Runner(host, port, user, password, key string) (Runner, error)
}

func newSharedRunnerFactory() RunnerFactory {
	return &runnerFactory{
		mu:      sync.Mutex{},
		runners: map[string]Runner{},
	}
}

type runnerFactory struct {
	mu      sync.Mutex
	runners map[string]Runner
}

func (f *runnerFactory) Runner(host, port, user, password, key string) (Runner, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	runner, exists := f.runners[host]
	if exists {
		return runner, nil
	}
	var err error
	runner, err = newRunner()
	if err != nil {
		return nil, err
	}
	f.runners[host] = runner
	return runner, nil
}

type Runner interface {
	Run(cmd string) error
	RunOut(cmd string) ([]byte, error)
}

type runner struct {
	client *ssh.Client
}

func (r *runner) Run(cmd string) error {
	return r.client.Run(cmd)
}

func (r *runner) RunOut(cmd string) ([]byte, error) {
	return r.client.RunOut(cmd)
}

func newRunner() (Runner, error) {
	// TODO: set runner
	return &runner{}, nil
}
