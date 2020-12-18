package entity

import (
	"cloud.google.com/go/firestore"
	"crypto/md5"
	"fmt"
	"github.com/tpreischadt/ProjetoJupiter/db"
)

// Course represents a course/major (example: BCC)
type Course struct {
	Name        string                   `firestore:"name"`
	Code        string                   `firestore:"code"`
	SubjectsRef *firestore.CollectionRef `firestore:"subjects"`
	Subjects    []Subject
}

func (c Course) Hash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(c.Code)))
}

func (c Course) Insert(DB db.Env, collection string) error {
	_, err := DB.Client.Collection(collection).Doc(c.Hash()).Set(DB.Ctx, c)
	if err != nil {
		return err
	}

	return nil
}
