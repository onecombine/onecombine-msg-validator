package src

type Validator interface {
	Sign(data string, options ...string) string
	Verify(data, signature string) bool
}
