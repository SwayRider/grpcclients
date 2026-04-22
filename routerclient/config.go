package routerclient

import "time"

var deadline = 60 * time.Second

func Deadline() time.Duration {
	return deadline
}

func SetDeadline(d time.Duration) {
	deadline = d
}
