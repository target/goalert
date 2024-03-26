package alert

import (
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

const TypeAlertMetaV1 = "alert_meta_v1"

type MetaData map[string]string

type Meta struct {
	Type        string
	AlertMetaV1 MetaData
}

func (m MetaData) Normalize() error {
	var totalSize int
	for k, v := range m {
		err := validate.ASCII("Meta[<key>]", k, 1, 255)
		if err != nil {
			return err
		}

		totalSize += len(k) + len(v)
	}

	if totalSize > 32768 {
		return validation.NewFieldError("Meta", "cannot exceed 32KiB in size")
	}
	return nil
}
