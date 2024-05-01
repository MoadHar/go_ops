package dataformats

import (
	"errors"
)

type record []string

func (r record) Validate() error {
	if len(r) != 2 {
		return errors.New("data format incorrect")
	}
	return nil
}
func (r record) first() string {
	return r[0]
}
func (r record) last() string {
	return r[1]
}
