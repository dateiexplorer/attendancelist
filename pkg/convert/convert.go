package convert

// A Converter provides the functions to convert data in any file format.
type Converter interface {
	Header() []string
	NextEntry() <-chan []string
}
