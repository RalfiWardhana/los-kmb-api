package common

import (
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/go-playground/validator.v9"
)

var Key, Gender, Tenor, Marital, ProfID, ClientKey string

func NewValidator() *Validator {

	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return &Validator{
		validator: validate,
	}
}

type Validator struct {
	validator *validator.Validate
	sync      sync.RWMutex
}

func (v *Validator) Validate(i interface{}) error {

	v.sync.Lock()
	v.validator.RegisterValidation("key", checkClientKey)
	v.validator.RegisterValidation("dateformat", dateFormatValidation)
	v.validator.RegisterValidation("allowcharsname", allowedCharsInName)
	v.validator.RegisterValidation("marital", maritalValidation)
	v.validator.RegisterValidation("gender", genderValidation)
	v.validator.RegisterValidation("spouse_gender", spouseGenderValidation)
	v.validator.RegisterValidation("profession", professionValidation)
	v.validator.RegisterValidation("bpkbname", checkBpkbname)
	v.validator.RegisterValidation("number", numberValidation)
	v.validator.RegisterValidation("customer_status", checkCustomerStatus)
	v.validator.RegisterValidation("customer_category", checkCustomerCategory)
	v.validator.RegisterValidation("result_pefindo", checkResultPefindo)
	v.sync.Unlock()

	return v.validator.Struct(i)
}

func dateFormatValidation(fl validator.FieldLevel) bool {

	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

	return re.MatchString(fl.Field().String())
}

func allowedCharsInName(fl validator.FieldLevel) bool {

	re := regexp.MustCompile("^[a-zA-Z.,'` ]*$")

	return re.MatchString(fl.Field().String())

}

func checkClientKey(fl validator.FieldLevel) (validator bool) {

	arrayKey := []string{os.Getenv("CLIENT_KEY")}

	validator = contains(arrayKey, fl.Field().String())

	ClientKey = os.Getenv("CLIENT_KEY")

	return
}

func checkBpkbname(fl validator.FieldLevel) (validator bool) {

	namaSama := utils.AizuArrayString(os.Getenv("NAMA_SAMA"))
	namaBeda := utils.AizuArrayString(os.Getenv("NAMA_BEDA"))

	if contains(namaSama, fl.Field().String()) || contains(namaBeda, fl.Field().String()) {
		validator = true
	} else {
		validator = false
	}

	return
}

func genderValidation(fl validator.FieldLevel) (validator bool) {

	gender, err := utils.ValidatorFromCache("group_gender")

	if err != nil {
		return
	}

	arrGender := strings.Split(gender.Value, ",")

	validator = contains(arrGender, fl.Field().String())

	Gender = gender.Value

	return
}

func maritalValidation(fl validator.FieldLevel) bool {

	marital, err := utils.ValidatorFromCache("group_marital_status")

	if err != nil {
		return false
	}

	arrMarital := strings.Split(marital.Value, ",")

	validator := contains(arrMarital, fl.Field().String())

	Marital = marital.Value

	return validator
}

func spouseGenderValidation(fl validator.FieldLevel) bool {

	return fl.Field().Bool()
}

func professionValidation(fl validator.FieldLevel) bool {

	profID, err := utils.ValidatorFromCache("group_professionID")

	if err != nil {
		return false
	}

	arrProfID := strings.Split(profID.Value, ",")

	validator := contains(arrProfID, fl.Field().String())

	ProfID = profID.Value

	return validator
}

func numberValidation(fl validator.FieldLevel) bool {

	re := regexp.MustCompile(`^[0-9]*$`)

	return re.MatchString(fl.Field().String())
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func checkCustomerStatus(fl validator.FieldLevel) (validator bool) {

	customer_new := constant.STATUS_KONSUMEN_NEW
	customer_roao := constant.STATUS_KONSUMEN_RO_AO

	if customer_new == fl.Field().String() || customer_roao == fl.Field().String() {
		validator = true
	} else {
		validator = false
	}

	return
}

func checkCustomerCategory(fl validator.FieldLevel) (validator bool) {

	regular := constant.RO_AO_REGULAR
	prime := constant.RO_AO_PRIME
	priority := constant.RO_AO_PRIORITY

	if fl.Field().String() == "" || regular == fl.Field().String() || prime == fl.Field().String() || priority == fl.Field().String() {
		validator = true
	} else {
		validator = false
	}

	return
}

func checkResultPefindo(fl validator.FieldLevel) (validator bool) {

	pass := constant.DECISION_PASS
	reject := constant.DECISION_REJECT

	if pass == fl.Field().String() || reject == fl.Field().String() {
		validator = true
	} else {
		validator = false
	}

	return
}
