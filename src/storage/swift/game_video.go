package swift

type GameVideo struct {
	client *Client
}

func NewGameVideo(client *Client) *GameVideo {
	return &GameVideo{
		client: client,
	}
}
