package allocdir

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/tomb.v1"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/hpcloud/tail/watch"
)

var (
	// The name of the directory that is shared across tasks in a task group.
	SharedAllocName = "alloc"

	// Name of the directory where logs of Tasks are written
	LogDirName = "logs"

	// The set of directories that exist inside eache shared alloc directory.
	SharedAllocDirs = []string{LogDirName, "tmp", "data"}

	// The name of the directory that exists inside each task directory
	// regardless of driver.
	TaskLocal = "local"

	// TaskSecrets is the name of the secret directory inside each task
	// directory
	TaskSecrets = "secrets"

	// TaskDirs is the set of directories created in each tasks directory.
	TaskDirs = []string{"tmp"}
)

type AllocDir struct {
	// AllocDir is the directory used for storing any state
	// of this allocation. It will be purged on alloc destroy.
	AllocDir string

	// The shared directory is available to all tasks within the same task
	// group.
	SharedDir string

	// TaskDirs is a mapping of task names to their non-shared directory.
	TaskDirs map[string]string
}

// AllocFileInfo holds information about a file inside the AllocDir
type AllocFileInfo struct {
	Name     string
	IsDir    bool
	Size     int64
	FileMode string
	ModTime  time.Time
}

// AllocDirFS exposes file operations on the alloc dir
type AllocDirFS interface {
	List(path string) ([]*AllocFileInfo, error)
	Stat(path string) (*AllocFileInfo, error)
	ReadAt(path string, offset int64) (io.ReadCloser, error)
	Snapshot(w io.Writer) error
	BlockUntilExists(path string, t *tomb.Tomb) (chan error, error)
	ChangeEvents(path string, curOffset int64, t *tomb.Tomb) (*watch.FileChanges, error)
}

// NewAllocDir initializes the AllocDir struct with allocDir as base path for
// the allocation directory.
func NewAllocDir(allocDir string) *AllocDir {
	d := &AllocDir{
		AllocDir: allocDir,
		TaskDirs: make(map[string]string),
	}
	d.SharedDir = filepath.Join(d.AllocDir, SharedAllocName)
	return d
}

// Snapshot creates an archive of the files and directories in the data dir of
// the allocation and the task local directories
func (d *AllocDir) Snapshot(w io.Writer) error {
	allocDataDir := filepath.Join(d.SharedDir, "data")
	rootPaths := []string{allocDataDir}
	for _, path := range d.TaskDirs {
		taskLocaPath := filepath.Join(path, "local")
		rootPaths = append(rootPaths, taskLocaPath)
	}

	tw := tar.NewWriter(w)
	defer tw.Close()

	walkFn := func(path string, fileInfo os.FileInfo, err error) error {
		// Ignore if the file is a symlink
		if fileInfo.Mode() == os.ModeSymlink {
			return nil
		}

		// Include the path of the file name relative to the alloc dir
		// so that we can put the files in the right directories
		relPath, err := filepath.Rel(d.AllocDir, path)
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return fmt.Errorf("error creating file header: %v", err)
		}
		hdr.Name = relPath
		tw.WriteHeader(hdr)

		// If it's a directory we just write the header into the tar
		if fileInfo.IsDir() {
			return nil
		}

		// Write the file into the archive
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
		return nil
	}

	// Walk through all the top level directories and add the files and
	// directories in the archive
	for _, path := range rootPaths {
		if err := filepath.Walk(path, walkFn); err != nil {
			return err
		}
	}

	return nil
}

// Move moves the shared data and task local dirs
func (d *AllocDir) Move(other *AllocDir, tasks []*structs.Task) error {
	// Move the data directory
	otherDataDir := filepath.Join(other.SharedDir, "data")
	dataDir := filepath.Join(d.SharedDir, "data")
	if fileInfo, err := os.Stat(otherDataDir); fileInfo != nil && err == nil {
		if err := os.Rename(otherDataDir, dataDir); err != nil {
			return fmt.Errorf("error moving data dir: %v", err)
		}
	}

	// Move the task directories
	for _, task := range tasks {
		taskDir := filepath.Join(other.AllocDir, task.Name)
		otherTaskLocal := filepath.Join(taskDir, TaskLocal)

		if fileInfo, err := os.Stat(otherTaskLocal); fileInfo != nil && err == nil {
			if taskDir, ok := d.TaskDirs[task.Name]; ok {
				if err := os.Rename(otherTaskLocal, filepath.Join(taskDir, TaskLocal)); err != nil {
					return fmt.Errorf("error moving task local dir: %v", err)
				}
			}
		}
	}

	return nil
}

// Tears down previously build directory structure.
func (d *AllocDir) Destroy() error {

	// Unmount all mounted shared alloc dirs.
	var mErr multierror.Error
	if err := d.UnmountAll(); err != nil {
		mErr.Errors = append(mErr.Errors, err)
	}

	if err := os.RemoveAll(d.AllocDir); err != nil {
		mErr.Errors = append(mErr.Errors, err)
	}

	return mErr.ErrorOrNil()
}

func (d *AllocDir) UnmountAll() error {
	var mErr multierror.Error
	for _, dir := range d.TaskDirs {
		// Check if the directory has the shared alloc mounted.
		taskAlloc := filepath.Join(dir, SharedAllocName)
		if d.pathExists(taskAlloc) {
			if err := d.unmountSharedDir(taskAlloc); err != nil {
				mErr.Errors = append(mErr.Errors,
					fmt.Errorf("failed to unmount shared alloc dir %q: %v", taskAlloc, err))
			} else if err := os.RemoveAll(taskAlloc); err != nil {
				mErr.Errors = append(mErr.Errors,
					fmt.Errorf("failed to delete shared alloc dir %q: %v", taskAlloc, err))
			}
		}

		taskSecret := filepath.Join(dir, TaskSecrets)
		if d.pathExists(taskSecret) {
			if err := d.removeSecretDir(taskSecret); err != nil {
				mErr.Errors = append(mErr.Errors,
					fmt.Errorf("failed to remove the secret dir %q: %v", taskSecret, err))
			}
		}

		// Unmount dev/ and proc/ have been mounted.
		d.unmountSpecialDirs(dir)
	}

	return mErr.ErrorOrNil()
}

// Given a list of a task build the correct alloc structure.
func (d *AllocDir) Build(tasks []*structs.Task) error {
	// Make the alloc directory, owned by the nomad process.
	if err := os.MkdirAll(d.AllocDir, 0755); err != nil {
		return fmt.Errorf("Failed to make the alloc directory %v: %v", d.AllocDir, err)
	}

	// Make the shared directory and make it available to all user/groups.
	if err := os.MkdirAll(d.SharedDir, 0777); err != nil {
		return err
	}

	// Make the shared directory have non-root permissions.
	if err := d.dropDirPermissions(d.SharedDir); err != nil {
		return err
	}

	for _, dir := range SharedAllocDirs {
		p := filepath.Join(d.SharedDir, dir)
		if err := os.MkdirAll(p, 0777); err != nil {
			return err
		}
		if err := d.dropDirPermissions(p); err != nil {
			return err
		}
	}

	// Make the task directories.
	for _, t := range tasks {
		taskDir := filepath.Join(d.AllocDir, t.Name)
		if err := os.MkdirAll(taskDir, 0777); err != nil {
			return err
		}

		// Make the task directory have non-root permissions.
		if err := d.dropDirPermissions(taskDir); err != nil {
			return err
		}

		// Create a local directory that each task can use.
		local := filepath.Join(taskDir, TaskLocal)
		if err := os.MkdirAll(local, 0777); err != nil {
			return err
		}

		if err := d.dropDirPermissions(local); err != nil {
			return err
		}

		d.TaskDirs[t.Name] = taskDir

		// Create the directories that should be in every task.
		for _, dir := range TaskDirs {
			local := filepath.Join(taskDir, dir)
			if err := os.MkdirAll(local, 0777); err != nil {
				return err
			}

			if err := d.dropDirPermissions(local); err != nil {
				return err
			}
		}

		// Create the secret directory
		secret := filepath.Join(taskDir, TaskSecrets)
		if err := d.createSecretDir(secret); err != nil {
			return err
		}

		if err := d.dropDirPermissions(secret); err != nil {
			return err
		}
	}

	return nil
}

// Embed takes a mapping of absolute directory or file paths on the host to
// their intended, relative location within the task directory. Embed attempts
// hardlink and then defaults to copying. If the path exists on the host and
// can't be embedded an error is returned.
func (d *AllocDir) Embed(task string, entries map[string]string) error {
	taskdir, ok := d.TaskDirs[task]
	if !ok {
		return fmt.Errorf("Task directory doesn't exist for task %v", task)
	}

	subdirs := make(map[string]string)
	for source, dest := range entries {
		// Check to see if directory exists on host.
		s, err := os.Stat(source)
		if os.IsNotExist(err) {
			continue
		}

		// Embedding a single file
		if !s.IsDir() {
			if err := d.createDir(taskdir, filepath.Dir(dest)); err != nil {
				return fmt.Errorf("Couldn't create destination directory %v: %v", dest, err)
			}

			// Copy the file.
			taskEntry := filepath.Join(taskdir, dest)
			if err := d.linkOrCopy(source, taskEntry, s.Mode().Perm()); err != nil {
				return err
			}

			continue
		}

		// Create destination directory.
		destDir := filepath.Join(taskdir, dest)

		if err := d.createDir(taskdir, dest); err != nil {
			return fmt.Errorf("Couldn't create destination directory %v: %v", destDir, err)
		}

		// Enumerate the files in source.
		dirEntries, err := ioutil.ReadDir(source)
		if err != nil {
			return fmt.Errorf("Couldn't read directory %v: %v", source, err)
		}

		for _, entry := range dirEntries {
			hostEntry := filepath.Join(source, entry.Name())
			taskEntry := filepath.Join(destDir, filepath.Base(hostEntry))
			if entry.IsDir() {
				subdirs[hostEntry] = filepath.Join(dest, filepath.Base(hostEntry))
				continue
			}

			// Check if entry exists. This can happen if restarting a failed
			// task.
			if _, err := os.Lstat(taskEntry); err == nil {
				continue
			}

			if !entry.Mode().IsRegular() {
				// If it is a symlink we can create it, otherwise we skip it.
				if entry.Mode()&os.ModeSymlink == 0 {
					continue
				}

				link, err := os.Readlink(hostEntry)
				if err != nil {
					return fmt.Errorf("Couldn't resolve symlink for %v: %v", source, err)
				}

				if err := os.Symlink(link, taskEntry); err != nil {
					// Symlinking twice
					if err.(*os.LinkError).Err.Error() != "file exists" {
						return fmt.Errorf("Couldn't create symlink: %v", err)
					}
				}
				continue
			}

			if err := d.linkOrCopy(hostEntry, taskEntry, entry.Mode().Perm()); err != nil {
				return err
			}
		}
	}

	// Recurse on self to copy subdirectories.
	if len(subdirs) != 0 {
		return d.Embed(task, subdirs)
	}

	return nil
}

// MountSharedDir mounts the shared directory into the specified task's
// directory. Mount is documented at an OS level in their respective
// implementation files.
func (d *AllocDir) MountSharedDir(task string) error {
	taskDir, ok := d.TaskDirs[task]
	if !ok {
		return fmt.Errorf("No task directory exists for %v", task)
	}

	taskLoc := filepath.Join(taskDir, SharedAllocName)
	if err := d.mountSharedDir(taskLoc); err != nil {
		return fmt.Errorf("Failed to mount shared directory for task %v: %v", task, err)
	}

	return nil
}

// LogDir returns the log dir in the current allocation directory
func (d *AllocDir) LogDir() string {
	return filepath.Join(d.AllocDir, SharedAllocName, LogDirName)
}

// List returns the list of files at a path relative to the alloc dir
func (d *AllocDir) List(path string) ([]*AllocFileInfo, error) {
	if escapes, err := structs.PathEscapesAllocDir(path); err != nil {
		return nil, fmt.Errorf("Failed to check if path escapes alloc directory: %v", err)
	} else if escapes {
		return nil, fmt.Errorf("Path escapes the alloc directory")
	}

	p := filepath.Join(d.AllocDir, path)
	finfos, err := ioutil.ReadDir(p)
	if err != nil {
		return []*AllocFileInfo{}, err
	}
	files := make([]*AllocFileInfo, len(finfos))
	for idx, info := range finfos {
		files[idx] = &AllocFileInfo{
			Name:     info.Name(),
			IsDir:    info.IsDir(),
			Size:     info.Size(),
			FileMode: info.Mode().String(),
			ModTime:  info.ModTime(),
		}
	}
	return files, err
}

// Stat returns information about the file at a path relative to the alloc dir
func (d *AllocDir) Stat(path string) (*AllocFileInfo, error) {
	if escapes, err := structs.PathEscapesAllocDir(path); err != nil {
		return nil, fmt.Errorf("Failed to check if path escapes alloc directory: %v", err)
	} else if escapes {
		return nil, fmt.Errorf("Path escapes the alloc directory")
	}

	p := filepath.Join(d.AllocDir, path)
	info, err := os.Stat(p)
	if err != nil {
		return nil, err
	}

	return &AllocFileInfo{
		Size:     info.Size(),
		Name:     info.Name(),
		IsDir:    info.IsDir(),
		FileMode: info.Mode().String(),
		ModTime:  info.ModTime(),
	}, nil
}

// ReadAt returns a reader for a file at the path relative to the alloc dir
func (d *AllocDir) ReadAt(path string, offset int64) (io.ReadCloser, error) {
	if escapes, err := structs.PathEscapesAllocDir(path); err != nil {
		return nil, fmt.Errorf("Failed to check if path escapes alloc directory: %v", err)
	} else if escapes {
		return nil, fmt.Errorf("Path escapes the alloc directory")
	}

	p := filepath.Join(d.AllocDir, path)

	// Check if it is trying to read into a secret directory
	for _, dir := range d.TaskDirs {
		sdir := filepath.Join(dir, TaskSecrets)
		if filepath.HasPrefix(p, sdir) {
			return nil, fmt.Errorf("Reading secret file prohibited: %s", path)
		}
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	if _, err := f.Seek(offset, 0); err != nil {
		return nil, fmt.Errorf("can't seek to offset %q: %v", offset, err)
	}
	return f, nil
}

// BlockUntilExists blocks until the passed file relative the allocation
// directory exists. The block can be cancelled with the passed tomb.
func (d *AllocDir) BlockUntilExists(path string, t *tomb.Tomb) (chan error, error) {
	if escapes, err := structs.PathEscapesAllocDir(path); err != nil {
		return nil, fmt.Errorf("Failed to check if path escapes alloc directory: %v", err)
	} else if escapes {
		return nil, fmt.Errorf("Path escapes the alloc directory")
	}

	// Get the path relative to the alloc directory
	p := filepath.Join(d.AllocDir, path)
	watcher := getFileWatcher(p)
	returnCh := make(chan error, 1)
	go func() {
		returnCh <- watcher.BlockUntilExists(t)
		close(returnCh)
	}()
	return returnCh, nil
}

// ChangeEvents watches for changes to the passed path relative to the
// allocation directory. The offset should be the last read offset. The tomb is
// used to clean up the watch.
func (d *AllocDir) ChangeEvents(path string, curOffset int64, t *tomb.Tomb) (*watch.FileChanges, error) {
	if escapes, err := structs.PathEscapesAllocDir(path); err != nil {
		return nil, fmt.Errorf("Failed to check if path escapes alloc directory: %v", err)
	} else if escapes {
		return nil, fmt.Errorf("Path escapes the alloc directory")
	}

	// Get the path relative to the alloc directory
	p := filepath.Join(d.AllocDir, path)
	watcher := getFileWatcher(p)
	return watcher.ChangeEvents(t, curOffset)
}

// getFileWatcher returns a FileWatcher for the given path.
func getFileWatcher(path string) watch.FileWatcher {
	return watch.NewPollingFileWatcher(path)
}

func fileCopy(src, dst string, perm os.FileMode) error {
	// Do a simple copy.
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Couldn't open src file %v: %v", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, perm)
	if err != nil {
		return fmt.Errorf("Couldn't create destination file %v: %v", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("Couldn't copy %v to %v: %v", src, dst, err)
	}

	return nil
}

// pathExists is a helper function to check if the path exists.
func (d *AllocDir) pathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (d *AllocDir) GetSecretDir(task string) (string, error) {
	if t, ok := d.TaskDirs[task]; !ok {
		return "", fmt.Errorf("Allocation directory doesn't contain task %q", task)
	} else {
		return filepath.Join(t, TaskSecrets), nil
	}
}

// createDir creates a directory structure inside the basepath. This functions
// preserves the permissions of each of the subdirectories in the relative path
// by looking up the permissions in the host.
func (d *AllocDir) createDir(basePath, relPath string) error {
	filePerms, err := d.splitPath(relPath)
	if err != nil {
		return err
	}

	// We are going backwards since we create the root of the directory first
	// and then create the entire nested structure.
	for i := len(filePerms) - 1; i >= 0; i-- {
		fi := filePerms[i]
		destDir := filepath.Join(basePath, fi.Name)
		if err := os.MkdirAll(destDir, fi.Perm); err != nil {
			return err
		}
	}
	return nil
}

// fileInfo holds the path and the permissions of a file
type fileInfo struct {
	Name string
	Perm os.FileMode
}

// splitPath stats each subdirectory of a path. The first element of the array
// is the file passed to this method, and the last element is the root of the
// path.
func (d *AllocDir) splitPath(path string) ([]fileInfo, error) {
	var mode os.FileMode
	i, err := os.Stat(path)

	// If the path is not present in the host then we respond with the most
	// flexible permission.
	if err != nil {
		mode = os.ModePerm
	} else {
		mode = i.Mode()
	}
	var dirs []fileInfo
	dirs = append(dirs, fileInfo{Name: path, Perm: mode})
	currentDir := path
	for {
		dir := filepath.Dir(filepath.Clean(currentDir))
		if dir == currentDir {
			break
		}

		// We try to find the permission of the file in the host. If the path is not
		// present in the host then we respond with the most flexible permission.
		i, err = os.Stat(dir)
		if err != nil {
			mode = os.ModePerm
		} else {
			mode = i.Mode()
		}
		dirs = append(dirs, fileInfo{Name: dir, Perm: mode})
		currentDir = dir
	}
	return dirs, nil
}
