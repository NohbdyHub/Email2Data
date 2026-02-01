package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"path"

	"parking/parsers"
	_ "parking/parsers/premiumparking"
	"parking/retrievers"
	"parking/util"

	"github.com/modernice/pdfire"
	"github.com/yosuke-furukawa/json5/encoding/json5"
)

func pdfPrint(HTMLpages []string, outfile string) {
	merge := pdfire.NewMergeOptions()
	for _, raw := range HTMLpages {
		page := pdfire.NewConversionOptions()
		page.HTML = raw
		merge.Documents = append(merge.Documents, page)
	}

	r := util.Must(os.Create(outfile))
	if err := pdfire.Merge(context.Background(), r, merge); err != nil {
		panic(err)
	}
	r.Close()
}

//go:embed example.json
var example embed.FS
// loads config from given file. Creates config file populated with initial contents of {q} if none already exists
var queries []retrievers.Query
func init() {
	qFiles, err := os.ReadDir("searches")

	if err != nil {
		err = os.Mkdir("searches", 0777)
		if err != nil {
			panic(err)
		}
	}

	if len(qFiles) == 0 {
		b := util.Must(example.ReadFile("example.json"))
		e := util.Must(os.Create(path.Join("searches", "example.json")))
		defer e.Close()

		util.Must(e.Write(b))
	}

	for _, f := range qFiles {
		if !f.IsDir() {
			p := path.Join("searches", f.Name())
			qfile := util.Must(os.Open(p))
			defer qfile.Close()

			qJSON, err := io.ReadAll(qfile)
			if err != nil {
				fmt.Println("Error Reading from ", p, err)
				continue
			}

			q := retrievers.Query{}
			err = json5.Unmarshal(qJSON, &q)
			if err != nil {
				fmt.Println("Error Loading ", p, err)
				continue
			}
			queries = append(queries, q)
		}
	}
}

func main() {
	r := retrievers.NewRetriever("gmail")

	skip := 0
	for _, q := range queries {
		if q.Sender == "example@address.com" {
			skip++
			continue
		}
		raw := r.Retrieve(q)
		pdfPrint(raw, "receipts.pdf")
		for _, p := range parsers.Available() {
			p.Parse(q, raw)
		}
	}

	if skip == len(queries) {
		fmt.Println("Nothing happened!")
		fmt.Println("Make sure you add a search in the \"searches\" folder!")
	}
}
