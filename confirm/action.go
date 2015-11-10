package confirm

import (
	"bytes"
	"io"
	"strconv"
	"strings"
)

type Type int

const (
	TypeCreate = iota + 1
	TypeUpdate
	TypeDelete
)

func NewUpdate(title string, diff []byte, f ActionFunc) *Action {
	return New(TypeUpdate, title, diff, f)
}

func NewDelete(title string, diff []byte, f ActionFunc) *Action {
	return New(TypeDelete, title, diff, f)
}

func NewCreate(title string, diff []byte, f ActionFunc) *Action {
	return New(TypeCreate, title, diff, f)
}

func New(t Type, title string, diff []byte, f ActionFunc) *Action {
	if t == 0 {
		panic("Type " + strconv.Itoa(int(t)) + " not valid")
	}
	if title == "" {
		panic("title must be present")
	}
	if f == nil {
		panic("ActionFunc must not be nil")
	}
	return &Action{Type: t, Title: title, Diff: diff, Call: f}
}

type Action struct {
	Title string
	Diff  []byte
	Type  Type
	Call  ActionFunc
}

var types = map[Type]string{
	TypeCreate: "CREATE",
	TypeUpdate: "UPDATE",
	TypeDelete: "DELETE",
}

func (s *Action) String() string {
	out := []string{}
	a, ok := types[s.Type]
	if !ok {
		panic("type " + strconv.Itoa(int(s.Type)) + " not mapped")
	}
	out = append(out, a, s.Title)
	return strings.Join(out, " ")
}

type ActionFunc func() error

type Actions []*Action

func (list Actions) Exec() error {
	for _, a := range list {
		if err := a.Call(); err != nil {
			return err
		}
	}
	return nil
}

func (list *Actions) Create(title string, diff []byte, f ActionFunc) {
	*list = append(*list, NewCreate(title, diff, f))
}

func (list *Actions) Delete(title string, diff []byte, f ActionFunc) {
	*list = append(*list, NewDelete(title, diff, f))
}

func (list *Actions) Update(title string, diff []byte, f ActionFunc) {
	*list = append(*list, NewUpdate(title, diff, f))
}

func (list Actions) String() string {
	buf := &bytes.Buffer{}
	for i, a := range list {
		io.WriteString(buf, a.String())
		if i != len(list)-1 {
			io.WriteString(buf, "\n")
		}
	}
	return buf.String()
}
