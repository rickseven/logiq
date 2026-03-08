package cmdintel

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GenerateExecutionID creates a unique identifier for a command run
func GenerateExecutionID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x-%s", b, time.Now().Format("20060102150405"))
}
