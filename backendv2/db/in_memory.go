package db

type InMemoryOptions struct {
}

type InMemoryClient struct {
}

func NewInMemoryClient() (*InMemoryClient, error) {
	return &InMemoryClient{}, nil
}

func (client *InMemoryClient) Close() {
}
