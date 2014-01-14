package git

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

func (repo *Repository) Clean() error {
	log.Println("cleaning repository")
	_, e := repo.executeGitCommand("clean -d -x -f")
	return e
}

func (repo *Repository) cachePath() string {
	chunks := strings.Split(repo.Origin, "/")
	return repo.cacheDir() + "/" + chunks[len(chunks)-1]
}

func (repo *Repository) clone() error {
	log.Printf("cloning %s into %s", repo.Origin, repo.cachePath())
	cmd := exec.Command("git", "clone", repo.Origin, repo.cachePath())
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

func (repo *Repository) Checkout(revision string) error {
	e := repo.Init()
	if e != nil {
		return e
	}
	log.Printf("checking out revision %q", revision)
	_, e = repo.executeGitCommand(checkoutCommand(revision))
	return e
}

func (repo *Repository) executeGitCommand(gitCommand string) (b []byte, e error) {
	cmd := exec.Command("bash", "-c", "cd "+repo.cachePath()+" && "+gitCommand)
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

func (repo *Repository) Tar(w *tar.Writer) error {
	return repo.addFileToArchive(repo.cachePath(), w)
}

func (repo *Repository) addFileToArchive(file string, w *tar.Writer) (e error) {
	if strings.Contains(file, "/.git/") {
		return nil
	}
	f, e := os.Open(file)
	if e != nil {
		return e
	}
	defer f.Close()
	stat, e := f.Stat()
	if e != nil {
		return e
	}
	header := &tar.Header{Name: "./" + strings.TrimPrefix(file, repo.cachePath()+"/"), Size: 0}
	header.ModTime = stat.ModTime()
	header.Mode = 0644
	if stat.IsDir() {
		header.Mode = 0755
		header.Typeflag = tar.TypeDir
		e := w.WriteHeader(header)
		if e != nil {
			return e
		}
		files, e := filepath.Glob(file + "/*")
		if e != nil {
			return e
		}
		for _, file := range files {
			e := repo.addFileToArchive(file, w)
			if e != nil {
				return e
			}
		}
		return nil
	} else {
		header.Size = stat.Size()
		e = w.WriteHeader(header)
		if e != nil {
			return e
		}
		_, e = io.Copy(w, f)
		return e
	}
}

func (repo *Repository) Commits(options *CommitOptions) (commits []*Commit, e error) {
	if options == nil {
		options = &CommitOptions{Limit: 10}
	}
	cmd := fmt.Sprintf("git log -n %d --pretty=format:'%%H\t%%at\t%%s' %s", options.Limit, options.Pattern)
	path := repo.LocalPath
	if path == "" {
		path = repo.cachePath()
	}
	b, e := repo.executeGitCommand(cmd)
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

func checkoutCommand(revision string) string {
	return "git fetch && git reset --hard " + revision
}
