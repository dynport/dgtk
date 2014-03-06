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

var logger = log.New(os.Stderr, "", 0)

type Repository struct {
	Origin    string
	LocalPath string
}

func (repo *Repository) cacheDir() string {
	return os.Getenv("HOME") + "/.dgtk/cache/git_repositories"
}

func (repo *Repository) Fetch() error {
	logger.Println("fetching origin")
	_, e := repo.executeGitCommand("fetch")
	return e
}

const githubPrefix = "git@github.com:"

func (repo *Repository) cachePath() string {
	name := repo.Name()
	if strings.HasPrefix(repo.Origin, githubPrefix) {
		name = strings.TrimPrefix(repo.Origin, githubPrefix)
		return repo.cacheDir() + "/github.com/" + name
	}
	return repo.cacheDir() + "/" + repo.Name()
}

func (repo *Repository) clone() error {
	logger.Printf("cloning %s into %s", repo.Origin, repo.cachePath())
	cmd := exec.Command("git", "clone", "--bare", repo.Origin, repo.cachePath())
	if b, e := cmd.CombinedOutput(); e != nil {
		logger.Printf("ERROR: %s", strings.TrimSpace(string(b)))
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
		logger.Printf("already cloned %s to %s", repo.Origin, repo.cachePath())
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
		logger.Printf("ERROR: %s (%v)", strings.TrimSpace(string(b)), cmd)
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

	mtime, e := repo.DateOf(revision, ".")
	if e != nil {
		return e
	}

	e = repo.addArchiveToTar(revision, mtime, w)
	if e != nil {
		return e
	}

	return addFileToArchive("REVISION", []byte(revision), mtime, w)
}

func (repo *Repository) WriteFilesToTar(revision string, w *tar.Writer, files ...string) (e error) {
	if len(files) == 0 {
		return fmt.Errorf("empty file list given")
	}

	if !validTar.MatchString(revision) {
		return fmt.Errorf("revision %q not valid (must be 40 digit git sha)", revision)
	}

	for _, file := range files {
		mtime, e := repo.DateOf(revision, file)
		if e != nil {
			return e
		}

		buf, e := repo.getFileAtRevision(revision, file)
		if e != nil {
			return e
		}

		if e = addFileToArchive(file, buf, mtime, w); e != nil {
			return e
		}
	}

	return nil
}

func (repo *Repository) getFileAtRevision(revision, file string) (content []byte, e error) {
	buf := bytes.NewBuffer(nil)

	cmd := repo.createGitCommand("show", revision+":"+file)
	cmd.Stdout = buf

	if e = cmd.Run(); e != nil {
		return nil, e
	}

	return buf.Bytes(), nil
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

func (repo *Repository) DateOf(revision, file string) (time.Time, error) {
	b, e := repo.executeGitCommand("log", "-1", "--format='%ct'", revision, "--", file)
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
				logger.Printf("ERROR: %s", e.Error())
			}
		}
	}
	return commits, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}
