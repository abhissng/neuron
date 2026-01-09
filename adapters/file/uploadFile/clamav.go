package uploadFile

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/abhissng/neuron/utils/helpers"
)

type ClamAVScanner struct {
	Address string
}

func NewClamAVScanner(addr string) *ClamAVScanner {
	return &ClamAVScanner{Address: addr}
}

func (c *ClamAVScanner) Scan(r io.Reader) (bool, error) {
	conn, err := net.Dial("tcp", c.Address)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = conn.Close()
	}()

	// INSTREAM command
	if _, err := conn.Write([]byte("zINSTREAM\000")); err != nil {
		return false, err
	}

	buf := make([]byte, 8192)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			size := []byte{
				byte(n >> 24),
				byte(n >> 16),
				byte(n >> 8),
				byte(n),
			}
			if _, err := conn.Write(size); err != nil {
				return false, err
			}
			if _, err := conn.Write(buf[:n]); err != nil {
				return false, err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}
	}

	// End of stream
	if _, err := conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return false, err
	}

	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	if helpers.MatchesAny(resp, "FOUND") {
		return false, nil
	}

	if helpers.MatchesAny(resp, "OK") {
		return true, nil
	}

	return false, fmt.Errorf("unexpected clamav response: %s", resp)
}
