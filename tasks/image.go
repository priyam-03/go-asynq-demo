// tasks/image.go
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
	TypeImageResize = "image:resize"
)

//---------------------------------------------------------------
// Task payload structs
//---------------------------------------------------------------

type ImageResizePayload struct {
	ImageID string
	Width   int
	Height  int
	Format  string
	UserID  int
}

//---------------------------------------------------------------
// Task creator functions
//---------------------------------------------------------------

// NewImageResizeTask creates a new task for resizing an image.
func NewImageResizeTask(imageID string, width, height int, format string, userID int) (*asynq.Task, error) {
	payload, err := json.Marshal(ImageResizePayload{
		ImageID: imageID,
		Width:   width,
		Height:  height,
		Format:  format,
		UserID:  userID,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeImageResize, payload), nil
}

//---------------------------------------------------------------
// Task handler functions
//---------------------------------------------------------------

// HandleImageResizeTask handles image resize tasks.
func HandleImageResizeTask(ctx context.Context, t *asynq.Task) error {
	var p ImageResizePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	// Simulate image processing
	fmt.Printf("Processing image: %s for user: %d\n", p.ImageID, p.UserID)
	fmt.Printf("Resizing to %dx%d, format: %s\n", p.Width, p.Height, p.Format)

	// Simulate work that takes time
	time.Sleep(3 * time.Second)

	fmt.Printf("Image processing completed for %s\n", p.ImageID)
	return nil
}
