package retrievers

type Retriever interface {
	Retrieve(q Query) (rawHTML []string)
}

type Query struct {
	Sender string
	Date string
	Subject string
}

func NewRetriever(kind string) (r Retriever) {
	switch kind {
		case "gmail":
			r = newGmail()
		default:
			r = newGmail()
	}
	return
}
