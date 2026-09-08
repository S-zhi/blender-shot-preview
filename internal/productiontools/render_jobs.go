package productiontools

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"sync"
	"time"
)

type RenderStatus string

const (
	RenderQueued    RenderStatus = "queued"
	RenderRunning   RenderStatus = "running"
	RenderSucceeded RenderStatus = "succeeded"
	RenderFailed    RenderStatus = "failed"
	RenderCancelled RenderStatus = "cancelled"
)

type RenderJob struct {
	ID           string       `json:"job_id"`
	Status       RenderStatus `json:"status"`
	ArtifactPath string       `json:"artifact_path,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
}

type renderRecord struct {
	RenderJob
	cancel  context.CancelFunc
	process CommandProcess
}

// RenderJobManager stores the lifetime of one asynchronous Blender render.
// Persistence/recovery belongs to the future job repository boundary.
type RenderJobManager struct {
	ws      workspace
	blender string
	exec    CommandExecutor
	mu      sync.RWMutex
	jobs    map[string]*renderRecord
}

func NewRenderJobManager(ws workspace, blender string, exec CommandExecutor) *RenderJobManager {
	return &RenderJobManager{ws: ws, blender: blender, exec: exec, jobs: make(map[string]*renderRecord)}
}

// NewRenderManager creates a job manager for callers that do not need the
// complete registered tool set.
func NewRenderManager(workspaceRoot, blender string, exec CommandExecutor) (*RenderJobManager, error) {
	ws, err := newWorkspace(workspaceRoot)
	if err != nil {
		return nil, err
	}
	if exec == nil {
		exec = OSCommandExecutor{}
	}
	return NewRenderJobManager(ws, blender, exec), nil
}

func (m *RenderJobManager) Submit(ctx context.Context, project, output string, first, last int) (RenderJob, error) {
	if err := ctx.Err(); err != nil {
		return RenderJob{}, err
	}
	id := newRenderID()
	started := time.Now().UTC()
	// The Eino guard cancels the short submit call after this method returns.
	// A long render must instead be owned by its job and stopped explicitly by
	// Cancel, while still rejecting a request already cancelled before submit.
	runCtx, cancel := context.WithCancel(context.Background())
	record := &renderRecord{RenderJob: RenderJob{ID: id, Status: RenderQueued, ArtifactPath: m.ws.show(output), CreatedAt: started}, cancel: cancel}
	m.mu.Lock()
	m.jobs[id] = record
	m.mu.Unlock()

	command := Command{Name: m.blender, Dir: m.ws.root, Args: []string{
		"--background", project, "--render-output", output,
		"--frame-start", strconv.Itoa(first), "--frame-end", strconv.Itoa(last), "--animation",
	}}
	process, err := m.exec.Start(runCtx, command)
	if err != nil {
		m.finish(id, RenderFailed)
		return record.RenderJob, err
	}
	m.mu.Lock()
	record.process = process
	record.Status = RenderRunning
	job := record.RenderJob
	m.mu.Unlock()
	waitDone := make(chan struct{})
	go func() {
		select {
		case <-runCtx.Done():
			_ = process.Kill()
		case <-waitDone:
		}
	}()
	go func() {
		result, waitErr := process.Wait()
		close(waitDone)
		if runCtx.Err() != nil {
			m.finish(id, RenderCancelled)
		} else if waitErr != nil || result.ExitCode != 0 {
			m.finish(id, RenderFailed)
		} else {
			m.finish(id, RenderSucceeded)
		}
	}()
	return job, nil
}

func (m *RenderJobManager) Get(id string) (RenderJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	record := m.jobs[id]
	if record == nil {
		return RenderJob{}, errors.New("render job not found")
	}
	return record.RenderJob, nil
}

func (m *RenderJobManager) Cancel(id string) (RenderJob, error) {
	m.mu.Lock()
	record := m.jobs[id]
	if record == nil {
		m.mu.Unlock()
		return RenderJob{}, errors.New("render job not found")
	}
	if record.Status == RenderSucceeded || record.Status == RenderFailed || record.Status == RenderCancelled {
		job := record.RenderJob
		m.mu.Unlock()
		return job, errors.New("render job is already final")
	}
	record.cancel()
	if record.process != nil {
		_ = record.process.Kill()
	}
	now := time.Now().UTC()
	record.Status = RenderCancelled
	record.CompletedAt = &now
	job := record.RenderJob
	m.mu.Unlock()
	return job, nil
}

func (m *RenderJobManager) finish(id string, status RenderStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	record := m.jobs[id]
	if record == nil || record.Status == RenderCancelled || record.Status == RenderSucceeded || record.Status == RenderFailed {
		return
	}
	now := time.Now().UTC()
	record.Status = status
	record.CompletedAt = &now
}

func newRenderID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (p *ProductionTools) renderSubmit(ctx context.Context, raw string) string {
	var in struct {
		ProjectPath string `json:"project_path"`
		OutputPath  string `json:"output_path"`
		FrameStart  int    `json:"frame_start"`
		FrameEnd    int    `json:"frame_end"`
	}
	if err := decode(raw, &in); err != nil {
		return decodeFailure(err)
	}
	project, err := p.ws.resolve(in.ProjectPath, ".blend")
	if err != nil || !regular(project) {
		return failure("TOOL_INVALID_ARGUMENT", "project_path must be an existing .blend workspace file.")
	}
	output, err := p.ws.resolve(in.OutputPath)
	if err != nil || in.FrameStart < 1 || in.FrameEnd < in.FrameStart || in.FrameEnd > maxFrame {
		return failure("TOOL_ARGUMENT_OUT_OF_RANGE", "output path or frame range is invalid.")
	}
	fingerprint := digest(ToolBlenderRenderSubmit + project + output + strconv.Itoa(in.FrameStart) + strconv.Itoa(in.FrameEnd))
	return p.once(ctx, fingerprint, func() string {
		job, submitErr := p.jobs.Submit(ctx, project, output, in.FrameStart, in.FrameEnd)
		if submitErr != nil {
			return failure("TOOL_EXECUTION_FAILED", "The render could not be submitted.")
		}
		return success(job)
	})
}

func (p *ProductionTools) renderStatus(_ context.Context, raw string) string {
	var in struct {
		JobID string `json:"job_id"`
	}
	if err := decode(raw, &in); err != nil {
		return decodeFailure(err)
	}
	if in.JobID == "" {
		return failure("TOOL_INVALID_ARGUMENT", "job_id is required.")
	}
	job, err := p.jobs.Get(in.JobID)
	if err != nil {
		return failure("TOOL_NOT_FOUND", "The render job was not found.")
	}
	return success(job)
}

func (p *ProductionTools) renderCancel(_ context.Context, raw string) string {
	var in struct {
		JobID string `json:"job_id"`
	}
	if err := decode(raw, &in); err != nil {
		return decodeFailure(err)
	}
	if in.JobID == "" {
		return failure("TOOL_INVALID_ARGUMENT", "job_id is required.")
	}
	job, err := p.jobs.Cancel(in.JobID)
	if err != nil {
		return failure("TOOL_CONFLICT", "The render job cannot be cancelled.")
	}
	return success(job)
}

func (p *ProductionTools) renderArtifact(_ context.Context, raw string) string {
	var in struct {
		JobID string `json:"job_id"`
	}
	if err := decode(raw, &in); err != nil {
		return decodeFailure(err)
	}
	if in.JobID == "" {
		return failure("TOOL_INVALID_ARGUMENT", "job_id is required.")
	}
	job, err := p.jobs.Get(in.JobID)
	if err != nil {
		return failure("TOOL_NOT_FOUND", "The render job was not found.")
	}
	if job.Status != RenderSucceeded {
		return failure("TOOL_PRECONDITION_FAILED", "The render artifact is not available yet.")
	}
	return success(map[string]string{"artifact_path": job.ArtifactPath})
}
