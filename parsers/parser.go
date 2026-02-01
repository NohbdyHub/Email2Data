package parsers

import (
	"fmt"

	html "golang.org/x/net/html"

	"parking/retrievers"
)

var parsers = make([]Parser, 0)

type Parser struct {
	query retrievers.Query
	parse func([]string)
}

func Available() []Parser {
	return parsers
}

func Register(query retrievers.Query, parse func([]string)) {
	fmt.Println("Registering ", query.Sender)
	parsers = append(parsers, Parser{query, parse})
}


func (p *Parser) Parse(query retrievers.Query, raw []string) {
	if p.query.Sender == query.Sender && p.query.Subject == query.Subject {
		p.parse(raw)
	}
}

func Breadth(n *html.Node, op func(*html.Node)) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		op(c)

		res := Breadth(c, op)

		if res != nil {

			return res
		}
	}

	return nil
}
