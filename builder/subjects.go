// package builder contains useful functions for building the Firestore database
// Use with caution, because it can overwrite most data present in the database, including reviews and statistics
package builder

import (
	"github.com/tpreischadt/ProjetoJupiter/crawler/icmc/subject"
	"github.com/tpreischadt/ProjetoJupiter/db"
	"log"
	"sync"
)

type SubjectBuilder struct{}

// SubjectBuilder.Build scrapes all icmc courses and subjects and builds them onto Firestore
func (SubjectBuilder) Build(DB db.Env) error {
	log.Println("scraping subjects")
	courses, err := subject.ScrapeICMCCourses()
	log.Println("done")
	if err != nil {
		return err
	}

	objs := make([]db.Object, 0)
	for _, course := range courses {
		courseSubNames := make(map[string]string)
		for _, sub := range course.Subjects {
			sub.Stats = map[string]int{
				"worth_it": 0,
				"total":    0,
			}
			objs = append(objs, db.Object{Collection: "subjects", Doc: sub.Hash(), Data: sub})
			courseSubNames[sub.Code] = sub.Name
		}
		course.SubjectCodes = courseSubNames
		objs = append(objs, db.Object{Collection: "courses", Doc: course.Hash(), Data: course})
	}

	var wg sync.WaitGroup
	for _, o := range objs {
		var err error

		wg.Add(1)
		go func(group *sync.WaitGroup) {
			defer group.Done()
			err = DB.Insert(o.Data, o.Collection)
		}(&wg)

		log.Printf("inserting %v into %v\n", o.Doc, o.Collection)
		if err != nil {
			return err
		}
	}

	wg.Wait()
	log.Printf("inserted %d total objects\n", len(objs))

	return nil
}
