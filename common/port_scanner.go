package common

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"syscall"
	"time"
)

func GetPortInRange(start, end int) (int, error) {
	if start > end {
		return -1, fmt.Errorf("bad input range")
	}
	for i := start; i < end; i++ {
		ok := freePort(i)
		if ok {
			return i, nil
		}
	}
	return -1, fmt.Errorf("no open port in %d-%d", start, end)
}

// true if port is not in use
func freePort(port int) bool {
	address := ":" + strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", address, 10*time.Millisecond)

	if err != nil {
		return errors.Is(err, syscall.ECONNREFUSED)
	}
	defer conn.Close()
	return true
}
