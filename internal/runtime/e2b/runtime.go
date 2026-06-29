package e2b

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"dunk/internal/runtime"
)

type Runtime struct {
	api apiClient
	cli cliBridge
}

func New() *Runtime {
	return &Runtime{api: newAPIClient(os.Getenv("E2B_API_KEY")), cli: cliBridge{}}
}

func (r *Runtime) Name() string { return "e2b" }

func (r *Runtime) Capabilities(context.Context) runtime.RuntimeCapabilities {
	return runtime.RuntimeCapabilities{
		CanCreate:      true,
		CanStop:        true,
		CanPersist:     true,
		CanAttachTTY:   true,
		CanDetach:      true,
		CanReattach:    true,
		CanUploadFiles: true,
		MaxLifetime:    24 * time.Hour,
	}
}

func (r *Runtime) Ensure(ctx context.Context, req runtime.EnsureRequest) (*runtime.Workspace, error) {
	var st providerState
	if len(req.State) > 0 {
		_ = json.Unmarshal(req.State, &st)
	}
	if st.SandboxID != "" {
		if sb, err := r.api.connectSandbox(ctx, st.SandboxID, int(req.Timeout.Seconds())); err == nil {
			return workspace(req.ProjectName, req.Workdir, providerState{SandboxID: st.SandboxID, EnvdAccessToken: sb.EnvdAccessToken})
		}
	}
	sb, err := r.api.createSandbox(ctx, req.Template, int(req.Timeout.Seconds()), req.Metadata)
	if err != nil {
		return nil, err
	}
	return workspace(req.ProjectName, req.Workdir, providerState{SandboxID: sb.SandboxID, EnvdAccessToken: sb.EnvdAccessToken})
}

func (r *Runtime) Stop(ctx context.Context, ws *runtime.Workspace, _ runtime.StopRequest) error {
	return r.api.deleteSandbox(ctx, ws.ID)
}

func (r *Runtime) Push(ctx context.Context, ws *runtime.Workspace, manifest runtime.TransferManifest) error {
	for i, item := range manifest.Items {
		if err := r.api.uploadFile(ctx, ws, item.LocalPath, item.RemotePath); err != nil {
			return uploadError(item.LocalPath, i+1, len(manifest.Items), err)
		}
	}
	return nil
}

func (r *Runtime) Run(ctx context.Context, ws *runtime.Workspace, cmd runtime.CommandSpec) (*runtime.CommandResult, error) {
	return r.cli.run(ctx, ws, cmd)
}

func (r *Runtime) Attach(ctx context.Context, ws *runtime.Workspace, session runtime.SessionSpec, opts runtime.AttachOptions) error {
	return r.cli.attach(ctx, ws, session, opts)
}

func workspace(name, workdir string, st providerState) (*runtime.Workspace, error) {
	raw, err := json.Marshal(st)
	if err != nil {
		return nil, err
	}
	return &runtime.Workspace{ID: st.SandboxID, Provider: "e2b", Name: name, Workdir: workdir, ProviderState: raw}, nil
}
