package cloudcommon

type Action int

const (
	Create Action = iota
	Delete
)

func (a Action) String() string {
	return [...]string{"Create", "Delete"}[a]
}
