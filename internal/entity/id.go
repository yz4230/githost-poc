package entity

import (
	"strconv"

	"github.com/samber/lo"
)

type ID string

func NewID(id any) ID {
	switch v := id.(type) {
	case string:
		return ID(v)
	case uint:
		return ID(strconv.FormatUint(uint64(v), 10))
	}
	panic("unsupported ID type")
}
func (id ID) String() string { return string(id) }
func (id ID) Uint() uint     { return uint(lo.Must(strconv.ParseUint(id.String(), 10, 64))) }
