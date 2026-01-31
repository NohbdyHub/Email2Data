package parsers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bojanz/currency"
	html "golang.org/x/net/html"
)

type parkingInfo struct {
	date time.Time
	address string
	amount currency.Amount
}

func (p *parkingInfo) String() string {
	var b strings.Builder
	b.WriteString(p.date.Format("01/02/2006"))
	b.WriteRune(';')
	b.WriteString(p.address)
	b.WriteRune(';')
	b.WriteString(p.amount.Number())
	b.WriteRune('\n')

	return  b.String()
}

type ParkingParser struct {
	info []parkingInfo
}

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
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
		t, err := total.Add(i.amount)
		if err != nil {
			panic(err)
		}

		total = t
	}

	csv := Must(os.Create("receipts.csv"))
	csv.WriteString("Date of Service;Name of Parking Garage/Lot;Amount Paid\n")
	for _, r := range p.info {
		csv.WriteString(r.String())
	}
	fmt.Fprintf(csv, ";;=SUM(C2:C%d)", len(p.info) + 1)
	csv.Close()
}

func (p *ParkingParser) IsParseable(addr string) bool {
	return addr == "noreply@premiumparking.com"
}

func init() {
	Parsers = append(Parsers, &ParkingParser{})
}
