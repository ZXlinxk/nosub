package envguard

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Finding represents a detected secret or sensitive file.
type Finding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Rule    string `json:"rule"`
	Preview string `json:"preview"`
	Commit  string `json:"commit,omitempty"`
}

// ErrHookExists is returned when an unmanaged pre-commit hook is present.
var ErrHookExists = errors.New("a different pre-commit hook already exists")

// ScanText scans the given text content for secrets.
func ScanText(path, content string) []Finding {
	var findings []Finding
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		clean := strings.TrimRight(line, "\r")
		findings = append(findings, scanLine(path, clean, i+1)...)
	}
	return findings
}

// ScanDir scans a directory recursively for secrets and exposed env files.
func ScanDir(root string) ([]Finding, error) {
	patterns := loadNosubignore(root)

	var findings []Finding
	isGit := isGitRepo(root)

	if isGit {
		trackedOut, err := gitCmd(root, "ls-files", "-z")
		if err != nil {
			return nil, err
		}
		tracked := make(map[string]bool)
		for _, f := range strings.Split(trackedOut, "\x00") {
			if f != "" {
				tracked[filepath.ToSlash(f)] = true
			}
		}

		allOut, err := gitCmd(root, "ls-files", "-co", "--exclude-standard", "-z")
		if err != nil {
			return nil, err
		}
		for _, file := range strings.Split(allOut, "\x00") {
			if file == "" {
				continue
			}
			rel := filepath.ToSlash(file)
			if shouldSkipPath(rel) || isIgnored(patterns, rel) {
				continue
			}
			if tracked[rel] && IsEnvFile(rel) {
				findings = append(findings, Finding{
					File:    rel,
					Line:    0,
					Rule:    "env-file",
					Preview: rel,
				})
			}
			full := filepath.Join(root, file)
			if skip, _ := isBinaryOrLarge(full); skip {
				continue
			}
			b, err := os.ReadFile(full)
			if err != nil {
				continue
			}
			findings = append(findings, ScanText(rel, string(b))...)
		}
	} else {
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if p == root {
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return nil
			}
			relSlash := filepath.ToSlash(rel)
			if d.IsDir() {
				if shouldSkipDir(d.Name()) || isIgnored(patterns, relSlash) {
					return filepath.SkipDir
				}
				return nil
			}
			if shouldSkipPath(relSlash) || isIgnored(patterns, relSlash) {
				return nil
			}
			if skip, _ := isBinaryOrLarge(p); skip {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			findings = append(findings, ScanText(relSlash, string(b))...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sortFindings(findings)
	return findings, nil
}

// ScanStaged scans git staged files for secrets and env files.
func ScanStaged(root string) ([]Finding, error) {
	if err := checkGit(); err != nil {
		return nil, err
	}
	patterns := loadNosubignore(root)

	out, err := gitCmd(root, "diff", "--cached", "--name-only", "--diff-filter=ACMR", "-z")
	if err != nil {
		return nil, err
	}

	var findings []Finding
	for _, file := range strings.Split(out, "\x00") {
		if file == "" {
			continue
		}
		rel := filepath.ToSlash(file)
		if shouldSkipPath(rel) || isIgnored(patterns, rel) {
			continue
		}
		if IsEnvFile(rel) {
			findings = append(findings, Finding{
				File:    rel,
				Line:    0,
				Rule:    "env-file",
				Preview: rel,
			})
		}
		showOut, err := gitCmd(root, "show", ":"+file)
		if err != nil {
			continue
		}
		raw := []byte(showOut)
		if len(raw) > maxFileSize || isBinaryBytes(raw) {
			continue
		}
		findings = append(findings, ScanText(rel, showOut)...)
	}

	sortFindings(findings)
	return findings, nil
}

// ScanHistory scans git commit history for committed secrets.
func ScanHistory(root string) ([]Finding, error) {
	if err := checkGit(); err != nil {
		return nil, err
	}
	patterns := loadNosubignore(root)

	out, err := gitCmd(root, "-c", "core.quotePath=false", "log", "-p", "--all", "--no-color", "--unified=0", "--format=commit:%H")
	if err != nil {
		return nil, err
	}

	type dedupKey struct {
		file    string
		rule    string
		preview string
	}
	seen := make(map[dedupKey]bool)

	var (
		findings    []Finding
		currCommit  string
		currFile    string
		currLine    int
		skipCurFile bool
	)

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "commit:") {
			hash := strings.TrimPrefix(line, "commit:")
			if len(hash) > 12 {
				currCommit = hash[:12]
			} else {
				currCommit = hash
			}
			continue
		}
		if strings.HasPrefix(line, "+++ b/") {
			currFile = filepath.ToSlash(strings.TrimRight(strings.TrimPrefix(line, "+++ b/"), "	"))
			skipCurFile = shouldSkipPath(currFile) || isIgnored(patterns, currFile)
			continue
		}
		if skipCurFile {
			continue
		}
		if strings.HasPrefix(line, "@@ ") {
			currLine = parseHunkLine(line)
			continue
		}
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			content := line[1:]
			matched := scanLine(currFile, content, currLine)
			for _, m := range matched {
				m.Commit = currCommit
				key := dedupKey{file: m.File, rule: m.Rule, preview: m.Preview}
				if !seen[key] {
					seen[key] = true
					findings = append(findings, m)
				}
			}
			currLine++
		}
	}

	sortFindings(findings)
	return findings, nil
}

func sortFindings(fs []Finding) {
	sort.Slice(fs, func(i, j int) bool {
		if fs[i].File != fs[j].File {
			return fs[i].File < fs[j].File
		}
		if fs[i].Line != fs[j].Line {
			return fs[i].Line < fs[j].Line
		}
		return fs[i].Rule < fs[j].Rule
	})
}
