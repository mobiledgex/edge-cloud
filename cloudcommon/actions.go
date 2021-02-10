package cloudcommon

type Action int

const (
	Create Action = iota
	Delete
	Update
)

func (a Action) String() string {
	return [...]string{"Create", "Delete", "Update"}[a]
}
