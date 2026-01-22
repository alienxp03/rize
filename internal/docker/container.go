package docker

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alienxp03/rize/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/moby/term"
)

const (
	ImageName       = "alienxp03/rize:latest"
	ContainerHome   = "/home/agent"
	ClaudeConfigDir = "/home/agent/.agents/claude"
)

var defaultContainerCmd = []string{"sleep", "infinity"}

func (c *Client) isComposeServiceRunning(networkName, serviceName string) bool {
	args := filters.NewArgs()
	args.Add("status", "running")
	args.Add("label", fmt.Sprintf("com.docker.compose.service=%s", serviceName))

	if networkName != "" {
		args.Add("network", networkName)
		containers, err := c.cli.ContainerList(c.ctx, container.ListOptions{Filters: args})
		if err == nil && len(containers) > 0 {
			return true
		}
		args = filters.NewArgs()
		args.Add("status", "running")
		args.Add("label", fmt.Sprintf("com.docker.compose.service=%s", serviceName))
	}

	containers, err := c.cli.ContainerList(c.ctx, container.ListOptions{Filters: args})
	if err != nil {
		return false
	}

	return len(containers) > 0
}

func projectContainerName(cwd string) string {
	absPath, err := filepath.Abs(cwd)
	if err != nil {
		absPath = cwd
	}

	projectName := filepath.Base(absPath)
	safeName := sanitizeContainerName(projectName)
	hash := shortHash(absPath)

	return fmt.Sprintf("rize-%s-%s", safeName, hash)
}

func sanitizeContainerName(name string) string {
	name = strings.ToLower(name)
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('-')
	}

	safe := strings.Trim(b.String(), "-")
	if safe == "" {
		return "project"
	}

	return safe
}

func shortHash(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])[:6]
}

// RunContainer runs the rize container with the given command
func (c *Client) RunContainer(cfg *config.Config, cmd []string, interactive bool) error {
	// Ensure image exists
	if err := c.ensureImage(); err != nil {
		return err
	}

	// Ensure network exists
	networkCfg := config.DefaultNetworkConfig()
	if err := c.ensureNetwork(networkCfg); err != nil {
		return err
	}

	// Build container config
	containerName, workspaceDir, containerConfig, hostConfig, networkConfig := c.buildContainerConfigs(cfg)

	containerID, err := c.ensureProjectContainer(containerName, containerConfig, hostConfig, networkConfig)
	if err != nil {
		return err
	}

	if err := c.startContainerIfNeeded(containerID); err != nil {
		return err
	}

	c.ensureConnectedToServiceNetworks(containerID, cfg)

	return c.execInContainer(containerID, workspaceDir, cfg, cmd, interactive)
}

func (c *Client) ensureNetwork(netCfg config.NetworkConfig) error {
	if netCfg.Name == "" {
		return nil
	}

	_, err := c.cli.NetworkInspect(c.ctx, netCfg.Name, network.InspectOptions{})
	if err == nil {
		return nil
	}

	if !dockerclient.IsErrNotFound(err) {
		return fmt.Errorf("failed to inspect network %s: %w", netCfg.Name, err)
	}

	_, err = c.cli.NetworkCreate(c.ctx, netCfg.Name, network.CreateOptions{
		Driver: netCfg.Driver,
	})
	if err != nil {
		return fmt.Errorf("failed to create network %s: %w", netCfg.Name, err)
	}

	return nil
}

// ensureImage ensures the rize image exists locally
func (c *Client) ensureImage() error {
	_, _, err := c.cli.ImageInspectWithRaw(c.ctx, ImageName)
	if err == nil {
		return nil
	}

	// Pull image
	reader, err := c.cli.ImagePull(c.ctx, ImageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	// Wait for pull to complete
	_, err = io.Copy(os.Stdout, reader)
	return err
}

// buildContainerConfigs builds container, host, and network configurations
func (c *Client) buildContainerConfigs(cfg *config.Config) (string, string, *container.Config, *container.HostConfig, *network.NetworkingConfig) {
	// Get current directory
	cwd, _ := os.Getwd()
	absPath, err := filepath.Abs(cwd)
	if err != nil {
		absPath = cwd
	}

	projectName := filepath.Base(absPath)
	projectDir := projectName
	workspaceDir := fmt.Sprintf("/workspace/%s", projectDir)
	containerName := projectContainerName(absPath)

	// Build environment variables
	env := []string{
		fmt.Sprintf("HOST_UID=%d", os.Getuid()),
		fmt.Sprintf("HOST_GID=%d", os.Getgid()),
		fmt.Sprintf("TERM=%s", os.Getenv("TERM")),
		fmt.Sprintf("COLORTERM=%s", os.Getenv("COLORTERM")),
		fmt.Sprintf("RIZE_PROJECT_NAME=%s", projectName),
		fmt.Sprintf("RIZE_PROJECT_DIR=%s", projectDir),
		fmt.Sprintf("RIZE_WORKSPACE_DIR=%s", workspaceDir),
		fmt.Sprintf("CLAUDE_CONFIG_DIR=%s", ClaudeConfigDir),
	}

	// Add custom environment variables from config
	for key, value := range cfg.Environment {
		if value != "" {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Add service URLs for enabled services
	networkName := config.DefaultNetworkConfig().Name
	for name, enabled := range cfg.Services {
		if !enabled {
			continue
		}

		switch name {
		case "postgres":
			env = append(env, fmt.Sprintf("DATABASE_URL=postgresql://dev:dev@%s:5432/dev", name))
		case "redis":
			env = append(env, fmt.Sprintf("REDIS_URL=redis://%s:6379", name))
		case "playwright":
			env = append(env, fmt.Sprintf("PLAYWRIGHT_URL=http://%s:3000", name))
		case "mitmproxy":
			if c.isComposeServiceRunning(networkName, name) {
				env = append(env, fmt.Sprintf("HTTP_PROXY=http://%s:8080", name))
				env = append(env, fmt.Sprintf("HTTPS_PROXY=http://%s:8080", name))
				env = append(env, "NO_PROXY=localhost,127.0.0.1")
			}
		}
	}

	// Build mounts
	mounts := []mount.Mount{
		// Workspace mount
		{
			Type:   mount.TypeBind,
			Source: cwd,
			Target: workspaceDir,
		},
		// Agent volume
		{
			Type:   mount.TypeVolume,
			Source: "rize-agents",
			Target: "/home/agent/.agents",
		},
	}

	// Add home directory mounts
	home, _ := os.UserHomeDir()

	// SSH directory (default mount)
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err == nil {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: sshDir,
			Target: filepath.Join(ContainerHome, ".ssh"),
		})
	}

	// Claude directories
	claudeDirs := []string{"commands", "agents", "skills"}
	for _, dir := range claudeDirs {
		hostPath := filepath.Join(home, ".claude", dir)
		os.MkdirAll(hostPath, 0755)
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: hostPath,
			Target: filepath.Join(ClaudeConfigDir, dir),
		})
	}

	// Optional mounts
	optionalMounts := map[string]string{
		".config/opencode": ".config/opencode",
		".netrc":           ".netrc",
		".gitconfig":       ".gitconfig",
		".env":             ".env",
	}

	for hostSuffix, containerSuffix := range optionalMounts {
		hostPath := filepath.Join(home, hostSuffix)
		if _, err := os.Stat(hostPath); err == nil {
			readOnly := true
			if strings.Contains(hostSuffix, "opencode") {
				readOnly = false
			}
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   hostPath,
				Target:   filepath.Join(ContainerHome, containerSuffix),
				ReadOnly: readOnly,
			})
		}
	}

	// Per-project state directory (includes zsh history, etc.)
	// Bind mount per-project state for isolation
	stateDir, err := config.ProjectStateDir(absPath)
	if err == nil {
		os.MkdirAll(stateDir, 0755)
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: stateDir,
			Target: filepath.Join(ContainerHome, ".local/share/rize"),
		})
	}

	// Docker socket
	if runtime.GOOS != "windows" {
		dockerSock := "/var/run/docker.sock"
		if _, err := os.Stat(dockerSock); err == nil {
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: dockerSock,
				Target: dockerSock,
			})
		}
	}

	// SSH Agent forwarding
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock != "" {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: sshAuthSock,
			Target: sshAuthSock,
		})
		env = append(env, fmt.Sprintf("SSH_AUTH_SOCK=%s", sshAuthSock))
	}

	// Container config
	containerConfig := &container.Config{
		Image:        ImageName,
		Cmd:          defaultContainerCmd,
		Env:          env,
		WorkingDir:   workspaceDir,
		Tty:          true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    true,
		Entrypoint:   []string{"/usr/local/bin/entrypoint.sh"},
	}

	// Host config
	hostConfig := &container.HostConfig{
		Mounts:      mounts,
		AutoRemove:  false,
		NetworkMode: container.NetworkMode(networkName),
	}

	// Network config
	networkConfig := &network.NetworkingConfig{}

	return containerName, workspaceDir, containerConfig, hostConfig, networkConfig
}

func (c *Client) ensureProjectContainer(name string, containerConfig *container.Config, hostConfig *container.HostConfig, networkConfig *network.NetworkingConfig) (string, error) {
	inspect, err := c.cli.ContainerInspect(c.ctx, name)
	if err == nil {
		if inspect.State != nil && inspect.State.Running {
			return inspect.ID, nil
		}

		if err := c.cli.ContainerRemove(c.ctx, inspect.ID, container.RemoveOptions{Force: true}); err != nil {
			return "", fmt.Errorf("failed to remove existing container %s: %w", name, err)
		}
	} else if !dockerclient.IsErrNotFound(err) {
		return "", fmt.Errorf("failed to inspect container %s: %w", name, err)
	}

	resp, err := c.cli.ContainerCreate(
		c.ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		name,
	)
	if err != nil {
		if strings.Contains(err.Error(), "Conflict") {
			inspect, inspectErr := c.cli.ContainerInspect(c.ctx, name)
			if inspectErr == nil {
				return inspect.ID, nil
			}
		}
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return resp.ID, nil
}

func (c *Client) startContainerIfNeeded(containerID string) error {
	inspect, err := c.cli.ContainerInspect(c.ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	if inspect.State != nil && inspect.State.Running {
		return nil
	}

	if err := c.cli.ContainerStart(c.ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

func (c *Client) execInContainer(containerID, workspaceDir string, cfg *config.Config, cmd []string, interactive bool) error {
	execEnv := c.buildExecEnv(cfg)
	execCmd := []string{"/usr/local/bin/entrypoint.sh"}
	execCmd = append(execCmd, cmd...)

	execConfig := container.ExecOptions{
		Cmd:          execCmd,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  interactive,
		Tty:          interactive,
		Env:          execEnv,
		WorkingDir:   workspaceDir,
	}

	resp, err := c.cli.ContainerExecCreate(c.ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	return c.attachExec(resp.ID, execConfig.Tty, interactive)
}

func (c *Client) ensureConnectedToServiceNetworks(containerID string, cfg *config.Config) {
	inspect, err := c.cli.ContainerInspect(c.ctx, containerID)
	if err != nil {
		return
	}

	attachedNetworks := map[string]struct{}{}
	for name := range inspect.NetworkSettings.Networks {
		attachedNetworks[name] = struct{}{}
	}

	for name, enabled := range cfg.Services {
		if !enabled {
			continue
		}

		serviceID := c.findComposeServiceContainerID(name)
		if serviceID == "" {
			continue
		}

		serviceInspect, err := c.cli.ContainerInspect(c.ctx, serviceID)
		if err != nil || serviceInspect.NetworkSettings == nil {
			continue
		}

		for networkName := range serviceInspect.NetworkSettings.Networks {
			if _, exists := attachedNetworks[networkName]; exists {
				continue
			}

			if err := c.cli.NetworkConnect(c.ctx, networkName, containerID, nil); err == nil {
				attachedNetworks[networkName] = struct{}{}
			}
		}
	}
}

func (c *Client) findComposeServiceContainerID(serviceName string) string {
	args := filters.NewArgs()
	args.Add("status", "running")
	args.Add("label", fmt.Sprintf("com.docker.compose.service=%s", serviceName))

	containers, err := c.cli.ContainerList(c.ctx, container.ListOptions{Filters: args})
	if err != nil || len(containers) == 0 {
		return ""
	}

	return containers[0].ID
}

func (c *Client) buildExecEnv(cfg *config.Config) []string {
	var env []string
	networkName := config.DefaultNetworkConfig().Name
	for name, enabled := range cfg.Services {
		if !enabled {
			continue
		}
		if name == "mitmproxy" && c.isComposeServiceRunning(networkName, name) {
			env = append(env, fmt.Sprintf("HTTP_PROXY=http://%s:8080", name))
			env = append(env, fmt.Sprintf("HTTPS_PROXY=http://%s:8080", name))
			env = append(env, fmt.Sprintf("http_proxy=http://%s:8080", name))
			env = append(env, fmt.Sprintf("https_proxy=http://%s:8080", name))
			env = append(env, "NO_PROXY=localhost,127.0.0.1")
			env = append(env, "no_proxy=localhost,127.0.0.1")
		}
	}
	return env
}

func (c *Client) attachExec(execID string, tty bool, interactive bool) error {
	var oldState *term.State
	var inFd uintptr

	if tty {
		var isInTerm bool
		var outFd uintptr
		var isOutTerm bool

		inFd, isInTerm = term.GetFdInfo(os.Stdin)
		outFd, isOutTerm = term.GetFdInfo(os.Stdout)

		if isInTerm {
			state, err := term.SetRawTerminal(inFd)
			if err == nil {
				oldState = state
			}
		}
		if isOutTerm {
			_, _ = term.SetRawTerminalOutput(outFd)
		}
	}
	if oldState != nil {
		defer term.RestoreTerminal(inFd, oldState)
	}

	attachResp, err := c.cli.ContainerExecAttach(c.ctx, execID, container.ExecAttachOptions{Tty: tty})
	if err != nil {
		return fmt.Errorf("failed to attach exec: %w", err)
	}
	defer attachResp.Close()

	if interactive {
		go io.Copy(attachResp.Conn, os.Stdin)
	}

	if tty {
		go io.Copy(os.Stdout, attachResp.Reader)
	} else {
		go stdcopy.StdCopy(os.Stdout, os.Stderr, attachResp.Reader)
	}

	return c.waitExec(execID)
}

func (c *Client) waitExec(execID string) error {
	for {
		inspect, err := c.cli.ContainerExecInspect(c.ctx, execID)
		if err != nil {
			return fmt.Errorf("failed to inspect exec: %w", err)
		}

		if !inspect.Running {
			if inspect.ExitCode != 0 {
				return fmt.Errorf("command exited with code %d", inspect.ExitCode)
			}
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// attachInteractive attaches to the container interactively
func (c *Client) attachInteractive(containerID string, tty bool) error {
	var oldState *term.State
	var inFd uintptr
	if tty {
		var isInTerm bool
		var outFd uintptr
		var isOutTerm bool

		inFd, isInTerm = term.GetFdInfo(os.Stdin)
		outFd, isOutTerm = term.GetFdInfo(os.Stdout)

		if isInTerm {
			state, err := term.SetRawTerminal(inFd)
			if err == nil {
				oldState = state
			}
		}
		if isOutTerm {
			_, _ = term.SetRawTerminalOutput(outFd)
		}
	}
	if oldState != nil {
		defer term.RestoreTerminal(inFd, oldState)
	}

	// Attach to container
	attachResp, err := c.cli.ContainerAttach(c.ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	// Handle IO
	go io.Copy(attachResp.Conn, os.Stdin)
	if tty {
		go io.Copy(os.Stdout, attachResp.Reader)
	} else {
		go stdcopy.StdCopy(os.Stdout, os.Stderr, attachResp.Reader)
	}

	// Wait for container to finish
	statusCh, errCh := c.cli.ContainerWait(c.ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for container: %w", err)
		}
	case <-statusCh:
	}

	return nil
}

// waitContainer waits for the container to finish
func (c *Client) waitContainer(containerID string) error {
	statusCh, errCh := c.cli.ContainerWait(c.ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with code %d", status.StatusCode)
		}
	}

	return nil
}

// PullImage pulls the latest rize image
func (c *Client) PullImage() error {
	reader, err := c.cli.ImagePull(c.ctx, ImageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	_, err = io.Copy(os.Stdout, reader)
	return err
}

// RemoveImage removes the rize image
func (c *Client) RemoveImage() error {
	_, err := c.cli.ImageRemove(context.Background(), ImageName, image.RemoveOptions{Force: true})
	return err
}
