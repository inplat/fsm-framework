package model

type Model interface {
	Name() string
	States() []State
	Resolve(name string) State
	Has(state State) bool
	SetEngine(engine Engine)
	Engine() Engine
	SetService(svc interface{}) error
	Service() interface{}
}
