package fsmengine

import (
	"context"
	"fmt"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	zlog "github.com/rs/zerolog"

	callback_manager "fsm-framework/fsm-engine/callback-manager"
	"fsm-framework/fsm-engine/lock"
	"fsm-framework/fsm-engine/model"
	"fsm-framework/fsm-engine/queue"
)

// Engine Машина состояний отвечает за переход транзакции между состояниями, работу с очередью и локами
type Engine struct {
	// broker система управления очередями сообщений
	broker queue.Broker
	// locker инструмент для взятия эксклюзивных блокировок обработки транзакций
	locker lock.Locker
	// repo репозиторий со всеми, необходимыми в ходе процессинга события, методами
	repo model.Repository
	// verboseTracing подробное логгирование в трейсинг
	verboseTracing bool
	// cm управление отправкой обратного вызова (sync/async)
	cm callback_manager.CallbackManager

	models []model.Model
	states map[model.State]*StateProcessor
}

type Config struct {
	Repository      model.Repository
	Locker          lock.Locker
	Broker          queue.Broker
	CallbackManager callback_manager.CallbackManager
	VerboseTracing  bool
}

// New создает машину состояний
func New(cfg Config) *Engine {
	fsm := &Engine{
		broker:         cfg.Broker,
		locker:         cfg.Locker,
		repo:           cfg.Repository,
		cm:             cfg.CallbackManager,
		verboseTracing: cfg.VerboseTracing,
	}

	fsm.states = make(map[model.State]*StateProcessor, 128)

	return fsm
}

// Stop закрывает все соединения, останавливая обработку событий
func (e *Engine) Stop(ctx context.Context) {
	var err error

	// сначала останавливаем консюминг всех очередей (паралелльно)
	wg := sync.WaitGroup{}

	for _, stateProcessor := range e.states {
		wg.Add(1)

		go func(stateProcessor *StateProcessor) {
			defer wg.Done()

			closeErr := stateProcessor.consumerCh.Close()
			if closeErr != nil {
				zlog.Ctx(ctx).Error().Err(err).Msg("error while stopping consumer in fsm")
			}
		}(stateProcessor)
	}

	// Останавливаем обработку уведомлений
	wg.Add(1)

	go func() {
		defer wg.Done()

		stopErr := e.cm.Stop()
		if stopErr != nil {
			zlog.Ctx(ctx).Error().Err(err).Msg("error while stopping callback consumer in fsm")
		}
	}()

	wg.Wait()

	// потом останавливаем работу с брокером полностью
	e.broker.Close(ctx)

	err = e.locker.Close()
	if err != nil {
		zlog.Ctx(ctx).Error().Err(err).Msg("error while closing broker in fsm")
	}

	zlog.Ctx(ctx).Info().Msg("fsm stopped")
}

// AddModel инициализирует очередную fsm модель
func (e *Engine) AddModel(ctx context.Context, newModel model.Model) error {
	// проверяем, нет ли такой модели в списке инициализации
	for _, mdl := range e.models {
		if newModel == mdl {
			return fmt.Errorf("fsm-model has initialized already")
		}
	}

	// инициализируем все состояния переданной модели
	for _, s := range newModel.States() {
		publisherCh, err := e.broker.Channel()
		if err != nil {
			return fmt.Errorf("state channel creation error: %w", err)
		}

		cfg := &pipelineConfig{
			ch:             publisherCh,
			locker:         e.locker,
			repo:           e.repo,
			verboseTracing: e.verboseTracing,
			cm:             e.cm,
		}

		// consume
		consumerCh, err := e.broker.Channel()
		if err != nil {
			return err
		}

		sp := &StateProcessor{
			state:       s,
			consumerCh:  consumerCh,
			publisherCh: publisherCh,
		}

		e.states[s] = sp

		err = sp.StartConsume(ctx, cfg)
		if err != nil {
			return err
		}
	}

	newModel.SetEngine(e)

	e.models = append(e.models, newModel)

	zlog.Ctx(ctx).Info().Str("model", newModel.Name()).Msg("fsm consumers started")

	return nil
}

// Resolve ищем состояние по названию среди инициализированных моделей, либо nil
func (e *Engine) Resolve(ctx context.Context, state string) (model.State, model.Model) {
	span, _ := opentracing.StartSpanFromContext(ctx, "resolving state", opentracing.Tag{
		Key:   "state",
		Value: state,
	})
	defer span.Finish()

	for _, mdl := range e.models {
		if found := mdl.Resolve(state); found != nil {
			span.LogFields(log.String("model", mdl.Name()))

			return found, mdl
		}
	}

	span.LogFields(log.Message("state not found in any model"))

	return nil, nil
}

// CreateTx задает транзакции начальное состояние и отправляет событие на его обработку
func (e *Engine) CreateTx(ctx context.Context, tx model.Tx, initState model.State) error {
	// транзакция должна быть передана
	if tx == nil {
		return status.Error(codes.NotFound, "trying to move empty tx")
	}

	// нельзя создаться в пустое состояние
	if initState == nil {
		return status.Error(codes.Internal, "trying to move in nil state")
	}

	createTxSpan, ctx := opentracing.StartSpanFromContext(ctx, "creating tx", opentracing.Tag{
		Key:   "init_state",
		Value: initState.Name(),
	}, opentracing.Tag{
		Key:   "tx_id",
		Value: tx.ID().String(),
	})
	defer createTxSpan.Finish()

	// определяем модель состояния и инициализирована ли она
	_, initModel := e.Resolve(ctx, initState.Name())
	if initModel == nil {
		createTxSpan.LogFields(log.Message("state has no model"))

		return status.Errorf(codes.Internal, "%s has no model", initState.Name())
	}

	// проверяем можем ли сделать
	if !initState.IsInitial() {
		createTxSpan.LogFields(log.Message("state isn't initial state"))

		return status.Errorf(codes.PermissionDenied, "state isn't initial state. state: %s", initState.Name())
	}

	// обработчик нового состояния
	initStateProcessor, ok := e.states[initState]
	if !ok {
		createTxSpan.LogFields(log.Message("state not initialized"))

		return status.Errorf(codes.Internal, "%s not initialized", initState.Name())
	}

	// устанавливаем параметры транзакции, отвечающие за состояние
	tx.SetState(initState)
	tx.SetStatus(model.TxStatusPending)

	// создаем корневой span всего процессинга
	var parentTraceID, parentSpanID string

	fsmEventsSpan, _ := opentracing.StartSpanFromContext(ctx, "fsm "+initModel.Name())
	defer fsmEventsSpan.Finish()

	//carrier := make(opentracing.TextMapCarrier)
	//opentracing.GlobalTracer().Inject(fsmEventsSpan.Context(), opentracing.TextMap, carrier)
	// todo: правильное извлечение трейсID
	if spanCtx, ok := fsmEventsSpan.Context().(jaeger.SpanContext); ok {
		parentTraceID = spanCtx.TraceID().String()
		parentSpanID = spanCtx.SpanID().String()
	}

	tx.SetTraceID(parentTraceID)
	tx.SetSpanID(parentSpanID)

	// сохраняем сведения в БД
	err := e.repo.CreateTransaction(ctx, tx)
	if err != nil {
		return status.Errorf(codes.Internal, "create transaction error: %s", err.Error())
	}

	// создаем событие для обработки
	ev := model.NewEvent(initState, tx, 0)

	// отправляем сообщение в очередь
	initStateProcessor.Publish(ctx, ev)

	return nil
}

// Transit создает событие на проведение транзакции из одного состояния в другое
func (e *Engine) Transit(ctx context.Context, tx model.Tx, newState model.State) error {
	// транзакция должна быть передана
	if tx == nil {
		return status.Error(codes.NotFound, "trying to move empty tx")
	}

	// нельзя двинуться в пустое состояние
	if newState == nil {
		return status.Error(codes.PermissionDenied, "trying to move in nil state")
	}

	transitTx, ctx := opentracing.StartSpanFromContext(ctx, "transit tx", opentracing.Tag{
		Key:   "from_state",
		Value: tx.State().Name(),
	}, opentracing.Tag{
		Key:   "to_state",
		Value: newState.Name(),
	}, opentracing.Tag{
		Key:   "tx_id",
		Value: tx.ID().String(),
	})
	defer transitTx.Finish()

	// текущее состояние
	currState := tx.State()

	if currState == newState {
		transitTx.LogFields(log.Message("tx already in this state"))

		return status.Error(codes.PermissionDenied, "tx in this state already")
	}

	// проверяем можем ли сделать
	if !currState.CanTransitIn(newState) {
		transitTx.LogFields(log.Message("transition is illegal"))

		return status.Errorf(codes.PermissionDenied, "transition is illegal from %s to %s",
			currState.Name(), newState.Name())
	}

	// обработчик нового состояния
	nextStateProcessor, ok := e.states[newState]
	if !ok {
		transitTx.LogFields(log.Message("state processor not initialized"))

		return status.Errorf(codes.Internal, "%s not initialized", newState.Name())
	}

	tx.SetState(newState)

	// создаем событие для обработки
	ev := model.NewEvent(newState, tx, 0)

	// обновляем транзакцию в БД
	err := e.repo.UpdateTransaction(ctx, tx, currState.Name())
	if err != nil {
		return status.Errorf(codes.Internal, "update transaction error: %s", err.Error())
	}

	// отправляем событие в очередь
	nextStateProcessor.Publish(ctx, ev)

	return nil
}
