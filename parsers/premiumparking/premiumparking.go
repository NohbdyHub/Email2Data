package premiumparking

import (
	"fmt"
	"os"
	"strings"
	"time"

	"parking/parsers"
	"parking/retrievers"
	"parking/util"

	"github.com/bojanz/currency"
	html "golang.org/x/net/html"
)

type parkingInfo struct {
	date time.Time
	property string
	amount currency.Amount
}

func (p *parkingInfo) String() string {
	var b strings.Builder
	b.WriteString(p.date.Format("01/02/2006"))
	b.WriteRune(';')
	b.WriteString(p.property)
	b.WriteRune(';')
	b.WriteString(p.amount.Number())
	b.WriteRune('\n')

	return  b.String()
}

func parsePremiumParking(s []string) {
	allInfo := make([]parkingInfo, 0)
	for _, raw := range s {
		doc := util.Must(html.Parse(strings.NewReader(raw)))
		info := parkingInfo{}

		op := func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "td" {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "strong" {
						switch c.FirstChild.Data {
							case "Start:":
								d, err := time.Parse("01/02/2006  3:04 PM (MST)", strings.Trim(c.NextSibling.Data, " \n\x0a"))
								if err != nil {
									panic(err)
								}
								info.date = d
							case "Property name:":
								info.property = strings.Trim(c.NextSibling.Data, " \n\x0a")
							case "Amount:":
								a, err := currency.NewAmount(strings.Trim(c.NextSibling.Data, "$ \n\x0a"), "USD")
								if err != nil {
									panic(err)
								}

								info.amount = a
						}
					}
				}
			}
		}

		parsers.Breadth(doc, op)
		allInfo = append(allInfo, info)
	}

	csv := util.Must(os.Create("receipts.csv"))
	csv.WriteString("Date of Service;Name of Parking Garage/Lot;Amount Paid\n")
	for _, r := range allInfo {
		csv.WriteString(r.String())
	}
	fmt.Fprintf(csv, ";;=SUM(C2:C%d)", len(allInfo) + 1)
	csv.Close()
}

func init() {
	q := retrievers.Query{Sender: "noreply@premiumparking.com", Date: "auto", Subject: "Expired"}
	q.Load("premiumparking.json")
	parsers.Register(q, parsePremiumParking)
}
