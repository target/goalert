package escalation

import (
	"github.com/target/goalert/validation/validate"
)

type Policy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Repeat      int    `json:"repeat"`
}

func (p Policy) Normalize() (*Policy, error) {
	err := validate.Many(
		validate.IDName("Name", p.Name),
		validate.Text("Description", p.Description, 1, 255),
		validate.Range("Repeat", p.Repeat, 0, 5),
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
