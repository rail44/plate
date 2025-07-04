package common


type State struct {
	Tables map[string]struct{}
	Params []any
}
