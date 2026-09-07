package productiontools

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/S-zhi/blender-shot-preview/internal/agent"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const (
	ToolAssetInspect           = "asset.inspect"
	ToolAssetSearch            = "asset.search"
	ToolBlenderCreateModel     = "blender.create_model"
	ToolBlenderCreateMaterial  = "blender.create_material"
	ToolBlenderCreateRig       = "blender.create_rig"
	ToolBlenderImportReference = "blender.import_reference"
	ToolBlenderCreateProject   = "blender.create_project"
	ToolBlenderImportAsset     = "blender.import_asset"
	ToolBlenderApplyScenePatch = "blender.apply_scene_patch"
	ToolBlenderApplyShotPlan   = "blender.apply_shot_plan"
	ToolBlenderSaveProject     = "blender.save_project"
	ToolBlenderRenderSubmit    = "blender.render.submit"
	ToolBlenderRenderStatus    = "blender.render.status"
	ToolBlenderRenderCancel    = "blender.render.cancel"
	ToolBlenderRenderArtifact  = "blender.render.artifact"
	ToolFFmpegEncode           = "ffmpeg.encode"
	ToolFFprobeInspect         = "ffprobe.inspect"
	maxArgumentBytes           = 64 << 10
	maxOutputBytes             = 1 << 20
	maxSearchResults           = 100
	maxFrame                   = 1000000
)

type Command struct {
	Name string
	Args []string
	Dir  string
}
type CommandResult struct {
	Stdout   []byte
	ExitCode int
}

// BlenderCommand builds the only command shape accepted for a scene operation.
func BlenderCommand(binary, workspaceRoot, project, script, output string) Command {
	args := []string{"--background"}
	if project != "" {
		args = append(args, project)
	}
	args = append(args, "--python-exit-code", "1", "--python", script)
	if output != "" {
		args = append(args, "--", "--output", output)
	}
	return Command{Name: binary, Args: args, Dir: workspaceRoot}
}

// FFmpegCommand builds a constrained argv for image-sequence encoding.
func FFmpegCommand(binary, workspaceRoot, input, output string, fps int, codec string) Command {
	return Command{Name: binary, Dir: workspaceRoot, Args: []string{"-nostdin", "-y", "-framerate", strconv.Itoa(fps), "-i", input, "-c:v", codec, "-pix_fmt", "yuv420p", output}}
}

// FFprobeCommand requests JSON metadata without allowing arbitrary flags.
func FFprobeCommand(binary, workspaceRoot, path string) Command {
	return Command{Name: binary, Dir: workspaceRoot, Args: []string{"-v", "error", "-print_format", "json", "-show_format", "-show_streams", path}}
}

type CommandProcess interface {
	Wait() (CommandResult, error)
	Kill() error
}
type CommandExecutor interface {
	Run(context.Context, Command) (CommandResult, error)
	Start(context.Context, Command) (CommandProcess, error)
}
type OSCommandExecutor struct{}

func (OSCommandExecutor) Run(ctx context.Context, c Command) (CommandResult, error) {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Dir = c.Dir
	stdout := &limitedBuffer{remaining: maxOutputBytes}
	cmd.Stdout = stdout
	cmd.Stderr = io.Discard
	err := cmd.Run()
	return CommandResult{Stdout: stdout.Bytes(), ExitCode: exitCode(err)}, err
}
func (OSCommandExecutor) Start(ctx context.Context, c Command) (CommandProcess, error) {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Dir = c.Dir
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return osProcess{cmd}, nil
}

type osProcess struct{ cmd *exec.Cmd }

func (p osProcess) Wait() (CommandResult, error) {
	e := p.cmd.Wait()
	return CommandResult{ExitCode: exitCode(e)}, e
}
func (p osProcess) Kill() error {
	if p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}
func exitCode(e error) int {
	if e == nil {
		return 0
	}
	var x *exec.ExitError
	if errors.As(e, &x) {
		return x.ExitCode()
	}
	return -1
}

type limitedBuffer struct {
	bytes.Buffer
	remaining int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.remaining > 0 {
		n := len(p)
		if n > b.remaining {
			n = b.remaining
		}
		_, _ = b.Buffer.Write(p[:n])
		b.remaining -= n
	}
	return len(p), nil
}

type Config struct {
	Workspace, BlenderBinary, FFmpegBinary, FFprobeBinary string
	Executor                                              CommandExecutor
}
type ProductionTools struct {
	ws                             workspace
	blenderBinary, ffmpeg, ffprobe string
	executor                       CommandExecutor
	jobs                           *RenderJobManager
	ledger                         *idempotencyLedger
}

func New(c Config) (*ProductionTools, error) {
	ws, e := newWorkspace(c.Workspace)
	if e != nil {
		return nil, e
	}
	if c.Executor == nil {
		c.Executor = OSCommandExecutor{}
	}
	if c.BlenderBinary == "" {
		c.BlenderBinary = "blender"
	}
	if c.FFmpegBinary == "" {
		c.FFmpegBinary = "ffmpeg"
	}
	if c.FFprobeBinary == "" {
		c.FFprobeBinary = "ffprobe"
	}
	p := &ProductionTools{ws: ws, blenderBinary: c.BlenderBinary, ffmpeg: c.FFmpegBinary, ffprobe: c.FFprobeBinary, executor: c.Executor, ledger: newIdempotencyLedger()}
	p.jobs = NewRenderJobManager(ws, c.BlenderBinary, c.Executor)
	return p, nil
}

func RegisterProductionTools(r agent.SkillRegistry, c Config) (*ProductionTools, error) {
	if r == nil {
		return nil, errors.New("production tool registry is required")
	}
	p, e := New(c)
	if e != nil {
		return nil, e
	}
	read := map[string]bool{ToolAssetInspect: true, ToolAssetSearch: true, ToolBlenderRenderStatus: true, ToolBlenderRenderArtifact: true, ToolFFprobeInspect: true}
	ids := []string{ToolAssetInspect, ToolAssetSearch, ToolBlenderCreateModel, ToolBlenderCreateMaterial, ToolBlenderCreateRig, ToolBlenderImportReference, ToolBlenderCreateProject, ToolBlenderImportAsset, ToolBlenderApplyScenePatch, ToolBlenderApplyShotPlan, ToolBlenderSaveProject, ToolBlenderRenderSubmit, ToolBlenderRenderStatus, ToolBlenderRenderCancel, ToolBlenderRenderArtifact, ToolFFmpegEncode, ToolFFprobeInspect}
	for _, id := range ids {
		id := id
		k := agent.ToolKindWrite
		if read[id] {
			k = agent.ToolKindRead
		}
		if e := r.RegisterWithKind(id, k, func(context.Context) (tool.BaseTool, error) { return p.tool(id), nil }); e != nil {
			return nil, fmt.Errorf("register %s: %w", id, e)
		}
	}
	return p, nil
}

type einoTool struct {
	info *schema.ToolInfo
	run  func(context.Context, string) string
}

func (t einoTool) Info(context.Context) (*schema.ToolInfo, error) { return t.info, nil }
func (t einoTool) InvokableRun(c context.Context, a string, _ ...tool.Option) (string, error) {
	return t.run(c, a), nil
}
func (p *ProductionTools) tool(id string) tool.BaseTool {
	params := map[string]*schema.ParameterInfo{}
	str := func(required bool) *schema.ParameterInfo {
		return &schema.ParameterInfo{Type: schema.String, Required: required}
	}
	integer := func() *schema.ParameterInfo { return &schema.ParameterInfo{Type: schema.Integer, Required: true} }
	switch id {
	case ToolAssetInspect, ToolFFprobeInspect:
		params["path"] = str(true)
	case ToolAssetSearch:
		params["directory"], params["extension"], params["limit"] = str(true), str(false), integer()
	case ToolBlenderRenderSubmit:
		params["project_path"], params["output_path"] = str(true), str(true)
		params["frame_start"], params["frame_end"] = integer(), integer()
	case ToolBlenderRenderStatus, ToolBlenderRenderCancel, ToolBlenderRenderArtifact:
		params["job_id"] = str(true)
	case ToolFFmpegEncode:
		params["input_path"], params["output_path"], params["codec"] = str(true), str(true), str(true)
		params["fps"] = integer()
	default:
		params["project_path"], params["script_path"], params["output_path"] = str(false), str(true), str(false)
	}
	return einoTool{info: &schema.ToolInfo{Name: id, Desc: "Controlled production tool; paths are workspace-relative.", ParamsOneOf: schema.NewParamsOneOfByParams(params)}, run: p.dispatch(id)}
}
func (p *ProductionTools) dispatch(id string) func(context.Context, string) string {
	switch id {
	case ToolAssetInspect:
		return p.assetInspect
	case ToolAssetSearch:
		return p.assetSearch
	case ToolBlenderRenderSubmit:
		return p.renderSubmit
	case ToolBlenderRenderStatus:
		return p.renderStatus
	case ToolBlenderRenderCancel:
		return p.renderCancel
	case ToolBlenderRenderArtifact:
		return p.renderArtifact
	case ToolFFmpegEncode:
		return p.ffmpegEncode
	case ToolFFprobeInspect:
		return p.ffprobeInspect
	}
	return func(c context.Context, a string) string { return p.blender(c, id, a) }
}

type result struct {
	OK    bool       `json:"ok"`
	Data  any        `json:"data,omitempty"`
	Error *toolError `json:"error,omitempty"`
}
type toolError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

func success(v any) string { b, _ := json.Marshal(result{OK: true, Data: v}); return string(b) }
func failure(c, m string) string {
	b, _ := json.Marshal(result{Error: &toolError{Code: c, Message: m}})
	return string(b)
}
func decode(a string, v any) error {
	if len(a) == 0 || len(a) > maxArgumentBytes {
		return errInvalidJSON
	}
	d := json.NewDecoder(strings.NewReader(a))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		var syntax *json.SyntaxError
		if errors.As(e, &syntax) || errors.Is(e, io.ErrUnexpectedEOF) {
			return errInvalidJSON
		}
		return errInvalidArgument
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		return errInvalidJSON
	}
	return nil
}

var errInvalidJSON = errors.New("invalid JSON")
var errInvalidArgument = errors.New("invalid tool argument")

func decodeFailure(err error) string {
	if errors.Is(err, errInvalidJSON) {
		return failure("TOOL_INVALID_JSON", "Tool arguments must be one valid JSON object.")
	}
	return failure("TOOL_INVALID_ARGUMENT", "Tool arguments do not match the required schema.")
}

type workspace struct{ root, canonical string }

var errPath = errors.New("invalid workspace path")

func newWorkspace(root string) (workspace, error) {
	abs, e := filepath.Abs(root)
	if e != nil {
		return workspace{}, e
	}
	i, e := os.Stat(abs)
	if e != nil || !i.IsDir() {
		return workspace{}, errors.New("workspace must be an existing directory")
	}
	real, e := filepath.EvalSymlinks(abs)
	if e != nil {
		return workspace{}, e
	}
	return workspace{abs, real}, nil
}
func (w workspace) resolve(path string, exts ...string) (string, error) {
	if path == "" || filepath.IsAbs(path) || strings.TrimSpace(path) != path {
		return "", errPath
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errPath
	}
	full := filepath.Join(w.root, clean)
	rel, e := filepath.Rel(w.root, full)
	if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errPath
	}
	// For a not-yet-created output, walk up to the first existing component.
	// Any symlink on that route must resolve back into the canonical workspace.
	for current := full; ; current = filepath.Dir(current) {
		if real, evalErr := filepath.EvalSymlinks(current); evalErr == nil {
			realRel, relErr := filepath.Rel(w.canonical, real)
			if relErr != nil || realRel == ".." || strings.HasPrefix(realRel, ".."+string(filepath.Separator)) {
				return "", errPath
			}
			break
		}
		if current == w.root || filepath.Dir(current) == current {
			break
		}
	}
	if len(exts) > 0 {
		got := strings.ToLower(filepath.Ext(full))
		for _, x := range exts {
			if got == x {
				return full, nil
			}
		}
		return "", errPath
	}
	return full, nil
}
func (w workspace) show(path string) string {
	r, e := filepath.Rel(w.root, path)
	if e != nil {
		return ""
	}
	return filepath.ToSlash(r)
}
func regular(path string) bool { i, e := os.Stat(path); return e == nil && i.Mode().IsRegular() }

func (p *ProductionTools) assetInspect(_ context.Context, a string) string {
	var in struct {
		Path string `json:"path"`
	}
	if err := decode(a, &in); err != nil {
		return decodeFailure(err)
	}
	path, e := p.ws.resolve(in.Path)
	if e != nil || !regular(path) {
		return failure("TOOL_INVALID_ARGUMENT", "path must be an existing regular file inside the task workspace.")
	}
	i, _ := os.Stat(path)
	return success(map[string]any{"path": p.ws.show(path), "size_bytes": i.Size(), "extension": filepath.Ext(path)})
}
func (p *ProductionTools) assetSearch(_ context.Context, a string) string {
	var in struct {
		Directory string `json:"directory"`
		Extension string `json:"extension"`
		Limit     int    `json:"limit"`
	}
	if err := decode(a, &in); err != nil {
		return decodeFailure(err)
	}
	if in.Limit < 1 || in.Limit > maxSearchResults {
		return failure("TOOL_ARGUMENT_OUT_OF_RANGE", "limit must be between 1 and 100.")
	}
	dir, e := p.ws.resolve(in.Directory)
	i, e2 := os.Stat(dir)
	if e != nil || e2 != nil || !i.IsDir() {
		return failure("TOOL_INVALID_ARGUMENT", "directory must be an existing workspace directory.")
	}
	var paths []string
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, e error) error {
		if e != nil || d.IsDir() || d.Type()&os.ModeSymlink != 0 || !d.Type().IsRegular() {
			return nil
		}
		if in.Extension != "" && filepath.Ext(path) != in.Extension {
			return nil
		}
		if len(paths) < in.Limit {
			paths = append(paths, p.ws.show(path))
		}
		return nil
	})
	sort.Strings(paths)
	return success(map[string]any{"paths": paths, "limited": len(paths) == in.Limit})
}

func (p *ProductionTools) blender(c context.Context, id, a string) string {
	var in struct {
		ProjectPath string `json:"project_path"`
		ScriptPath  string `json:"script_path"`
		OutputPath  string `json:"output_path"`
	}
	if err := decode(a, &in); err != nil {
		return decodeFailure(err)
	}
	script, e := p.ws.resolve(in.ScriptPath, ".py")
	if e != nil || !regular(script) {
		return failure("TOOL_INVALID_ARGUMENT", "script_path must be an existing .py workspace file.")
	}
	project := ""
	if in.ProjectPath != "" {
		project, e = p.ws.resolve(in.ProjectPath, ".blend")
		if e != nil || !regular(project) {
			return failure("TOOL_INVALID_ARGUMENT", "project_path must be an existing .blend workspace file.")
		}
	}
	if id != ToolBlenderCreateProject && project == "" {
		return failure("TOOL_INVALID_ARGUMENT", "project_path is required for this operation.")
	}
	if id == ToolBlenderCreateProject && project != "" {
		return failure("TOOL_INVALID_ARGUMENT", "project_path must be omitted when creating a project.")
	}
	output := ""
	if in.OutputPath != "" {
		output, e = p.ws.resolve(in.OutputPath, ".blend")
		if e != nil {
			return failure("TOOL_INVALID_ARGUMENT", "output_path must be a workspace .blend path.")
		}
	}
	command := BlenderCommand(p.blenderBinary, p.ws.root, project, script, output)
	args := command.Args
	return p.once(c, digest(id+strings.Join(args, "\x00")), func() string {
		r, e := p.executor.Run(c, command)
		if e != nil || r.ExitCode != 0 {
			return failure("TOOL_EXECUTION_FAILED", "The external Blender command did not complete successfully.")
		}
		return success(map[string]any{"operation": id, "project_path": p.ws.show(project), "output_path": p.ws.show(output)})
	})
}
func (p *ProductionTools) ffmpegEncode(c context.Context, a string) string {
	var in struct {
		InputPath  string `json:"input_path"`
		OutputPath string `json:"output_path"`
		FPS        int    `json:"fps"`
		Codec      string `json:"codec"`
	}
	if err := decode(a, &in); err != nil {
		return decodeFailure(err)
	}
	if in.FPS < 1 || in.FPS > 120 || (in.Codec != "libx264" && in.Codec != "prores_ks") {
		return failure("TOOL_ARGUMENT_OUT_OF_RANGE", "fps or codec is outside the supported range.")
	}
	input, e := p.ws.resolve(in.InputPath)
	if e != nil || !regular(input) {
		return failure("TOOL_INVALID_ARGUMENT", "input_path must be an existing workspace file.")
	}
	output, e := p.ws.resolve(in.OutputPath, ".mp4", ".mov", ".mkv")
	if e != nil || input == output {
		return failure("TOOL_INVALID_ARGUMENT", "output_path must be a distinct supported media path.")
	}
	command := FFmpegCommand(p.ffmpeg, p.ws.root, input, output, in.FPS, in.Codec)
	return p.once(c, digest(ToolFFmpegEncode+strings.Join(command.Args, "\x00")), func() string {
		r, e := p.executor.Run(c, command)
		if e != nil || r.ExitCode != 0 {
			return failure("TOOL_EXECUTION_FAILED", "The external FFmpeg command did not complete successfully.")
		}
		return success(map[string]any{"output_path": p.ws.show(output), "codec": in.Codec, "fps": in.FPS})
	})
}
func (p *ProductionTools) ffprobeInspect(c context.Context, a string) string {
	var in struct {
		Path string `json:"path"`
	}
	if err := decode(a, &in); err != nil {
		return decodeFailure(err)
	}
	path, e := p.ws.resolve(in.Path)
	if e != nil || !regular(path) {
		return failure("TOOL_INVALID_ARGUMENT", "path must be an existing workspace file.")
	}
	command := FFprobeCommand(p.ffprobe, p.ws.root, path)
	r, e := p.executor.Run(c, command)
	if e != nil || r.ExitCode != 0 {
		return failure("TOOL_EXECUTION_FAILED", "ffprobe did not complete successfully.")
	}
	metadata, err := ParseFFprobeOutput(r.Stdout)
	if err != nil {
		return failure("TOOL_EXECUTION_FAILED", "ffprobe returned invalid metadata.")
	}
	metadata["path"] = p.ws.show(path)
	return success(metadata)
}

// ParseFFprobeOutput keeps the model-facing metadata small and rejects output
// that is not a JSON object containing a format object and stream list.
func ParseFFprobeOutput(raw []byte) (map[string]any, error) {
	var out struct {
		Format  map[string]any   `json:"format"`
		Streams []map[string]any `json:"streams"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.Format == nil {
		return nil, errors.New("invalid ffprobe metadata")
	}
	if out.Streams == nil {
		out.Streams = []map[string]any{}
	}
	return map[string]any{"format": out.Format, "streams": out.Streams}, nil
}

// ParseFFprobeResult is retained as a descriptive alias for integrations.
func ParseFFprobeResult(raw []byte) (map[string]any, error) { return ParseFFprobeOutput(raw) }

type idempotencyLedger struct {
	mu    sync.Mutex
	items map[string]*idempotencyEntry
}
type idempotencyEntry struct {
	fp     string
	done   chan struct{}
	result string
}

func newIdempotencyLedger() *idempotencyLedger {
	return &idempotencyLedger{items: map[string]*idempotencyEntry{}}
}
func (p *ProductionTools) once(c context.Context, fp string, run func() string) string {
	x, yes := agent.ToolExecutionFromContext(c)
	if !yes || x.IdempotencyKey == "" {
		return run()
	}
	return p.ledger.execute(c, x.IdempotencyKey, fp, run)
}

func (l *idempotencyLedger) execute(c context.Context, key, fp string, run func() string) string {
	l.mu.Lock()
	if e := l.items[key]; e != nil {
		l.mu.Unlock()
		if e.fp != fp {
			return failure("TOOL_IDEMPOTENCY_CONFLICT", "The idempotency key was used with different arguments.")
		}
		select {
		case <-c.Done():
			return failure("TOOL_CANCELLED", "The tool call was cancelled.")
		case <-e.done:
			return e.result
		}
	}
	e := &idempotencyEntry{fp: fp, done: make(chan struct{})}
	l.items[key] = e
	l.mu.Unlock()
	e.result = run()
	close(e.done)
	return e.result
}
func digest(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
