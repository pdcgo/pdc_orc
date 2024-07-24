package node_proj

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pdcgo/pdc_orc/pkg/common"
)

type PackageJson struct {
	Scripts map[string]string `json:"scripts"`
}

type GitConfig struct {
	Uri    string
	Branch string
}

type NodeProject struct {
	*common.ExecutionChain
	Git            *GitConfig
	Dir            string
	PackageManager string
}

func NewNodeProject(
	git *GitConfig,
	dir string,
) *NodeProject {
	return &NodeProject{
		ExecutionChain: common.NewExecutionChain(),
		Git:            git,
		Dir:            dir,
	}
}

func (d *NodeProject) Exec(handler func(seterr func(err error))) *NodeProject {
	d.ExecutionChain.Exec(handler)
	return d
}

func (d *NodeProject) Pulling() *NodeProject {
	return d.Exec(func(seterr func(err error)) {

		seterr(d.CreateCmd("git", "pull").Run())
	})
}

func (d *NodeProject) Check() *NodeProject {
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

func (d *NodeProject) InstallPackage() *NodeProject {
	d.PackageManager = ""

	return d.
		Exec(func(seterr func(err error)) {
			lockMap := map[string]string{
				"npm":  "package-lock.json",
				"pnpm": "pnpm-lock.yaml",
			}

			for key, file := range lockMap {
				if d.IsExist(filepath.Join(d.Dir, file)) {
					d.PackageManager = key
				}
			}
		}).
		Exec(func(seterr func(err error)) { // installing library
			if d.PackageManager == "" {
				seterr(errors.New("cannot detect package manager"))
				return
			}

			err := d.CreateCmd(d.PackageManager, "install").Run()
			seterr(err)
		})
}

func (d *NodeProject) Serve() *NodeProject {
	return d.
		Exec(func(seterr func(err error)) {
			log.Println("serving", d.Dir)
		})
}

func (d *NodeProject) Folder(folder string) *NodeProject {
	newd := NewNodeProject(d.Git, filepath.Join(d.Dir, folder))
	newd.ExecutionChain = d.ExecutionChain
	return newd
}

func (d *NodeProject) DotEnv(envs map[string]string) *NodeProject {
	return d.Exec(func(seterr func(err error)) {
		filename := filepath.Join(d.Dir, ".env")
		log.Println("creating environtment", filename)

		f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			seterr(err)
			return
		}
		defer f.Close()
		for key, value := range envs {
			line := strings.Join([]string{key, "=", value}, " ") + "\r\n"
			f.WriteString(line)
		}

	})

}

func (d *NodeProject) PackageRunScript(cmd string, args ...string) *NodeProject {
	// PackageJson
	packageManifest := PackageJson{}
	return d.
		Exec(func(seterr func(err error)) { // load package json
			fname := filepath.Join(d.Dir, "package.json")
			data, err := os.ReadFile(fname)
			if err != nil {
				seterr(err)
				return
			}

			err = json.Unmarshal(data, &packageManifest)
			seterr(err)

		}).
		Exec(func(seterr func(err error)) {
			if packageManifest.Scripts[cmd] == "" {
				seterr(errors.New(cmd + " not found in package.json"))
			}

			if d.PackageManager == "" {
				seterr(errors.New(cmd + " not detect library manager"))
			}
		}).
		Exec(func(seterr func(err error)) {
			cmds := []string{}
			switch d.PackageManager {
			case "pnpm":
				cmds = []string{d.PackageManager, cmd}
				if len(args) != 0 {
					cmds = append(cmds, args...)
				}
			case "npm":
				cmds = []string{d.PackageManager, "run", cmd}
				if len(args) != 0 {
					cmds = append(cmds, "--")
					cmds = append(cmds, args...)
				}

			}

			seterr(d.CreateCmd(cmds...).Run())
		})
}

func (ex *NodeProject) CreateCmd(args ...string) *exec.Cmd {

	log.Println("running", strings.Join(args, " "))

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)
	cmd := exec.Cmd{
		Path:   ex.Where(args[0]),
		Dir:    ex.Dir,
		Stdout: os.Stdout,
		Args:   args,
	}

	cmd.Stdout = mw
	cmd.Stderr = mw

	return &cmd
}
