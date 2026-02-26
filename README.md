# secretexec

`secretexec` fetches a secret from AWS Secrets Manager and injects its key-value pairs as environment variables before executing a given command via `execve(2)`. The current process is replaced by the target command — no wrapper process remains.

## Requirements

- Unix-like OS
- AWS credentials available via the default credential chain (environment, `~/.aws`, IAM role, etc.)

## Usage

```
secretexec -s secret-id-or-arn /path/to/command [args...]
```

The secret must be a JSON object with string values, e.g.:

```json
{
  "DB_PASSWORD": "hunter2",
  "API_KEY": "abc123"
}
```

Each key-value pair is injected into the environment of the executed command. Keys must consist only of ASCII letters, digits, hyphens, and underscores. Values must not contain Unicode control/format characters. Invalid entries are skipped with a warning.

If the command name contains no path separator, it is resolved via `PATH` lookup.

## Example

```sh
secretexec -s myapp/production/env -- ./myapp --serve
```

## Flags

`-s` — Secret name or ARN (required)

## Notes

- Execution timeout is 10 seconds (covers secret fetch only; the spawned process is not affected).
- Binary secrets (`SecretBinary`) are not supported.
- The program is Unix-only (`//go:build unix`).

## Build

```sh
go build -o secretexec .
```
