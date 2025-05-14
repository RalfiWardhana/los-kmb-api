package common

import (
	"errors"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/gorm"
)

var Key, ClientKey, Gender, StatusKonsumen, Channel, Lob, Incoming, Home, Education, Marital, ProfID, Photo, Relationship, AppSource, Address, Tenor, Relation, Decision string

var DB *gorm.DB

func SetDB(db *gorm.DB) {
	DB = db
}

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
	v.validator.RegisterValidation("prospect_id", prospectIDValidation)
	v.validator.RegisterValidation("key", checkClientKey)
	v.validator.RegisterValidation("dateformat", dateFormatValidation)
	v.validator.RegisterValidation("allowcharsname", allowedCharsInName)
	v.validator.RegisterValidation("isValidNameChars", isValidNameChars)
	v.validator.RegisterValidation("marital", maritalValidation)
	v.validator.RegisterValidation("gender", genderValidation)
	v.validator.RegisterValidation("spouse_gender", spouseGenderValidation)
	v.validator.RegisterValidation("profession", professionValidation)
	v.validator.RegisterValidation("bpkbname", checkBpkbname)
	v.validator.RegisterValidation("number", numberValidation)
	v.validator.RegisterValidation("id_number", idNumberValidation)
	v.validator.RegisterValidation("branch_id", branchIDValidation)
	v.validator.RegisterValidation("allow_name", allowedName)
	v.validator.RegisterValidation("customer_status", checkCustomerStatus)
	v.validator.RegisterValidation("customer_category", checkCustomerCategory)
	v.validator.RegisterValidation("result_pefindo", checkResultPefindo)
	v.validator.RegisterValidation("required_baki_debet", checkBakiDebet)
	v.validator.RegisterValidation("url", urlFormatValidation)
	v.validator.RegisterValidation("status_konsumen", consumentStatusValidation)
	v.validator.RegisterValidation("recom", recomValidation)
	v.validator.RegisterValidation("channel", channelValidation)
	v.validator.RegisterValidation("lob", lobValidation)
	v.validator.RegisterValidation("incoming", incomingValidation)
	v.validator.RegisterValidation("ktp", ktpValidation)
	v.validator.RegisterValidation("home", homeValidation)
	v.validator.RegisterValidation("npwp", npwpValidation)
	v.validator.RegisterValidation("address", addressValidation)
	v.validator.RegisterValidation("education", educationValidation)
	v.validator.RegisterValidation("marital", maritalValidation)
	v.validator.RegisterValidation("profession", professionValidation)
	v.validator.RegisterValidation("photo", photoValidation)
	v.validator.RegisterValidation("relationship", relationshipValidation)
	v.validator.RegisterValidation("relation", relationValidation)
	v.validator.RegisterValidation("appsource", appsourceValidation)
	v.validator.RegisterValidation("prospectID", ftrProspectIDValidation)
	v.validator.RegisterValidation("tenor", tenorValidation)
	v.validator.RegisterValidation("notnull", notNullValidation)
	v.validator.RegisterValidation("mustnull", mustNullValidation)
	v.validator.RegisterValidation("decision", DecisionValidation)
	v.validator.RegisterValidation("mobile_phone", MobilePhoneValidation)
	v.validator.RegisterValidation("tipe_usaha", tipeUsahaValidation)
	v.validator.RegisterValidation("af_principle", afValidationPrinciple)
	v.validator.RegisterValidation("admin_fee_principle", adminFeeValidationPrinciple)
	v.validator.RegisterValidation("installment_amount_principle", installmentAmountValidationPrinciple)
	v.validator.RegisterValidation("ntf_principle", ntfValidationPrinciple)
	v.validator.RegisterValidation("otr_principle", otrValidationPrinciple)
	v.validator.RegisterValidation("dealer_principle", dealerValidationPrinciple)
	v.validator.RegisterValidation("asset_category_id_principle", assetCategoryIDValidationPrinciple)
	v.validator.RegisterValidation("allowcharsaddress", allowedCharsInAddress)
	v.validator.RegisterValidation("htmlValidation", htmlValidation)
	v.validator.RegisterValidation("prospect_id_asset_principle", prospectIdAssetPrincipleNotExists)
	v.validator.RegisterValidation("prospect_id_pemohon_principle", prospectIdPemohonPrincipleNotExists)
	v.validator.RegisterValidation("prospect_id_pembiayaan_principle", prospectIdPembiayaanPrincipleNotExists)
	v.validator.RegisterValidation("prospect_id_emcon_principle", prospectIdEmconPrincipleNotExists)
	v.validator.RegisterValidation("allowcharstipeusaha", allowedCharsInTipeUsaha)
	v.sync.Unlock()

	return v.validator.Struct(i)
}

func prospectIDValidation(fl validator.FieldLevel) (validator bool) {

	prospectID := fl.Field().String()
	re := regexp.MustCompile(`^[A-Z]{2}[A-Z0-9-]*$`)
	if len(prospectID) < 10 || len(prospectID) > 20 {
		validator = false
	} else if re.MatchString(prospectID) {
		validator = true
	}

	return validator
}

func htmlValidation(fl validator.FieldLevel) (validator bool) {

	// Regular expression to detect HTML tags
	re := regexp.MustCompile(`<.*?>`)

	return !re.MatchString(fl.Field().String())
}

func dateFormatValidation(fl validator.FieldLevel) (validator bool) {

	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	layout := "2006-01-02"
	_, err := time.Parse(layout, fl.Field().String())

	if re.MatchString(fl.Field().String()) && err == nil {
		validator = true
	}
	return validator
}

func allowedCharsInName(fl validator.FieldLevel) bool {

	re := regexp.MustCompile("^[a-zA-Z.,'` ]*$")

	return re.MatchString(fl.Field().String())

}

func isValidNameChars(fl validator.FieldLevel) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.,'`_ -]*$")
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

	gender := entity.AppConfig{
		Value: "M,F",
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

func idNumberValidation(fl validator.FieldLevel) (validator bool) {

	s := fl.Field().String()
	idnumber, err := utils.PlatformDecryptText(s)
	if err != nil {
		validator = false
	} else if !regexp.MustCompile(`^[0-9]*$`).MatchString(idnumber) {
		validator = false
	} else if len(idnumber) != 16 {
		validator = false
	} else if idnumber[0:1] == "0" {
		validator = false
	} else {
		validator = true
	}

	return validator
}

func branchIDValidation(fl validator.FieldLevel) (validator bool) {

	branchID := fl.Field().String()
	if !regexp.MustCompile(`^[0-9]*$`).MatchString(branchID) {
		validator = false
	} else if len(branchID) != 3 {
		validator = false
	} else {
		validator = true
	}

	return validator
}

func allowedName(fl validator.FieldLevel) (validator bool) {

	s := fl.Field().String()
	name, err := utils.PlatformDecryptText(s)
	if err != nil {
		validator = false
	} else if len(name) > 100 {
		validator = false
	} else if !regexp.MustCompile("^[a-zA-Z.,'` ]*$").MatchString(name) {
		validator = false
	} else {
		validator = true
	}

	return validator
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
	customer_status := fl.Parent().FieldByName("CustomerStatus").String()

	if customer_status == constant.STATUS_KONSUMEN_RO_AO {
		if regular == fl.Field().String() || prime == fl.Field().String() || priority == fl.Field().String() {
			validator = true
		}
	} else {
		validator = true
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

func checkBakiDebet(fl validator.FieldLevel) (validator bool) {

	result_pefindo := fl.Parent().FieldByName("ResultPefindo").String()

	if result_pefindo == constant.DECISION_PASS {
		validator = true
	} else {
		if fl.Field().Interface() != nil {
			validator = true
		} else {
			validator = false
		}
	}

	return
}

func notNullValidation(fl validator.FieldLevel) bool {
	fmt.Println(fl.Field().Bool())
	return fl.Field().Bool()
}

func mustNullValidation(fl validator.FieldLevel) bool {
	fmt.Println(fl.Field().Bool())
	return fl.Field().Bool()
}

func ftrProspectIDValidation(fl validator.FieldLevel) bool {

	arr := strings.Split(fl.Field().String(), " - ")
	validator, _ := strconv.ParseBool(arr[1])

	return validator
}

func urlFormatValidation(fl validator.FieldLevel) bool {

	re := regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/)[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)

	return re.MatchString(fl.Field().String())
}

func ktpValidation(fl validator.FieldLevel) bool {
	return fl.Field().String() == "KTP"
}

func tenorValidation(fl validator.FieldLevel) bool {

	tenor, err := utils.ValidatorFromCache("group_tenor_kmob")

	if err != nil {
		return false
	}

	arrTenor := strings.Split(tenor.Value, ",")

	validator := contains(arrTenor, fl.Field().String())

	Tenor = tenor.Value

	return validator
}

func photoValidation(fl validator.FieldLevel) bool {

	Photo = "AKTA_CERAI,AKTA_KEMATIAN,ASSET_BELAKANG,ASSET_DEPAN,ASSET_KANAN,ASSET_KIRI,BPKB,BUKU_NIKAH,FOTO_JUAL_BELI,KK,KONSUMEN_KTP_CMO,KTP,LAINNYA_I,LAINNYA_II,NPWP,PLAT_NOMOR,SELFIE,SHM,SLIPGAJI,SPOUSE_KTP,STNK,RESULT_SURVEY,PENGHASILAN_PASANGAN,PENGHASILAN_LAIN"

	arrPhoto := strings.Split(Photo, ",")

	validator := contains(arrPhoto, fl.Field().String())

	return validator
}

func educationValidation(fl validator.FieldLevel) bool {

	education, err := utils.ValidatorFromCache("group_education")

	if err != nil {
		return false
	}

	arrEducation := strings.Split(education.Value, ",")

	validator := contains(arrEducation, fl.Field().String())

	Education = education.Value

	return validator

}

func homeValidation(fl validator.FieldLevel) bool {

	home, err := utils.ValidatorFromCache("group_home_status")

	if err != nil {
		return false
	}

	arrHome := strings.Split(home.Value, ",")

	validator := contains(arrHome, fl.Field().String())

	Home = home.Value

	return validator
}

func npwpValidation(fl validator.FieldLevel) (validator bool) {

	s := fl.Field().String()
	validator = true
	if s != "" {
		if !regexp.MustCompile(`^[0-9]*$`).MatchString(s) {
			validator = false
		} else if len(s) < 15 || len(s) > 16 {
			validator = false
		}
	}

	return validator
}

func lobValidation(fl validator.FieldLevel) bool {

	arrLob := []string{"KMB"}

	validator := contains(arrLob, fl.Field().String())

	Lob = "KMB"

	return validator
}

func incomingValidation(fl validator.FieldLevel) bool {

	incoming, err := utils.ValidatorFromCache("los_incoming_source")

	if err != nil {
		return false

	}

	arrIncoming := strings.Split(incoming.Value, ",")

	validator := contains(arrIncoming, fl.Field().String())

	Incoming = incoming.Value

	return validator
}

func channelValidation(fl validator.FieldLevel) bool {

	channel, err := utils.ValidatorFromCache("group_channel")

	if err != nil {
		return false

	}

	arrChannel := strings.Split(channel.Value, ",")

	validator := contains(arrChannel, fl.Field().String())

	Channel = channel.Value

	return validator
}

func consumentStatusValidation(fl validator.FieldLevel) (validator bool) {

	consument := "NEW,RO,AO"

	arrConsument := strings.Split(consument, ",")

	validator = contains(arrConsument, fl.Field().String())

	StatusKonsumen = consument

	return
}

func recomValidation(fl validator.FieldLevel) (validator bool) {

	if fl.Field().String() == "0" || fl.Field().String() == "1" {
		validator = true
		return
	}

	return

}

func addressValidation(fl validator.FieldLevel) (validator bool) {

	conf, err := utils.ValidatorFromCache("group_type_address_kmob")

	if err != nil {
		return false

	}

	arrAddress := strings.Split(conf.Value, ",")

	validator = contains(arrAddress, fl.Field().String())

	Address = conf.Value

	return validator

}

func relationshipValidation(fl validator.FieldLevel) bool {

	conf, err := utils.ValidatorFromCache("group_relationship")

	if err != nil {
		return false

	}

	arrRelation := strings.Split(conf.Value, ",")

	validator := contains(arrRelation, fl.Field().String())

	Relationship = conf.Value

	return validator
}

func appsourceValidation(fl validator.FieldLevel) bool {

	app, err := utils.ValidatorFromCache("group_application_source")

	if err != nil {
		return false

	}

	arrApp := strings.Split(app.Value, ",")

	validator := contains(arrApp, fl.Field().String())

	AppSource = app.Value

	return validator

}

func relationValidation(fl validator.FieldLevel) bool {

	conf, err := utils.ValidatorFromCache("group_relation_kmob")

	if err != nil {
		return false

	}

	arrRelation := strings.Split(conf.Value, ",")

	validator := contains(arrRelation, fl.Field().String())

	Relation = conf.Value

	return validator

}

func DecisionValidation(fl validator.FieldLevel) (validator bool) {

	decision := "APPROVE,REJECT"

	arrDecision := strings.Split(decision, ",")

	validator = contains(arrDecision, fl.Field().String())

	Decision = decision

	return
}

func MobilePhoneValidation(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		regex := regexp.MustCompile(`^08\d{7,12}$`)
		return regex.MatchString(fl.Field().String())
	}
	return true
}

func tipeUsahaValidation(fl validator.FieldLevel) bool {
	tipeUsaha := fl.Field().String()
	dealer := fl.Parent().FieldByName("Dealer").String()

	if dealer == "NON PSA" && tipeUsaha == "" {
		return false
	}
	return true
}

func afValidationPrinciple(fl validator.FieldLevel) bool {
	af := fl.Field().Float()

	prospectID := fl.Parent().FieldByName("ProspectID").String()

	var marketingProgram entity.TrxPrincipleMarketingProgram
	query := fmt.Sprintf(`SELECT TOP 1 FinanceAmount FROM trx_principle_marketing_program WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err := DB.Raw(query).Scan(&marketingProgram).Error; err != nil {
		return false
	}

	if af != marketingProgram.FinanceAmount {
		return false
	}

	return true
}

func adminFeeValidationPrinciple(fl validator.FieldLevel) bool {
	adminFee := fl.Field().Float()

	prospectID := fl.Parent().FieldByName("ProspectID").String()

	var marketingProgram entity.TrxPrincipleMarketingProgram
	query := fmt.Sprintf(`SELECT TOP 1 AdminFee FROM trx_principle_marketing_program WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err := DB.Raw(query).Scan(&marketingProgram).Error; err != nil {
		return false
	}

	if adminFee != marketingProgram.AdminFee {
		return false
	}

	return true
}

func installmentAmountValidationPrinciple(fl validator.FieldLevel) bool {
	installmentAmount := fl.Field().Float()

	prospectID := fl.Parent().FieldByName("ProspectID").String()

	var marketingProgram entity.TrxPrincipleMarketingProgram
	query := fmt.Sprintf(`SELECT TOP 1 InstallmentAmount FROM trx_principle_marketing_program WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err := DB.Raw(query).Scan(&marketingProgram).Error; err != nil {
		return false
	}

	if installmentAmount != marketingProgram.InstallmentAmount {
		return false
	}

	return true
}

func ntfValidationPrinciple(fl validator.FieldLevel) bool {
	ntf := fl.Field().Float()

	prospectID := fl.Parent().FieldByName("ProspectID").String()

	var marketingProgram entity.TrxPrincipleMarketingProgram
	query := fmt.Sprintf(`SELECT TOP 1 NTF FROM trx_principle_marketing_program WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err := DB.Raw(query).Scan(&marketingProgram).Error; err != nil {
		return false
	}

	if ntf != marketingProgram.NTF {
		return false
	}

	return true
}

func otrValidationPrinciple(fl validator.FieldLevel) bool {
	otr := fl.Field().Float()

	prospectID := fl.Parent().FieldByName("ProspectID").String()

	var marketingProgram entity.TrxPrincipleMarketingProgram
	query := fmt.Sprintf(`SELECT TOP 1 OTR FROM trx_principle_marketing_program WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err := DB.Raw(query).Scan(&marketingProgram).Error; err != nil {
		return false
	}

	if otr != marketingProgram.OTR {
		return false
	}

	return true
}

func dealerValidationPrinciple(fl validator.FieldLevel) bool {
	dealer := fl.Field().String()

	prospectID := fl.Parent().FieldByName("ProspectID").String()

	var marketingProgram entity.TrxPrincipleMarketingProgram
	query := fmt.Sprintf(`SELECT TOP 1 Dealer FROM trx_principle_marketing_program WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err := DB.Raw(query).Scan(&marketingProgram).Error; err != nil {
		return false
	}

	if dealer != marketingProgram.Dealer {
		return false
	}

	return true
}

func assetCategoryIDValidationPrinciple(fl validator.FieldLevel) bool {
	assetCategoryID := fl.Field().String()

	prospectID := fl.Parent().FieldByName("ProspectID").String()

	var marketingProgram entity.TrxPrincipleMarketingProgram
	query := fmt.Sprintf(`SELECT TOP 1 AssetCategoryID FROM trx_principle_marketing_program WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err := DB.Raw(query).Scan(&marketingProgram).Error; err != nil {
		return false
	}

	if assetCategoryID != marketingProgram.AssetCategoryID {
		return false
	}

	return true
}

func allowedCharsInAddress(fl validator.FieldLevel) bool {

	re := regexp.MustCompile("^[a-zA-Z0-9., ]+$")

	return re.MatchString(fl.Field().String())

}

func prospectIdAssetPrincipleNotExists(fl validator.FieldLevel) bool {
	prospectID := fl.Field().String()

	var principleStepOne entity.TrxPrincipleStepOne
	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_one WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	result := DB.Raw(query).Scan(&principleStepOne)

	if result.Error != nil {
		return errors.Is(result.Error, gorm.ErrRecordNotFound)
	}

	if principleStepOne != (entity.TrxPrincipleStepOne{}) {
		return false
	}

	return true
}

func prospectIdPemohonPrincipleNotExists(fl validator.FieldLevel) bool {
	prospectID := fl.Field().String()

	var principleStepTwo entity.TrxPrincipleStepTwo
	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_two WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	result := DB.Raw(query).Scan(&principleStepTwo)

	if result.Error != nil {
		return errors.Is(result.Error, gorm.ErrRecordNotFound)
	}

	if principleStepTwo != (entity.TrxPrincipleStepTwo{}) {
		return false
	}

	return true
}

func prospectIdPembiayaanPrincipleNotExists(fl validator.FieldLevel) bool {
	prospectID := fl.Field().String()

	var principleStepThree entity.TrxPrincipleStepThree
	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_three WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	result := DB.Raw(query).Scan(&principleStepThree)

	if result.Error != nil {
		return errors.Is(result.Error, gorm.ErrRecordNotFound)
	}

	if principleStepThree != (entity.TrxPrincipleStepThree{}) {
		return false
	}

	return true
}

func prospectIdEmconPrincipleNotExists(fl validator.FieldLevel) bool {
	prospectID := fl.Field().String()

	var principleEmergencyContact entity.TrxPrincipleEmergencyContact
	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_emergency_contact WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	result := DB.Raw(query).Scan(&principleEmergencyContact)

	if result.Error != nil {
		return errors.Is(result.Error, gorm.ErrRecordNotFound)
	}

	if principleEmergencyContact != (entity.TrxPrincipleEmergencyContact{}) {
		return false
	}

	return true
}

func allowedCharsInTipeUsaha(fl validator.FieldLevel) bool {

	re := regexp.MustCompile(`^[a-zA-Z.,'/ ` + "`" + `]*$`)

	return re.MatchString(fl.Field().String())
}
