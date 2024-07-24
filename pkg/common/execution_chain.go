package common

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

type ExecutionChain struct {
	Err    error
	AppMap map[string]string
}

func NewExecutionChain() *ExecutionChain {
	return &ExecutionChain{
		AppMap: map[string]string{},
	}
}

func (ex *ExecutionChain) GetErr() error {
	return ex.Err
}

func (ex *ExecutionChain) SetErr(err error) {
	if err == nil {
		return
	}
	ex.Err = err
}

func (ex *ExecutionChain) Where(cli string) string {
	if ex.AppMap[cli] == "" {
		cmd := exec.Command("where", cli)
		data, err := cmd.Output()

		if err != nil {
			log.Println(cli, "not found")
			log.Panicln(err)
			return ""
		}
		pathlist := strings.Split(string(data), "\n")
		if len(pathlist) == 0 {
			log.Panicln(cli, "tidak terinstall")
		}

		datastr := strings.ReplaceAll(pathlist[0], "\r", "")

		if datastr == "" {
			log.Panicln(cli, "tidak terinstall")
			return ""
		}

		ex.AppMap[cli] = datastr
	}

	return ex.AppMap[cli]

}

func (ex *ExecutionChain) Exec(handler func(seterr func(err error))) {
	if ex.Err != nil {
		return
	}
	handler(ex.SetErr)
}

func (ex *ExecutionChain) IsExist(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}
