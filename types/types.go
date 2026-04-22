package types

type PbOneOf[R any] interface {
	PbStruct() R
}

type OneOf[R any, V any] struct {
	Value  V
	pbCtor func(V) R
}

func NewOneOf[R any, V any](
	value V,
	pbCtor func(V) R,
) PbOneOf[R] {
	return &OneOf[R, V]{
		Value:  value,
		pbCtor: pbCtor,
	}
}

func (of OneOf[R, V]) PbStruct() R {
	return of.pbCtor(of.Value)
}
