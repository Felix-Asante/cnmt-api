package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

var (
	validate *validator.Validate
	trans    ut.Translator
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Fields []FieldError `json:"fields"`
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

func InitValidator() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	locale := en.New()
	uni := ut.New(locale, locale)
	translator, ok := uni.GetTranslator("en")
	if !ok {
		panic("failed to get English translator")
	}
	trans = translator

	if err := enTranslations.RegisterDefaultTranslations(validate, trans); err != nil {
		panic(err)
	}
}

func DecodeAndValidate[T any](r *http.Request) (T, error) {
	var body T
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&body); err != nil {
		return body, InvalidJSONError
	}
	if err := validate.Struct(body); err != nil {
		return body, translateValidationError(err)
	}
	return body, nil
}

func translateValidationError(err error) error {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return err
	}

	fields := make([]FieldError, 0, len(verrs))
	for _, fe := range verrs {
		fields = append(fields, FieldError{
			Field:   fieldPath(fe),
			Message: fe.Translate(trans),
		})
	}
	return &ValidationError{Fields: fields}
}

func fieldPath(fe validator.FieldError) string {
	ns := fe.Namespace()
	if i := strings.Index(ns, "."); i >= 0 {
		return ns[i+1:]
	}
	return fe.Field()
}
