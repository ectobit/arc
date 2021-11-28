// Package send contains message sender abstraction.
package send

// Sender abstracts message sender methods.
type Sender interface {
	// Send sends message.
	Send(recipient, subject, message string) error
}

// Fake implements Sender interface doing nothing.
type Fake struct{}

// Send accretes the message.
func (s *Fake) Send(_, _, _ string) error {
	return nil
}
