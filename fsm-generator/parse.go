package fsm_generator

import (
	"bytes"
	"crypto/md5" // #nosec
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func ParseModel(file *os.File) (*Model, error) {
	d := yaml.NewDecoder(file)
	d.KnownFields(false)

	model := &Model{}

	err := d.Decode(model)
	if err != nil {
		return nil, err
	}

	model.ETag = ETag(model)

	states := make([]*State, 0, len(model.States))

	// разборка модели
	for _, state := range model.States {
		if state.MaxRetryCount == 0 {
			state.MaxRetryCount = model.DefaultConfig.MaxRetryCount
		}

		if state.MinRetryDelay == time.Duration(0) {
			state.MinRetryDelay = model.DefaultConfig.MinRetryDelay
		}

		if state.CancellationTTL == time.Duration(0) {
			state.CancellationTTL = model.DefaultConfig.CancellationTTL
		}

		for _, transition := range state.Transitions {
			if transition.StateName == state.Name {
				return nil, errors.New("state can transit in itself")
			}

			for _, otherState := range model.States {
				if otherState.Name == transition.StateName {
					transition.State = otherState
					break
				}
			}

			if transition.State == nil {
				// todo: дописать название состояний
				return nil, fmt.Errorf("no state found for transition: model=%s state=%s", model.Title, state.Name)
			}
		}

		states = append(states, state)

		if !state.SuccessFinal && !state.FailFinal && !state.DisableFallbackState {
			fallbackState := &State{
				Name:            state.Name + "_FAILED",
				Description:     "Перехват ошибки для " + state.Name,
				FailFinal:       true,
				MaxRetryCount:   state.MaxRetryCount,
				MinRetryDelay:   state.MinRetryDelay,
				CancellationTTL: state.CancellationTTL,
			}
			states = append(states, fallbackState)
			state.Transitions = append(state.Transitions, &Transition{
				StateName: fallbackState.Name,
				Condition: "Исчерпаны попытки исполнить событие состояния " + state.Name,
				State:     fallbackState,
			})
		}
	}

	model.States = states

	return model, nil
}

func ETag(model *Model) string {
	buf := bytes.NewBuffer(nil)

	err := yaml.NewEncoder(buf).Encode(model)
	if err != nil {
		log.Fatalf("etag can't be generated: %v", err)
	}

	h := md5.Sum(buf.Bytes()) // #nosec

	return hex.EncodeToString(h[:3])
}
