package controller

import (
	"io"
)

type Transport interface {
	io.ReadWriter
}
