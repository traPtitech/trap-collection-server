package swift

type GameImage struct {
	client *Client
}

func NewGameImage(client *Client) *GameImage {
	return &GameImage{
		client: client,
	}
}
