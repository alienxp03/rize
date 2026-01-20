package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alienxp03/rize/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
)

const (
	ImageName      = "alienxp03/rize:latest"
	ContainerHome  = "/home/agent"
	ClaudeConfigDir = "/home/agent/.agents/claude"
)

// RunContainer runs the rize container with the given command
func (c *Client) RunContainer(cfg *config.Config, cmd []string, interactive bool) error {
	// Ensure image exists
	if err := c.ensureImage(); err != nil {
		return err
	}

	// Build container config
	containerConfig, hostConfig, networkConfig := c.buildContainerConfigs(cfg, cmd)

	// Create container
	resp, err := c.cli.ContainerCreate(
		c.ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		"",
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Remove container when done
	defer c.cli.ContainerRemove(c.ctx, resp.ID, container.RemoveOptions{Force: true})

	// Start container
	if err := c.cli.ContainerStart(c.ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Attach to container
	if interactive {
		return c.attachInteractive(resp.ID)
	}

	return c.waitContainer(resp.ID)
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
func (c *Client) buildContainerConfigs(cfg *config.Config, cmd []string) (*container.Config, *container.HostConfig, *network.NetworkingConfig) {
	// Get current directory
	cwd, _ := os.Getwd()
	projectName := filepath.Base(cwd)
	projectDir := projectName
	workspaceDir := fmt.Sprintf("/workspace/%s", projectDir)

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
	for name, svc := range cfg.Services {
		if svc.Enabled {
			switch name {
			case "postgres":
				user := svc.Environment["POSTGRES_USER"]
				password := svc.Environment["POSTGRES_PASSWORD"]
				db := svc.Environment["POSTGRES_DB"]
				env = append(env, fmt.Sprintf("DATABASE_URL=postgresql://%s:%s@%s:5432/%s", user, password, name, db))
			case "redis":
				env = append(env, fmt.Sprintf("REDIS_URL=redis://%s:6379", name))
			case "playwright":
				env = append(env, fmt.Sprintf("PLAYWRIGHT_URL=http://%s:3000", name))
			case "mitmproxy":
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
		".ssh/known_hosts": ".ssh/known_hosts",
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

	// .rize directory
	rizePath := filepath.Join(home, ".rize")
	os.MkdirAll(rizePath, 0755)
	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: rizePath,
		Target: filepath.Join(ContainerHome, ".local/share/rize"),
	})

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
		Cmd:          cmd,
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
		NetworkMode: container.NetworkMode(cfg.Network.Name),
	}

	// Network config
	networkConfig := &network.NetworkingConfig{}

	return containerConfig, hostConfig, networkConfig
}

// attachInteractive attaches to the container interactively
func (c *Client) attachInteractive(containerID string) error {
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
	go stdcopy.StdCopy(os.Stdout, os.Stderr, attachResp.Reader)

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
