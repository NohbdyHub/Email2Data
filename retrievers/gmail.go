package retrievers

import (
	"context"
	"encoding/base64"
	"os"
	"strings"
	"time"

	"parking/auth"
	"parking/util"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func GmailQuery(q Query) string {
	var b strings.Builder

	var t time.Time
	if q.Date == "auto" {
		t = time.Now()
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	} else {
		t = util.Must(time.Parse("01/2006", q.Date))
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

type Gmail struct {
	*gmail.Service
}

func newGmail() Gmail {
	// google auth setup
	b := util.Must(os.ReadFile("credentials.json"))

	oauth := util.Must(google.ConfigFromJSON(b, gmail.GmailReadonlyScope))
	client := auth.GetClient(oauth)
	g := util.Must(gmail.NewService(context.Background(), option.WithHTTPClient(client)))

	return Gmail{g}
}


func (g Gmail) Retrieve(q Query) (rawHTML []string) {
	msgs := util.Must(g.Users.Messages.List("me").Q(GmailQuery(q)).Do())
	for _, msg := range msgs.Messages {
		m := util.Must(g.Users.Messages.Get("me", msg.Id).Format("full").Do())
		for _, p := range m.Payload.Parts {
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

			rawHTML = append(rawHTML, string(util.Must(base64.URLEncoding.DecodeString(p.Body.Data))))
		}
	}
	return
}
