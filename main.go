package main

import (
	"context"
	"embed"
	"errors"
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

type skip struct {
	q *retrievers.Query
	r error
}

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

func createDir(name string) []os.DirEntry {
	dirs, err := os.ReadDir(name)

	if err != nil {
		err = os.Mkdir(name, 0777)
		if err != nil {
			panic(err)
		}
	}
	return dirs
}

//go:embed example.json
var example embed.FS

// loads config from given file. Creates config file populated with initial contents of {q} if none already exists
var queries []retrievers.Query

func init() {
	qDir := createDir("searches")
	_ = createDir("config")
	_ = createDir("output")

	if len(qDir) == 0 {
		b := util.Must(example.ReadFile("example.json"))
		e := util.Must(os.Create(path.Join("searches", "example.json")))
		defer e.Close()

		util.Must(e.Write(b))
	}

	for _, f := range qDir {
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

	skips := make([]skip, 0)
	for _, q := range queries {
		if q.Sender == "example@address.com" {
			skips = append(skips, skip{&q, errors.New("Just an example!")})
			continue
		}

		spin := util.Spinner(fmt.Sprintf("Getting mail from <%s>", q.Sender), fmt.Sprintf("Got mail from <%s>", q.Sender))
		raw := r.Retrieve(q)
		spin.Close()

		if len(raw) == 0 {
			skips = append(skips, skip{&q, errors.New("No results")})
			continue
		}

		p := path.Join("output", "receipts.pdf")
		spin = util.Spinner(fmt.Sprintf("Saving to <%s>", p), fmt.Sprintf("Saved to <%s>", p))
		pdfPrint(raw, p)
		spin.Close()

		for _, p := range parsers.Available() {
			p.Parse(q, raw)
		}
	}

	if len(skips) == 1 && len(queries) == 1 {
		fmt.Println("See searches/example.json to set up a search")
	}

	for _, s := range skips {
		fmt.Println(s.r, "for search:", s.q)
	}
}
