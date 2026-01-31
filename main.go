package main

import (
	"fmt"

	"parking/parsers"
	_ "parking/parsers/premiumparking"
	"parking/retrievers"
)

func main() {
	fmt.Println("Startng!")
	g := retrievers.NewGmail()
	fmt.Println("Made Gmail API!")

	for _, p := range parsers.Available() {
		fmt.Println("Fetching!")
		p.Fetch(g)
		fmt.Println("PDFing!")
		p.PDF("receipts.pdf")
		fmt.Println("Customing!")
		p.CustomParse()
	}
}
