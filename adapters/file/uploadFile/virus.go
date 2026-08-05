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

// VirusScanner inspects raw file bytes. Scan must return (true, nil) when clean,
// (false, nil) when infected, and (false, err) when scanning cannot complete.
type VirusScanner interface {
	Scan(r io.Reader) (clean bool, err error)
}
