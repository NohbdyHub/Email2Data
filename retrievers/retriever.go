package retrievers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"parking/util"
)

type Retriever interface {
	Retrieve(q Query) (rawHTML []string)
}

type Query struct {
	Sender string
	Date string
	Subject string
}

// loads config from given file. Creates config file populated with initial contents of {q} if none already exists
func (q *Query) Load(filename string) {
	qfile := util.Must(os.OpenFile(filename, os.O_CREATE | os.O_RDWR, 0666))
	defer qfile.Close()

	qJSON := util.Must(io.ReadAll(qfile))
	if len(qJSON) == 0 {
		qfile.Write(util.Must(json.MarshalIndent(q, "", "	")))
		fmt.Printf("Please ensure `%s` contains the correct information\n", filename)
		fmt.Println("Defaults loaded")
		fmt.Println("Date: \"auto\" to use current realtime month, MM/YYYY to filter to a specific month")
		fmt.Println("Subject: Only retrieve mail including this phrase in the Subject line")

		return
	}

	err := json.Unmarshal(qJSON, &q)
	if err != nil {
		panic(err)
	}
}
