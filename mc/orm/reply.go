package orm

type M map[string]interface{}

type Result struct {
	Message string `json:"message"`
}

func Msg(msg string) *Result {
	return &Result{Message: msg}
}

type ResultID struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
}

func MsgID(msg string, id int64) *ResultID {
	return &ResultID{
		Message: msg,
		ID:      id,
	}
}

type ResultName struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

func MsgName(msg, name string) *ResultName {
	return &ResultName{
		Message: msg,
		Name:    name,
	}
}
