package internal

func Predicate(pick bool) func() bool {
	return func() bool { return pick }
}
