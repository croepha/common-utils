package minicgroups

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"os"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"gitlab.com/croepha/common-utils/loggingsugar"
	"gitlab.com/croepha/common-utils/lostandfound"
	"gitlab.com/croepha/common-utils/syscallextra"
)

var l loggingsugar.L

// TODO: Cgroup namespace?

type mount struct {
	mntPath  string
	rootPath string
	nextId   atomic.Uint64
}

func NewMount(ctx context.Context) (*mount, error) {

	mntPath, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}

	m := &mount{mntPath: mntPath}

	if err := syscallextra.WrapEINTR(func() error {
		return syscall.Mount("pvt-cgroup", mntPath, "cgroup2", 0, "")
	}); err != nil {
		l.Err(err).Error(ctx, "error")
		return nil, err
	}

	rootAttempts := 5
	for range rootAttempts {
		rootPath := fmt.Sprintf("%s/mcgroot-%d", mntPath, time.Now().Unix())
		m.rootPath = rootPath
		err := os.Mkdir(rootPath, 0)
		if err == nil {
			goto rootCreated
		}
		if !errors.Is(err, syscall.EEXIST) {
			defer m.Done(ctx)
			return nil, err
		}
	}

	defer m.Done(ctx)
	return nil, fmt.Errorf("ran out of attempts to create root directory: %+q", m.rootPath)

rootCreated:

	return &mount{mntPath: mntPath}, nil
}

func (m *mount) Done(ctx context.Context) error {
	if m.mntPath == "" {
		panic("mount not opened")
	}
	if err := syscallextra.WrapEINTR(func() error {
		return syscall.Unmount(m.mntPath, 0)
	}); err != nil {
		return err
	}
	if err := os.Remove(m.mntPath); err != nil {
		return err
	}

	m.mntPath = ""
	return nil
}

func (m *mount) CreateGroup(ctx context.Context, controllers []string) (*group, error) {
	if m.mntPath == "" {
		panic("mount not opened")
	}

try:
	path := fmt.Sprintf("%s/mcg-%d", m.mntPath, m.nextId.Add(1)-1)
	group, err := Create(ctx, path, controllers)

	if errors.Is(err, syscall.EEXIST) {
		goto try
	}

	return group, nil
}

type group struct {
	path string
}

// Create (if it doesn't exist) a v2 cgroup at path
// path must be under a cgroup v2 mount
// Also ensure that the given controllers are enabled
func Create(ctx context.Context, path string, controllers []string) (*group, error) {

	ctx = l.A("mcgPath", path).Context(ctx)

	if err := os.MkdirAll(path, 0); err != nil {
		return nil, err
	}

	cg := &group{path: path}
	if err := cg.EnableControllers(ctx, controllers); err != nil {
		if err2 := cg.Delete(ctx); err2 != nil {
			return cg, fmt.Errorf(
				"second error: %w while cleaning up from original error: %w",
				err2, err)
		}
	}

	l.Debug(ctx, "mcg Create")
	return cg, nil
}

func (cg *group) EnableControllers(ctx context.Context, controllers []string) error {
	if cg.path == "" {
		panic("cgroup not opened")
	}
	return enableControllers(ctx, cg.path, controllers)
}

func (cg *group) WriteFiles(ctx context.Context, fileContents map[string]string) error {
	if cg.path == "" {
		panic("cgroup not opened")
	}
	for name, content := range fileContents {
		path := cg.path + "/" + name
		if err := os.WriteFile(path, []byte(content), 0); err != nil {
			return err
		}
	}
	return nil
}

func (cg *group) ReadFiles(ctx context.Context, files []string) ([]string, error) {
	if cg.path == "" {
		panic("cgroup not opened")
	}
	contents := []string{}
	for _, name := range files {
		path := cg.path + "/" + name

		if fb, err := os.ReadFile(path); err != nil {
			return nil, err
		} else {
			contents = append(contents, string(fb))
		}
	}
	return contents, nil
}

func (cg *group) FD(ctx context.Context) iter.Seq2[int, error] {
	if cg.path == "" {
		panic("cgroup not opened")
	}
	return syscallextra.PathFD(ctx, cg.path)
}

func (cg *group) Delete(ctx context.Context) error {
	if cg.path == "" {
		panic("cgroup not opened")
	}
	err := os.Remove(cg.path)
	if err != nil {
		return err
	}
	cg.path = ""
	return nil
}

// TODO: Mutex? so that global cgroup operations are serialized?
// TODO: Inotify to listen for changes?

func enableControllers(ctx context.Context, path string, needed []string) error {

	if len(needed) == 0 {
		return nil
	}

	fb, err := os.ReadFile(path + "/cgroup.controllers")
	if err != nil {
		return err
	}

	current := strings.Split(string(fb), " ")
	missing := lostandfound.SliceSubtract(needed, current)

	if len(missing) == 0 {
		return nil
	}

	if !lostandfound.FileExists(path + "/cgroup.type") {
		return fmt.Errorf("needed controller not present on root: %+q controllers: %s", path, missing)
	}

	// TODO: We could cache the controllers that are enabled on each controllers?
	parentPath := path + "/.."
	if err := enableControllers(ctx, parentPath, missing); err != nil {
		return err
	}

	content := ""
	for _, s := range missing {
		content += " +" + s
	}

	return os.WriteFile(parentPath+"/cgroup.subtree_control", []byte(content), 0)

}
