package healthclient

type ServiceStatus string

const (
	Unknown ServiceStatus = "UNKNOWN"
	Ok      ServiceStatus = "UP"
	Error   ServiceStatus = "DOWN"
)
