package mock

import (
	"sync"

	"github.com/m12r/email"
)

type Opt func(*Sender)

func SetValidator(v Validator) Opt {
	return func(s *Sender) {
		s.validator = v
	}
}

func SetError(err error) Opt {
	return func(s *Sender) {
		s.err = err
	}
}

type Sender struct {
	counter   int
	validator Validator
	err       error
	mu        sync.Mutex
}

func NewSender(opts ...Opt) *Sender {
	s := &Sender{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Sender) Send(m *email.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++

	if s.err != nil {
		return s.err
	}

	if s.validator == nil {
		return nil
	}
	if err := s.validator(m); err != nil {
		return err
	}
	return nil
}

func (s *Sender) MessageCount() int {
	return s.counter
}

func (s *Sender) SetError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.err = err
}

func (s *Sender) Reset(opts ...Opt) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter = 0

	for _, opt := range opts {
		opt(s)
	}
}

type Validator func(*email.Message) error
