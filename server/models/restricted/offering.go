package restricted

import (
	"fmt"
	"net/http"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/Projeto-USPY/uspy-backend/db"
	"github.com/Projeto-USPY/uspy-backend/entity/controllers"
	"github.com/Projeto-USPY/uspy-backend/entity/models"
	"github.com/Projeto-USPY/uspy-backend/server/views/restricted"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetOfferingComments(ctx *gin.Context, DB db.Env, off *controllers.Offering) {
	collectionMask := "subjects/%s/offerings/%s/comments"
	subHash := models.Subject{
		Code:           off.Subject.Code,
		CourseCode:     off.Subject.CourseCode,
		Specialization: off.Subject.Specialization,
	}.Hash()

	// check if offering exists
	if _, err := DB.Restore("subjects/"+subHash+"/offerings", off.Hash); err != nil {
		if status.Code(err) == codes.NotFound {
			ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("could not find comments: %s", err.Error()))
			return
		}

		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to fetch comments: %s", err.Error()))
		return
	}

	// get comments
	snaps, err := DB.RestoreCollection(fmt.Sprintf(collectionMask, subHash, off.Hash))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to fetch comments: %s", err.Error()))
		return
	}

	comments := make([]*models.Comment, 0)
	for _, s := range snaps {
		var comm models.Comment
		if err := s.DataTo(&comm); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("could not bind data to comment: %s", err.Error()))
			return
		}

		comments = append(comments, &comm)
	}

	restricted.GetOfferingComments(ctx, comments)
}

// GetOfferings is a closure for the GET /api/restricted/offerings endpoint
func GetOfferingsWithStats(ctx *gin.Context, DB db.Env, sub *controllers.Subject) {
	model := models.NewSubjectFromController(sub)

	offerings := make([]*models.Offering, 0, 20)
	IDs := make([]string, 0, 20)
	stats := make([]*models.OfferingStats, 0, 20)

	offeringsMask := "subjects/%s/offerings"
	offeringsPath := fmt.Sprintf(offeringsMask, model.Hash())

	snaps, err := DB.RestoreCollection(offeringsPath)

	if err != nil {
		if status.Code(err) == codes.NotFound {
			ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("could not find collection offerings: %s", err.Error()))
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to fetch offerings: %s", err.Error()))
		return
	} else if len(snaps) == 0 {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	var wg sync.WaitGroup
	offsChannel := make(chan *models.Offering, len(snaps))
	IDchan := make(chan string, len(snaps))
	statsChan := make(chan *models.OfferingStats, len(snaps))

	for _, s := range snaps {
		wg.Add(1)
		go func(snap *firestore.DocumentSnapshot, wg *sync.WaitGroup) {
			defer wg.Done()
			var off models.Offering
			if err := snap.DataTo(&off); err != nil {
				ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("could not bind offering: %s", err.Error()))
				return
			}

			commentsPath := fmt.Sprintf("%s/%s/comments", offeringsPath, snap.Ref.ID)
			commentsCol := DB.Client.Collection(commentsPath)

			commentSnaps, err := commentsCol.Documents(DB.Ctx).GetAll()
			if err != nil {
				if status.Code(err) == codes.NotFound {
					ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("could not find collection comments: %s", err.Error()))
					return
				}
				ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("could not get offering comments: %s", err.Error()))
				return
			}

			var posQt, negQt, neutQt int
			for _, snap := range commentSnaps {
				if rating, err := snap.DataAt("rating"); err != nil {
					ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("could not get comment rating: %s", err.Error()))
					return
				} else {
					value := rating.(int64)

					if value < 3 {
						negQt++
					} else if value > 3 {
						posQt++
					} else {
						neutQt++
					}
				}
			}

			offsChannel <- &off
			IDchan <- snap.Ref.ID
			statsChan <- &models.OfferingStats{
				Approval:    posQt,
				Disapproval: negQt,
				Neutral:     neutQt,
			}
		}(s, &wg)
	}

	wg.Wait()

	close(offsChannel)
	close(IDchan)
	close(statsChan)

	for i := 0; i < len(snaps); i++ {
		offerings = append(offerings, <-offsChannel)
		IDs = append(IDs, <-IDchan)
		stats = append(stats, <-statsChan)
	}

	limit := len(IDs)
	if sub.Limit > 0 {
		limit = sub.Limit
	}

	restricted.GetOfferingsWithStats(ctx, IDs, offerings, stats, limit)
}
