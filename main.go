//go:build unix

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("secretexec: ")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var secretID string
	flag.StringVar(&secretID, "s", secretID, "Secrets Manager secret name or ARN")
	flag.Parse()
	if err := run(ctx, secretID, flag.Args()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, secretID string, rest []string) error {
	if secretID == "" {
		return errors.New("secret id must be set")
	}
	if len(rest) == 0 {
		return errors.New("need a command to execute")
	}

	if !strings.ContainsRune(rest[0], os.PathSeparator) {
		if name, err := exec.LookPath(rest[0]); err != nil {
			return err
		} else {
			rest[0] = name
		}
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	svc := secretsmanager.NewFromConfig(cfg)

	secret, err := svc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &secretID})
	if err != nil {
		return err
	}
	if secret.SecretString == nil {
		return errors.New("secret.SecretString is nil (not set)")
	}
	var m map[string]any // to gracefully handle entries code doesn't support
	if err := json.Unmarshal([]byte(*secret.SecretString), &m); err != nil {
		return err
	}

	env := os.Environ()

	for _, k := range slices.Sorted(maps.Keys(m)) {
		if !validKey(k) {
			log.Printf("key %q is not valid for KEY=VALUE format, skipping", k)
			continue
		}
		switch v := m[k].(type) {
		case string:
			if !validValue(v) {
				log.Printf("key %q value is not valid for KEY=VALUE format, skipping", k)
				continue
			}
			env = append(env, k+"="+v)
		default:
			log.Printf("key %q value is of unsupported type, skipping", k)
			continue
		}
	}
	return syscall.Exec(rest[0], rest, env)
}

func validValue(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Other, r) {
			return false
		}
	}
	return true
}

func validKey(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case 'A' <= r && r <= 'Z':
		case 'a' <= r && r <= 'z':
		case r == '-' || r == '_':
		case '0' <= r && r <= '9':
		default:
			return false
		}
	}
	return true
}

func init() {
	flag.Usage = func() {
		const usage = "usage: secretexec -s secret-id-or-arn /path/to/command -command -args\n"
		flag.CommandLine.Output().Write([]byte(usage))
		flag.PrintDefaults()
	}
}
