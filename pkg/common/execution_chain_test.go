package common_test

import (
	"errors"
	"testing"

	"github.com/pdcgo/pdc_orc/pkg/common"
	"github.com/stretchr/testify/assert"
)

type Dummy struct {
	*common.ExecutionChain
	Data string
}

func (d *Dummy) Exec(handler func(seterr func(err error))) *Dummy {
	d.ExecutionChain.Exec(handler)
	return d
}

func TestExecution(t *testing.T) {
	dum := Dummy{
		Data:           "asdasd",
		ExecutionChain: &common.ExecutionChain{},
	}

	dum.Exec(func(seterr func(err error)) {
		seterr(errors.New("dummy error"))
	})

	assert.NotNil(t, dum.Err)
}
