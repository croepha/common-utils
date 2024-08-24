package lostandfound

// Truly misc things that have no home of their own

// Small helper to make it easy to reset a value
func SetReset[T any](variable *T, newValue T) func() {
	originalValue := *variable
	*variable = newValue
	return func() {
		// TODO: Add a runtime check to ensure that func gets called?
		*variable = originalValue
	}
}

// Removes all pending elements from a channel,
// returns when channel is exhausted
func DainChannel[T any](c <-chan T) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}
