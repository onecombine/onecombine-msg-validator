package src

type Validator interface {
	Sign(data string) string
	Verify(data, signature string) bool
}
