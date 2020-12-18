package entity

import "github.com/tpreischadt/ProjetoJupiter/db"

type Grade struct {
	Grade int `firestore:"value,omitempty"`
}

func (g Grade) Insert(DB db.Env, collection string) error {
	_, _, err := DB.Client.Collection(collection).Add(DB.Ctx, g)
	if err != nil {
		return err
	}
	return nil
}
