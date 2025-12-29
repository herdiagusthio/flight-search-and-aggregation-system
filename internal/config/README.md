# Configuration

This package handles application configuration from environment variables.

## Environment Variables

| Variable                | Default | Description                    |
| ----------------------- | ------- | ------------------------------ |
| `SERVER_PORT`           | `8080`  | HTTP server port               |
| `TIMEOUT_GLOBAL_SEARCH` | `5s`    | Maximum search duration        |
| `TIMEOUT_PER_PROVIDER`  | `2s`    | Per-provider timeout           |
| `LOG_LEVEL`             | `info`  | Log level (debug/info/warn/error) |
| `LOG_FORMAT`            | `json`  | Log format (json/console)      |

## Usage

```go
cfg := config.MustLoad()
```
