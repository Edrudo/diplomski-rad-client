package application

type CommandName int

const (
	SendPhoto CommandName = iota
)

type Command struct {
	Name CommandName
	Args []string
}
