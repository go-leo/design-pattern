package prototype

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

//type ElfMage struct {
//	HelpType string `clone:"help_type"`
//}
//
//type OrcMage struct {
//	Weapon string `clone:"weapon"`
//}
//
//func TestClone(t *testing.T) {
//	var orc *OrcMage
//	obj := &ElfMage{HelpType: "cooking"}
//	orc = Clone[*ElfMage, *OrcMage](obj, Cloned(orc))
//
//}

func TestInvalidToError(t *testing.T) {
	var from bool
	var to bool

	err := Clone(to, from)
	t.Log(err)
	assert.NotNil(t, err)

	err = Clone(nil, from)
	t.Log(err)
	assert.NotNil(t, err)

	var nilTo *bool
	err = Clone(nilTo, from)
	t.Log(err)
	assert.NotNil(t, err)

}

func TestInvalidFromError(t *testing.T) {
	var to bool
	err := Clone(&to, nil)
	t.Log(err)
	assert.NotNil(t, err)

	var e error
	err = Clone(&to, e)
	t.Log(err)
	assert.NotNil(t, err)
}

func TestCloneBoolBool(t *testing.T) {
	var from bool
	var to bool
	err := Clone(&to, from)
	assert.Nil(t, err)
	assert.Equal(t, from, to)
}

func TestJson(t *testing.T) {
	var to bool
	v, err := json.Marshal(&to)
	t.Log(string(v), err)
	assert.NotNil(t, err)
}
