package util

type Tuple[T1 any, T2 any] struct {
	Item1 T1
	Item2 T2
}

func NewTuple[T1 any, T2 any](item1 T1, item2 T2) Tuple[T1, T2] {
	return Tuple[T1, T2]{item1, item2}
}
func (t Tuple[T1, T2]) GetItem1() T1 {
	return t.Item1
}

func (t Tuple[T1, T2]) GetItem2() T2 {
	return t.Item2
}

type Tuple3[T1 any, T2 any, T3 any] struct {
	Item1 T1
	Item2 T2
	Item3 T3
}

func NewTuple3[T1 any, T2 any, T3 any](item1 T1, item2 T2, item3 T3) Tuple3[T1, T2, T3] {
	return Tuple3[T1, T2, T3]{item1, item2, item3}
}
func (t Tuple3[T1, T2, T3]) GetItem1() T1 {
	return t.Item1
}

func (t Tuple3[T1, T2, T3]) GetItem2() T2 {
	return t.Item2
}

func (t Tuple3[T1, T2, T3]) GetItem3() T3 {
	return t.Item3
}
