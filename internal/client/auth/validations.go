package auth

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

func (c *controller) Struct(params interface{}) (map[string]string, error) {
	report := make(map[string]string)

	err := c.validate.Struct(params)
	if err != nil {
		for _, fe := range err.(validator.ValidationErrors) {
			text := getErrorMessage(fe)
			if text != "" {
				report[fe.Field()] = text
			}
		}

		return report, err
	}

	return nil, nil
}

func (c *controller) registerTagNameFunc(validate *validator.Validate) {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "обязательное поле"
	case "url":
		return "неправильный формат ссылки"
	case "email":
		return "неправильный формат электронной почты"
	case "oneof":
		return "допустимые значения - " + fe.Param()
	case "lte":
		return "значение не должно превышать - " + fe.Param()
	case "gte":
		return "значение должно быть не меньше - " + fe.Param()
	case "number":
		return "должно быть число"
	case "eqfield":
		return "значение должно совпадать с полем " + fe.Param()
	default:
		return "неизвестная ошибка"
	}
}
