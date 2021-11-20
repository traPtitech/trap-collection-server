package swift

type GameFile struct {
	client *Client
}

func NewGameFile(client *Client) *GameFile {
	return &GameFile{
		client: client,
	}
}
