package golang_proj

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pdcgo/pdc_orc/pkg/common"
)

type MapEnvs map[string]string

func (m MapEnvs) ToSlice() []string {
	hasil := os.Environ()
	for key, val := range m {
		hasil = append(hasil, strings.Join([]string{key, val}, "="))
	}

	return hasil
}

type GitConfig struct {
	Uri    string
	Branch string
}
type GolangProject struct {
	*common.ExecutionChain
	Git  *GitConfig
	Dir  string
	Envs MapEnvs
}

func NewGolangProject(
	git *GitConfig,
	dir string,
) *GolangProject {
	return &GolangProject{
		ExecutionChain: common.NewExecutionChain(),
		Git:            git,
		Dir:            dir,
		Envs:           map[string]string{},
	}
}
func (d *GolangProject) SetEnvs(m MapEnvs) *GolangProject {
	return d.
		Exec(func(seterr func(err error)) {
			d.Envs = m
		})
}

func (d *GolangProject) Exec(handler func(seterr func(err error))) *GolangProject {
	d.ExecutionChain.Exec(handler)
	return d
}

func (d *GolangProject) Check() *GolangProject {
	isNotExist := false

	log.Println("checking", d.Dir)

	d.
		Exec(func(seterr func(err error)) {
			_, err := os.Stat(d.Dir)
			if os.IsNotExist(err) {
				isNotExist = true
				return
			}

			seterr(err)
		})

	if isNotExist {
		d.Exec(func(seterr func(err error)) { // cloning from git
			log.Println("cloning", d.Git.Uri)
			var cmd *exec.Cmd
			if d.Git.Branch == "" {
				cmd = exec.Command("git", "clone", d.Git.Uri, d.Dir)
			} else {
				cmd = exec.Command("git", "clone", "-b", d.Git.Branch, d.Git.Uri, d.Dir)
			}
			err := cmd.Run()
			seterr(err)
		})
	}

	return d
}

func (d *GolangProject) Run(args ...string) *GolangProject {
	return d.Exec(func(seterr func(err error)) {
		seterr(d.CreateCmd(args...).Run())
	})
}

func (ex *GolangProject) CreateCmd(args ...string) *exec.Cmd {

	log.Println("running", strings.Join(args, " "))
	path := args[0]
	if !ex.IsExist(filepath.Join(ex.Dir, args[0])) || (args[0] == "cmd") {
		path = ex.Where(args[0])
	}

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)
	cmd := exec.Cmd{
		Path:   path,
		Dir:    ex.Dir,
		Stdout: os.Stdout,
		Args:   args,
		Env:    ex.Envs.ToSlice(),
	}

	cmd.Stdout = mw
	cmd.Stderr = mw

	return &cmd
}
