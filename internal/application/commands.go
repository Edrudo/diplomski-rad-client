package application

type CommandName int

const (
	SendGeoshot CommandName = iota
)

type Command struct {
	Name CommandName
	Args []string
}
