package progress

type Printer interface {
	Print(*Status)
}
