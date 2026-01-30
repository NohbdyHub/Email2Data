package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
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

var mailService *gmail.Service
var sender string
var date string
var subject string

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
	sender = "noreply@premiumparking.com"
	date = "2026/01/01"
	subject = "\"Expired\""

	// google auth setup
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config := Must(google.ConfigFromJSON(b, gmail.GmailReadonlyScope))
	client := auth.GetClient(config)
	mailService = Must(gmail.NewService(context.Background(), option.WithHTTPClient(client)))
}

func main() {
	tPDF := make([]string, 0)
	t := Must(os.MkdirTemp("", "receipts"))
	defer os.RemoveAll(t)
	fmt.Println(t)

	// retrieve all mail from {sender} that was received after {date}, only including mail containing {subject} in the Subject line
	var b strings.Builder
	b.WriteString("from:")
	b.WriteString(sender)
	b.WriteString(" after:")
	b.WriteString(date)
	b.WriteString(" subject:")
	b.WriteString(subject)
	msgs := Must(mailService.Users.Messages.List("me").Q(b.String()).Do())
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
	fmt.Println(tPDF)
	pdfcpu.MergeCreateFile(tPDF, "Receipts.pdf", false, nil)
}
