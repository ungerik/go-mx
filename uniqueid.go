package mx

import (
	"context"
	"strconv"
)

func UniqueID() Attrib {
	return uniqueID(idCounter.Add(1))
}

type uniqueID uint64

func (id uniqueID) AttribName() string {
	return "id"
}

func (id uniqueID) AttribValue(context.Context) (string, error) {
	return "_" + strconv.FormatUint(uint64(id), 36), nil
}
