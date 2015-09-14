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
	// -f is required to make sure force pushes will properly update.
	_, e := repo.executeGitCommand("fetch", "-f")
	return e
}

const githubPrefix = "git@github.com:"

func (repo *Repository) clone() error {
	logger.Printf("cloning %s into %s", repo.Origin, repo.localPath())
	cmd := exec.Command("git", "clone", "--bare", repo.Origin, repo.localPath())
	if b, e := cmd.CombinedOutput(); e != nil {
		logger.Printf("ERROR: %s", strings.TrimSpace(string(b)))
		return e
	}
	cmd = exec.Command("git", "--git-dir="+repo.localPath(), "config", "remote.origin.fetch", "refs/heads/*:refs/heads/*")
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

	if !fileExists(repo.localPath()) {
		if e := repo.clone(); e != nil {
			return e
		}
	} else {
		logger.Printf("already cloned %s to %s", repo.Origin, repo.localPath())
	}
	return nil
}

func (repo *Repository) createGitCommand(gitCommand ...string) *exec.Cmd {
	return exec.Command("git", append([]string{"--git-dir=" + repo.localPath()}, gitCommand...)...)
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

// write tar.gz archive to writer
func (repo *Repository) Archive(revision string, w io.Writer, files ...string) (int64, error) {
	if err := repo.Init(); err != nil {
		return 0, err
	}
	args := append([]string{"archive", "--format=tar.gz", revision}, files...)
	cmd := repo.createGitCommand(args...)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	cnt := &counter{}
	cmd.Stdout = io.MultiWriter(w, cnt)
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("%s: %s", err, stderr.String())
	}
	return cnt.cnt, nil
}

type counter struct {
	cnt int64
}

func (c *counter) Write(b []byte) (int, error) {
	c.cnt += int64(len(b))
	return len(b), nil

}

func (repo *Repository) addArchiveToTar(revision string, mtime time.Time, w *tar.Writer) (e error) {
	filename := repo.Name() + ".tar.gz"

	buf := &bytes.Buffer{}

	if _, err := repo.Archive(revision, buf); err != nil {
		return err
	}

	e = w.WriteHeader(&tar.Header{Name: filename, Size: int64(buf.Len()), ModTime: mtime, Mode: 0644})
	if e != nil {
		return e
	}
	_, err := io.Copy(w, buf)
	return err
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

func (repo *Repository) localPath() string {
	if repo.LocalPath == "" {
		name := repo.Name()
		if strings.HasPrefix(repo.Origin, githubPrefix) {
			name = strings.TrimPrefix(repo.Origin, githubPrefix)
			repo.LocalPath = repo.cacheDir() + "/github.com/" + name
		} else {
			repo.LocalPath = repo.cacheDir() + "/" + repo.Name()
		}
	}
	return repo.LocalPath
}

func (repo *Repository) Commits(options *CommitOptions) (commits []*Commit, e error) {
	if options == nil {
		options = &CommitOptions{Limit: 10}
	}

	parts := []string{"log", "-n", strconv.Itoa(options.Limit), "--pretty=format:%H\t%at\t%s"}
	if options.Pattern != "" {
		parts = append(parts, options.Pattern)
	}
	b, e := repo.executeGitCommand(parts...)
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
