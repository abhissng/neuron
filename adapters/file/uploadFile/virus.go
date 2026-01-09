package uploadFile

import (
	"errors"
	"io"
)

/*
========================================
 Errors
========================================
*/

var (
	ErrVirusDetected = errors.New("virus detected in file")
)

type VirusScanner interface {
	Scan(r io.Reader) (clean bool, err error)
}
