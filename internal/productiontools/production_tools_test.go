package productiontools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/S-zhi/blender-shot-preview/internal/agent"
)

func TestRegisterProductionToolsRegistersExpectedIDs(t *testing.T) {
	root := testWorkspace(t)
	registry := agent.NewMemorySkillRegistry()
	_, err := RegisterProductionTools(registry, Config{Workspace: root, Executor: &fakeExecutor{}})
	if err != nil {
		t.Fatalf("register tools: %v", err)
	}
	for _, id := range []string{ToolAssetInspect, ToolAssetSearch, ToolBlenderCreateModel, ToolBlenderCreateMaterial, ToolBlenderCreateRig, ToolBlenderImportReference, ToolBlenderCreateProject, ToolBlenderImportAsset, ToolBlenderApplyScenePatch, ToolBlenderApplyShotPlan, ToolBlenderSaveProject, ToolBlenderRenderSubmit, ToolBlenderRenderStatus, ToolBlenderRenderCancel, ToolBlenderRenderArtifact, ToolFFmpegEncode, ToolFFprobeInspect} {
		if !registry.Exists(id) {
			t.Errorf("missing registered tool %q", id)
		}
	}
	if kind, _ := registry.Kind(ToolBlenderRenderSubmit); kind != agent.ToolKindWrite {
		t.Fatalf("render submit kind = %q", kind)
	}
	if kind, _ := registry.Kind(ToolFFprobeInspect); kind != agent.ToolKindRead {
		t.Fatalf("ffprobe kind = %q", kind)
	}
	toolValue, err := registry.Resolve(context.Background(), ToolBlenderRenderSubmit)
	if err != nil {
		t.Fatal(err)
	}
	toolInfo, err := toolValue.Info(context.Background())
	if err != nil || toolInfo.ParamsOneOf == nil {
		t.Fatalf("render submit schema missing: %#v, %v", toolInfo, err)
	}
}

func TestMalformedJSONHasStableValidationCode(t *testing.T) {
	p, err := New(Config{Workspace: testWorkspace(t), Executor: &fakeExecutor{}})
	if err != nil {
		t.Fatal(err)
	}
	if code := resultCode(t, p.assetInspect(context.Background(), `{`)); code != "TOOL_INVALID_JSON" {
		t.Fatalf("code = %q", code)
	}
}

func TestToolValidationAndWorkspaceConfinement(t *testing.T) {
	root := testWorkspace(t)
	writeFile(t, root, "scripts/build.py")
	writeFile(t, root, "projects/source.blend")
	p, err := New(Config{Workspace: root, BlenderBinary: "blender-test", Executor: &fakeExecutor{}})
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		`{"script_path":"scripts/build.py","project_path":"projects/source.blend","unknown":true}`,
		`{"script_path":"../outside.py","project_path":"projects/source.blend"}`,
		`{"script_path":"scripts/build.py"}`,
	} {
		if got := p.blender(context.Background(), ToolBlenderCreateModel, raw); resultCode(t, got) != "TOOL_INVALID_ARGUMENT" {
			t.Errorf("validation result = %s", got)
		}
	}
	external := t.TempDir()
	writeFile(t, external, "outside.py")
	if err := os.Symlink(external, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	if got := p.blender(context.Background(), ToolBlenderCreateProject, `{"script_path":"linked/outside.py"}`); resultCode(t, got) != "TOOL_INVALID_ARGUMENT" {
		t.Errorf("symlink escape validation: %s", got)
	}
	if got := p.blender(context.Background(), ToolBlenderCreateProject, `{"script_path":"scripts/build.py","project_path":"projects/source.blend"}`); resultCode(t, got) != "TOOL_INVALID_ARGUMENT" {
		t.Errorf("create project project-path validation: %s", got)
	}
}

func TestIdempotencyLedgerReturnsFirstResultAndRejectsConflict(t *testing.T) {
	ledger := newIdempotencyLedger()
	called := 0
	first := ledger.execute(context.Background(), "key", "request-a", func() string { called++; return success(map[string]string{"value": "first"}) })
	second := ledger.execute(context.Background(), "key", "request-a", func() string { called++; return success(map[string]string{"value": "second"}) })
	if called != 1 || first != second {
		t.Fatalf("called=%d first=%s second=%s", called, first, second)
	}
	if resultCode(t, ledger.execute(context.Background(), "key", "request-b", func() string { return "unexpected" })) != "TOOL_IDEMPOTENCY_CONFLICT" {
		t.Fatal("expected conflict")
	}
}

func TestCommandsAreArgumentArrays(t *testing.T) {
	root := testWorkspace(t)
	writeFile(t, root, "scripts/build.py")
	writeFile(t, root, "projects/source.blend")
	exec := &fakeExecutor{}
	p, err := New(Config{Workspace: root, BlenderBinary: "blender-test", FFmpegBinary: "ffmpeg-test", Executor: exec})
	if err != nil {
		t.Fatal(err)
	}
	if got := p.blender(context.Background(), ToolBlenderCreateModel, `{"project_path":"projects/source.blend","script_path":"scripts/build.py","output_path":"projects/out.blend"}`); resultCode(t, got) != "" {
		t.Fatalf("blender: %s", got)
	}
	writeFile(t, root, "renders/input.png")
	if got := p.ffmpegEncode(context.Background(), `{"input_path":"renders/input.png","output_path":"renders/movie.mp4","fps":24,"codec":"libx264"}`); resultCode(t, got) != "" {
		t.Fatalf("ffmpeg: %s", got)
	}
	commands := exec.Commands()
	if len(commands) != 2 {
		t.Fatalf("commands = %#v", commands)
	}
	if commands[0].Name != "blender-test" || !reflect.DeepEqual(commands[0].Args, []string{"--background", filepath.Join(root, "projects/source.blend"), "--python-exit-code", "1", "--python", filepath.Join(root, "scripts/build.py"), "--", "--output", filepath.Join(root, "projects/out.blend")}) {
		t.Fatalf("blender command = %#v", commands[0])
	}
	if commands[1].Name != "ffmpeg-test" || commands[1].Args[0] != "-nostdin" || commands[1].Args[4] != "-i" {
		t.Fatalf("ffmpeg command = %#v", commands[1])
	}
}

func TestFFprobeParsing(t *testing.T) {
	root := testWorkspace(t)
	writeFile(t, root, "renders/movie.mp4")
	exec := &fakeExecutor{runResult: CommandResult{Stdout: []byte(`{"format":{"format_name":"mov,mp4","duration":"2.5"},"streams":[{"codec_type":"video","codec_name":"h264","width":1920,"height":1080}]}`)}}
	p, err := New(Config{Workspace: root, FFprobeBinary: "ffprobe-test", Executor: exec})
	if err != nil {
		t.Fatal(err)
	}
	var response result
	if err := json.Unmarshal([]byte(p.ffprobeInspect(context.Background(), `{"path":"renders/movie.mp4"}`)), &response); err != nil {
		t.Fatal(err)
	}
	if !response.OK {
		t.Fatalf("ffprobe failure: %#v", response)
	}
	data := response.Data.(map[string]any)
	if data["path"] != "renders/movie.mp4" {
		t.Fatalf("path = %#v", data)
	}
	if exec.Commands()[0].Name != "ffprobe-test" {
		t.Fatalf("command = %#v", exec.Commands()[0])
	}
}

func TestRenderJobTransitionsAndCancel(t *testing.T) {
	root := testWorkspace(t)
	writeFile(t, root, "projects/source.blend")
	process := newBlockingProcess()
	exec := &fakeExecutor{process: process}
	p, err := New(Config{Workspace: root, BlenderBinary: "blender-test", Executor: exec})
	if err != nil {
		t.Fatal(err)
	}
	job, err := p.jobs.Submit(context.Background(), filepath.Join(root, "projects/source.blend"), filepath.Join(root, "renders/frame_"), 1, 10)
	if err != nil || job.Status != RenderRunning {
		t.Fatalf("submit = %#v, %v", job, err)
	}
	cancelled, err := p.jobs.Cancel(job.ID)
	if err != nil || cancelled.Status != RenderCancelled || !process.Killed() {
		t.Fatalf("cancel = %#v, %v, killed=%v", cancelled, err, process.Killed())
	}
	if _, err := p.jobs.Cancel(job.ID); err == nil {
		t.Fatal("expected final job cancellation to fail")
	}
	close(process.done)
}

func testWorkspace(t *testing.T) string { t.Helper(); return t.TempDir() }
func writeFile(t *testing.T, root, name string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
}
func resultCode(t *testing.T, raw string) string {
	t.Helper()
	var got result
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatal(err)
	}
	if got.Error == nil {
		return ""
	}
	return got.Error.Code
}

type fakeExecutor struct {
	mu        sync.Mutex
	commands  []Command
	runResult CommandResult
	runErr    error
	process   *blockingProcess
}

func (f *fakeExecutor) Run(_ context.Context, command Command) (CommandResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.commands = append(f.commands, command)
	return f.runResult, f.runErr
}
func (f *fakeExecutor) Start(_ context.Context, command Command) (CommandProcess, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.commands = append(f.commands, command)
	if f.process == nil {
		return nil, errors.New("process was not configured")
	}
	return f.process, nil
}
func (f *fakeExecutor) Commands() []Command {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Command(nil), f.commands...)
}

type blockingProcess struct {
	done   chan struct{}
	mu     sync.Mutex
	killed bool
}

func newBlockingProcess() *blockingProcess              { return &blockingProcess{done: make(chan struct{})} }
func (p *blockingProcess) Wait() (CommandResult, error) { <-p.done; return CommandResult{}, nil }
func (p *blockingProcess) Kill() error                  { p.mu.Lock(); p.killed = true; p.mu.Unlock(); return nil }
func (p *blockingProcess) Killed() bool                 { p.mu.Lock(); defer p.mu.Unlock(); return p.killed }
