package parsers

import (
	"context"
	"fmt"
	"os"

	html "golang.org/x/net/html"

	"parking/retrievers"

	"github.com/modernice/pdfire"
)

var parsers = make([]Parser, 0)

type Parser struct {
	sender string
	parse func([]string)
	htmlmail []string
}

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func Available() []Parser {
	return parsers
}

func Register(sender string, parse func([]string)) {
	fmt.Println("Registering ", sender)
	parsers = append(parsers, Parser{sender, parse, make([]string, 0)})
}

func (p *Parser) Fetch(r retrievers.Retriever) {
	p.htmlmail = r.Retrieve(p.sender)
}

func (p *Parser) PDF(filename string) {
	merge := pdfire.NewMergeOptions()
	for _, raw := range p.htmlmail {
		page := pdfire.NewConversionOptions()
		page.HTML = raw
		merge.Documents = append(merge.Documents, page)
	}

	r := Must(os.Create(filename))
	if err := pdfire.Merge(context.Background(), r, merge); err != nil {
		panic(err)
	}
	r.Close()
}

func (p *Parser) CustomParse() {
	p.parse(p.htmlmail)
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
