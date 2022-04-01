package prettyuuid

import (
	"github.com/google/uuid"
)

// New генерирует uuid.UUID с заменой первых байт на переданные
func New(prefix ...byte) uuid.UUID {
	return Prefix(uuid.New(), prefix...)
}

// Prefix заменяет первые байты на переданные
func Prefix(UUID uuid.UUID, prefix ...byte) uuid.UUID {
	if len(prefix) > len(UUID) {
		return UUID
	}

	for i, b := range prefix {
		UUID[i] = b
	}

	return UUID
}
