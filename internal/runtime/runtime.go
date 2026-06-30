package runtime

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

type EnsureRequest struct {
	ProjectName string
	ProjectRoot string
	Workdir     string
	Template    string
	Timeout     time.Duration
	Metadata    map[string]string
	State       json.RawMessage
}

type StopRequest struct{}

type Workspace struct {
	ID            string          `json:"id"`
	Provider      string          `json:"provider"`
	Name          string          `json:"name"`
	Workdir       string          `json:"workdir"`
	ProviderState json.RawMessage `json:"provider_state,omitempty"`
}

type RuntimeCapabilities struct {
	CanCreate      bool
	CanStop        bool
	CanPersist     bool
	CanAttachTTY   bool
	CanDetach      bool
	CanReattach    bool
	CanUploadFiles bool
	MaxLifetime    time.Duration
}

type TransferItem struct {
	LocalPath  string
	RemotePath string
	Mode       string
	Size       int64
	Sensitive  bool
}

type TransferManifest struct {
	Items []TransferItem
}

type CommandSpec struct {
	Command string
	Workdir string
	Env     map[string]string
}

type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

type SessionSpec struct {
	Name    string
	Command string
	Workdir string
	Env     map[string]string
	TTY     bool
}

type AttachOptions struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type Runtime interface {
	Name() string
	Ensure(ctx context.Context, req EnsureRequest) (*Workspace, error)
	Stop(ctx context.Context, ws *Workspace, req StopRequest) error
	Push(ctx context.Context, ws *Workspace, manifest TransferManifest) error
	Run(ctx context.Context, ws *Workspace, cmd CommandSpec) (*CommandResult, error)
	Attach(ctx context.Context, ws *Workspace, session SessionSpec, opts AttachOptions) error
	Capabilities(ctx context.Context) RuntimeCapabilities
}
