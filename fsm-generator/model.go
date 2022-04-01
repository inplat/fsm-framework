package fsm_generator

import (
	"fmt"
	"strings"
	"time"

	"github.com/goccy/go-graphviz/cgraph"
)

// ModelDefaultConfig настройки
type ModelDefaultConfig struct {
	// MaxRetryCount максимальное количество повторений
	MaxRetryCount int `yaml:"max_retry_count"`
	// MinRetryDelay минимальная задержка в обработке событий текущего состояния
	MinRetryDelay time.Duration `yaml:"min_retry_delay"`
	// CancellationTTL время после которого неизмененная в текущем состоянии транзакция считается отмененной
	CancellationTTL time.Duration `yaml:"cancellation_ttl"`
}

// Transition разрешенный переход из одного состояния в другое
type Transition struct {
	// StateName название состояния, в которое осуществляется переход
	StateName string `yaml:"to"`
	// Condition описание условий перехода
	Condition string `yaml:"condition"`

	// State структура, найденная по названию состояния
	State *State `yaml:"-"`
}

type State struct {
	// Name название состояния в SCREAMING_SNAKE_CASE
	Name string `yaml:"name"`
	// Description описание состояния
	Description string `yaml:"description"`
	// Initial является ли состояние начальным
	Initial bool `yaml:"initial"`
	// SuccessFinal является ли состояние успешным
	SuccessFinal bool `yaml:"success_final"`
	// FailFinal является ли состояние конечным неудачным
	FailFinal bool `yaml:"fail_final"`
	// DisableFallbackState флаг, отключающий автоматическое создание fallback failed state
	DisableFallbackState bool `yaml:"disable_fallback_state"`
	// MaxRetryCount максимальное количество повторений
	MaxRetryCount int `yaml:"max_retry_count"`
	// MinRetryDelay минимальная задержка в обработке событий текущего состояния
	MinRetryDelay time.Duration `yaml:"min_retry_delay"`
	// CancellationTTL время после которого неизмененная в текущем состоянии транзакция считается отмененной
	CancellationTTL time.Duration `yaml:"cancellation_ttl"`
	// Transitions список разрешенных переходов из текущего состояния
	Transitions []*Transition `yaml:"transitions"`

	// GraphNode структура, создаваемая в рамках отрисовки графа
	GraphNode *cgraph.Node `yaml:"-"`
}

// Model fsm-модель, содержащая описания состояний и переходы между ними
type Model struct {
	// Name название модели на английском в lower_case (берется как название yaml файла)
	Name string `yaml:"-"`
	// ETag hash в hex, по нему можно понять изменилась ли модель
	ETag string `yaml:"-"`
	// Title кириллическое название модели
	Title string `yaml:"title"`
	// DefaultConfig настройки переходов между состояниями по умолчанию
	DefaultConfig *ModelDefaultConfig `yaml:"default_config"`
	// States список состояний модели
	States []*State `yaml:"states"`
}

func (m *Model) Prefix() string {
	return strings.ToUpper(m.Name) + "_TX_"
}

func (m *Model) PrintStates() {
	fmt.Printf("| --- %-20s", m.Title+" ("+m.Name+" v."+m.ETag+") ---\n")

	prefix := m.Prefix()

	for _, state := range m.States {
		fmt.Printf("- %-50s -> %s\n", prefix+state.Name, state.Description)
	}

	fmt.Printf("| ---------------------------------------\n\n")
}
