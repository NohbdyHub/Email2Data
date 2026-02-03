package retrievers

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path"
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
	cred := path.Join("config", "credentials.json")
	b, err := os.ReadFile(cred)
	if err != nil {
		fmt.Println("No API credentials found in", cred, ". To get credentials, complete each of the steps at the following link according to these instructions: ")
		fmt.Println("https://developers.google.com/workspace/guides/get-started")
		fmt.Println("")

		fmt.Println("Step 1: Create a Google Cloud Project")
		fmt.Println("- Click \"Go to Create a Project\", Name can be whatever, and leave Location as \"No organization\".")
		fmt.Println("")

		fmt.Println("Step 2: Enable the APIs you want to use")
		fmt.Println("- Click \"Go to Product Library\", click \"Gmail API\", click \"Enable\".")
		fmt.Println("")

		fmt.Println("Step 3: Learn how authentication and authorization works")
		fmt.Println("- SKIP. Optional reading.")
		fmt.Println("")

		fmt.Println("Step 4: Configure OAuth consent")
		fmt.Println("- Click \"Get started\". App Information can be whatever. Audience must be EXTERNAL. Contact Information is your own.")
		fmt.Println("- Click \"Create OAuth client\". Application type is \"Web application\". Name can be whatever.")
		fmt.Println("- Under \"Authorized redirect URIs\", click \"Add URI\", and paste the following: https://localhost:6969")
		fmt.Println("- Click \"Create\". Click \"Download JSON\". Rename to \"credentials.json\", and move to the \"config\" folder.")
		fmt.Println("- If you lose or delete your credentials.json file, delete the client you just made and repeat this step.")
		fmt.Println("")

		fmt.Println("Step 5: Let yourself use your own app (dumb I know)")
		fmt.Println("- Navigate to \"https://console.cloud.google.com/auth/audience\". ")
		fmt.Println("- Under \"Test users\", click \"Add users\". Enter any email addresses you will want to use this program with.")
		fmt.Println("")

		fmt.Println("DONE. Run this program again!")
		os.Exit(1)
	}

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
