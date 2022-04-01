package fsm_generator

import (
	"errors"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
)

func (m *Model) ValidateModel() error {
	if !strings.EqualFold(m.Name, strcase.ToSnake(m.Name)) {
		return errors.New("model name should be in snake_case: " + strcase.ToSnake(m.Name))
	}

	var (
		initStates []*State
		hasFinals  bool
	)

	for _, state := range m.States {
		if state.Initial {
			initStates = append(initStates, state)
		}

		if state.SuccessFinal || state.FailFinal {
			hasFinals = true
		}

		if state.MinRetryDelay != 0 && state.MinRetryDelay < time.Second {
			return errors.New(state.Name + ": min retry delay should be times of 1 second (or be equal zero)")
		}

		if state.CancellationTTL != 0 && state.CancellationTTL < time.Second {
			return errors.New(state.Name + ": cancellation ttl should be times of 1 second (or be equal zero)")
		}
	}

	// есть начальные состояния (минимум 1)
	if len(initStates) == 0 {
		return errors.New("model should have at least one init state")
	}
	// есть конечные состояния (минимум 1)
	if !hasFinals {
		return errors.New("model should have at least one final state")
	}

	// todo: из начального состояния можно попасть в любое другое состояние (покрасить граф)

	return nil
}
