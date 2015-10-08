package agent

// Agent interface
type Agent interface {
	Get(id string) (text string, err error)
	Run() (bool, error)
}
