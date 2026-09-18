package pkg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// RenderKind identifies the artifact requested from a renderer. It deliberately
// describes the artifact, rather than an OpenSCAD command-line invocation.
type RenderKind string

const (
	RenderKindSTL   RenderKind = "stl"
	RenderKindImage RenderKind = "image"
)

// RenderCamera is only meaningful for image requests. Keeping it separate from
// RenderRequest makes it possible for a renderer without camera support to
// reject the request explicitly instead of interpreting CLI arguments.
type RenderCamera struct {
	Name        string
	Coordinates string
	ImageSize   string
}

// RenderRequest is the portable rendering boundary. It carries source bytes for
// engines without filesystem access while retaining paths for the local CLI and
// output naming. No renderer-specific flags or executable paths leak into this
// contract.
type RenderRequest struct {
	Kind       RenderKind
	SourcePath string
	// Source is the primary SCAD source for engines that do not have access to
	// the host filesystem (for example a browser/WASM worker). CLI engines may
	// use SourcePath directly. Include dependencies can be supplied by a future
	// virtual-filesystem resolver without changing the artifact contract.
	Source      []byte
	OutputPath  string
	Definitions map[string]interface{}
	Render      bool
	Debug       bool
	Camera      *RenderCamera
}

type RenderResponse struct {
	Command  string // Optional diagnostic text; not required for non-CLI renderers.
	Output   string
	Duration time.Duration
	// Artifact is optional for filesystem-native renderers. Portable engines
	// return it and the orchestration layer writes it to the requested target.
	Artifact []byte
}

func materializeRenderArtifact(outputPath string, response RenderResponse) error {
	if response.Artifact == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create render artifact directory: %w", err)
	}
	if err := os.WriteFile(outputPath, response.Artifact, 0o644); err != nil {
		return fmt.Errorf("write render artifact: %w", err)
	}
	return nil
}

// RenderEngine is the execution seam used by instance and image processing.
// It owns how a request is rendered; Process owns orchestration, reporting and
// output policy. Implementations must not shell-interpret request values.
type RenderEngine interface {
	Render(context.Context, RenderRequest) (RenderResponse, error)
}

// renderContext bridges the existing cancellation channel to context-aware
// renderers. The local CLI adapter uses it to terminate an in-flight child
// process, while remote/WASM adapters can stop their own work promptly.
func renderContext(cancel <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, stop := context.WithCancel(context.Background())
	if cancel == nil {
		return ctx, stop
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-cancel:
			stop()
		case <-done:
		}
	}()
	return ctx, func() {
		close(done)
		stop()
	}
}

// OpenSCADRenderer adapts the typed render contract to the local OpenSCAD CLI.
// It is intentionally the only place that knows OpenSCAD command-line flags.
type OpenSCADRenderer struct {
	Executable          string
	AdditionalArguments []string
}

func (r OpenSCADRenderer) Render(ctx context.Context, request RenderRequest) (RenderResponse, error) {
	if r.Executable == "" {
		return RenderResponse{}, fmt.Errorf("openscad command not found")
	}
	args, err := openSCADArguments(request)
	if err != nil {
		return RenderResponse{}, err
	}
	args = append(args[:2], append(r.AdditionalArguments, args[2:]...)...)
	start := time.Now()
	cmd := exec.CommandContext(ctx, r.Executable, args...)
	cmd.Env = managedOpenSCADEnvironment()
	output, err := cmd.CombinedOutput()
	response := RenderResponse{
		Command:  displayCommand(r.Executable, args),
		Output:   string(output),
		Duration: time.Since(start),
	}
	if err != nil {
		return response, fmt.Errorf("OpenSCAD command failed: %w", err)
	}
	return response, nil
}

func openSCADArguments(request RenderRequest) ([]string, error) {
	args := []string{"-o", request.OutputPath}
	switch request.Kind {
	case RenderKindSTL:
		if request.Render {
			args = append(args, "--render")
		}
		if request.Debug {
			args = append(args, "--debug", "all", "--summary", "all")
		}
	case RenderKindImage:
		if request.Camera == nil {
			return nil, fmt.Errorf("image render request is missing camera settings")
		}
		imageSize := request.Camera.ImageSize
		if imageSize == "" {
			imageSize = "1920,1080"
		}
		args = append(args, "--imgsize", imageSize, "--projection", "perspective", "--camera", request.Camera.Coordinates, "--preview")
		request.Definitions = copyDefinitions(request.Definitions)
		request.Definitions["genType"] = "image"
	default:
		return nil, fmt.Errorf("unsupported render kind %q", request.Kind)
	}
	var err error
	args, err = appendOpenSCADDefinitions(args, request.Definitions)
	if err != nil {
		return nil, err
	}
	return append(args, request.SourcePath), nil
}

func copyDefinitions(definitions map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(definitions)+1)
	for key, value := range definitions {
		copy[key] = value
	}
	return copy
}
