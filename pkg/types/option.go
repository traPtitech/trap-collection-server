package types

// Option
// optionalな値を表す型。
// 初期値はinvalid、validな値を作る場合はNewOptionを使う。
type Option[T any] struct {
	valid bool
	value T
}

// NewOption
// validなOptionを作る。
func NewOption[T any](value T) Option[T] {
	return Option[T]{
		valid: true,
		value: value,
	}
}

// Value
// 値を取得する。
// 値がinvalidの場合、falseを返す。
func (o Option[T]) Value() (T, bool) {
	return o.value, o.valid
}
