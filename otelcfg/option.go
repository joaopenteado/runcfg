package otelcfg

type option[T any] interface {
	apply(cfg *T)
}

type optionFunc[T any] func(cfg *T)

func (fn optionFunc[T]) apply(cfg *T) {
	fn(cfg)
}
