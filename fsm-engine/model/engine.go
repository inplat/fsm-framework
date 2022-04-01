package model

import (
	"context"
	"errors"
)

var (
	// ErrServiceNotImplementedInterface возникает в случае, если переданный при инициализации модели сервис
	// не подходит под описанный в модели интерфейс
	ErrServiceNotImplementedInterface = errors.New("service doesn't implement model's interface")
)

type Engine interface {
	// Stop закрывает все соединения, останавливая обработку событий
	Stop(ctx context.Context)
	// AddModel инициализирует очередную fsm модель
	AddModel(ctx context.Context, newModel Model) error
	// Resolve ищем состояние по названию среди инициализированных моделей, либо nil
	Resolve(ctx context.Context, state string) (State, Model)
	// CreateTx задает транзакции начальное состояние и отправляет событие на его обработку
	CreateTx(ctx context.Context, tx Tx, initState State) error
	// Transit создает событие на проведение транзакции из одного состояния в другое
	Transit(ctx context.Context, tx Tx, newState State) error
}
