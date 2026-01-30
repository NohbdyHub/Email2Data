package parsers

import (
	"fmt"
	"strings"
	"time"

	"github.com/bojanz/currency"
	html "golang.org/x/net/html"
)

type Parser interface {
	Parse(html string)
	ParseBatch()
}

type parkingInfo struct {
	date time.Time
	address string
	amount currency.Amount
}

type ParkingParser struct {
	info []parkingInfo
}

func (p *ParkingParser) Parse(h string) {
	doc, err := html.Parse(strings.NewReader(h))
	if err != nil {
		panic(err)
	}
	i := parkingInfo{}

	td := make([]*html.Node, 0)

	var traverse func(n *html.Node) *html.Node

	traverse = func(n *html.Node) *html.Node {
		for c := n.FirstChild; c != nil; c = c.NextSibling {

			if c.Type == html.ElementNode && c.Data == "td" {
				td = append(td, c)
			}

			res := traverse(c)

			if res != nil {

				return res
			}
		}

		return nil
	}

	traverse(doc)

	for _, r := range td {
		for c := r.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "strong" {
				switch c.FirstChild.Data {
					case "Start:":
						d, err := time.Parse("01/02/2006  3:04 PM (MST)", strings.Trim(c.NextSibling.Data, " \n\x0a"))
						if err != nil {
							panic(err)
						}
						i.date = d
					case "Address:":
						i.address = strings.Trim(c.NextSibling.Data, " \n\x0a")
					case "Amount:":
						a, err := currency.NewAmount(strings.Trim(c.NextSibling.Data, "$ \n\x0a"), "USD")
						if err != nil {
							panic(err)
						}

						i.amount = a
				}
			}
		}
	}

	p.info = append(p.info, i)
}

func (p *ParkingParser) ParseBatch() {
	var total currency.Amount
	for _, i := range p.info {
		fmt.Println(i.amount.Number())
		fmt.Println(i.address)
		fmt.Println(i.date.Format("01/02/2006"))
		t, err := total.Add(i.amount)
		if err != nil {
			panic(err)
		}

		total = t
	}

	fmt.Println(total.Number())
}
