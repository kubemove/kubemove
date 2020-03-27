package plugin

type Plugin interface {
	Init(map[string]string) error
	Sync(map[string]string) (string, error)
	Status(map[string]string) (int32, error)
}

const (
	Completed = iota
	InProgress
	Invalid
	Canceled
	Errored
	Failed
	Unknown
)
