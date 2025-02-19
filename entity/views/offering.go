package views

import (
	"sort"

	"github.com/Projeto-USPY/uspy-backend/entity/models"
)

type Offering struct {
	ProfessorName string   `json:"professor"`
	ProfessorCode string   `json:"code"`
	Years         []string `json:"years"`

	Approval    float64 `json:"approval"`
	Neutral     float64 `json:"neutral"`
	Disapproval float64 `json:"disapproval"`
}

func SortOfferings(results []*Offering) {
	sort.SliceStable(results,
		func(i, j int) bool {
			// sort by ratings
			ithApproval, jthApproval := (results[i].Approval + results[i].Neutral), (results[j].Approval + results[j].Neutral)

			if ithApproval == jthApproval {
				if results[i].Disapproval == results[j].Disapproval {
					// if ratings are the same, show latest or most offerings
					sizeI, sizeJ := len(results[i].Years), len(results[j].Years)
					if results[i].Years[sizeI-1] == results[j].Years[sizeJ-1] {
						return len(results[i].Years) > len(results[j].Years)
					}

					return results[i].Years[sizeI-1] > results[j].Years[sizeJ-1]
				}

				return results[i].Disapproval < results[j].Disapproval
			}

			return ithApproval > jthApproval
		},
	)
}

func NewOfferingFromModel(ID string, model *models.Offering, approval, disapproval, neutral int) *Offering {
	total := (approval + disapproval + neutral)

	approvalRate := 0.0
	disapprovalRate := 0.0
	neutralRate := 0.0

	if total != 0 {
		approvalRate = float64(approval) / float64(total)
		disapprovalRate = float64(disapproval) / float64(total)
		neutralRate = float64(neutral) / float64(total)
	}

	return &Offering{
		ProfessorName: model.Professor,
		ProfessorCode: ID,
		Years:         model.Years,
		Approval:      approvalRate,
		Disapproval:   disapprovalRate,
		Neutral:       neutralRate,
	}
}

func NewPartialOfferingFromModel(ID string, model *models.Offering) *Offering {
	return &Offering{
		ProfessorName: model.Professor,
		ProfessorCode: ID,
		Years:         model.Years,
	}
}
