// tasks/email.go
package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	TypeEmailDelivery = "email:deliver"
	TypeEmailWelcome  = "email:welcome"
)

//---------------------------------------------------------------
// Task payload structs
//---------------------------------------------------------------

type EmailDeliveryPayload struct {
	UserID     int
	TemplateID string
	Data       map[string]interface{}
}

type EmailWelcomePayload struct {
	UserID int
}

//---------------------------------------------------------------
// Task creator functions
//---------------------------------------------------------------

// NewEmailDeliveryTask creates a new task for sending an email.
func NewEmailDeliveryTask(userID int, templateID string, data map[string]interface{}) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailDeliveryPayload{
		UserID:     userID,
		TemplateID: templateID,
		Data:       data,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEmailDelivery, payload), nil
}

// NewEmailWelcomeTask creates a new task for sending a welcome email.
func NewEmailWelcomeTask(userID int) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailWelcomePayload{UserID: userID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEmailWelcome, payload), nil
}

//---------------------------------------------------------------
// Task handler functions
//---------------------------------------------------------------

// HandleEmailDeliveryTask handles email delivery tasks.
func HandleEmailDeliveryTask(ctx context.Context, t *asynq.Task) error {
	var p EmailDeliveryPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	// Simulate email sending with delay
	fmt.Printf("Sending Email to User: %d with Template: %s and Data: %v\n", p.UserID, p.TemplateID, p.Data)
	time.Sleep(2 * time.Second)
	fmt.Printf("Email sent to User: %d\n", p.UserID)

	return nil
}

// HandleEmailWelcomeTask handles welcome email tasks.
func HandleEmailWelcomeTask(ctx context.Context, t *asynq.Task) error {
	var p EmailWelcomePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}
	fmt.Printf("Preparing to send welcome email to User: %v\n", p)

	// Simulate welcome email sending
	fmt.Printf("Sending Welcome Email to User: %d\n", p.UserID)
	time.Sleep(1 * time.Second)
	fmt.Printf("Welcome Email sent to User: %d\n", p.UserID)

	return nil
}
