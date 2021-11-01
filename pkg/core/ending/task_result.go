package ending

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type TaskResult struct {
	mu        sync.Mutex
	Errors    []HostError
	Status    ResultStatus
	StartTime time.Time
	EndTime   time.Time
}

type HostError struct {
	Host  connector.Host
	Error error
}

func NewTaskResult() *TaskResult {
	return &TaskResult{Errors: make([]HostError, 0, 0), Status: NULL, StartTime: time.Now()}
}

func (t *TaskResult) ErrResult() {
	if t.Status != NULL {
		return
	}
	t.EndTime = time.Now()
	t.Status = FAILED
}

func (t *TaskResult) NormalResult() {
	if t.Status != NULL {
		return
	}
	t.EndTime = time.Now()
	t.Status = SUCCESS
}

func (t *TaskResult) SkippedResult() {
	if t.Status != NULL {
		return
	}
	t.EndTime = time.Now()
	t.Status = SKIPPED
}

func (t *TaskResult) AppendErr(host connector.Host, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e := HostError{
		Host:  host,
		Error: err,
	}

	t.Errors = append(t.Errors, e)
	t.EndTime = time.Now()
	t.Status = FAILED
}

func (t *TaskResult) IsFailed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status == FAILED {
		return true
	}
	return false
}

func (t *TaskResult) CombineErr() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.Errors) != 0 {
		var str string
		for i := range t.Errors {
			str += fmt.Sprintf("\nfailed: [%s] %s", t.Errors[i].Host.GetName(), t.Errors[i].Error.Error())
		}
		return errors.New(str)
	}
	return nil
}
