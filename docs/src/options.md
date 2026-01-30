# Additional options

## Daemon

```yaml
daemon:
  debounce: 500ms
```

| property | type | default | description |
|----------|------|---------|-------------|
| `debounce` | duration | `500ms` | Wait for filesystem activity to settle before executing rules |

The debounce prevents rapid re-execution when files are being written or modified in quick succession. Decreasing it will make rule invocations more responsive, but may reduce performance.

## Logging

```yaml
logging:
  level: warn
```

| property | type | default | description |
|----------|------|---------|-------------|
| `level` | string | `warn` | Log level: `debug`, `info`, `warn`, `error` |
