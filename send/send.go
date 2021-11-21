// Package send contains message sender abstraction.
package send

// Sender abstracts message sender methods.
type Sender interface {
	// Send sends message.
	Send(recipient, subject, message string) error
}
