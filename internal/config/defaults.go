package config

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Services: map[string]Service{
			"playwright": {
				Enabled: true,
				Image:   "mcr.microsoft.com/playwright:v1.40.0",
				Ports:   []string{"8381:3000"},
				Environment: map[string]string{
					"PLAYWRIGHT_BROWSERS_PATH": "/ms-playwright",
				},
			},
			"postgres": {
				Enabled: true,
				Image:   "postgres:16-alpine",
				Environment: map[string]string{
					"POSTGRES_PASSWORD": "dev",
					"POSTGRES_USER":     "dev",
					"POSTGRES_DB":       "dev",
				},
				Volumes: []string{"rize-postgres:/var/lib/postgresql/data"},
				HealthCheck: &HealthCheck{
					Test:     []string{"CMD-SHELL", "pg_isready -U dev"},
					Interval: "5s",
					Timeout:  "5s",
					Retries:  5,
				},
			},
			"redis": {
				Enabled: true,
				Image:   "redis:7-alpine",
				Volumes: []string{"rize-redis:/data"},
				HealthCheck: &HealthCheck{
					Test:     []string{"CMD", "redis-cli", "ping"},
					Interval: "5s",
					Timeout:  "3s",
					Retries:  5,
				},
			},
			"mitmproxy": {
				Enabled: true,
				Image:   "mitmproxy/mitmproxy:latest",
				Ports:   []string{"8080:8080", "8081:8081"},
				Volumes: []string{"rize-mitmproxy:/home/mitmproxy/.mitmproxy"},
				Command: []string{
					"/bin/sh",
					"-c",
					`cat > /tmp/rize-noauth.py <<'PY'
from mitmproxy import ctx

class DisableWebAuth:
    def running(self):
        app = getattr(ctx.master, "app", None)
        if app:
            app.settings["is_valid_password"] = lambda _password: True

addons = [DisableWebAuth()]
PY
exec mitmweb --web-host 0.0.0.0 --set block_global=false --set web_password= --set web_open_browser=false -s /tmp/rize-noauth.py`,
				},
			},
		},
		Environment: map[string]string{
			"ANTHROPIC_API_KEY": "",
			"OPENAI_API_KEY":    "",
			"GOOGLE_API_KEY":    "",
		},
		Network: NetworkConfig{
			Name:   "rize",
			Driver: "bridge",
		},
		Volumes: []string{
			"rize-postgres",
			"rize-redis",
			"rize-mitmproxy",
		},
	}
}
