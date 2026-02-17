package logs

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Alijeyrad/simorq_backend/config"
)

// lokiWriter implements io.Writer that pushes JSON log lines to Loki's push API.
// Each Write() call is one log line. This keeps dependencies minimal.
type lokiWriter struct {
	endpoint string
	username string
	password string
	client   *http.Client
	labels   string // e.g. `{service="simorq_backend",env="production"}`
}

func newLokiHandler(cfg *config.Config, level slog.Level) slog.Handler {
	lw := &lokiWriter{
		endpoint: cfg.Logging.Output.Loki.Endpoint + "/loki/api/v1/push",
		username: cfg.Logging.Output.Loki.Username,
		password: cfg.Logging.Output.Loki.Password,
		client:   &http.Client{Timeout: 3 * time.Second},
		labels:   fmt.Sprintf(`{service="%s",env="%s"}`, cfg.Observability.ServiceName, cfg.Server.Environment),
	}
	return slog.NewJSONHandler(lw, &slog.HandlerOptions{Level: level})
}

func (lw *lokiWriter) Write(p []byte) (n int, err error) {
	now := fmt.Sprintf("%d", time.Now().UnixNano())
	line := strings.TrimRight(string(p), "\n")

	payload := fmt.Sprintf(`{"streams":[{"stream":%s,"values":[["%s",%q]]}]}`,
		// convert {k="v"} label syntax to JSON object for the stream field
		lokiLabelsToJSON(lw.labels),
		now,
		line,
	)

	req, err := http.NewRequest(http.MethodPost, lw.endpoint, strings.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if lw.username != "" {
		req.SetBasicAuth(lw.username, lw.password)
	}

	resp, err := lw.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return len(p), nil
}

func lokiLabelsToJSON(labels string) string {
	// {service="foo",env="bar"} -> {"service":"foo","env":"bar"}
	s := strings.TrimPrefix(labels, "{")
	s = strings.TrimSuffix(s, "}")
	parts := strings.Split(s, ",")
	var sb strings.Builder
	sb.WriteString("{")
	for i, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(`"` + strings.TrimSpace(kv[0]) + `":` + kv[1])
	}
	sb.WriteString("}")
	return sb.String()
}
