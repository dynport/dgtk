package git

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	log.SetFlags(0)
}

type Repository struct {
	Origin    string
	LocalPath string
}

func (repo *Repository) cacheDir() string {
	return os.Getenv("HOME") + "/.dgtk/cache/git_repositories"
}

func (repo *Repository) Fetch() error {
	log.Println("fetching origin")
	_, e := repo.executeGitCommand("fetch")
	return e
}

func (repo *Repository) cachePath() string {
	return repo.cacheDir() + "/" + repo.Name()
}

func (repo *Repository) clone() error {
	log.Printf("cloning %s into %s", repo.Origin, repo.cachePath())
	cmd := exec.Command("git", "clone", "--bare", repo.Origin, repo.cachePath())
	if b, e := cmd.CombinedOutput(); e != nil {
		log.Printf("ERROR: %s", strings.TrimSpace(string(b)))
		return e
	}
	return nil
}

func (repo *Repository) Init() error {
	e := os.MkdirAll(repo.cacheDir(), 0755)
	if e != nil {
		return e
	}

	if !fileExists(repo.cachePath()) {
		if e := repo.clone(); e != nil {
			return e
		}
	} else {
		log.Printf("already cloned %s to %s", repo.Origin, repo.cachePath())
	}
	return nil
}

func (repo *Repository) createGitCommand(gitCommand ...string) *exec.Cmd {
	return exec.Command("git", append([]string{"--git-dir=" + repo.cachePath()}, gitCommand...)...)
}

func (repo *Repository) executeGitCommand(gitCommand ...string) (b []byte, e error) {
	cmd := repo.createGitCommand(gitCommand...)
	b, e = cmd.CombinedOutput()
	if e != nil {
		log.Printf("ERROR: %s", strings.TrimSpace(string(b)))
		return b, e
	}
	return b, nil
}

func (repo *Repository) MostRecentCommitFor(pattern string) (commit string, e error) {
	commits, e := repo.Commits(&CommitOptions{Limit: 1, Pattern: pattern})
	if e != nil {
		return "", e
	}
	if len(commits) == 0 {
		return "", e
	}
	return commits[0].Checksum, nil
}

var validTar = regexp.MustCompile("^([0-9a-f]{40})$")

// Writes tgz archive to the given tar writer.
func (repo *Repository) WriteArchiveToTar(revision string, w *tar.Writer) (e error) {
	if !validTar.MatchString(revision) {
		return fmt.Errorf("revision %q not valid (must be 40 digit git sha)", revision)
	}

	mtime, e := repo.DateOfRevision(revision)
	if e != nil {
		return e
	}

	e = repo.addArchiveToTar(revision, mtime, w)
	if e != nil {
		return e
	}

	return addFileToArchive("REVISION", []byte(revision), mtime, w)
}

func addFileToArchive(name string, content []byte, modTime time.Time, w *tar.Writer) error {
	e := w.WriteHeader(&tar.Header{Name: name, Size: int64(len(content)), ModTime: modTime, Mode: 0644})
	if e != nil {
		return e
	}
	_, e = w.Write(content)
	return e
}

func (repo *Repository) addArchiveToTar(revision string, mtime time.Time, w *tar.Writer) (e error) {
	filename := repo.Name() + ".tar.gz"

	buf := bytes.NewBuffer(nil)

	cmd := repo.createGitCommand("archive", "--format=tar.gz", revision)
	cmd.Stdout = buf

	if e = cmd.Run(); e != nil {
		return e
	}

	e = w.WriteHeader(&tar.Header{Name: filename, Size: int64(buf.Len()), ModTime: mtime, Mode: 0644})
	if e != nil {
		return e
	}

	_, e = io.Copy(w, buf)
	return e
}

func (repo *Repository) Name() string {
	return strings.TrimSuffix(filepath.Base(repo.Origin), ".git")
}

func (repo *Repository) DateOfRevision(revision string) (time.Time, error) {
	b, e := repo.executeGitCommand("log", "-1", "--format='%ct'", revision)
	if e != nil {
		return time.Now(), e
	}
	d, e := strconv.Atoi(strings.Trim(string(b), "'\n"))
	if e != nil {
		return time.Now(), e
	}

	return time.Unix(int64(d), 0), nil
}

func (repo *Repository) Commits(options *CommitOptions) (commits []*Commit, e error) {
	if options == nil {
		options = &CommitOptions{Limit: 10}
	}
	path := repo.LocalPath
	if path == "" {
		path = repo.cachePath()
	}
	b, e := repo.executeGitCommand("log", "-n", strconv.Itoa(options.Limit), "--pretty=format:'%%H\t%%at\t%%s'", options.Pattern)
	if e != nil {
		return nil, e
	}
	lines := strings.Split(string(b), "\n")
	commits = make([]*Commit, 0, len(lines))

	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) == 3 {
			if t, e := strconv.ParseInt(parts[1], 10, 64); e == nil {
				commits = append(commits, &Commit{Checksum: parts[0], AuthorDate: time.Unix(t, 0), Message: parts[2]})
			} else {
				log.Printf("ERROR: %s", e.Error())
			}
		}
	}
	return commits, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}
