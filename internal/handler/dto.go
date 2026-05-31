package handler

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Data struct {
	Website string `json:"website" validate:"required,url"`
}

var validate = validator.New()

func (d *Data) Validate() error {
	err := validate.Struct(d)
	if err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			return fmt.Errorf("validation failed: %w", err)
		}
		return err
	}

	return nil
}

func HandleValidationErr(err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			fmt.Printf("Field '%s' failed on the '%s' tag\n",
				fieldErr.Field(),
				fieldErr.Tag(),
			)
		}
	}
}
