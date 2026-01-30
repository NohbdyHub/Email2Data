package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	//	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"parking/auth"

	"github.com/ledongthuc/pdf"
	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/api"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GmailQuery struct {
	Sender string
	Date string
	Subject string
}

func (q *GmailQuery) String() string {
	var b strings.Builder

	var t time.Time
	if q.Date == "auto" {
		t = time.Now()
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	} else {
		t = Must(time.Parse("01/2006", q.Date))
	}

	b.WriteString("from:")
	b.WriteString(q.Sender)
	b.WriteString(" after:")
	b.WriteString(t.Format("2006/01/02"))
	b.WriteString(" before:")
	b.WriteString(t.AddDate(0, 1, 0).Format("2006/01/02"))
	b.WriteString(" subject:")
	b.WriteString("\"" + q.Subject + "\"")

	return b.String()
}

var mailService *gmail.Service
var query = GmailQuery{}

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func getHeader(h []*gmail.MessagePartHeader, n string) string {
	for _, header := range h {
		if header.Name == n {
			return header.Value
		}
	}
	return ""
}

func init() {
	pdf.DebugOn = true

	queryFile := Must(os.OpenFile("query.json", os.O_CREATE | os.O_RDWR, 0666))
	defer queryFile.Close()

	queryJSON := Must(io.ReadAll(queryFile))

	if len(queryJSON) == 0 {
		query = GmailQuery{Sender: "noreply@premiumparking.com", Date: "auto", Subject: "Expired"}
		queryFile.Write(Must(json.MarshalIndent(query, "", "	")))
		fmt.Println("Please ensure `query.json` has the correct information")
		fmt.Println("The default query will find all PremiumParking receipts from the current month")
		fmt.Println("Sender: Email address you recieve parking confirmations from")
		fmt.Println("Date: \"auto\" to use current realtime month, MM/YYYY to filter to a specific month")
		fmt.Println("Subject: Only retrieve mail including this phrase in the Subject line")
		os.Exit(0)
	} else {
		err := json.Unmarshal(queryJSON, &query)
		if err != nil {
			panic(err)
		}
	}

	// google auth setup
	b := Must(os.ReadFile("credentials.json"))

	oauth := Must(google.ConfigFromJSON(b, gmail.GmailReadonlyScope))
	client := auth.GetClient(oauth)
	mailService = Must(gmail.NewService(context.Background(), option.WithHTTPClient(client)))
}

func main() {
	tPDF := make([]string, 0)
	t := Must(os.MkdirTemp("", "receipts"))
	defer os.RemoveAll(t)
	fmt.Println(t)

	// retrieve all mail from {sender} that was received after {date}, only including mail containing {subject} in the Subject line
	msgs := Must(mailService.Users.Messages.List("me").Q(query.String()).Do())
	for _, msg := range msgs.Messages {
		email := Must(mailService.Users.Messages.Get("me", msg.Id).Format("full").Do())

		for j, p := range email.Payload.Parts {
			isHTML := false
			for _, h := range p.Headers {
				if h.Name == "Content-Type" && strings.Split(h.Value, ";")[0] == "text/html" {
					isHTML = true
					break
				}
			}
			if !isHTML {
				continue
			}

			// trim 2 bytes off html, invalid base64 enoding error at that byte?
			d := Must(base64.RawURLEncoding.DecodeString(p.Body.Data[0:len(p.Body.Data) - 2]))

			f := path.Join(t, msg.Id + strconv.FormatInt(int64(j), 10))
			o := Must(os.Create(f + ".html"))
			Must(o.WriteString(string(d)))
			o.Close()

			// render to pdf
			err := exec.Command("wkhtmltopdf", "--enable-local-file-access", f + ".html", f + ".pdf").Run()
			fmt.Println(err)
			tPDF = append(tPDF, f + ".pdf")
		}
	}

	// merge PDFs
	pdfcpu.MergeCreateFile(tPDF, "Receipts.pdf", false, nil)
}
