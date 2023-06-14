package algorithms

type Validator interface {
	Sign(data string, options ...string) string
	Verify(data []byte, signature string) bool
}
