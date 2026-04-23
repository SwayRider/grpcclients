package searchclient

import "time"

var deadline = 5 * time.Second

func Deadline() time.Duration {
	return deadline
}

func SetDeadline(d time.Duration) {
	deadline = d
}
