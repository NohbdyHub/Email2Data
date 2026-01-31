package parsers

type Parser interface {
	Parse(html string)
	ParseBatch()
	IsParseable(addr string) bool
}

var Parsers = make([]Parser, 0)
