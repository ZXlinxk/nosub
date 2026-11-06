package schema

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func dumpCommand(target string) (bin string, args, env []string, err error) {
	u, err := url.Parse(target)
	if err != nil {
		// The parse error quotes the whole URL, password included.
		return "", nil, nil, errors.New("invalid connection url")
	}
	password, _ := u.User.Password()

	switch u.Scheme {
	case "postgres", "postgresql":
		bin = "pg_dump"
		// The password stays out of argv, where other users could read it.
		if password != "" {
			env = []string{"PGPASSWORD=" + password}
			u.User = url.User(u.User.Username())
		}
		return bin, []string{"--schema-only", "--no-owner", "--no-privileges", u.String()}, env, nil

	case "mysql":
		bin = "mysqldump"
		args = []string{"--no-data", "--skip-comments", "--single-transaction"}

		if host := u.Hostname(); host != "" {
			args = append(args, "-h", host)
		}
		if port := u.Port(); port != "" && port != "3306" {
			args = append(args, "-P", port)
		}
		if u.User != nil {
			if user := u.User.Username(); user != "" {
				args = append(args, "-u", user)
			}
		}
		db := strings.TrimPrefix(u.Path, "/")
		if db != "" {
			args = append(args, db)
		}

		if password != "" {
			env = []string{"MYSQL_PWD=" + password}
		}
		return bin, args, env, nil

	default:
		return "", nil, nil, fmt.Errorf("unsupported database scheme: %s", u.Scheme)
	}
}

// LoadReader parses SQL text from r.
func LoadReader(r io.Reader) (*Schema, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseDDL(string(data))
}

// Load loads a schema from a database connection string or a file.
func Load(target string) (*Schema, error) {
	if strings.HasPrefix(target, "postgres://") ||
		strings.HasPrefix(target, "postgresql://") ||
		strings.HasPrefix(target, "mysql://") {
		bin, args, env, err := dumpCommand(target)
		if err != nil {
			return nil, err
		}
		if _, err := exec.LookPath(bin); err != nil {
			return nil, fmt.Errorf("%s not found in PATH, run it yourself and pass the .sql file", bin)
		}
		cmd := exec.Command(bin, args...)
		if len(env) > 0 {
			cmd.Env = append(os.Environ(), env...)
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("%s: %w: %s", bin, err, stderr.String())
		}
		return ParseDDL(stdout.String())
	}

	f, err := os.Open(target)
	if err != nil {
		return nil, err
	}
	var hdr [16]byte
	n, _ := io.ReadFull(f, hdr[:])
	f.Close()

	if n == 16 && string(hdr[:]) == "SQLite format 3\x00" {
		return ReadSQLite(target)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		return nil, err
	}
	return ParseDDL(string(data))
}
