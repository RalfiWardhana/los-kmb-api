package request

import (
	"los-kmb-api/models/entity"
	"time"
)

type FilteringRequest struct {
	ClientKey string `json:"client_key" validate:"required,key"`
	Data      Data   `json:"data" validate:"required"`
}

type BodyRequestKreditmu struct {
	ClientKey      string `json:"client_key" validate:"required,key"`
	Data           Data   `json:"data" validate:"required"`
	StatusKonsumen string `json:"status_konsumen"`
}

type Blacklist struct {
	IDNumber          string `json:"IDNumber" validate:"required"`
	LegalName         string `json:"LegalName" validate:"required,allowcharsname"`
	BirthPlace        string `json:"BirthPlace" validate:"required,allowcharsname"`
	BirthDate         string `json:"BirthDate" validate:"required,"`
	SurgateMotherName string `json:"SurgateMotherName" validate:"required,allowcharsname"`
	Gender            string `json:"Gender" validate:"required,"`
	MaritalStatus     string `json:"MaritalStatus" validate:"required,"`
	Spouse            Spouse `json:"Spouse" validate:"omitempty"`
}

type Data struct {
	BPKBName          string  `json:"BPKBName" validate:"required,bpkbname"`
	ProspectID        string  `json:"ProspectID" validate:"required,max=20"`
	BranchID          string  `json:"BranchID" validate:"required"`
	IDNumber          string  `json:"IDNumber" validate:"required,number,len=16"`
	LegalName         string  `json:"LegalName" validate:"required,allowcharsname"`
	BirthPlace        string  `json:"BirthPlace" validate:"required,allowcharsname"`
	BirthDate         string  `json:"BirthDate" validate:"required,dateformat"`
	SurgateMotherName string  `json:"SurgateMotherName" validate:"required,allowcharsname"`
	Gender            string  `json:"Gender" validate:"required,gender"`
	MaritalStatus     string  `json:"MaritalStatus" validate:"required,marital"`
	ProfessionID      string  `json:"ProfessionID" validate:"required,profession"`
	Spouse            *Spouse `json:"Spouse" validate:"omitempty"`
	MobilePhone       string  `json:"MobilePhone" validate:"required,number"`
}

type Spouse struct {
	IDNumber          string `json:"Spouse_IDNumber"  validate:"required,number,len=16"`
	LegalName         string `json:"Spouse_LegalName"  validate:"required,allowcharsname"`
	BirthPlace        string `json:"Spouse_BirthPlace"  validate:"required,allowcharsname"`
	BirthDate         string `json:"Spouse_BirthDate"  validate:"required,dateformat"`
	SurgateMotherName string `json:"Spouse_SurgateMotherName"  validate:"required,allowcharsname"`
	Gender            string `json:"Spouse_Gender"  validate:"required,gender"`
}

type Pefindo struct {
	ClientKey               string `json:"ClientKey" `
	IDMember                string `json:"IDMember" `
	User                    string `json:"user" `
	ProspectID              string `json:"ProspectID" validate:"required,max=20"`
	BranchID                string `json:"BranchID" validate:"required"`
	BPKBName                string `json:"BPKBName" validate:"required,bpkbname"`
	IDNumber                string `json:"IDNumber" validate:"required,number,len=16"`
	LegalName               string `json:"LegalName" validate:"required,allowcharsname"`
	BirthDate               string `json:"BirthDate" validate:"required,dateformat"`
	SurgateMotherName       string `json:"SurgateMotherName" validate:"required,allowcharsname"`
	Gender                  string `json:"Gender" validate:"required,gender"`
	MaritalStatus           string `json:"MaritalStatus" validate:"required,marital"`
	SpouseIDNumber          string `json:"Spouse_IDNumber"  validate:"required,number,len=16"`
	SpouseLegalName         string `json:"Spouse_LegalName"  validate:"required,allowcharsname"`
	SpouseBirthPlace        string `json:"Spouse_BirthPlace"  validate:"required,allowcharsname"`
	SpouseBirthDate         string `json:"Spouse_BirthDate"  validate:"required,dateformat"`
	SpouseSurgateMotherName string `json:"Spouse_SurgateMotherName"  validate:"required,allowcharsname"`
	SpouseGender            string `json:"Spouse_Gender"  validate:"required,gender"`
}

type GenderCompare struct {
	Gender bool `json:"Spouse_Gender" validate:"spouse_gender"`
}

type BodyRequestElaborate struct {
	ClientKey string        `json:"client_key" validate:"required,key"`
	Data      DataElaborate `json:"data" validate:"required"`
}

type DataElaborate struct {
	ProspectID        string      `json:"ProspectID" validate:"required,max=20"`
	BranchID          string      `json:"BranchID" validate:"required"`
	BPKBName          string      `json:"BPKBName" validate:"required,bpkbname"`
	CustomerStatus    string      `json:"CustomerStatus" validate:"required,customer_status" ex:"NEW or RO/AO"`
	CategoryCustomer  string      `json:"CategoryCustomer" validate:"customer_category" ex:"REGULAR, PRIME or PRIORITY"`
	ResultPefindo     string      `json:"ResultPefindo" validate:"required,result_pefindo" ex:"PASS or REJECT"`
	TotalBakiDebet    interface{} `json:"TotalBakiDebet" validate:"required_baki_debet"`
	Tenor             int         `json:"Tenor" validate:"required"`
	ManufacturingYear string      `json:"ManufacturingYear" validate:"required,len=4,number"`
	OTR               float64     `json:"OTR" validate:"required"`
	NTF               float64     `json:"NTF" validate:"required"`
}

type DupcheckApi struct {
	ProspectID            string              `json:"prospect_id" validate:"required"`
	BranchID              string              `json:"branch_id" validate:"required"`
	ImageSelfie           string              `json:"image_selfie" validate:"required"`
	ImageKtp              string              `json:"ktp_url" validate:"required"`
	MonthlyFixedIncome    float64             `json:"monthly_fixed_income" validate:"required"`
	HomeStatus            string              `json:"home_status" validate:"required,max=2"`
	MonthlyVariableIncome float64             `json:"monthly_variable_income"`
	SpouseIncome          float64             `json:"spouse_income"`
	JobType               string              `json:"job_type" validate:"required"`
	JobPosition           string              `json:"job_position" validate:"required"`
	ProfessionID          string              `json:"profession_id" validate:"required"`
	EmploymentSinceYear   string              `json:"employment_since_year" validate:"required,len=4"`
	EmploymentSinceMonth  string              `json:"employment_since_month" validate:"required,len=2"`
	StaySinceYear         string              `json:"stay_since_year" validate:"required,len=4"`
	StaySinceMonth        string              `json:"stay_since_month" validate:"required,len=2"`
	BirthDate             string              `json:"birth_date" validate:"required,dateformat"`
	BirthPlace            string              `json:"birth_place" validate:"required,allowcharsname"`
	Tenor                 int                 `json:"tenor" validate:"required"`
	IDNumber              string              `json:"id_number" validate:"required,len=16,number"`
	LegalName             string              `json:"legal_name" validate:"required,allowcharsname"`
	MotherName            string              `json:"surgate_mother_name" validate:"required,allowcharsname"`
	Spouse                *DupcheckApiSpouse  `json:"spouse" validate:"omitempty"`
	MobilePhone           string              `json:"mobile_phone" validate:"min=9,max=14" example:"085689XXX01"`
	EngineNo              string              `json:"no_engine" validate:"required"`
	RangkaNo              string              `json:"no_rangka" validate:"required"`
	ManufactureYear       string              `json:"manufacture_year" validate:"required,len=4,number"`
	BPKBName              string              `json:"bpkb_name" validate:"required,bpkbname"`
	NumOfDependence       int                 `json:"num_of_dependence" validate:"required"`
	DPAmount              float64             `json:"down_payment_amount" validate:"required,max=999999999999" example:"22000000"`
	OTRPrice              float64             `json:"otr" validate:"required"`
	NTF                   float64             `json:"ntf" validate:"required"`
	LegalZipCode          string              `json:"legal_zip_code" validate:"required"`
	CompanyZipCode        string              `json:"company_zip_code" validate:"required"`
	Gender                string              `json:"gender" validate:"required"`
	InstallmentAmount     float64             `json:"installment_amount" validate:"required"`
	MaritalStatus         string              `json:"marital_status"`
	CustomerSegment       string              `json:"customer_segment"`
	Dealer                string              `json:"dealer"`
	AdminFee              *float64            `json:"admin_fee" validate:"required,max=999999999999" example:"1500000"`
	Cluster               string              `json:"-"`
	CMOCluster            string              `json:"-"`
	AF                    float64             `json:"-"`
	Filtering             entity.FilteringKMB `json:"-"`
}

type SpouseDupcheck struct {
	IDNumber   string `json:"spouse_id_number" validate:"required,len=16,number"`
	LegalName  string `json:"spouse_legal_name" validate:"required,allowcharsname"`
	BirthDate  string `json:"spouse_birth_date" validate:"required,dateformat"`
	MotherName string `json:"spouse_surgate_mother_name" validate:"required,allowcharsname"`
}

type NegativeCustomer struct {
	ProspectID        string `json:"prospect_id"`
	IDNumber          string `json:"id_number"`
	LegalName         string `json:"legal_name"`
	BirthDate         string `json:"birth_date"`
	SurgateMotherName string `json:"surgate_mother_name"`
	ProfessionID      string `json:"profession_id"`
	JobType           string `json:"job_type_id"`
	JobPosition       string `json:"job_position_id"`
}

type DupcheckApiSpouse struct {
	IDNumber   string `json:"spouse_id_number" validate:"required,len=16"`
	LegalName  string `json:"spouse_legal_name" validate:"required,allowcharsname"`
	BirthDate  string `json:"spouse_birth_date" validate:"required,dateformat"`
	Gender     string `json:"spouse_gender" validate:"required"`
	MotherName string `json:"spouse_surgate_mother_name" validate:"required,allowcharsname"`
}

type CustomerData struct {
	TransactionID   string `json:"transaction_id" validate:"required"`
	StatusKonsumen  string `json:"status_konsumen"`
	CustomerSegment string `json:"customer_segment"`
	IDNumber        string `json:"id_number"`
	LegalName       string `json:"legal_name"`
	BirthDate       string `json:"birth_date"`
	MotherName      string `json:"surgate_mother_name"`
	CustomerID      string `json:"customer_id"`
}

type ReqCustomerDomain struct {
	IDNumber   string `json:"id_number" validate:"required,len=16,number" example:"1234XXXXXXXX0001"`
	LegalName  string `json:"legal_name" validate:"required,allowcharsname,max=50" example:"LEGAL NAME"`
	BirthDate  string `json:"birth_date" validate:"required,dateformat" example:"YYYY-MM-DD"`
	MotherName string `json:"surgate_mother_name" validate:"required,allowcharsname,max=50" example:"SURGATE MOTHER NAME"`
	LobID      int    `json:"lob_id,omitempty"  example:"3"`
}

type ReqLatestPaidInstallment struct {
	IDNumber   string `json:"id_number"`
	CustomerID string `json:"customer_id"`
}

type FaceCompareRequest struct {
	ProspectID  string  `json:"prospect_id" validate:"required"`
	ImageSelfie string  `json:"image_selfie" validate:"required"`
	ImageKtp    string  `json:"ktp_url" validate:"required"`
	Lob         string  `json:"lob"  validate:"required"`
	IDNumber    string  `json:"id_number" validate:"required,len=16,number"`
	BirthDate   string  `json:"birth_date" validate:"required,dateformat"`
	BirthPlace  string  `json:"birth_place" validate:"required"`
	LegalName   string  `json:"legal_name" validate:"min=2,allowcharsname" example:"JONATHAN"`
	FaceType    *string `json:"type" validate:"omitempty,oneof=null PIN DEVICE"`
}

type ReqRejectTenor struct {
	ProspectID string `json:"prospect_id" validate:"required"`
	IDNumber   string `json:"id_number" validate:"required,len=16,number" example:"1234XXXXXXXX0001"`
}

type GetImagePlatform struct {
	CustomerID int    `json:"customer_id" validate:"required"`
	ImageURL   string `json:"image_url" validate:"required"`
}

type CustomerDomain struct {
	IDNumber   string `json:"id_number" validate:"required,len=16,number" example:"1234XXXXXXXX0001"`
	LegalName  string `json:"legal_name" validate:"required,allowcharsname,max=50" example:"LEGAL NAME"`
	BirthDate  string `json:"birth_date" validate:"required,dateformat" example:"YYYY-MM-DD"`
	MotherName string `json:"surgate_mother_name" validate:"required,allowcharsname,max=50" example:"SURGATE MOTHER NAME"`
	LobID      int    `json:"lob_id,omitempty"  example:"3"`
}

type LatestPaidInstallment struct {
	IDNumber   string `json:"id_number"`
	CustomerID string `json:"customer_id"`
}

type FacePlus struct {
	ProspectID string `json:"prospect_id" validate:"required"`
	Image1     string `json:"ktp_image" validate:"required,url"`
	Image2     string `json:"selfie_image" validate:"required,url"`
}

type FacePlusEncode struct {
	Encode interface{} `json:"encode"`
}

type PMK struct {
	MonthlyFixedIncome    float64 `json:"monthly_fixed_income" validate:"required"`
	MonthlyVariableIncome float64 `json:"monthly_variable_income"`
	HomeStatus            string  `json:"home_status" validate:"required,max=2"`
	SpouseIncome          float64 `json:"spouse_income"`
	JobPosition           string  `json:"job_position" validate:"required"`
	EmploymentSinceYear   string  `json:"employment_since_year" validate:"required,len=4"`
	EmploymentSinceMonth  string  `json:"employment_since_month" validate:"required,len=2"`
	StaySinceYear         string  `json:"stay_since_year" validate:"required,len=4"`
	StaySinceMonth        string  `json:"stay_since_month" validate:"required,len=2"`
	BirthDate             string  `json:"birth_date" validate:"required,dateformat"`
	Tenor                 int     `json:"tenor" validate:"required"`
	MaritalStatus         string  `json:"marital_status" validate:"required"`
}

type Dupcheck struct {
	ProspectID      string          `json:"transaction_id" validate:"required"`
	IDNumber        string          `json:"id_number" validate:"required,len=16,number"`
	LegalName       string          `json:"legal_name" validate:"required,allowcharsname"`
	BirthDate       string          `json:"birth_date" validate:"required,dateformat"`
	MotherName      string          `json:"surgate_mother_name" validate:"required,allowcharsname"`
	Spouse          *SpouseDupcheck `json:"spouse" validate:"omitempty"`
	EngineNo        string          `json:"engine_number" validate:"required"`
	ManufactureYear string          `json:"manufacture_year" validate:"required,len=4,number"`
}

type IntegratorPefindo struct {
	BiroName                string      `json:"biro_credit_name"`
	AppID                   string      `json:"app_id,omitempty"`
	RequestID               string      `json:"request_id,omitempty"`
	TransactionID           string      `json:"transaction_id,omitempty"`
	ProspectID              string      `json:"prospect_id,omitempty"`
	BranchID                string      `json:"branch_id"`
	IDNumber                string      `json:"id_number"`
	LegalName               string      `json:"legal_name"`
	BirthDate               string      `json:"birth_date"`
	Gender                  string      `json:"gender"`
	SurgateMotherName       string      `json:"surgate_mother_name"`
	Spouse                  interface{} `json:"spouse"`
	StatusKonsumen          string      `json:"status_konsumen"`
	OSInstallmentDue        float64     `json:"os_installmentdue"`
	NumberOfPaidInstallment int         `json:"number_of_paid_installment"`
}

type IntegratorPefindoAlco struct {
	BiroName          string      `json:"biro_credit_name"`
	AppID             string      `json:"app_id,omitempty"`
	RequestID         string      `json:"request_id,omitempty"`
	TransactionID     string      `json:"transaction_id,omitempty"`
	ProspectID        string      `json:"prospect_id,omitempty"`
	BranchID          string      `json:"branch_id"`
	IDNumber          string      `json:"id_number"`
	LegalName         string      `json:"legal_name"`
	BirthDate         string      `json:"birth_date"`
	Gender            string      `json:"gender"`
	SurgateMotherName string      `json:"surgate_mother_name"`
	Spouse            interface{} `json:"spouse"`
	StatusKonsumen    string      `json:"status_konsumen"`
	CustomerID        interface{} `json:"customer_id"`
	OTR               float64     `json:"otr"`
	NTF               float64     `json:"ntf"`
	ResidenceZipCode  string      `json:"residence_zip_code"`
	CompanyZipCode    string      `json:"company_zip_code"`
	StaySinceYear     string      `json:"stay_since_year"`
	StaySinceMonth    string      `json:"stay_since_month"`
	ProfessionID      string      `json:"profession_id"`
	BPKBName          string      `json:"bpkb_name"`
	Tenor             int         `json:"tenor"`
	Education         string      `json:"education"`
	MobilePhone       string      `json:"mobile_phone"`
	Brand             string      `json:"brand"`
	ExpiredDate       string      `json:"expired_date"`
	HomeStatus        string      `json:"home_status"`
	CategoryID        string      `json:"category_id"`
	DPAmount          float64     `json:"down_payment_amount"`
	SupplierID        string      `json:"supplier_id"`
}

type Dsr struct {
	ProspectID                     string     `json:"prospect_id" validate:"required"`
	IDNumber                       string     `json:"id_number"  validate:"required"`
	StatusKonsumen                 string     `json:"status_confins_customer" validate:"required,status_konsumen"`
	EngineNo                       string     `json:"engine_number" validate:"required"`
	InstallmentAmountConfins       float64    `json:"installment_amount_confins"`
	InstallmentAmountConfinsSpouse float64    `json:"installment_amount_confins_spouse"`
	InstallmentAmount              float64    `json:"installment_amount" validate:"required"`
	MonthlyFixedIncome             float64    `json:"monthly_fixed_income" validate:"required"`
	MonthlyVariableIncome          float64    `json:"monthly_variable_income"`
	SpouseIncome                   float64    `json:"spouse_income"`
	LegalName                      string     `json:"legal_name"`
	MotherName                     string     `json:"mother_name"`
	BirthDate                      string     `json:"birth_date"`
	Spouse                         *DsrSpouse `json:"spouse"`
}

type DsrSpouse struct {
	IDNumber   string `json:"id_number"`
	LegalName  string `json:"legal_name"`
	MotherName string `json:"mother_name"`
	BirthDate  string `json:"birth_date"`
}

type SyncGoLive struct {
	ProspectID           string  `json:"prospect_id"`
	AgreementNo          string  `json:"agreement_no"`
	ApplicationID        string  `json:"application_id"`
	FinalDisbursedAmount float64 `json:"final_disbursed_amount"`
	GoLiveStatus         string  `json:"go_live_status"`
	GoLiveDate           string  `json:"go_live_date"`
	PayToAgentAmount     float64 `json:"pay_to_agent_amount"`
	PayToAgentDate       string  `json:"pay_to_agent_date"`
}

type AfterPrescreening struct {
	ProspectID string `json:"prospect_id" validate:"required,max=20" example:"SAL042600001"`
}

type MetricsEkyc struct {
	CustomerStatus  string
	CustomerSegment string
	CBFound         bool
}

type Metrics struct {
	Transaction        Transaction        `json:"transaction" validate:"required"`
	Apk                Apk                `json:"apk" validate:"required"`
	CustomerPersonal   CustomerPersonal   `json:"customer_personal" validate:"required"`
	CustomerEmployment CustomerEmployment `json:"customer_employment" validate:"required"`
	Address            []Address          `json:"address" validate:"len=7,dive"`
	CustomerPhoto      []CustomerPhoto    `json:"customer_photo" validate:"dive"`
	CustomerEmcon      CustomerEmcon      `json:"customer_emcon" validate:"required"`
	CustomerSpouse     *CustomerSpouse    `json:"customer_spouse"`
	CustomerOmset      *[]CustomerOmset   `json:"customer_omset"`
	Item               Item               `json:"item" validate:"required"`
	Agent              Agent              `json:"agent" validate:"required"`
	Surveyor           []Surveyor         `json:"surveyor" validate:"required"`
}

type MetricsNE struct {
	Transaction        TransactionNE      `json:"transaction" validate:"required"`
	Apk                ApkNE              `json:"apk" validate:"required"`
	CustomerPersonal   CustomerPersonalNE `json:"customer_personal" validate:"required"`
	CustomerEmployment CustomerEmployment `json:"customer_employment" validate:"required"`
	Address            []Address          `json:"address" validate:"dive"`
	CustomerPhoto      []CustomerPhoto    `json:"customer_photo" validate:"dive"`
	CustomerEmcon      CustomerEmconNE    `json:"customer_emcon" validate:"required"`
	CustomerSpouse     *CustomerSpouseNE  `json:"customer_spouse"`
	CustomerOmset      *[]CustomerOmsetNE `json:"customer_omset"`
	Item               ItemNE             `json:"item" validate:"required"`
	Agent              AgentNE            `json:"agent" validate:"required"`
	CreatedBy          CreatedBy          `json:"created_by" validate:"required"`
}

type CreatedBy struct {
	CreatedByID   string `json:"created_by_id" validate:"required"`
	CreatedByName string `json:"created_by_name" validate:"required"`
}

type ValidateOmset struct {
	CustomerOmset []CustomerOmset `json:"customer_omset" validate:"required,len=3"`
}

type CustomerOmset struct {
	MonthlyOmsetYear  string  `json:"monthly_omset_year" validate:"len=4" example:"2021"`
	MonthlyOmsetMonth string  `json:"monthly_omset_month" validate:"len=2" example:"01"`
	MonthlyOmset      float64 `json:"monthly_omset" validate:"max=999999999999999" example:"5000000"`
}

type CustomerOmsetNE struct {
	MonthlyOmsetYear  string  `json:"monthly_omset_year" validate:"len=4" example:"2021"`
	MonthlyOmsetMonth string  `json:"monthly_omset_month" validate:"len=2" example:"01"`
	MonthlyOmset      float64 `json:"monthly_omset" validate:"max=999999999999999" example:"5000000"`
}

type Surveyor struct {
	Destination  string `json:"destination" validate:"required" example:"HOME"`
	RequestDate  string `json:"request_date" validate:"required,dateformat" example:"2021-07-29"`
	AssignDate   string `json:"assign_date" validate:"required,dateformat" example:"2021-07-29"`
	SurveyorName string `json:"surveyor_name" validate:"required" example:"TOTO SURYA"`
	ResultDate   string `json:"result_date" validate:"required,dateformat" example:"2021-07-29"`
	Status       string `json:"surveyor_status" validate:"required" example:"APPROVE"`
}

type Transaction struct {
	ProspectID        string `json:"prospect_id" validate:"prospect_id,noHTML" example:"SAL042600001"`
	BranchID          string `json:"branch_id" validate:"branch_id,noHTML" example:"426"`
	ApplicationSource string `json:"application_source" validate:"required,max=10,noHTML" example:"H"`
	Channel           string `json:"channel" validate:"channel,noHTML" example:"OFF"`
	Lob               string `json:"lob" validate:"lob,noHTML" example:"KMB"`
	OrderAt           string `json:"order_at" validate:"required,noHTML" example:"2021-07-15T11:44:05+07:00"`
	IncomingSource    string `json:"incoming_source" validate:"incoming,noHTML" example:"SLY"`
}

type TransactionNE struct {
	ProspectID        string `json:"prospect_id" validate:"prospect_id,noHTML" example:"SAL042600001"`
	BranchID          string `json:"branch_id" validate:"branch_id,noHTML" example:"426"`
	BranchName        string `json:"branch_name" validate:"required,noHTML" example:"BANDUNG"`
	ApplicationSource string `json:"application_source" validate:"required,max=10,noHTML" example:"H"`
	Channel           string `json:"channel" validate:"channel,noHTML" example:"OFF"`
	Lob               string `json:"lob" validate:"lob,noHTML" example:"KMB"`
	OrderAt           string `json:"order_at" validate:"required,noHTML" example:"2021-07-15T11:44:05+07:00"`
	IncomingSource    string `json:"incoming_source" validate:"incoming,noHTML" example:"SLY"`
}

type CustomerPersonal struct {
	IDType                     string   `json:"id_type" validate:"ktp,noHTML" example:"KTP"`
	IDNumber                   string   `json:"id_number" validate:"required,id_number,noHTML" example:"ENCRYPTED"`
	FullName                   string   `json:"full_name" validate:"required,allow_name,noHTML" example:"ENCRYPTED"`
	LegalName                  string   `json:"legal_name" validate:"required,allow_name,noHTML" example:"ENCRYPTED"`
	BirthPlace                 string   `json:"birth_place" validate:"min=3,max=100,noHTML" example:"JAKARTA"`
	BirthDate                  string   `json:"birth_date" validate:"dateformat" example:"1991-01-12"`
	SurgateMotherName          string   `json:"surgate_mother_name" validate:"required,allow_name,noHTML" example:"ENCRYPTED"`
	Gender                     string   `json:"gender" validate:"gender" example:"M"`
	MobilePhone                string   `json:"mobile_phone" validate:"min=9,max=14" example:"085689XXX01"`
	Email                      string   `json:"email" validate:"email,max=100,noHTML" example:"jonathaxx@gmail.com"`
	StaySinceYear              string   `json:"stay_since_year" validate:"len=4,noHTML" example:"2018"`
	StaySinceMonth             string   `json:"stay_since_month" validate:"len=2,noHTML" example:"03"`
	HomeStatus                 string   `json:"home_status" validate:"home,noHTML" example:"KL"`
	NPWP                       *string  `json:"npwp" validate:"omitempty,npwp,noHTML" example:"994646808XXX895"`
	Education                  string   `json:"education" validate:"education"  example:"S1"`
	MaritalStatus              string   `json:"marital_status" validate:"marital"  example:"M"`
	NumOfDependence            *int     `json:"num_of_dependence" validate:"required,max=50"  example:"1"`
	IDTypeIssueDate            *string  `json:"id_type_isssue_date" validate:"omitempty,dateformat"  example:"2021-07-29"`
	Religion                   string   `json:"religion" validate:"required,len=1"  example:"1"`
	PersonalCustomerType       string   `json:"personal_customer_type" validate:"required,max=20,noHTML"  example:"M"`
	ExpiredDate                *string  `json:"expired_date" validate:"omitempty,dateformat" example:"2021-07-29"`
	Nationality                string   `json:"nationality" validate:"required,max=3,noHTML"  example:"WNI"`
	WNACountry                 string   `json:"wna_country" validate:"required,max=50,noHTML"  example:"-"`
	HomeLocation               string   `json:"home_location" validate:"required,max=10,noHTML"  example:"N"`
	CustomerGroup              string   `json:"customer_group" validate:"required,max=1"  example:"2"`
	KKNo                       string   `json:"kk_no" validate:"required,len=16,number"  example:"97846094XXX34346"`
	BankID                     string   `gorm:"column:BankID" validate:"omitempty,max=5,noHTML" json:"bank_id" example:"BCA"`
	AccountNo                  string   `gorm:"column:AccountNo" validate:"omitempty,max=20,noHTML" json:"account_no" example:"567XX021"`
	AccountName                string   `gorm:"column:AccountName" validate:"omitempty,max=50,allowcharsname,noHTML" json:"account_name" example:"JONATHAN"`
	LivingCostAmount           *float64 `json:"living_cost_amount" validate:"required,max=999999999999,gte=0" example:"0"`
	Counterpart                *int     `json:"counterpart" validate:"omitempty,max=9999,gte=0" example:"169"`
	DebtBusinessScale          string   `json:"debt_business_scale" validate:"required,max=10,noHTML" example:"01"`
	DebtGroup                  string   `json:"debt_group" validate:"required,max=10,noHTML" example:"303"`
	IsAffiliateWithPP          string   `json:"is_affiliate_with_pp" validate:"required,max=1" example:"N"`
	AgreetoAcceptOtherOffering int      `json:"agree_to_accept_other_offering" validate:"required,max=1" example:"1"`
	DataType                   string   `json:"data_type" validate:"required,max=1" example:"G"`
	Status                     string   `json:"status" validate:"required,max=1" example:"F"`
	RentFinishDate             *string  `json:"rent_finish_date" validate:"omitempty,dateformat"  example:"2021-07-29"`
}

type CustomerPersonalNE struct {
	IDNumber          string  `json:"id_number" validate:"required,len=16,number" example:"1234567890123456"`
	FullName          string  `json:"full_name" validate:"required,min=2,allowcharsname,noHTML" example:"ENCRYPTED"`
	LegalName         string  `json:"legal_name" validate:"required,min=2,allowcharsname,noHTML" example:"ENCRYPTED"`
	BirthPlace        string  `json:"birth_place" validate:"min=3,max=100" example:"JAKARTA"`
	BirthDate         string  `json:"birth_date" validate:"dateformat" example:"1991-01-12"`
	SurgateMotherName string  `json:"surgate_mother_name" validate:"required,min=2,allowcharsname,noHTML" example:"ENCRYPTED"`
	Gender            string  `json:"gender" validate:"gender" example:"M"`
	MobilePhone       string  `json:"mobile_phone" validate:"min=9,max=14,noHTML" example:"085689XXX01"`
	Email             string  `json:"email" validate:"email,max=100,noHTML" example:"jonathaxx@gmail.com"`
	StaySinceYear     string  `json:"stay_since_year" validate:"len=4,noHTML" example:"2018"`
	StaySinceMonth    string  `json:"stay_since_month" validate:"len=2,noHTML" example:"03"`
	HomeStatus        string  `json:"home_status" validate:"home,noHTML" example:"KL"`
	NPWP              *string `json:"npwp" validate:"omitempty,npwp,noHTML" example:"994646808XXX895"`
	Education         string  `json:"education" validate:"education"  example:"S1"`
	MaritalStatus     string  `json:"marital_status" validate:"marital"  example:"M"`
	NumOfDependence   *int    `json:"num_of_dependence" validate:"required,max=50"  example:"1"`
}

type CustomerEmployment struct {
	ProfessionID          string   `json:"profession_id" validate:"profession,noHTML" example:"WRST"`
	EmploymentSinceYear   string   `json:"employment_since_year" validate:"len=4,noHTML" example:"2020"`
	EmploymentSinceMonth  string   `json:"employment_since_month" validate:"len=2,noHTML" example:"02"`
	MonthlyFixedIncome    float64  `json:"monthly_fixed_income" validate:"gt=0,max=999999999999" example:"5000000"`
	JobType               string   `json:"job_type" validate:"required,max=10,noHTML" example:"008"`
	JobPosition           string   `json:"job_position" validate:"required,max=10,noHTML" example:"S"`
	MonthlyVariableIncome *float64 `json:"monthly_variable_income" validate:"required,max=999999999999,gte=0" example:"3000000"`
	SpouseIncome          *float64 `json:"spouse_income" validate:"required,max=999999999999,gte=0" example:"6000000"`
	CompanyName           string   `json:"company_name" validate:"required,max=50,noHTML" example:"PT.KIMIA FARMA"`
	IndustryTypeID        string   `json:"industry_type_id" validate:"required,max=10,noHTML" example:"9990"`
	ExtCompanyPhone       *string  `json:"company_phone_ext" validate:"omitempty,max=4,noHTML" example:"442"`
	SourceOtherIncome     *string  `json:"source_other_income" validate:"omitempty,max=30,noHTML" example:"TOKO MAKMUR"`
}

type Address struct {
	Type      string `json:"type" validate:"address,noHTML"  example:"RESIDENCE"`
	Address   string `json:"address" validate:"required,noHTML"  example:"JL.PEGANGSAAN 1"`
	Rt        string `json:"rt" validate:"min=1,max=3,noHTML"  example:"008"`
	Rw        string `json:"rw" validate:"min=1,max=3,noHTML"  example:"017"`
	Kelurahan string `json:"kelurahan" validate:"required,noHTML"  example:"TEGAL PARANG"`
	Kecamatan string `json:"kecamatan" validate:"required,noHTML"  example:"MAMPANG PRAPATAN"`
	City      string `json:"city" validate:"required,noHTML"  example:"JAKARTA SELATAN"`
	ZipCode   string `json:"zip_code" validate:"required,noHTML"  example:"12790"`
	AreaPhone string `json:"area_phone" validate:"min=2,max=4,noHTML"  example:"021"`
	Phone     string `json:"phone" validate:"required,max=10,noHTML"  example:"84522"`
}

type CustomerPhoto struct {
	ID  string `json:"id" validate:"photo,noHTML" example:"KTP"`
	Url string `json:"url" validate:"url,max=250,noHTML" example:"https://dev-media.kreditplus.com/media/reference/20000/KPM-3677/ktp_EFM-3677.jpg"`
}

type CustomerEmcon struct {
	Name                   string `json:"name" validate:"min=2,allowcharsname,noHTML" example:"MULYADI"`
	Relationship           string `json:"relationship" validate:"relationship,noHTML" example:"FM"`
	MobilePhone            string `json:"mobile_phone" validate:"min=9,max=14,noHTML" example:"0856789XXX1"`
	AreaPhone              string `json:"emergency_area_phone_office" validate:"min=2,max=4,noHTML" example:"021"`
	Phone                  string `json:"emergency_phone_office" validate:"required,max=20,noHTML" example:"567892"`
	VerificationWith       string `json:"verification_with" validate:"required,max=100,noHTML" example:"JONO"`
	ApplicationEmconSesuai string `json:"application_emcon_sesuai" validate:"required,max=1,noHTML" example:"1"`
	VerifyBy               string `json:"verify_by" validate:"required,max=10,noHTML" example:"PHONE"`
	KnownCustomerAddress   string `json:"known_customer_address" validate:"required,max=1,noHTML" example:"1"`
	KnownCustomerJob       string `json:"known_customer_job" validate:"required,max=1,noHTML" example:"1"`
}

type CustomerEmconNE struct {
	Name         string `json:"name" validate:"min=2,allowcharsname,noHTML" example:"MULYADI"`
	Relationship string `json:"relationship" validate:"relationship,noHTML" example:"FM"`
	MobilePhone  string `json:"mobile_phone" validate:"min=9,max=14,noHTML" example:"0856789XXX1"`
}

type CustomerSpouse struct {
	IDNumber          string  `json:"id_number" validate:"required,id_number,noHTML" example:"177105550374XX01"`
	FullName          string  `json:"full_name" validate:"required,allow_name,noHTML" example:"SUSI BUNGA"`
	LegalName         string  `json:"legal_name" validate:"required,allow_name,noHTML" example:"SUSI BUNGA"`
	BirthDate         string  `json:"birth_date"  validate:"dateformat" example:"1991-01-29"`
	BirthPlace        string  `json:"birth_place" validate:"required,min=3,max=100,noHTML" example:"JAKARTA"`
	Gender            string  `json:"gender" validate:"gender" example:"F"`
	SurgateMotherName string  `json:"surgate_mother_name" validate:"required,allow_name,noHTML" example:"TUTI"`
	CompanyPhone      *string `json:"company_phone" validate:"omitempty,max=30,noHTML" example:"865542"`
	CompanyName       *string `json:"company_name" validate:"omitempty,max=50,noHTML" example:"PT.BUMI KARYA"`
	MobilePhone       string  `json:"mobile_phone" validate:"min=9,max=14,noHTML" example:"08772012XXX0"`
	ProfessionID      *string `json:"profession_id" validate:"omitempty,profession,noHTML" example:"KRYSW"`
}

type CustomerSpouseNE struct {
	IDNumber          string `json:"id_number" validate:"required,len=16,number" example:"177105550374XX01"`
	FullName          string `json:"full_name" validate:"required,min=2,allowcharsname" example:"SUSI BUNGA"`
	LegalName         string `json:"legal_name" validate:"required,min=2,allowcharsname" example:"SUSI BUNGA"`
	BirthDate         string `json:"birth_date"  validate:"dateformat" example:"1991-01-29"`
	Gender            string `json:"gender" validate:"gender" example:"F"`
	SurgateMotherName string `json:"surgate_mother_name" validate:"required,min=2,allowcharsname" example:"TUTI"`
}

type Apk struct {
	OtherFee                    float64  `json:"other_fee" validate:"min=0,max=999999999999,gte=0" example:"0"`
	Tenor                       int      `json:"tenor" validate:"required,max=60,gte=0" example:"36"`
	ProductOfferingID           string   `json:"product_offering_id" validate:"required,max=10,noHTML" example:"NLMKKAPSEP"`
	ProductOfferingDesc         string   `json:"product_offering_desc" validate:"omitempty,max=200,noHTML"`
	ProductID                   string   `json:"product_id" validate:"required,max=10,noHTML" example:"1SNLMK"`
	OTR                         float64  `json:"otr" validate:"required,max=999999999999,gte=0" example:"105000000"`
	DPAmount                    float64  `json:"down_payment_amount" validate:"required,max=999999999999,gte=0" example:"22000000"`
	NTF                         float64  `json:"ntf" validate:"required,max=999999999999,gte=0" example:"150528000"`
	AF                          float64  `json:"af" validate:"required,max=999999999999,gte=0" example:"84000000"`
	AoID                        string   `json:"aoid" validate:"required,max=20,noHTML" example:"81088"`
	AdminFee                    *float64 `json:"admin_fee" validate:"required,max=999999999999,gte=0" example:"1500000"`
	InstallmentAmount           float64  `json:"installment_amount" validate:"required,max=999999999999,gte=0" example:"4181333"`
	PercentDP                   *float64 `json:"down_payment_rate" validate:"required,max=99,gte=0" example:"20.95"`
	PremiumAmountToCustomer     float64  `json:"premium_amount_to_customer" validate:"min=0,max=999999999999,gte=0" example:"2184000"`
	FidusiaFee                  *float64 `json:"fidusia_fee" validate:"omitempty,min=0,max=999999999999,gte=0" example:"0"`
	InterestRate                *float64 `json:"interest_rate" validate:"required,max=99,gte=0" example:"2.2"`
	InterestAmount              *float64 `json:"interest_amount" validate:"required,max=999999999999,gte=0" example:"66528000"`
	InsuranceAmount             float64  `json:"insurance_amount" validate:"min=0,max=999999999999,gte=0" example:"3150000"`
	FirstInstallment            string   `json:"first_installment" validate:"required,max=2,noHTML" example:"AR"`
	PaymentMethod               string   `json:"payment_method" validate:"required,max=2,noHTML" example:"CR"`
	SurveyFee                   *float64 `json:"survey_fee" validate:"required,max=999999999999,gte=0" example:"0"`
	IsFidusiaCovered            string   `json:"is_fidusia_covered" validate:"required,len=1" example:"Y"`
	ProvisionFee                *float64 `json:"provision_fee" validate:"required,max=999999999999,gte=0" example:"2475000"`
	InsAssetPaidBy              string   `json:"ins_asset_paid_by" validate:"required,noHTML" example:"CU"`
	InsAssetPeriod              string   `json:"ins_asset_period" validate:"required,noHTML" example:"FT"`
	EffectiveRate               float64  `json:"effective_rate" validate:"required,max=99,gte=0" example:"26.4"`
	SalesmanID                  string   `json:"salesman_id" validate:"required,noHTML" example:"81088"`
	SupplierBankAccountID       string   `json:"supplier_bank_account_id" validate:"required,noHTML" example:"1"`
	LifeInsuranceCoyBranchID    string   `json:"life_insurance_coy_branch_id" validate:"max=20,noHTML" example:"426"`
	LifeInsuranceAmountCoverage float64  `json:"life_insurance_amount_coverage" validate:"min=0,max=999999999999,gte=0" example:"105000000"`
	CommisionSubsidy            float64  `json:"commision_subsidi" validate:"min=0,max=999999999999,gte=0" example:"0"`
	FinancePurpose              string   `json:"finance_purpose" validate:"required,max=100,noHTML"`
	Dealer                      string   `json:"dealer" validate:"omitempty,max=50,noHTML"`
	LoanAmount                  float64  `json:"loan_amount"  validate:"min=0,max=999999999999,gte=0" example:"105000000"`
	WayOfPayment                string   `json:"way_of_payment" validate:"required,max=2,noHTML" example:"CA"`
	StampDutyFee                float64  `json:"stamp_duty_fee" validate:"min=0,max=999999999999,gte=0" example:"250000"`
	AgentFee                    float64  `json:"agent_fee" validate:"min=0,max=999999999999,gte=0" example:"250000"`
}

type ApkNE struct {
	Tenor                   int      `json:"tenor" validate:"required,min=1,max=60" example:"36"`
	OTR                     float64  `json:"otr" validate:"required,gte=0,max=999999999999" example:"105000000"`
	DPAmount                float64  `json:"down_payment_amount" validate:"required,gte=0,max=999999999999" example:"22000000"`
	NTF                     float64  `json:"ntf" validate:"required,gte=0,max=999999999999" example:"150528000"`
	AF                      float64  `json:"af" validate:"required,gte=0,max=999999999999" example:"84000000"`
	AdminFee                *float64 `json:"admin_fee" validate:"required,gte=0,max=999999999999" example:"1500000"`
	InstallmentAmount       float64  `json:"installment_amount" validate:"required,gte=0,max=999999999999" example:"4181333"`
	PercentDP               *float64 `json:"down_payment_rate" validate:"required,gte=0,max=99" example:"20.95"`
	PremiumAmountToCustomer float64  `json:"premium_amount_to_customer" validate:"gte=0,max=999999999999" example:"2184000"`
	InsuranceAmount         float64  `json:"insurance_amount" validate:"gte=0,max=999999999999" example:"3150000"`
	ProvisionFee            *float64 `json:"provision_fee" validate:"required,gte=0,max=999999999999" example:"2475000"`
	FinancePurpose          string   `json:"finance_purpose" validate:"required,max=100,noHTML"`
	Dealer                  string   `json:"dealer" validate:"omitempty,max=50,noHTML"`
	LoanAmount              float64  `json:"loan_amount"  validate:"gte=0,max=999999999999" example:"105000000"`
}

type Item struct {
	SupplierID                   string  `json:"supplier_id" validate:"required,max=100,noHTML" example:"42600342"`
	AssetCode                    string  `json:"asset_code" validate:"required,max=200,noHTML" example:"SUZUKI,KMOBIL,GRAND VITARA.JLX 2,0 AT"`
	ManufactureYear              string  `json:"manufacture_year" validate:"len=4,number,noHTML" example:"2020"`
	NoChassis                    string  `json:"chassis_number" validate:"required,max=30,noHTML" example:"MHKV1AA2JBK107322"`
	NoEngine                     string  `json:"engine_number" validate:"required,max=30,noHTML" example:"73218JAJK"`
	Qty                          int     `json:"qty"  validate:"required,max=1" example:"1"`
	POS                          string  `json:"pos" validate:"required,max=10,noHTML" example:"426"`
	CC                           string  `json:"cc" validate:"required,max=10,noHTML" example:"1500"`
	Condition                    string  `json:"condition" validate:"required,max=10,noHTML" example:"U"`
	AssetUsage                   string  `json:"asset_usage" validate:"required,max=10,noHTML" example:"N"`
	Region                       string  `json:"region" validate:"required,max=10,noHTML" example:"0"`
	TaxDate                      string  `json:"tax_date" validate:"required,dateformat" example:"2022-03-02"`
	STNKExpiredDate              string  `json:"stnk_expired_date" validate:"required,dateformat" example:"2025-03-20"`
	CategoryID                   string  `json:"category_id" validate:"required,max=100,noHTML" example:"SEDAN"`
	AssetDescription             string  `json:"asset_description" validate:"required,max=200,noHTML" example:"SUZUKI.KMOBIL.GRAND VITARA.JLX 2,0 AT"`
	BPKBName                     string  `json:"bpkb_name" validate:"required,bpkbname" example:"K"`
	OwnerAsset                   string  `json:"owner_asset" validate:"required,max=50,noHTML" example:"JONATHAN"`
	LicensePlate                 string  `json:"license_plate" validate:"required,max=50,noHTML" example:"3006TBJ"`
	Color                        string  `json:"color" validate:"required,max=50,noHTML" example:"HITAM"`
	AssetInsuranceAmountCoverage float64 `json:"asset_insurance_amount_coverage" validate:"required,max=999999999999,gte=0" example:"105000000"`
	InsAssetInsuredBy            string  `json:"ins_asset_insured_by" validate:"required,max=10,noHTML" example:"CO"`
	InsuranceCoyBranchID         string  `json:"insurance_coy_branch_id" validate:"required,max=10,noHTML" example:"426"`
	CoverageType                 string  `json:"coverage_type" validate:"required,max=10,noHTML" example:"TLO"`
	OwnerKTP                     string  `json:"owner_ktp" validate:"required,len=16,number" example:"3172024508XXX002"`
	Brand                        string  `json:"brand" validate:"required,max=255,noHTML" example:"TOYOTA"`
	PremiumAmountToCustomer      float64 `json:"premium_amount_to_customer" validate:"min=0,max=999999999999,gte=0" example:"2184000"`
}

type ItemNE struct {
	AssetCode               string  `json:"asset_code" validate:"required,max=200,noHTML" example:"SUZUKI,KMOBIL,GRAND VITARA.JLX 2,0 AT"`
	ManufactureYear         string  `json:"manufacture_year" validate:"len=4,number" example:"2020"`
	LicensePlate            string  `json:"license_plate" validate:"required,max=15,noHTML" example:"DK 1234 ABC"`
	NoChassis               string  `json:"chassis_number" validate:"required,max=30,noHTML" example:"MHKV1AA2JBK107322"`
	NoEngine                string  `json:"engine_number" validate:"required,max=30,noHTML" example:"73218JAJK"`
	Condition               string  `json:"condition" validate:"required,max=10,noHTML" example:"U"`
	CategoryID              string  `json:"category_id" validate:"required,max=100,noHTML" example:"SEDAN"`
	AssetDescription        string  `json:"asset_description" validate:"required,max=200,noHTML" example:"SUZUKI.KMOBIL.GRAND VITARA.JLX 2,0 AT"`
	BPKBName                string  `json:"bpkb_name" validate:"required,bpkbname" example:"K"`
	OwnerAsset              string  `json:"owner_asset" validate:"required,max=50,noHTML" example:"JONATHAN"`
	Color                   string  `json:"color" validate:"required,max=50,noHTML" example:"HITAM"`
	Brand                   string  `json:"brand" validate:"required,max=255,noHTML" example:"TOYOTA"`
	PremiumAmountToCustomer float64 `json:"premium_amount_to_customer" validate:"min=0,max=999999999999" example:"2184000"`
}

type Agent struct {
	CmoRecom  string `json:"cmo_recom" validate:"recom" example:"1"`
	CmoName   string `json:"cmo_name" validate:"required,noHTML" example:"SETO MULYA"`
	CmoNik    string `json:"cmo_nik" validate:"required,noHTML" example:"93510"`
	RecomDate string `json:"recom_date" validate:"required,dateformat" example:"2021-07-15"`
}

type AgentNE struct {
	CmoRecom  string `json:"cmo_recom" validate:"recom" example:"1"`
	CmoName   string `json:"cmo_name" validate:"required,noHTML" example:"SETO MULYA"`
	CmoNik    string `json:"cmo_nik" validate:"required,noHTML" example:"93510"`
	RecomDate string `json:"recom_date" validate:"required,dateformat" example:"2021-07-15"`
}

type LockSystem struct {
	IDNumber      string `json:"id_number" validate:"required,id_number" example:"ENCRYPTED NIK"`
	ChassisNumber string `json:"chassis_number" validate:"max=50" example:"AWADW4221375G"`
	EngineNumber  string `json:"engine_number" validate:"max=50" example:"2AZE205717"`
}

type Recalculate struct {
	ProspectID                   string  `json:"prospect_id" validate:"prospect_id" example:"SAL042600001"`
	Tenor                        int     `json:"tenor" validate:"required,max=60" example:"36"`
	ProductOfferingID            string  `json:"product_offering_id" validate:"required,max=10" example:"NLMKKAPSEP"`
	ProductOfferingDesc          string  `json:"product_offering_desc" validate:"omitempty,max=200"`
	DPAmount                     float64 `json:"down_payment_amount" validate:"required,max=999999999999" example:"22000000"`
	NTF                          float64 `json:"ntf" validate:"required,max=999999999999" example:"150528000"`
	AF                           float64 `json:"af" validate:"required,max=999999999999" example:"84000000"`
	AdminFee                     float64 `json:"admin_fee" validate:"max=999999999999" example:"1500000"`
	InstallmentAmount            float64 `json:"installment_amount" validate:"required,max=999999999999" example:"4181333"`
	PercentDP                    float64 `json:"down_payment_rate" validate:"required,max=99" example:"20.95"`
	LifePremiumAmountToCustomer  float64 `json:"life_premium_amount_to_customer" validate:"min=0,max=999999999999" example:"2184000"`
	AssetPremiumAmountToCustomer float64 `json:"asset_premium_amount_to_customer" validate:"min=0,max=999999999999" example:"2184000"`
	FidusiaFee                   float64 `json:"fidusia_fee" validate:"max=999999999999" example:"0"`
	InterestRate                 float64 `json:"interest_rate" validate:"max=99" example:"2.2"`
	InterestAmount               float64 `json:"interest_amount" validate:"max=999999999999" example:"66528000"`
	ProvisionFee                 float64 `json:"provision_fee" validate:"max=999999999999" example:"2475000"`
	LoanAmount                   float64 `json:"loan_amount" validate:"max=999999999999" example:"105000000"`
}

type RequestGenerateFormAKKK struct {
	ProspectID string `json:"prospect_id" validate:"required" example:"TEST-DEV"`
	LOB        string `json:"lob" validate:"required" example:"new-kmb"`
	Source     string `json:"source" example:"SYSTEM"`
}

type Filtering struct {
	ProspectID    string           `json:"prospect_id" validate:"prospect_id" example:"SAL042600001"`
	BranchID      string           `json:"branch_id" validate:"required,branch_id" example:"426"`
	IDNumber      string           `json:"id_number" validate:"required,id_number" example:"ENCRYPTED NIK"`
	LegalName     string           `json:"legal_name" validate:"required,allow_name" example:"ENCRYPTED LEGAL NAME"`
	BirthDate     string           `json:"birth_date" validate:"required,dateformat" example:"YYYY-MM-DD"`
	Gender        string           `json:"gender" validate:"required,gender" example:"M"`
	MotherName    string           `json:"surgate_mother_name" validate:"required,allow_name" example:"ENCRYPTED SURGATE MOTHER NAME"`
	BPKBName      string           `json:"bpkb_name" validate:"required,bpkbname" example:"K"`
	CMOID         string           `json:"cmo_id" validate:"required,max=20,noHTML" example:"123456"`
	ChassisNumber *string          `json:"chassis_number" validate:"omitempty,max=50,noHTML" example:"AWADW4221375G"`
	EngineNumber  *string          `json:"engine_number" validate:"omitempty,max=50,noHTML" example:"2AZE205717"`
	Spouse        *FilteringSpouse `json:"spouse" validate:"omitempty"`
}

type FilteringSpouse struct {
	IDNumber   string `json:"spouse_id_number" validate:"required,id_number"  example:"ENCRYPTED NIK"`
	LegalName  string `json:"spouse_legal_name" validate:"required,allow_name" example:"ENCRYPTED LEGAL NAME"`
	BirthDate  string `json:"spouse_birth_date" validate:"required,dateformat" example:"YYYY-MM-DD"`
	Gender     string `json:"spouse_gender" validate:"required,gender" example:"F"`
	MotherName string `json:"spouse_surgate_mother_name" validate:"required,allow_name" example:"ENCRYPTED SURGATE MOTHER NAME"`
}

type ElaborateLTV struct {
	ProspectID        string `json:"prospect_id" validate:"prospect_id"`
	Tenor             int    `json:"tenor" validate:"required,max=60"`
	ManufacturingYear string `json:"manufacturing_year" validate:"required,len=4,number"`
}

type SallyFilteringCallback struct {
	ProspectID      string      `json:"prospect_id"`
	Decision        string      `json:"decision"`
	Reason          string      `json:"reason"`
	PbkReport       interface{} `json:"pbk_report"`
	PbkReportSpouse interface{} `json:"pbk_report_spouse"`
	MaxNTF          interface{} `json:"max_ntf"`
}

type ScoreProIntegrator struct {
	ProspectID       string      `json:"prospect_id"`
	CBFound          bool        `json:"cb_found"`
	StatusKonsumen   string      `json:"status_konsumen"`
	RequestorID      string      `json:"requestor_id"`
	Journey          string      `json:"journey"`
	PhoneNumber      string      `json:"phone_number"`
	ScoreGeneratorID string      `json:"score_generator_id"`
	Data             interface{} `json:"data"`
}

type OrderIDCheck struct {
	ProspectID string `json:"prospect_id" validate:"prospectID"`
}

type MarriedValidator struct {
	CustomerSpouse bool `json:"customer_spouse" validate:"notnull" default:"true"`
}

type SingleValidator struct {
	CustomerSpouse bool `json:"customer_spouse" validate:"mustnull" default:"true"`
}

type GetNTFAllLob struct {
	ProspectID string  `json:"prospect_id"`
	NTF        float64 `json:"ntf"`
	IDNumber   string  `json:"id_number"`
	BirthDate  string  `json:"birth_date"`
	LegalName  string  `json:"legal_name"`
	MotherName string  `json:"mother_name"`
}

type DataIDX struct {
	CBFoundCustomer   bool        `json:"cb_found_customer"`
	CBFoundSpouse     bool        `json:"cb_found_spouse"`
	PefindoIDCustomer interface{} `json:"pefindo_id_customer"`
	PefindoIDSpouse   interface{} `json:"pefindo_id_spouse"`
}

type PefindoIDX struct {
	ProspectID        string      `json:"prospect_id"`
	ModelType         string      `json:"model_type"`
	CBFoundCustomer   bool        `json:"cb_found_customer"`
	PefindoIDCustomer interface{} `json:"pefindo_id_customer"`
	CBFoundSpouse     bool        `json:"cb_found_spouse"`
	PefindoIDSpouse   interface{} `json:"pefindo_id_spouse"`
}

type FaceRecognitionRequest struct {
	TransactionID string `json:"transaction_id" validate:"required"`
	Threshold     string `json:"threshold" validate:"required"`
	IDNumber      string `json:"id_number" validate:"required,len=16,number"`
	SelfieImage   string `json:"selfie_image" validate:"required"`
}

type VerifyDataRequest struct {
	TransactionID     string `json:"transaction_id" validate:"required"`
	Threshold         string `json:"threshold" validate:"required"`
	IDNumber          string `json:"id_number" validate:"required,len=16"`
	LegalName         string `json:"legal_name" validate:"required"`
	Address           string `json:"address"`
	BirthPlace        string `json:"birth_place"`
	BirthDate         string `json:"birth_date" validate:"dateformat,omitempty" example:"1997-03-01"`
	SurgateMotherName string `json:"surgate_mother_name" validate:"required"`
	RT                string `json:"rt"`
	RW                string `json:"rw"`
	City              string `json:"city"`
	Province          string `json:"province"`
	Kabupaten         string `json:"kabupaten"`
	Kecamatan         string `json:"kecamatan"`
	Kelurahan         string `json:"kelurahan"`
	Gender            string `json:"gender" validate:"omitempty,oneof=F M" example:"F/M"`
	Profession_id     string `json:"profession_id" validate:"omitempty,oneof=KRYSW PNS ANG WRST" example:"KRYSW/PNS/ANG/WRST"`
	ReqID             string
}

type NewDupcheck struct {
	ProspectID string `json:"prospect_id" validate:"required,max=20" example:"TEST-DEV"`
	IDNumber   string `json:"id_number" validate:"required,len=16,number" example:"1234XXXXXXXX0001"`
	LegalName  string `json:"legal_name" validate:"required,allowcharsname,max=50" example:"LEGAL NAME"`
	BirthDate  string `json:"birth_date" validate:"required,dateformat" example:"YYYY-MM-DD"`
	MotherName string `json:"surgate_mother_name" validate:"required,allowcharsname,max=50" example:"SURGATE MOTHER NAME"`
}

type UpdateReason struct {
	ProspectID         string
	CustomerStatus     string
	Reason             string
	MaxOverdueDaysROAO int
}

type RequestPagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type ReqInquiryPrescreening struct {
	Search      string `json:"search"`
	UserID      string `json:"user_id" validate:"required,max=20"`
	BranchID    string `json:"branch_id" validate:"required,max=3"`
	MultiBranch string `json:"multi_branch" validate:"required,max=1"`
}

type ReqReasonPrescreening struct {
	ReasonID string `json:"reason_id"`
}

type ReqApprovalReason struct {
	Type string `json:"type" validate:"required"`
}

type ReqReviewPrescreening struct {
	ProspectID     string `json:"prospect_id" validate:"required,max=20" example:"TEST-DEV"`
	Decision       string `json:"decision" validate:"required,decision,max=7" example:"APPROVE,REJECT"`
	Reason         string `json:"reason" validate:"max=255"`
	DecisionBy     string `json:"decision_by" validate:"required,max=100"`
	DecisionByName string `json:"decision_by_name" validate:"required,max=250"`
}

type ReqInquiryCa struct {
	Search      string `json:"search"`
	BranchID    string `json:"branch_id" validate:"required,max=3"`
	MultiBranch string `json:"multi_branch" validate:"required,max=1"`
	Filter      string `json:"filter" validate:"max=15"`
	UserID      string `json:"user_id" validate:"required,max=20"`
}

type ReqInquiryNE struct {
	Search      string `json:"search"`
	BranchID    string `json:"branch_id" validate:"required,max=3"`
	MultiBranch string `json:"multi_branch" validate:"required,max=1"`
	Filter      string `json:"filter" validate:"max=15"`
	UserID      string `json:"user_id" validate:"required,max=20"`
}

type ReqSaveAsDraft struct {
	ProspectID string      `json:"prospect_id" validate:"required,max=20" example:"TEST-DEV"`
	Decision   string      `json:"decision" validate:"required,decision,max=7" example:"APPROVE,REJECT"`
	SlikResult string      `json:"slik_result" validate:"required,max=30"`
	Note       string      `json:"note" validate:"max=525"`
	CreatedBy  string      `json:"decision_by" validate:"required,max=100"`
	DecisionBy string      `json:"decision_by_name" validate:"required,max=250"`
	Edd        interface{} `json:"edd"`
}

type EDD struct {
	Pernyataan1 interface{} `json:"pernyataan_1"`
	Pernyataan2 interface{} `json:"pernyataan_2"`
	Pernyataan3 interface{} `json:"pernyataan_3"`
	Pernyataan4 interface{} `json:"pernyataan_4"`
	Pernyataan5 interface{} `json:"pernyataan_5"`
	Pernyataan6 interface{} `json:"pernyataan_6"`
}

type ReqSubmitDecision struct {
	ProspectID   string      `json:"prospect_id" validate:"required,max=20" example:"TEST-DEV"`
	NTFAkumulasi float64     `json:"ntf_akumulasi" validate:"required,max=9999999999999"`
	Decision     string      `json:"decision" validate:"required,decision,max=7" example:"APPROVE,REJECT"`
	SlikResult   string      `json:"slik_result" validate:"required,max=30"`
	Note         string      `json:"note" validate:"max=525"`
	CreatedBy    string      `json:"decision_by" validate:"required,max=100"`
	DecisionBy   string      `json:"decision_by_name" validate:"required,max=250"`
	Edd          interface{} `json:"edd"`
}

type ReqSubmitApproval struct {
	ProspectID     string  `json:"prospect_id" validate:"required,max=20" example:"TEST-DEV"`
	FinalApproval  string  `json:"final_approval" validate:"required,max=3"`
	Decision       string  `json:"decision" validate:"required,max=7" example:"APPROVE,REJECT,RETURN"`
	RuleCode       string  `json:"code" validate:"required,max=4"`
	Alias          string  `json:"alias" validate:"required,max=3"`
	Reason         string  `json:"reason" validate:"required,max=100"`
	NeedEscalation bool    `json:"need_escalation"`
	DPAmount       float64 `json:"dp_amount"`
	Note           string  `json:"note" validate:"max=525"`
	CreatedBy      string  `json:"decision_by" validate:"required,max=100"`
	DecisionBy     string  `json:"decision_by_name" validate:"required,max=250"`
}

type ReqSearchInquiry struct {
	UserID      string `json:"user_id" validate:"required,max=20"`
	BranchID    string `json:"branch_id" validate:"required,max=3"`
	MultiBranch string `json:"multi_branch" validate:"required,max=1"`
	Search      string `json:"search" validate:"required"`
}

type ReqCancelOrder struct {
	ProspectID   string `json:"prospect_id" validate:"required,max=20" example:"TEST-DEV"`
	CreatedBy    string `json:"decision_by" validate:"required,max=100"`
	DecisionBy   string `json:"decision_by_name" validate:"required,max=250"`
	CancelReason string `json:"reason" validate:"required,max=100"`
}

type ReqReturnOrder struct {
	ProspectID string `json:"prospect_id" validate:"required,max=20" example:"TEST-DEV"`
	CreatedBy  string `json:"decision_by" validate:"required,max=100"`
	DecisionBy string `json:"decision_by_name" validate:"required,max=250"`
}

type ReqRecalculateOrder struct {
	ProspectID string  `json:"prospect_id" validate:"required,max=20" example:"TEST-DEV"`
	DPAmount   float64 `json:"dp_amount" validate:"required" example:"245000"`
	CreatedBy  string  `json:"decision_by" validate:"required,max=100"`
	DecisionBy string  `json:"decision_by_name" validate:"required,max=250"`
}

type ReqInquiryApproval struct {
	Search      string `json:"search"`
	BranchID    string `json:"branch_id" validate:"required,max=3"`
	MultiBranch string `json:"multi_branch" validate:"required,max=1"`
	Filter      string `json:"filter" validate:"max=15"`
	UserID      string `json:"user_id" validate:"required,max=20"`
	Alias       string `json:"alias" validate:"required,max=3"`
}

type ReqListQuotaDeviasi struct {
	Search   string `json:"search"`
	BranchID string `json:"branch_id" example:"400"`
	IsActive string `json:"is_active" example:"1 / 0"`
}

type ReqListQuotaDeviasiBranch struct {
	BranchID   string `json:"branch_id" example:"400"`
	BranchName string `json:"customer_status" example:"BEKASI"`
}

type ReqUpdateQuotaDeviasi struct {
	BranchID      string  `json:"branch_id" validate:"required" example:"400"`
	QuotaAmount   float64 `json:"quota_amount" validate:"required,gt=0,max=99999999999" example:"97500000"`
	QuotaAccount  int     `json:"quota_account" validate:"required,gt=0,max=999" example:"65"`
	IsActive      bool    `json:"is_active" example:"true"`
	UpdatedByName string  `json:"updated_by_name" validate:"required,max=200" example:"MUHAMMAD RONALD"`
}

type ReqUploadSettingQuotaDeviasi struct {
	UpdatedByName string `form:"updated_by_name" validate:"required,max=200"`
}

type ReqResetQuotaDeviasiBranch struct {
	BranchID      string `json:"branch_id" validate:"required" example:"400"`
	UpdatedByName string `json:"updated_by_name" validate:"required,max=200" example:"MUHAMMAD RONALD"`
}

type ReqResetAllQuotaDeviasi struct {
	UpdatedByName string `json:"updated_by_name" validate:"required,max=200" example:"MUHAMMAD RONALD"`
}

type ReqInquiryListOrder struct {
	OrderDateStart string `json:"order_date_start" example:"2025-01-01"`
	OrderDateEnd   string `json:"order_date_end" example:"2025-01-30"`
	BranchID       string `json:"branch_id" validate:"max=3"`
	Decision       string `json:"decision" validate:"max=3" example:"APR,REJ,CAN,CPR"`
	IsHighRisk     string `json:"is_highrisk"`
	ProspectID     string `json:"prospect_id" validate:"max=20"`
	IDNumber       string `json:"id_number" validate:"max=16" example:"357810XXXXXX0003"`
	LegalName      string `json:"legal_name" validate:"max=200"`
}

type ReqListMappingCluster struct {
	Search         string `json:"search"`
	BranchID       string `json:"branch_id" example:"400"`
	CustomerStatus string `json:"customer_status" validate:"max=10" example:"AO/RO"`
	BPKBNameType   string `json:"bpkb_name_type" validate:"max=1" example:"0"`
	Cluster        string `json:"cluster" validate:"max=20" example:"Cluster A"`
}

type ReqUploadMappingCluster struct {
	UserID string `form:"user_id" validate:"required,max=20"`
}

type ReqListMappingClusterBranch struct {
	BranchID   string `json:"branch_id" example:"400"`
	BranchName string `json:"customer_status" example:"BEKASI"`
}

type ReqHrisCareerHistory struct {
	Limit     string `json:"limit"`
	Page      int    `json:"page"`
	Column    string `json:"column"`
	Ascending bool   `json:"ascending"`
	Query     string `json:"query"`
}

type PrincipleAsset struct {
	ProspectID         string `json:"prospect_id" validate:"required,max=20,prospect_id_asset_principle" example:"SAL-1140024080800004"`
	IDNumber           string `json:"id_number"  validate:"required,number,len=16" example:"3506126712000001"`
	SpouseIDNumber     string `json:"spouse_id_number"  validate:"omitempty,number,len=16" example:"3506126712000002"`
	ManufactureYear    int    `json:"manufacture_year" validate:"required" example:"2020"`
	NoChassis          string `json:"chassis_number" validate:"required,max=25,htmlValidation" example:"MHKV1AA2JBK107322"`
	NoEngine           string `json:"engine_number" validate:"required,max=20,htmlValidation" example:"73218JAJK"`
	BranchID           string `json:"branch_id" validate:"required,max=10,htmlValidation" example:"426"`
	CC                 int    `json:"cc" validate:"required,min=100,max=9999" example:"1500"`
	TaxDate            string `json:"tax_date" validate:"required,dateformat" example:"2022-03-02"`
	STNKExpiredDate    string `json:"stnk_expired_date" validate:"required,dateformat" example:"2025-03-20"`
	OwnerAsset         string `json:"owner_asset" validate:"required,min=2,max=50,allowcharsname" example:"JONATHAN"`
	LicensePlate       string `json:"license_plate" validate:"required,max=50,htmlValidation" example:"B3006TBJ"`
	Color              string `json:"color" validate:"required,min=4,max=30,allowcharsname" example:"HITAM"`
	Brand              string `json:"brand" validate:"required,max=50,htmlValidation" example:"HONDA"`
	ResidenceAddress   string `json:"residence_address" validate:"required,allowcharsaddress,max=100" example:"Dermaga Baru"`
	ResidenceRT        string `json:"residence_rt" validate:"required,min=1,max=3,number" example:"001"`
	ResidenceRW        string `json:"residence_rw" validate:"required,min=1,max=3,number" example:"002"`
	ResidenceProvince  string `json:"residence_province" validate:"required,max=50,allowcharsname" example:"Jakarta"`
	ResidenceCity      string `json:"residence_city" validate:"required,max=30,allowcharsname" example:"Jakarta Timur"`
	ResidenceKecamatan string `json:"residence_kecamatan" validate:"required,max=30,allowcharsname" example:"Duren Sawit"`
	ResidenceKelurahan string `json:"residence_kelurahan" validate:"required,max=30,allowcharsname" example:"Klender"`
	ResidenceZipCode   string `json:"residence_zipcode" validate:"required,max=5,number" example:"13470"`
	ResidenceAreaPhone string `json:"residence_area_phone" validate:"omitempty,max=5,number" example:"021"`
	ResidencePhone     string `json:"residence_phone" validate:"omitempty,min=5,max=14,number" example:"86605224"`
	HomeStatus         string `json:"home_status" validate:"required,max=2,allowcharsname" example:"SD"`
	StaySinceYear      int    `json:"stay_since_year" validate:"required" example:"2024"`
	StaySinceMonth     int    `json:"stay_since_month" validate:"required,min=1,max=12" example:"4"`
	AssetCode          string `json:"asset_code" validate:"required,max=200,htmlValidation" example:"SUZUKI,KMOBIL,GRAND VITARA.JLX 2,0 AT"`
	STNKPhoto          string `json:"stnk_photo" validate:"url,max=250" example:"https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1140024081400003/stnk_SAL-1140024081400003.jpg"`
	KPMID              int    `json:"kpm_id" validate:"required"`
}

type PrinciplePemohon struct {
	ProspectID              string  `json:"prospect_id" validate:"required,max=20,prospect_id_pemohon_principle" example:"SAL-1140024080800004"`
	IDNumber                string  `json:"id_number"  validate:"required,number,len=16" example:"3506126712000001"`
	SpouseIDNumber          string  `json:"spouse_id_number"  validate:"omitempty,number,len=16" example:"3506126712000002"`
	MobilePhone             string  `json:"mobile_phone" validate:"required,min=9,max=14,number" example:"085880529100"`
	Email                   string  `json:"email" validate:"required,email,max=100" example:"jonathaxx@gmail.com"`
	LegalName               string  `json:"legal_name" validate:"required,allowcharsname,max=50" example:"Arya Danu"`
	FullName                string  `json:"full_name" validate:"required,allowcharsname,max=50" example:"Arya Danu"`
	BirthDate               string  `json:"birth_date" validate:"required,dateformat" example:"1992-09-11"`
	BirthPlace              string  `json:"birth_place" validate:"required,max=100,allowcharsname" example:"Jakarta"`
	SurgateMotherName       string  `json:"surgate_mother_name" validate:"required,max=50,allowcharsname" example:"IBU"`
	Gender                  string  `json:"gender" validate:"required,max=1,allowcharsname" example:"M"`
	Religion                string  `json:"religion" validate:"required,len=1"  example:"1"`
	LegalAddress            string  `json:"legal_address" validate:"required,allowcharsaddress,max=100" example:"Dermaga Baru"`
	LegalRT                 string  `json:"legal_rt" validate:"required,min=1,max=3,number" example:"001"`
	LegalRW                 string  `json:"legal_rw" validate:"required,min=1,max=3,number" example:"003"`
	LegalProvince           string  `json:"legal_province" validate:"required,max=50,allowcharsname" example:"Jakarta"`
	LegalCity               string  `json:"legal_city" validate:"required,max=30,allowcharsname" example:"Jakarta Timur"`
	LegalKecamatan          string  `json:"legal_kecamatan" validate:"required,max=30,allowcharsname" example:"Duren Sawit"`
	LegalKelurahan          string  `json:"legal_kelurahan" validate:"required,max=30,allowcharsname" example:"Klender"`
	LegalZipCode            string  `json:"legal_zipcode" validate:"required,max=5,number" example:"13470"`
	LegalPhoneArea          string  `json:"legal_phone_area" validate:"required,min=2,max=4,number" example:"021"`
	LegalPhone              string  `json:"legal_phone" validate:"required,min=5,max=14,number" example:"86605224"`
	Education               string  `json:"education" validate:"required,max=10" example:"S1"`
	ProfessionID            string  `json:"profession_id" validate:"required,max=10" example:"KRYSW"`
	JobType                 string  `json:"job_type" validate:"required,max=10" example:"0012"`
	JobPosition             string  `json:"job_position" validate:"required,max=10" example:"M"`
	EmploymentSinceMonth    int     `json:"employement_since_month" validate:"required,min=1,max=12" example:"12"`
	EmploymentSinceYear     int     `json:"employement_since_year" validate:"required" example:"2020"`
	CompanyName             string  `json:"company_name" validate:"required,min=2,max=50,allowcharsaddress" example:"PT KB Finansia"`
	EconomySectorID         string  `json:"economy_sector" validate:"required,max=10" example:"06"`
	IndustryTypeID          string  `json:"industry_type_id" validate:"required,max=10" example:"1000"`
	CompanyAddress          string  `json:"company_address" validate:"required,allowcharsaddress,max=100" example:"Dermaga Baru"`
	CompanyRT               string  `json:"company_rt" validate:"required,min=1,max=3,number" example:"001"`
	CompanyRW               string  `json:"company_rw" validate:"required,min=1,max=3,number" example:"003"`
	CompanyProvince         string  `json:"company_province" validate:"required,max=50,allowcharsname" example:"Jakarta"`
	CompanyCity             string  `json:"company_city" validate:"required,max=30,allowcharsname" example:"Jakarta Timur"`
	CompanyKecamatan        string  `json:"company_kecamatan" validate:"required,max=30,allowcharsname" example:"Duren Sawit"`
	CompanyKelurahan        string  `json:"company_kelurahan" validate:"required,max=30,allowcharsname" example:"Klender"`
	CompanyZipCode          string  `json:"company_zipcode" validate:"required,max=5,number" example:"13470"`
	CompanyPhoneArea        string  `json:"company_phone_area" validate:"required,min=2,max=4,number" example:"021"`
	CompanyPhone            string  `json:"company_phone" validate:"required,min=5,max=14,number" example:"86605224"`
	MonthlyFixedIncome      float64 `json:"monthly_fixed_income" validate:"required" example:"5000000"`
	MaritalStatus           string  `json:"marital_status" validate:"required,max=10,allowcharsname" example:"M"`
	SpouseLegalName         string  `json:"spouse_legal_name" validate:"omitempty,allowcharsname,max=50" example:"YULINAR NIATI"`
	SpouseFullName          string  `json:"spouse_full_name" validate:"omitempty,allowcharsname,max=50" example:"YULINAR NIATI"`
	SpouseBirthDate         string  `json:"spouse_birth_date" validate:"omitempty,dateformat" example:"1992-09-11"`
	SpouseBirthPlace        string  `json:"spouse_birth_place" validate:"omitempty,max=100,allowcharsname" example:"Jakarta"`
	SpouseSurgateMotherName string  `json:"spouse_surgate_mother_name"  validate:"omitempty,max=100,allowcharsname"  example:"MAMA"`
	SpouseMobilePhone       string  `json:"spouse_mobile_phone" validate:"omitempty,min=9,max=14,number" example:"085880529111"`
	SpouseIncome            float64 `json:"spouse_income" example:"5000000"`
	SelfiePhoto             string  `json:"selfie_photo" validate:"url,max=250" example:"https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1140024081400003/selfie_SAL-1140024081400003.jpg"`
	KtpPhoto                string  `json:"ktp_photo" validate:"url,max=250" example:"https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1140024081400003/ktp_SAL-1140024081400003.jpg"`
	BpkbName                string  `json:"-"`
}

type PrincipleElaborateLTV struct {
	ProspectID     string  `json:"prospect_id" validate:"prospect_id"`
	Tenor          int     `json:"tenor" validate:"required,max=60"`
	FinancePurpose string  `json:"finance_purpose" validate:"required"`
	LoanAmount     float64 `json:"loan_amount"`
}

type PrinciplePembiayaan struct {
	ProspectID            string   `json:"prospect_id" validate:"required,max=20,prospect_id_pembiayaan_principle" example:"SAL-1140024080800004"`
	Tenor                 int      `json:"tenor" validate:"required,max=60" example:"36"`
	AF                    float64  `json:"af" validate:"required,max=999999999999,af_principle" example:"84000000"`
	NTF                   float64  `json:"ntf" validate:"required,max=999999999999,ntf_principle" example:"150528000"`
	OTR                   float64  `json:"otr" validate:"required,max=999999999999,otr_principle" example:"105000000"`
	DPAmount              float64  `json:"down_payment_amount" validate:"omitempty,max=999999999999" example:"22000000"`
	AdminFee              float64  `json:"admin_fee" validate:"required,max=999999999999,admin_fee_principle" example:"1500000"`
	InstallmentAmount     float64  `json:"installment_amount" validate:"required,max=999999999999,installment_amount_principle" example:"4181333"`
	Dealer                string   `json:"dealer" validate:"omitempty,max=50,dealer_principle"`
	MonthlyVariableIncome *float64 `json:"monthly_variable_income" validate:"omitempty,max=999999999999" example:"3000000"`
	AssetCategoryID       string   `json:"asset_category_id" validate:"required,max=100,asset_category_id_principle" example:"BEBEK"`
	FinancePurpose        string   `json:"finance_purpose" validate:"required,max=100,oneof='Multiguna Pembayaran dengan Angsuran' 'Modal Kerja Fasilitas Modal Usaha' 'Multiguna Pembayaran dengan Cara Fasilitas Dana'" example:"Multiguna Pembayaran dengan Angsuran"`
	TipeUsaha             string   `json:"tipe_usaha" validate:"tipe_usaha,max=100,allowcharstipeusaha" example:"Jasa Kesehatan"`
}

type PrincipleEmergencyContact struct {
	ProspectID   string `json:"prospect_id" validate:"required,max=20,prospect_id_emcon_principle" example:"SAL-1140024080800004"`
	Name         string `json:"name" validate:"required,min=2,max=50,allowcharsname" example:"MULYADI"`
	Relationship string `json:"relationship" validate:"required,relationship,max=10" example:"FM"`
	MobilePhone  string `json:"mobile_phone" validate:"required,min=9,max=14,number" example:"0856789XXX1"`
	Address      string `json:"address" validate:"required,allowcharsaddress,max=90" example:"JL.PEGANGSAAN 1"`
	Rt           string `json:"rt" validate:"required,min=1,max=3,number" example:"008"`
	Rw           string `json:"rw" validate:"required,min=1,max=3,number" example:"017"`
	Kelurahan    string `json:"kelurahan" validate:"required,max=30" example:"TEGAL PARANG"`
	Kecamatan    string `json:"kecamatan" validate:"required,max=30" example:"MAMPANG PRAPATAN"`
	City         string `json:"city" validate:"required,max=50" example:"JAKARTA SELATAN"`
	Province     string `json:"province" validate:"required,max=50" example:"DKI JAKARTA"`
	ZipCode      string `json:"zip_code" validate:"required,len=5,number,ne=0" example:"12790"`
	AreaPhone    string `json:"area_phone" validate:"omitempty,min=3,max=4,number" example:"021"`
	Phone        string `json:"phone" validate:"omitempty,min=5,max=14,number" example:"567892"`
}

type ValidateNik struct {
	IDNumber string `json:"id_number" validate:"required,number,len=16"`
}

type UserInformation struct {
	UserID    string `json:"user_id" validate:"required,max=50"`
	UserTitle string `json:"user_title" validate:"required,max=50"`
}

type ReqMarsevLoanAmount struct {
	BranchID      string  `json:"branch_id"`
	OTR           float64 `json:"otr"`
	MaxLTV        int     `json:"max_ltv_los"`
	IsRecalculate bool    `json:"is_recalculate"`
	LoanAmount    float64 `json:"loan_amount"`
	DPAmount      float64 `json:"dp_amount_los"`
}

type ReqMarsevFilterProgram struct {
	Page                   int     `json:"page"`
	Limit                  int     `json:"limit"`
	BranchID               string  `json:"branch_id"`
	FinancingTypeCode      string  `json:"financing_type_code"`
	CustomerOccupationCode string  `json:"customer_occupation_code"`
	BpkbStatusCode         string  `json:"bpkb_status_code"`
	SourceApplication      string  `json:"source_application"`
	CustomerType           string  `json:"customer_type"`
	AssetUsageTypeCode     string  `json:"asset_usage_type_code"`
	AssetCategory          string  `json:"asset_category"`
	AssetBrand             string  `json:"asset_brand"`
	AssetYear              int     `json:"asset_year"`
	LoanAmount             float64 `json:"loan_amount"`
	Search                 string  `json:"search"`
	Tenor                  int     `json:"tenor"`
	SalesMethodID          int     `json:"sales_method_id"`
}

type ReqMarsevCalculateInstallment struct {
	ProgramID              string  `json:"program_id"`
	BranchID               string  `json:"branch_id"`
	CustomerOccupationCode string  `json:"customer_occupation_code"`
	AssetUsageTypeCode     string  `json:"asset_usage_type_code"`
	AssetYear              int     `json:"asset_year"`
	BpkbStatusCode         string  `json:"bpkb_status_code"`
	LoanAmount             float64 `json:"loan_amount"`
	Otr                    float64 `json:"otr"`
	RegionCode             string  `json:"region_code"`
	AssetCategory          string  `json:"asset_category"`
	CustomerBirthDate      string  `json:"customer_birth_date"`
	Tenor                  int     `json:"tenor"`
}

type ReqSallySubmit2wPrinciple struct {
	Document         []SallySubmit2wPrincipleDocument       `json:"documents"`
	Order            SallySubmit2wPrincipleOrder            `json:"order"`
	Kop              SallySubmit2wPrincipleKop              `json:"kop"`
	ObjekSewa        SallySubmit2wPrincipleObjekSewa        `json:"objeksewa"`
	Biaya            SallySubmit2wPrincipleBiaya            `json:"biaya"`
	ProgramMarketing SallySubmit2wPrincipleProgramMarketing `json:"program_marketing"`
	Filtering        SallySubmit2wPrincipleFiltering        `json:"filtering"`
}

type SallySubmit2wPrincipleOrder struct {
	Application SallySubmit2wPrincipleApplication `json:"application"`
	Asset       SallySubmit2wPrincipleAsset       `json:"asset"`
	Customer    SallySubmit2wPrincipleCustomer    `json:"customer"`
}

type SallySubmit2wPrincipleApplication struct {
	BranchID          string  `json:"branch_id"`
	BranchName        string  `json:"branch_name"`
	CmoID             string  `json:"cmo_id"`
	CmoName           string  `json:"cmo_name"`
	InstallmentAmount float64 `json:"installment_amount"`
	ApplicationFormID int     `json:"application_form_id"`
	OrderTypeID       int     `json:"order_type_id"`
	ProspectID        string  `json:"prospect_id"`
}

type SallySubmit2wPrincipleAsset struct {
	BPKBName              string `json:"bpkb_name"`
	BPKBOwnershipStatusID int    `json:"bpkb_ownership_status_id"`
	PoliceNo              string `json:"police_no"`
}

type SallySubmit2wPrincipleCustomer struct {
	CustomerID string `json:"customer_id"`
}

type SallySubmit2wPrincipleDocument struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type SallySubmit2wPrincipleKop struct {
	IsPSA              bool   `json:"is_psa"`
	PurposeOfFinancing string `json:"purpose_of_financing"`
	FinancingObject    string `json:"financing_object"`
}

type SallySubmit2wPrincipleObjekSewa struct {
	AssetUsageID       string  `json:"asset_usage_id"`
	CategoryID         string  `json:"category_id"`
	AssetCode          string  `json:"asset_code"`
	ManufacturingYear  int     `json:"manufacturing_year"`
	Color              string  `json:"color"`
	CylinderVolume     int     `json:"cylinder_volume"`
	IsBBN              bool    `json:"is_bbn"`
	PlateAreaCode      string  `json:"plate_area_code"`
	ChassisNumber      string  `json:"chassis_number"`
	MachineNumber      string  `json:"machine_number"`
	OTRAmount          float64 `json:"otr_amount"`
	ExpiredSTNKDate    string  `json:"expired_stnk_date"`
	ExpiredSTNKTaxDate string  `json:"expired_stnk_tax_date"`
	UpdatedBy          string  `json:"updated_by"`
}

type SallySubmit2wPrincipleBiaya struct {
	TotalOTRAmount        float64 `json:"total_otr_amount"`
	Tenor                 int     `json:"tenor"`
	LoanAmount            float64 `json:"loan_amount"`
	AdminFee              float64 `json:"admin_fee"`
	ProvisionFee          float64 `json:"provision_fee"`
	TotalDPAmount         float64 `json:"total_dp_amount"`
	AmountFinance         float64 `json:"amount_finance"`
	CorrespondenceAddress string  `json:"correspondence_address"`
	PaymentDay            int     `json:"payment_day"`
	RentPaymentMethod     string  `json:"rent_payment_method"`
	PersonalNPWPNumber    string  `json:"personal_npwp_number"`
	MaxLTVLOS             int     `json:"max_ltv_los"`
	UpdatedBy             string  `json:"updated_by"`
	LoanAmountMaximum     float64 `json:"loan_amount_maximum"`
}

type SallySubmit2wPrincipleProgramMarketing struct {
	ProgramMarketingID   string `json:"program_marketing_id"`
	ProgramMarketingName string `json:"program_marketing_name"`
	ProductOfferingID    string `json:"product_offering_id"`
	ProductOfferingName  string `json:"product_offering_name"`
	UpdatedBy            string `json:"updated_by"`
}

type SallySubmit2wPrincipleFiltering struct {
	Decision          string  `json:"decision"`
	Reason            string  `json:"reason"`
	CustomerStatus    string  `json:"customer_status"`
	CustomerStatusKMB string  `json:"customer_status_kmb"`
	CustomerSegment   string  `json:"customer_segment"`
	IsBlacklist       bool    `json:"is_blacklist"`
	NextProcess       bool    `json:"next_process"`
	PBKReportCustomer string  `json:"pbk_report_customer"`
	PBKReportSpouse   string  `json:"pbk_report_spouse"`
	BakiDebet         float64 `json:"baki_debet"`
}

type Update2wPrincipleTransaction struct {
	OrderID                    string  `json:"order_id"`
	KpmID                      int     `json:"kpm_id"`
	Source                     int     `json:"source"`
	StatusCode                 string  `json:"status_code"`
	ProductName                string  `json:"product_name"`
	Amount                     float64 `json:"amount"`
	AssetTypeCode              string  `json:"asset_type_code"`
	BranchCode                 string  `json:"branch_code"`
	ReferralCode               string  `json:"referral_code"`
	Is2wPrincipleApprovalOrder bool    `json:"is_2w_principle_approval_order"`
}

type PrincipleGetData struct {
	Context        string `json:"context"  validate:"required"`
	ProspectID     string `json:"prospect_id" validate:"required"`
	FinancePurpose string `json:"finance_purpose"`
	KPMID          int    `json:"kpm_id" validate:"omitempty"`
}

type PrinciplePublish struct {
	StatusCode string `json:"status_code" validate:"required"`
	ProspectID string `json:"prospect_id" validate:"required"`
}

type PrincipleUpdateStatus struct {
	ProspectID  string `json:"prospect_id"`
	OrderStatus string `json:"order_status"`
}

type CheckStep2Wilen struct {
	IDNumber string `json:"id_number" validate:"required,id_number" example:"ENCRYPTED NIK"`
}

type GetMaxLoanAmount struct {
	ProspectID         string  `json:"prospect_id" validate:"required,prospect_id,max=20,htmlValidation" example:"SAL-1140024080800004"`
	BranchID           string  `json:"branch_id" validate:"required,max=10,htmlValidation" example:"426"`
	IDNumber           string  `json:"id_number"  validate:"required,number,len=16" example:"3506126712000001"`
	BirthDate          string  `json:"birth_date" validate:"required,dateformat" example:"1992-09-11"`
	SurgateMotherName  string  `json:"surgate_mother_name" validate:"required,max=50,allowcharsname" example:"IBU"`
	LegalName          string  `json:"legal_name" validate:"required,allowcharsname,max=50" example:"Arya Danu"`
	MobilePhone        string  `json:"mobile_phone" validate:"required,min=9,max=14,number" example:"085880529100"`
	BPKBNameType       string  `json:"bpkb_name_type" validate:"required,bpkbname"`
	ManufactureYear    string  `json:"manufacture_year" validate:"required,len=4,number" example:"2020"`
	AssetCode          string  `json:"asset_code" validate:"required,max=200,htmlValidation" example:"SUZUKI,KMOBIL,GRAND VITARA.JLX 2,0 AT"`
	AssetUsageTypeCode string  `json:"asset_usage_type_code" validate:"required,oneof=C N S,htmlValidation" example:"C"`
	ReferralCode       *string `json:"referral_code" validate:"omitempty,htmlValidation"`
}

type GetAvailableTenor struct {
	ProspectID         string  `json:"prospect_id" validate:"required,prospect_id,max=20,htmlValidation" example:"SAL-1140024080800004"`
	BranchID           string  `json:"branch_id" validate:"required,max=10,htmlValidation" example:"426"`
	IDNumber           string  `json:"id_number"  validate:"required,number,len=16" example:"3506126712000001"`
	BirthDate          string  `json:"birth_date" validate:"required,dateformat" example:"1992-09-11"`
	SurgateMotherName  string  `json:"surgate_mother_name" validate:"required,max=50,allowcharsname" example:"IBU"`
	LegalName          string  `json:"legal_name" validate:"required,allowcharsname,max=50" example:"Arya Danu"`
	MobilePhone        string  `json:"mobile_phone" validate:"required,min=9,max=14,number" example:"085880529100"`
	BPKBNameType       string  `json:"bpkb_name_type" validate:"required,bpkbname"`
	ManufactureYear    string  `json:"manufacture_year" validate:"required,len=4,number" example:"2020"`
	AssetCode          string  `json:"asset_code" validate:"required,max=200,htmlValidation" example:"SUZUKI,KMOBIL,GRAND VITARA.JLX 2,0 AT"`
	AssetUsageTypeCode string  `json:"asset_usage_type_code" validate:"required,oneof=C N S,htmlValidation" example:"C"`
	LicensePlate       string  `json:"license_plate" validate:"required,max=50,htmlValidation" example:"B3006TBJ"`
	LoanAmount         float64 `json:"loan_amount"  validate:"required,max=999999999999" example:"105000000"`
	ReferralCode       *string `json:"referral_code" validate:"omitempty,max=200,htmlValidation" example:"SUZUKI"`
}

type Submission2Wilen struct {
	ProspectID              string  `json:"prospect_id" validate:"required,prospect_id,max=20,htmlValidation" example:"SAL-1140024080800004"`
	IDNumber                string  `json:"id_number"  validate:"required,number,len=16" example:"3506126712000001"`
	LegalName               string  `json:"legal_name" validate:"required,allowcharsname,max=50" example:"Arya Danu"`
	MobilePhone             string  `json:"mobile_phone" validate:"required,min=9,max=14,number" example:"085880529100"`
	Email                   string  `json:"email" validate:"required,email,max=100" example:"jonathaxx@gmail.com"`
	BirthPlace              string  `json:"birth_place" validate:"required,max=100,allowcharsname" example:"Jakarta"`
	BirthDate               string  `json:"birth_date" validate:"required,dateformat" example:"1992-09-11"`
	SurgateMotherName       string  `json:"surgate_mother_name" validate:"required,max=50,allowcharsname" example:"IBU"`
	Gender                  string  `json:"gender" validate:"required,max=1,allowcharsname" example:"M"`
	ResidenceAddress        string  `json:"residence_address" validate:"required,allowcharsaddress,max=100" example:"Dermaga Baru"`
	ResidenceRT             string  `json:"residence_rt" validate:"required,min=1,max=3,number" example:"001"`
	ResidenceRW             string  `json:"residence_rw" validate:"required,min=1,max=3,number" example:"002"`
	ResidenceProvince       string  `json:"residence_province" validate:"required,max=50,allowcharsname" example:"Jakarta"`
	ResidenceCity           string  `json:"residence_city" validate:"required,max=30,allowcharsname" example:"Jakarta Timur"`
	ResidenceKecamatan      string  `json:"residence_kecamatan" validate:"required,max=30,allowcharsname" example:"Duren Sawit"`
	ResidenceKelurahan      string  `json:"residence_kelurahan" validate:"required,max=30,allowcharsname" example:"Klender"`
	ResidenceZipCode        string  `json:"residence_zipcode" validate:"required,max=5,number" example:"13470"`
	BranchID                string  `json:"branch_id" validate:"required,max=10,htmlValidation" example:"426"`
	AssetCode               string  `json:"asset_code" validate:"required,max=200,htmlValidation" example:"K-HND.MOTOR.ABSOLUTE REVO"`
	ManufactureYear         string  `json:"manufacture_year" validate:"required,len=4,number" example:"2020"`
	LicensePlate            string  `json:"license_plate" validate:"required,max=50,htmlValidation" example:"B3006TBJ"`
	AssetUsageTypeCode      string  `json:"asset_usage_type_code" validate:"required,oneof=C N S,htmlValidation" example:"C"`
	BPKBNameType            string  `json:"bpkb_name_type" validate:"required,bpkbname" example:"K"`
	OwnerAsset              string  `json:"owner_asset" validate:"required,min=2,max=50,allowcharsname" example:"JONATHAN"`
	LoanAmount              float64 `json:"loan_amount"  validate:"required,max=999999999999" example:"105000000"`
	MaxLoanAmount           float64 `json:"max_loan_amount"  validate:"required,max=999999999999" example:"105000000"`
	Tenor                   int     `json:"tenor" validate:"required,max=60" example:"12"`
	InstallmentAmount       float64 `json:"installment_amount" validate:"required,max=999999999999" example:"4181333"`
	NumOfDependence         int     `json:"num_of_dependence" validate:"omitempty,max=50"  example:"1"`
	MaritalStatus           string  `json:"marital_status" validate:"required,marital"  example:"M"`
	SpouseIDNumber          string  `json:"spouse_id_number"  validate:"omitempty,number,len=16" example:"3506126712000002"`
	SpouseLegalName         string  `json:"spouse_legal_name" validate:"omitempty,allowcharsname,max=50" example:"YULINAR NIATI"`
	SpouseBirthDate         string  `json:"spouse_birth_date" validate:"omitempty,dateformat" example:"1992-09-11"`
	SpouseBirthPlace        string  `json:"spouse_birth_place" validate:"omitempty,max=100,allowcharsname" example:"Jakarta"`
	SpouseSurgateMotherName string  `json:"spouse_surgate_mother_name"  validate:"omitempty,max=100,allowcharsname"  example:"MAMA"`
	SpouseMobilePhone       string  `json:"spouse_mobile_phone" validate:"omitempty,min=9,max=14,number" example:"085880529111"`
	Education               string  `json:"education" validate:"required,max=10" example:"S1"`
	ProfessionID            string  `json:"profession_id" validate:"required,max=10" example:"KRYSW"`
	JobType                 string  `json:"job_type" validate:"required,max=10" example:"0012"`
	JobPosition             string  `json:"job_position" validate:"required,max=10" example:"M"`
	EmploymentSinceMonth    int     `json:"employment_since_month" validate:"required,min=1,max=12" example:"12"`
	EmploymentSinceYear     int     `json:"employment_since_year" validate:"required" example:"2020"`
	MonthlyFixedIncome      float64 `json:"monthly_fixed_income" validate:"required" example:"5000000"`
	SpouseIncome            float64 `json:"spouse_income" example:"5000000"`
	NoChassis               string  `json:"chassis_number" validate:"required,max=25,htmlValidation" example:"MHKV1AA2JBK107322"`
	HomeStatus              string  `json:"home_status" validate:"required,max=2,allowcharsname" example:"SD"`
	StaySinceYear           int     `json:"stay_since_year" validate:"required" example:"2024"`
	StaySinceMonth          int     `json:"stay_since_month" validate:"required,min=1,max=12" example:"4"`
	KtpPhoto                string  `json:"ktp_photo" validate:"url,max=250" example:"https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1140024081400003/ktp_SAL-1140024081400003.jpg"`
	SelfiePhoto             string  `json:"selfie_photo" validate:"url,max=250" example:"https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1140024081400003/selfie_SAL-1140024081400003.jpg"`
	AF                      float64 `json:"af" validate:"required,max=999999999999" example:"84000000"`
	NTF                     float64 `json:"ntf" validate:"required,max=999999999999" example:"150528000"`
	OTR                     float64 `json:"otr" validate:"required,max=999999999999" example:"105000000"`
	DPAmount                float64 `json:"down_payment_amount" validate:"omitempty,max=999999999999" example:"22000000"`
	AdminFee                float64 `json:"admin_fee" validate:"required,max=999999999999" example:"1500000"`
	Dealer                  string  `json:"dealer" validate:"omitempty,max=50"`
	AssetCategoryID         string  `json:"asset_category_id" validate:"required,max=100" example:"BEBEK"`
	KPMID                   int     `json:"kpm_id" validate:"required"`
	RentFinishDate          string  `json:"rent_finish_date" validate:"omitempty,dateformat" example:"2021-07-29"`
	ReferralCode            string  `json:"referral_code" validate:"omitempty,max=20,htmlValidation" example:"TQ72AJ"`
}

type History2Wilen struct {
	ProspectID *string    `json:"prospect_id" validate:"omitempty,prospect_id,max=20,htmlValidation" example:"SAL-1140024080800004"`
	StartDate  *time.Time `json:"start_date" validate:"omitempty" example:"2021-07-29"`
	EndDate    *time.Time `json:"end_date" validate:"omitempty" example:"2021-07-29"`
	Status     *string    `json:"status" validate:"omitempty" example:"SD"`
}

type Publish2Wilen struct {
	StatusCode string `json:"status_code" validate:"required"`
	ProspectID string `json:"prospect_id" validate:"required"`
}
