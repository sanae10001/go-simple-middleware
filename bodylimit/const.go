package bodylimit

type Byte int64

const (
	B  Byte = 1
	KB      = 1024 * B
	MB      = 1024 * KB
	GB      = 1024 * MB
	TB      = 1024 * GB
	PB      = 1024 * TB
)
