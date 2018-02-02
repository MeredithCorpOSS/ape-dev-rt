package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/hashicorp/nomad/client/allocdir"
	"github.com/hashicorp/nomad/client/config"
	"github.com/hashicorp/nomad/client/driver"
	"github.com/hashicorp/nomad/client/vaultclient"
	"github.com/hashicorp/nomad/nomad/mock"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/hashicorp/nomad/testutil"

	ctestutil "github.com/hashicorp/nomad/client/testutil"
)

func testLogger() *log.Logger {
	return prefixedTestLogger("")
}

func prefixedTestLogger(prefix string) *log.Logger {
	return log.New(os.Stderr, prefix, log.LstdFlags)
}

type MockTaskStateUpdater struct {
	state  string
	failed bool
	events []*structs.TaskEvent
}

func (m *MockTaskStateUpdater) Update(name, state string, event *structs.TaskEvent) {
	if state != "" {
		m.state = state
	}
	if event != nil {
		if event.FailsTask {
			m.failed = true
		}
		m.events = append(m.events, event)
	}
}

func testTaskRunner(restarts bool) (*MockTaskStateUpdater, *TaskRunner) {
	return testTaskRunnerFromAlloc(restarts, mock.Alloc())
}

// Creates a mock task runner using the first task in the first task group of
// the passed allocation.
func testTaskRunnerFromAlloc(restarts bool, alloc *structs.Allocation) (*MockTaskStateUpdater, *TaskRunner) {
	logger := testLogger()
	conf := config.DefaultConfig()
	conf.StateDir = os.TempDir()
	conf.AllocDir = os.TempDir()
	upd := &MockTaskStateUpdater{}
	task := alloc.Job.TaskGroups[0].Tasks[0]
	// Initialize the port listing. This should be done by the offer process but
	// we have a mock so that doesn't happen.
	task.Resources.Networks[0].ReservedPorts = []structs.Port{{"", 80}}

	allocDir := allocdir.NewAllocDir(filepath.Join(conf.AllocDir, alloc.ID))
	allocDir.Build([]*structs.Task{task})

	vclient := vaultclient.NewMockVaultClient()
	ctx := driver.NewExecContext(allocDir, alloc.ID)
	tr := NewTaskRunner(logger, conf, upd.Update, ctx, alloc, task, vclient)
	if !restarts {
		tr.restartTracker = noRestartsTracker()
	}
	return upd, tr
}

func TestTaskRunner_SimpleRun(t *testing.T) {
	ctestutil.ExecCompatible(t)
	upd, tr := testTaskRunner(false)
	tr.MarkReceived()
	go tr.Run()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	select {
	case <-tr.WaitCh():
	case <-time.After(time.Duration(testutil.TestMultiplier()*15) * time.Second):
		t.Fatalf("timeout")
	}

	if len(upd.events) != 3 {
		t.Fatalf("should have 3 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStateDead {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStateDead)
	}

	if upd.events[0].Type != structs.TaskReceived {
		t.Fatalf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
	}

	if upd.events[1].Type != structs.TaskStarted {
		t.Fatalf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
	}

	if upd.events[2].Type != structs.TaskTerminated {
		t.Fatalf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskTerminated)
	}
}

func TestTaskRunner_Run_RecoverableStartError(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code":               0,
		"start_error":             "driver failure",
		"start_error_recoverable": true,
	}

	upd, tr := testTaskRunnerFromAlloc(true, alloc)
	tr.MarkReceived()
	go tr.Run()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	testutil.WaitForResult(func() (bool, error) {
		if l := len(upd.events); l < 3 {
			return false, fmt.Errorf("Expect at least three  events; got %v", l)
		}

		if upd.events[0].Type != structs.TaskReceived {
			return false, fmt.Errorf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
		}

		if upd.events[1].Type != structs.TaskDriverFailure {
			return false, fmt.Errorf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskDriverFailure)
		}

		if upd.events[2].Type != structs.TaskRestarting {
			return false, fmt.Errorf("Second Event was %v; want %v", upd.events[2].Type, structs.TaskRestarting)
		}

		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})
}

func TestTaskRunner_Destroy(t *testing.T) {
	ctestutil.ExecCompatible(t)
	upd, tr := testTaskRunner(true)
	tr.MarkReceived()
	defer tr.ctx.AllocDir.Destroy()

	// Change command to ensure we run for a bit
	tr.task.Config["command"] = "/bin/sleep"
	tr.task.Config["args"] = []string{"1000"}
	go tr.Run()

	testutil.WaitForResult(func() (bool, error) {
		if l := len(upd.events); l != 2 {
			return false, fmt.Errorf("Expect two events; got %v", l)
		}

		if upd.events[0].Type != structs.TaskReceived {
			return false, fmt.Errorf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
		}

		if upd.events[1].Type != structs.TaskStarted {
			return false, fmt.Errorf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
		}

		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})

	// Make sure we are collecting  afew stats
	time.Sleep(2 * time.Second)
	stats := tr.LatestResourceUsage()
	if len(stats.Pids) == 0 || stats.ResourceUsage == nil || stats.ResourceUsage.MemoryStats.RSS == 0 {
		t.Fatalf("expected task runner to have some stats")
	}

	// Begin the tear down
	tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))

	select {
	case <-tr.WaitCh():
	case <-time.After(time.Duration(testutil.TestMultiplier()*15) * time.Second):
		t.Fatalf("timeout")
	}

	if len(upd.events) != 4 {
		t.Fatalf("should have 4 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStateDead {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStateDead)
	}

	if upd.events[2].Type != structs.TaskKilling {
		t.Fatalf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskKilling)
	}

	if upd.events[3].Type != structs.TaskKilled {
		t.Fatalf("Third Event was %v; want %v", upd.events[3].Type, structs.TaskKilled)
	}
}

func TestTaskRunner_Update(t *testing.T) {
	ctestutil.ExecCompatible(t)
	_, tr := testTaskRunner(false)

	// Change command to ensure we run for a bit
	tr.task.Config["command"] = "/bin/sleep"
	tr.task.Config["args"] = []string{"100"}
	go tr.Run()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	// Update the task definition
	updateAlloc := tr.alloc.Copy()

	// Update the restart policy
	newTG := updateAlloc.Job.TaskGroups[0]
	newMode := "foo"
	newTG.RestartPolicy.Mode = newMode

	newTask := updateAlloc.Job.TaskGroups[0].Tasks[0]
	newTask.Driver = "foobar"

	// Update the kill timeout
	testutil.WaitForResult(func() (bool, error) {
		if tr.handle == nil {
			return false, fmt.Errorf("task not started")
		}
		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})

	oldHandle := tr.handle.ID()
	newTask.KillTimeout = time.Hour

	tr.Update(updateAlloc)

	// Wait for update to take place
	testutil.WaitForResult(func() (bool, error) {
		if tr.task == newTask {
			return false, fmt.Errorf("We copied the pointer! This would be very bad")
		}
		if tr.task.Driver != newTask.Driver {
			return false, fmt.Errorf("Task not copied")
		}
		if tr.restartTracker.policy.Mode != newMode {
			return false, fmt.Errorf("restart policy not updated")
		}
		if tr.handle.ID() == oldHandle {
			return false, fmt.Errorf("handle not updated")
		}
		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})
}

func TestTaskRunner_SaveRestoreState(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code": "0",
		"run_for":   "5s",
	}

	// Give it a Vault token
	task.Vault = &structs.Vault{Policies: []string{"default"}}

	upd, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	go tr.Run()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))

	// Wait for the task to be running and then snapshot the state
	testutil.WaitForResult(func() (bool, error) {
		if l := len(upd.events); l != 2 {
			return false, fmt.Errorf("Expect two events; got %v", l)
		}

		if upd.events[0].Type != structs.TaskReceived {
			return false, fmt.Errorf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
		}

		if upd.events[1].Type != structs.TaskStarted {
			return false, fmt.Errorf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
		}

		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})

	if err := tr.SaveState(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Read the token from the file system
	secretDir, err := tr.ctx.AllocDir.GetSecretDir(task.Name)
	if err != nil {
		t.Fatalf("failed to determine task %s secret dir: %v", err)
	}

	tokenPath := filepath.Join(secretDir, vaultTokenFile)
	data, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	token := string(data)
	if len(token) == 0 {
		t.Fatalf("Token not written to disk")
	}

	// Create a new task runner
	tr2 := NewTaskRunner(tr.logger, tr.config, upd.Update,
		tr.ctx, tr.alloc, &structs.Task{Name: tr.task.Name}, tr.vaultClient)
	tr2.restartTracker = noRestartsTracker()
	if err := tr2.RestoreState(); err != nil {
		t.Fatalf("err: %v", err)
	}
	go tr2.Run()
	defer tr2.Destroy(structs.NewTaskEvent(structs.TaskKilled))

	// Destroy and wait
	select {
	case <-tr2.WaitCh():
	case <-time.After(time.Duration(testutil.TestMultiplier()*15) * time.Second):
		t.Fatalf("timeout")
	}

	// Check that we recovered the token
	if act := tr2.vaultFuture.Get(); act != token {
		t.Fatalf("Vault token not properly recovered")
	}
}

func TestTaskRunner_Download_List(t *testing.T) {
	ctestutil.ExecCompatible(t)

	ts := httptest.NewServer(http.FileServer(http.Dir(filepath.Dir("."))))
	defer ts.Close()

	// Create an allocation that has a task with a list of artifacts.
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	f1 := "task_runner_test.go"
	f2 := "task_runner.go"
	artifact1 := structs.TaskArtifact{
		GetterSource: fmt.Sprintf("%s/%s", ts.URL, f1),
	}
	artifact2 := structs.TaskArtifact{
		GetterSource: fmt.Sprintf("%s/%s", ts.URL, f2),
	}
	task.Artifacts = []*structs.TaskArtifact{&artifact1, &artifact2}

	upd, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	go tr.Run()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	select {
	case <-tr.WaitCh():
	case <-time.After(time.Duration(testutil.TestMultiplier()*15) * time.Second):
		t.Fatalf("timeout")
	}

	if len(upd.events) != 4 {
		t.Fatalf("should have 4 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStateDead {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStateDead)
	}

	if upd.events[0].Type != structs.TaskReceived {
		t.Fatalf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
	}

	if upd.events[1].Type != structs.TaskDownloadingArtifacts {
		t.Fatalf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskDownloadingArtifacts)
	}

	if upd.events[2].Type != structs.TaskStarted {
		t.Fatalf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskStarted)
	}

	if upd.events[3].Type != structs.TaskTerminated {
		t.Fatalf("Fourth Event was %v; want %v", upd.events[3].Type, structs.TaskTerminated)
	}

	// Check that both files exist.
	taskDir := tr.ctx.AllocDir.TaskDirs[task.Name]
	if _, err := os.Stat(filepath.Join(taskDir, f1)); err != nil {
		t.Fatalf("%v not downloaded", f1)
	}
	if _, err := os.Stat(filepath.Join(taskDir, f2)); err != nil {
		t.Fatalf("%v not downloaded", f2)
	}
}

func TestTaskRunner_Download_Retries(t *testing.T) {
	ctestutil.ExecCompatible(t)

	// Create an allocation that has a task with bad artifacts.
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	artifact := structs.TaskArtifact{
		GetterSource: "http://127.1.1.111:12315/foo/bar/baz",
	}
	task.Artifacts = []*structs.TaskArtifact{&artifact}

	// Make the restart policy try one update
	alloc.Job.TaskGroups[0].RestartPolicy = &structs.RestartPolicy{
		Attempts: 1,
		Interval: 10 * time.Minute,
		Delay:    1 * time.Second,
		Mode:     structs.RestartPolicyModeFail,
	}

	upd, tr := testTaskRunnerFromAlloc(true, alloc)
	tr.MarkReceived()
	go tr.Run()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	select {
	case <-tr.WaitCh():
	case <-time.After(time.Duration(testutil.TestMultiplier()*15) * time.Second):
		t.Fatalf("timeout")
	}

	if len(upd.events) != 7 {
		t.Fatalf("should have 7 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStateDead {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStateDead)
	}

	if upd.events[0].Type != structs.TaskReceived {
		t.Fatalf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
	}

	if upd.events[1].Type != structs.TaskDownloadingArtifacts {
		t.Fatalf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskDownloadingArtifacts)
	}

	if upd.events[2].Type != structs.TaskArtifactDownloadFailed {
		t.Fatalf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskArtifactDownloadFailed)
	}

	if upd.events[3].Type != structs.TaskRestarting {
		t.Fatalf("Fourth Event was %v; want %v", upd.events[3].Type, structs.TaskRestarting)
	}

	if upd.events[4].Type != structs.TaskDownloadingArtifacts {
		t.Fatalf("Fifth Event was %v; want %v", upd.events[4].Type, structs.TaskDownloadingArtifacts)
	}

	if upd.events[5].Type != structs.TaskArtifactDownloadFailed {
		t.Fatalf("Sixth Event was %v; want %v", upd.events[5].Type, structs.TaskArtifactDownloadFailed)
	}

	if upd.events[6].Type != structs.TaskNotRestarting {
		t.Fatalf("Seventh Event was %v; want %v", upd.events[6].Type, structs.TaskNotRestarting)
	}
}

func TestTaskRunner_Validate_UserEnforcement(t *testing.T) {
	_, tr := testTaskRunner(false)
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	if err := tr.setTaskEnv(); err != nil {
		t.Fatalf("bad: %v", err)
	}

	// Try to run as root with exec.
	tr.task.Driver = "exec"
	tr.task.User = "root"
	if err := tr.validateTask(); err == nil {
		t.Fatalf("expected error running as root with exec")
	}

	// Try to run a non-blacklisted user with exec.
	tr.task.Driver = "exec"
	tr.task.User = "foobar"
	if err := tr.validateTask(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to run as root with docker.
	tr.task.Driver = "docker"
	tr.task.User = "root"
	if err := tr.validateTask(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskRunner_RestartTask(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code": "0",
		"run_for":   "10s",
	}

	upd, tr := testTaskRunnerFromAlloc(true, alloc)
	tr.MarkReceived()
	go tr.Run()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	go func() {
		time.Sleep(time.Duration(testutil.TestMultiplier()*300) * time.Millisecond)
		tr.Restart("test", "restart")
		time.Sleep(time.Duration(testutil.TestMultiplier()*300) * time.Millisecond)
		tr.Kill("test", "restart", false)
	}()

	select {
	case <-tr.WaitCh():
	case <-time.After(time.Duration(testutil.TestMultiplier()*15) * time.Second):
		t.Fatalf("timeout")
	}

	if len(upd.events) != 9 {
		t.Fatalf("should have 9 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStateDead {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStateDead)
	}

	if upd.events[0].Type != structs.TaskReceived {
		t.Fatalf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
	}

	if upd.events[1].Type != structs.TaskStarted {
		t.Fatalf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
	}

	if upd.events[2].Type != structs.TaskRestartSignal {
		t.Fatalf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskRestartSignal)
	}

	if upd.events[3].Type != structs.TaskKilling {
		t.Fatalf("Fourth Event was %v; want %v", upd.events[3].Type, structs.TaskKilling)
	}

	if upd.events[4].Type != structs.TaskKilled {
		t.Fatalf("Fifth Event was %v; want %v", upd.events[4].Type, structs.TaskKilled)
	}

	t.Logf("%+v", upd.events[5])
	if upd.events[5].Type != structs.TaskRestarting {
		t.Fatalf("Sixth Event was %v; want %v", upd.events[5].Type, structs.TaskRestarting)
	}

	if upd.events[6].Type != structs.TaskStarted {
		t.Fatalf("Seventh Event was %v; want %v", upd.events[7].Type, structs.TaskStarted)
	}
	if upd.events[7].Type != structs.TaskKilling {
		t.Fatalf("Eighth Event was %v; want %v", upd.events[7].Type, structs.TaskKilling)
	}

	if upd.events[8].Type != structs.TaskKilled {
		t.Fatalf("Nineth Event was %v; want %v", upd.events[8].Type, structs.TaskKilled)
	}
}

func TestTaskRunner_KillTask(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code": "0",
		"run_for":   "10s",
	}

	upd, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	go tr.Run()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	go func() {
		time.Sleep(100 * time.Millisecond)
		tr.Kill("test", "kill", true)
	}()

	select {
	case <-tr.WaitCh():
	case <-time.After(time.Duration(testutil.TestMultiplier()*15) * time.Second):
		t.Fatalf("timeout")
	}

	if len(upd.events) != 4 {
		t.Fatalf("should have 4 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStateDead {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStateDead)
	}

	if !upd.failed {
		t.Fatalf("TaskState should be failed: %+v", upd)
	}

	if upd.events[0].Type != structs.TaskReceived {
		t.Fatalf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
	}

	if upd.events[1].Type != structs.TaskStarted {
		t.Fatalf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
	}

	if upd.events[2].Type != structs.TaskKilling {
		t.Fatalf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskKilling)
	}

	if upd.events[3].Type != structs.TaskKilled {
		t.Fatalf("Fourth Event was %v; want %v", upd.events[3].Type, structs.TaskKilled)
	}
}

func TestTaskRunner_SignalFailure(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code":    "0",
		"run_for":      "10s",
		"signal_error": "test forcing failure",
	}

	_, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	go tr.Run()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	time.Sleep(100 * time.Millisecond)
	if err := tr.Signal("test", "test", syscall.SIGINT); err == nil {
		t.Fatalf("Didn't receive error")
	}
}

func TestTaskRunner_BlockForVault(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code": "0",
		"run_for":   "1s",
	}
	task.Vault = &structs.Vault{Policies: []string{"default"}}

	upd, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	// Control when we get a Vault token
	token := "1234"
	waitCh := make(chan struct{})
	handler := func(*structs.Allocation, []string) (map[string]string, error) {
		<-waitCh
		return map[string]string{task.Name: token}, nil
	}
	tr.vaultClient.(*vaultclient.MockVaultClient).DeriveTokenFn = handler

	go tr.Run()

	select {
	case <-tr.WaitCh():
		t.Fatalf("premature exit")
	case <-time.After(1 * time.Second):
	}

	if len(upd.events) != 1 {
		t.Fatalf("should have 1 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStatePending {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStatePending)
	}

	if upd.events[0].Type != structs.TaskReceived {
		t.Fatalf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
	}

	// Unblock
	close(waitCh)

	select {
	case <-tr.WaitCh():
	case <-time.After(time.Duration(testutil.TestMultiplier()*15) * time.Second):
		t.Fatalf("timeout")
	}

	if len(upd.events) != 3 {
		t.Fatalf("should have 3 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStateDead {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStateDead)
	}

	if upd.events[0].Type != structs.TaskReceived {
		t.Fatalf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
	}

	if upd.events[1].Type != structs.TaskStarted {
		t.Fatalf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
	}

	if upd.events[2].Type != structs.TaskTerminated {
		t.Fatalf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskTerminated)
	}

	// Check that the token is on disk
	secretDir, err := tr.ctx.AllocDir.GetSecretDir(task.Name)
	if err != nil {
		t.Fatalf("failed to determine task %s secret dir: %v", err)
	}

	// Read the token from the file system
	tokenPath := filepath.Join(secretDir, vaultTokenFile)
	data, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if act := string(data); act != token {
		t.Fatalf("Token didn't get written to disk properly, got %q; want %q", act, token)
	}
}

func TestTaskRunner_DeriveToken_Retry(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code": "0",
		"run_for":   "1s",
	}
	task.Vault = &structs.Vault{Policies: []string{"default"}}

	upd, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	// Control when we get a Vault token
	token := "1234"
	count := 0
	handler := func(*structs.Allocation, []string) (map[string]string, error) {
		if count > 0 {
			return map[string]string{task.Name: token}, nil
		}

		count++
		return nil, structs.NewRecoverableError(fmt.Errorf("Want a retry"), true)
	}
	tr.vaultClient.(*vaultclient.MockVaultClient).DeriveTokenFn = handler
	go tr.Run()

	select {
	case <-tr.WaitCh():
	case <-time.After(time.Duration(testutil.TestMultiplier()*15) * time.Second):
		t.Fatalf("timeout")
	}

	if len(upd.events) != 3 {
		t.Fatalf("should have 3 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStateDead {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStateDead)
	}

	if upd.events[0].Type != structs.TaskReceived {
		t.Fatalf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
	}

	if upd.events[1].Type != structs.TaskStarted {
		t.Fatalf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
	}

	if upd.events[2].Type != structs.TaskTerminated {
		t.Fatalf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskTerminated)
	}

	// Check that the token is on disk
	secretDir, err := tr.ctx.AllocDir.GetSecretDir(task.Name)
	if err != nil {
		t.Fatalf("failed to determine task %s secret dir: %v", err)
	}

	// Read the token from the file system
	tokenPath := filepath.Join(secretDir, vaultTokenFile)
	data, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if act := string(data); act != token {
		t.Fatalf("Token didn't get written to disk properly, got %q; want %q", act, token)
	}
}

func TestTaskRunner_DeriveToken_Unrecoverable(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code": "0",
		"run_for":   "10s",
	}
	task.Vault = &structs.Vault{
		Policies:   []string{"default"},
		ChangeMode: structs.VaultChangeModeRestart,
	}

	upd, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	// Error the token derivation
	vc := tr.vaultClient.(*vaultclient.MockVaultClient)
	vc.SetDeriveTokenError(alloc.ID, []string{task.Name}, fmt.Errorf("Non recoverable"))
	go tr.Run()

	// Wait for the task to start
	testutil.WaitForResult(func() (bool, error) {
		if l := len(upd.events); l != 2 {
			return false, fmt.Errorf("Expect two events; got %v", l)
		}

		if upd.events[0].Type != structs.TaskReceived {
			return false, fmt.Errorf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
		}

		if upd.events[1].Type != structs.TaskKilling {
			return false, fmt.Errorf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskKilling)
		}

		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})
}

func TestTaskRunner_Template_Block(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code": "0",
		"run_for":   "1s",
	}
	task.Templates = []*structs.Template{
		{
			EmbeddedTmpl: "{{key \"foo\"}}",
			DestPath:     "local/test",
			ChangeMode:   structs.TemplateChangeModeNoop,
		},
	}

	upd, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	go tr.Run()

	select {
	case <-tr.WaitCh():
		t.Fatalf("premature exit")
	case <-time.After(1 * time.Second):
	}

	if len(upd.events) != 1 {
		t.Fatalf("should have 1 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStatePending {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStatePending)
	}

	if upd.events[0].Type != structs.TaskReceived {
		t.Fatalf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
	}

	// Unblock
	tr.UnblockStart("test")

	select {
	case <-tr.WaitCh():
	case <-time.After(time.Duration(testutil.TestMultiplier()*15) * time.Second):
		t.Fatalf("timeout")
	}

	if len(upd.events) != 3 {
		t.Fatalf("should have 3 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStateDead {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStateDead)
	}

	if upd.events[0].Type != structs.TaskReceived {
		t.Fatalf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
	}

	if upd.events[1].Type != structs.TaskStarted {
		t.Fatalf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
	}

	if upd.events[2].Type != structs.TaskTerminated {
		t.Fatalf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskTerminated)
	}
}

func TestTaskRunner_Template_Artifact(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal("bad: %v", err)
	}

	ts := httptest.NewServer(http.FileServer(http.Dir(filepath.Join(dir, ".."))))
	defer ts.Close()

	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code": "0",
		"run_for":   "1s",
	}
	// Create an allocation that has a task that renders a template from an
	// artifact
	f1 := "CHANGELOG.md"
	artifact := structs.TaskArtifact{
		GetterSource: fmt.Sprintf("%s/%s", ts.URL, f1),
	}
	task.Artifacts = []*structs.TaskArtifact{&artifact}
	task.Templates = []*structs.Template{
		{
			SourcePath: "CHANGELOG.md",
			DestPath:   "local/test",
			ChangeMode: structs.TemplateChangeModeNoop,
		},
	}

	upd, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()

	go tr.Run()

	select {
	case <-tr.WaitCh():
	case <-time.After(time.Duration(testutil.TestMultiplier()*15) * time.Second):
		t.Fatalf("timeout")
	}

	if len(upd.events) != 4 {
		t.Fatalf("should have 4 updates: %#v", upd.events)
	}

	if upd.state != structs.TaskStateDead {
		t.Fatalf("TaskState %v; want %v", upd.state, structs.TaskStateDead)
	}

	if upd.events[0].Type != structs.TaskReceived {
		t.Fatalf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
	}

	if upd.events[1].Type != structs.TaskDownloadingArtifacts {
		t.Fatalf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskDownloadingArtifacts)
	}

	if upd.events[2].Type != structs.TaskStarted {
		t.Fatalf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskStarted)
	}

	if upd.events[3].Type != structs.TaskTerminated {
		t.Fatalf("Fourth Event was %v; want %v", upd.events[3].Type, structs.TaskTerminated)
	}

	// Check that both files exist.
	taskDir := tr.ctx.AllocDir.TaskDirs[task.Name]
	if _, err := os.Stat(filepath.Join(taskDir, f1)); err != nil {
		t.Fatalf("%v not downloaded", f1)
	}
	if _, err := os.Stat(filepath.Join(taskDir, allocdir.TaskLocal, "test")); err != nil {
		t.Fatalf("template not rendered")
	}
}

func TestTaskRunner_Template_NewVaultToken(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code": "0",
		"run_for":   "1s",
	}
	task.Templates = []*structs.Template{
		{
			EmbeddedTmpl: "{{key \"foo\"}}",
			DestPath:     "local/test",
			ChangeMode:   structs.TemplateChangeModeNoop,
		},
	}
	task.Vault = &structs.Vault{Policies: []string{"default"}}

	_, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()
	go tr.Run()

	// Wait for a Vault token
	var token string
	testutil.WaitForResult(func() (bool, error) {
		if token = tr.vaultFuture.Get(); token == "" {
			return false, fmt.Errorf("No Vault token")
		}

		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})

	// Error the token renewal
	vc := tr.vaultClient.(*vaultclient.MockVaultClient)
	renewalCh, ok := vc.RenewTokens[token]
	if !ok {
		t.Fatalf("no renewal channel")
	}

	originalManager := tr.templateManager

	renewalCh <- fmt.Errorf("Test killing")
	close(renewalCh)

	// Wait for a new Vault token
	var token2 string
	testutil.WaitForResult(func() (bool, error) {
		if token2 = tr.vaultFuture.Get(); token2 == "" || token2 == token {
			return false, fmt.Errorf("No new Vault token")
		}

		if originalManager == tr.templateManager {
			return false, fmt.Errorf("Template manager not updated")
		}

		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})
}

func TestTaskRunner_VaultManager_Restart(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code": "0",
		"run_for":   "10s",
	}
	task.Vault = &structs.Vault{
		Policies:   []string{"default"},
		ChangeMode: structs.VaultChangeModeRestart,
	}

	upd, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()
	go tr.Run()

	// Wait for the task to start
	testutil.WaitForResult(func() (bool, error) {
		if l := len(upd.events); l != 2 {
			return false, fmt.Errorf("Expect two events; got %v", l)
		}

		if upd.events[0].Type != structs.TaskReceived {
			return false, fmt.Errorf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
		}

		if upd.events[1].Type != structs.TaskStarted {
			return false, fmt.Errorf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
		}

		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})

	// Error the token renewal
	vc := tr.vaultClient.(*vaultclient.MockVaultClient)
	renewalCh, ok := vc.RenewTokens[tr.vaultFuture.Get()]
	if !ok {
		t.Fatalf("no renewal channel")
	}

	renewalCh <- fmt.Errorf("Test killing")
	close(renewalCh)

	// Ensure a restart
	testutil.WaitForResult(func() (bool, error) {
		if l := len(upd.events); l != 7 {
			return false, fmt.Errorf("Expect seven events; got %#v", upd.events)
		}

		if upd.events[0].Type != structs.TaskReceived {
			return false, fmt.Errorf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
		}

		if upd.events[1].Type != structs.TaskStarted {
			return false, fmt.Errorf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
		}

		if upd.events[2].Type != structs.TaskRestartSignal {
			return false, fmt.Errorf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskRestartSignal)
		}

		if upd.events[3].Type != structs.TaskKilling {
			return false, fmt.Errorf("Fourth Event was %v; want %v", upd.events[3].Type, structs.TaskKilling)
		}

		if upd.events[4].Type != structs.TaskKilled {
			return false, fmt.Errorf("Fifth Event was %v; want %v", upd.events[4].Type, structs.TaskKilled)
		}

		if upd.events[5].Type != structs.TaskRestarting {
			return false, fmt.Errorf("Sixth Event was %v; want %v", upd.events[5].Type, structs.TaskRestarting)
		}

		if upd.events[6].Type != structs.TaskStarted {
			return false, fmt.Errorf("Seventh Event was %v; want %v", upd.events[6].Type, structs.TaskStarted)
		}

		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})
}

func TestTaskRunner_VaultManager_Signal(t *testing.T) {
	alloc := mock.Alloc()
	task := alloc.Job.TaskGroups[0].Tasks[0]
	task.Driver = "mock_driver"
	task.Config = map[string]interface{}{
		"exit_code": "0",
		"run_for":   "10s",
	}
	task.Vault = &structs.Vault{
		Policies:     []string{"default"},
		ChangeMode:   structs.VaultChangeModeSignal,
		ChangeSignal: "SIGUSR1",
	}

	upd, tr := testTaskRunnerFromAlloc(false, alloc)
	tr.MarkReceived()
	defer tr.Destroy(structs.NewTaskEvent(structs.TaskKilled))
	defer tr.ctx.AllocDir.Destroy()
	go tr.Run()

	// Wait for the task to start
	testutil.WaitForResult(func() (bool, error) {
		if l := len(upd.events); l != 2 {
			return false, fmt.Errorf("Expect two events; got %v", l)
		}

		if upd.events[0].Type != structs.TaskReceived {
			return false, fmt.Errorf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
		}

		if upd.events[1].Type != structs.TaskStarted {
			return false, fmt.Errorf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
		}

		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})

	// Error the token renewal
	vc := tr.vaultClient.(*vaultclient.MockVaultClient)
	renewalCh, ok := vc.RenewTokens[tr.vaultFuture.Get()]
	if !ok {
		t.Fatalf("no renewal channel")
	}

	renewalCh <- fmt.Errorf("Test killing")
	close(renewalCh)

	// Ensure a restart
	testutil.WaitForResult(func() (bool, error) {
		if l := len(upd.events); l != 3 {
			return false, fmt.Errorf("Expect three events; got %#v", upd.events)
		}

		if upd.events[0].Type != structs.TaskReceived {
			return false, fmt.Errorf("First Event was %v; want %v", upd.events[0].Type, structs.TaskReceived)
		}

		if upd.events[1].Type != structs.TaskStarted {
			return false, fmt.Errorf("Second Event was %v; want %v", upd.events[1].Type, structs.TaskStarted)
		}

		if upd.events[2].Type != structs.TaskSignaling {
			return false, fmt.Errorf("Third Event was %v; want %v", upd.events[2].Type, structs.TaskSignaling)
		}

		return true, nil
	}, func(err error) {
		t.Fatalf("err: %v", err)
	})
}
