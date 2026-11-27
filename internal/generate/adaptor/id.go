package adaptor

import (
	"github.com/google/uuid"
)

type Uuid struct{}

func (m *Uuid) Gen(args ...string) any {
	return uuid.NewString()
}
