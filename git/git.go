package git

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MeredithCorpOSS/ape-dev-rt/ui"
	"github.com/mitchellh/go-homedir"
)

var dateLayout = "2006-01-02 15:04:05 -0700"

const AppConfigOrgName = "MeredithCorpOSS"
const AppConfigRepoName = "ape-dev-rt-apps"
const AppConfigRepoUrl = "git@github.com:" + AppConfigOrgName + "/" + AppConfigRepoName + ".git"
const GitLogTimezone = "Local"

var GitLogFormat = "format:" + strings.Join([]string{
	"%h",  // abbreviated commit hash
	"%s",  // subject
	"%an", // author name
	"%ae", // author email
	"%ad", // author date (format respects --date= option)
	"%cn", // committer name
	"%ce", // committer email
	"%cd", // committer date (format respects --date= option)
}, "%n") + "%n%n"

type Git struct {
	RepositoryURL  string
	RepositoryPath string
	UserInterface  *ui.StreamedUi
}

type Directory struct {
	Name       string
	LastCommit *GitCommit
	GitError   error
}
type GitCommit struct {
	AbbreviatedSHA string
	Message        string

	AuthorName     string
	AuthorEmail    string
	AuthorshipDate time.Time

	CommitterName  string
	CommitterEmail string
	CommitDate     time.Time
}

type CommitSorter []*GitCommit

func (a CommitSorter) Len() int           { return len(a) }
func (a CommitSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommitSorter) Less(i, j int) bool { return a[i].CommitDate.Before(a[j].CommitDate) }

func NewGit(url, path string) *Git {
	g := &Git{url, path, new(ui.StreamedUi)}
	return g
}

func GetRepositoryPath() (string, error) {
	// os.TempDir() isn't durable in OS X as an atomic unit
	// See "man confstr" -> _CS_DARWIN_USER_TEMP_DIR
	hd, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("Unable to resolve current user homedir: %s", err)
	}

	return filepath.Join(hd, ".rt", AppConfigRepoName), nil
}

func (g *Git) git(args ...string) (string, error) {
	log.Printf("Given arguments: %q\n", args)

	var arguments []string

	subCommand := args[0]

	if subcommandSupportsPathArgument(subCommand) {
		arguments = append(arguments, "-C", g.RepositoryPath)
	}
	arguments = append(arguments, args...)

	if subCommand == "clone" || subCommand == "pull" {
		arguments = append(arguments, "--progress")
		progPipe := ui.GitProgressPipe(subCommand, os.Stdout)
		g.UserInterface.ReplaceErrorWriter(progPipe)
		g.UserInterface.ReplaceOutputWriter(progPipe)
		defer progPipe.Close()
	} else {
		g.UserInterface.ReplaceErrorWriter(new(bytes.Buffer))
		g.UserInterface.ReplaceOutputWriter(new(bytes.Buffer))
	}

	log.Printf("Executing: git %q\n", arguments)

	cmd := exec.Command("git", arguments...)

	stdoutPipe, _ := cmd.StdoutPipe()
	stdout, doneChan := g.UserInterface.AttachOutputReadCloser(stdoutPipe)

	stderrPipe, _ := cmd.StderrPipe()
	g.UserInterface.AttachErrorReadCloser(stderrPipe)
	var stderr *bytes.Buffer
	stderr = g.UserInterface.ErrorBuffer

	g.UserInterface.FlushBuffers()
	cmd.Start()
	err := cmd.Wait()
	if err != nil {
		return stdout.String(), fmt.Errorf(
			"Git %q: %s\n\n%s", arguments, err.Error(), stderr.String())
	}
	<-doneChan
	return stdout.String(), nil
}

func subcommandSupportsPathArgument(subcommand string) bool {
	versionRegexp := regexp.MustCompile("^[-]{0,2}version$")
	return subcommand != "clone" && !versionRegexp.MatchString(subcommand)
}

func (g *Git) Update() (string, error) {
	log.Printf("Checking existence of %s\n", g.RepositoryPath)
	if _, err := os.Stat(g.RepositoryPath); err != nil {
		return g.git("clone", g.RepositoryURL, g.RepositoryPath)
	}

	return g.git("pull", "origin", "master")
}

func (g *Git) Checkout(ref string) (string, error) {
	return g.git("checkout", ref)
}

func (g *Git) Clean() (string, error) {
	return g.git("clean", "-xfd")
}

func (g *Git) RevParse(args []string) (string, error) {
	gitArgs := []string{"rev-parse"}
	gitArgs = append(gitArgs, args...)
	stdout, err := g.git(gitArgs...)
	return strings.Trim(stdout, "\n"), err
}

func (g *Git) GetRevision(ref string) (string, error) {
	isDirty, err := g.IsIndexDirty()
	if err != nil {
		return "", err
	}

	dirtySuffix := ""
	if isDirty {
		dirtySuffix = "+"
	}
	rev, err := g.RevParse([]string{"--short", ref})
	if err != nil {
		return "", err
	}

	return rev + dirtySuffix, nil
}

func (g *Git) IsIndexDirty() (bool, error) {
	stdout, err := g.Status([]string{"--porcelain"})
	if err != nil {
		return false, err
	}
	out := strings.Trim(stdout, "\n")

	if len(out) > 0 {
		return true, nil
	}

	return false, nil
}

func (g *Git) Status(args []string) (string, error) {
	gitArgs := []string{"status"}
	gitArgs = append(gitArgs, args...)
	stdout, err := g.git(gitArgs...)
	return strings.Trim(stdout, "\n"), err
}

func (g *Git) Reset(args []string) (string, error) {
	gitArgs := []string{"reset"}
	gitArgs = append(gitArgs, args...)
	stdout, err := g.git(gitArgs...)
	return strings.Trim(stdout, "\n"), err
}

func (g *Git) FreshCheckout(ref string) (string, error) {
	// Try cleaning in existing repository
	if _, err := os.Stat(g.RepositoryPath); err == nil {
		stdout, err := g.Clean()
		if err != nil {
			return stdout, err
		}

		stdout, err = g.Reset([]string{"--hard"})
		if err != nil {
			return stdout, err
		}
	}

	stdout, err := g.Update()
	if err != nil {
		return stdout, err
	}

	stdout, err = g.Checkout(ref)
	if err != nil {
		return stdout, err
	}

	return stdout, nil
}

func (g *Git) ListDirs() (dirs []*Directory, err error) {
	files, err := ioutil.ReadDir(g.RepositoryPath)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			lastCommit, err := g.LastCommit(file.Name())

			dir := Directory{
				Name:       file.Name(),
				LastCommit: lastCommit,
				GitError:   err,
			}
			dirs = append(dirs, &dir)
		}
	}
	return
}

func (g *Git) LastCommit(path string) (*GitCommit, error) {
	commits, err := g.Log(path, 1)
	if err != nil {
		return nil, err
	}

	if len(commits) < 1 {
		return nil, fmt.Errorf("No commits found for %q", path)
	}

	return commits[0], nil
}

func (g *Git) Log(path string, count int) ([]*GitCommit, error) {
	stdout, err := g.git("log", "--max-count="+strconv.Itoa(count), "--date=iso", "--format="+GitLogFormat, path)
	if err != nil {
		return nil, err
	}

	tz, err := time.LoadLocation(GitLogTimezone)
	if err != nil {
		return nil, fmt.Errorf("Failed loading timezone: %#v", err)
	}
	rawCommits := g.splitRawCommits(strings.TrimSpace(stdout))

	var commits []*GitCommit
	for _, rawCommit := range rawCommits {
		commit, err := g.parseCommit(rawCommit, tz)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func (g *Git) LogContainsSha(path string, sha string) (bool, error) {
	log.Printf("[DEBUG] Checking whether git log contains SHA")
	commits, err := g.Log(path, -1)
	if err != nil {
		return false, err
	}
	for _, commit := range commits {
		if commit.AbbreviatedSHA == sha {
			log.Printf("[DEBUG] Found that git log contains SHA")
			return true, nil
		}
	}
	log.Printf("[DEBUG] Found that git log does not contains SHA")
	return false, nil
}

func (g *Git) Version() (string, error) {
	stdout, err := g.git("--version")
	if err != nil {
		return stdout, err
	}
	return stdout, nil
}

func (g *Git) splitRawCommits(rawLog string) []string {
	return strings.Split(rawLog, "\n\n")
}

func (g *Git) parseCommit(rawCommit string, timeLocation *time.Location) (*GitCommit, error) {
	rawCommit = strings.TrimSpace(rawCommit)
	lines := strings.Split(rawCommit, "\n")
	authorDate, err := time.Parse(dateLayout, lines[4])
	if err != nil {
		return nil, err
	}
	commitDate, err := time.Parse(dateLayout, lines[7])
	if err != nil {
		return nil, err
	}

	commit := GitCommit{
		AbbreviatedSHA: lines[0],
		Message:        lines[1],

		AuthorName:     lines[2],
		AuthorEmail:    lines[3],
		AuthorshipDate: authorDate.In(timeLocation),

		CommitterName:  lines[5],
		CommitterEmail: lines[6],
		CommitDate:     commitDate.In(timeLocation),
	}

	return &commit, nil
}
