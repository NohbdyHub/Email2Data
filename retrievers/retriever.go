package retrievers

type Retriever interface {
	Retrieve(from string) (rawHTML []string)
}

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}
