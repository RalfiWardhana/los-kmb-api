package request

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
	ProspectID            string             `json:"prospect_id" validate:"required"`
	ImageSelfie           string             `json:"image_selfie" validate:"required"`
	ImageKtp              string             `json:"ktp_url" validate:"required"`
	MonthlyFixedIncome    float64            `json:"monthly_fixed_income" validate:"required"`
	HomeStatus            string             `json:"home_status" validate:"required,max=2"`
	MonthlyVariableIncome float64            `json:"monthly_variable_income"`
	SpouseIncome          float64            `json:"spouse_income"`
	JobPosition           string             `json:"job_position" validate:"required"`
	ProfessionID          string             `json:"profession_id" validate:"required"`
	EmploymentSinceYear   string             `json:"employment_since_year" validate:"required,len=4"`
	EmploymentSinceMonth  string             `json:"employment_since_month" validate:"required,len=2"`
	StaySinceYear         string             `json:"stay_since_year" validate:"required,len=4"`
	StaySinceMonth        string             `json:"stay_since_month" validate:"required,len=2"`
	BirthDate             string             `json:"birth_date" validate:"required,dateformat"`
	BirthPlace            string             `json:"birth_place" validate:"required,allowcharsname"`
	Tenor                 int                `json:"tenor" validate:"required"`
	IDNumber              string             `json:"id_number" validate:"required,len=16,number"`
	LegalName             string             `json:"legal_name" validate:"required,allowcharsname"`
	MotherName            string             `json:"surgate_mother_name" validate:"required,allowcharsname"`
	Spouse                *DupcheckApiSpouse `json:"spouse" validate:"omitempty"`
	EngineNo              string             `json:"no_engine" validate:"required"`
	RangkaNo              string             `json:"no_rangka" validate:"required"`
	ManufactureYear       string             `json:"manufacture_year" validate:"required,len=4,number"`
	BPKBName              string             `json:"bpkb_name" validate:"required,bpkbname"`
	NumOfDependence       int                `json:"num_of_dependence" validate:"required"`
	OTRPrice              float64            `json:"otr" validate:"required"`
	NTF                   float64            `json:"ntf" validate:"required"`
	LegalZipCode          string             `json:"legal_zip_code" validate:"required"`
	CompanyZipCode        string             `json:"company_zip_code" validate:"required"`
	Gender                string             `json:"gender" validate:"required"`
	InstallmentAmount     float64            `json:"installment_amount" validate:"required"`
	MaritalStatus         string             `json:"marital_status"`
}

type SpouseDupcheck struct {
	IDNumber   string `json:"spouse_id_number" validate:"required,len=16,number"`
	LegalName  string `json:"spouse_legal_name" validate:"required,allowcharsname"`
	BirthDate  string `json:"spouse_birth_date" validate:"required,dateformat"`
	MotherName string `json:"spouse_surgate_mother_name" validate:"required,allowcharsname"`
}

type DupcheckApiSpouse struct {
	IDNumber   string `json:"spouse_id_number" validate:"required,len=16"`
	LegalName  string `json:"spouse_legal_name" validate:"required,allowcharsname"`
	BirthDate  string `json:"spouse_birth_date" validate:"required,dateformat"`
	Gender     string `json:"spouse_gender" validate:"required"`
	MotherName string `json:"spouse_surgate_mother_name" validate:"required,allowcharsname"`
}

type CustomerData struct {
	StatusKonsumen string `json:"status_konsumen"`
	IDNumber       string `json:"id_number"`
	LegalName      string `json:"legal_name"`
	BirthDate      string `json:"birth_date"`
	MotherName     string `json:"mother_name"`
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
	Metadata           Metadata           `json:"metadata" validate:"required"`
	Surveyor           []Surveyor         `json:"surveyor" validate:"required"`
}

type ValidateOmset struct {
	CustomerOmset []CustomerOmset `json:"customer_omset" validate:"required,len=3"`
}

type CustomerOmset struct {
	MonthlyOmsetYear  string  `json:"monthly_omset_year" example:"2021"`
	MonthlyOmsetMonth string  `json:"monthly_omset_month" example:"01"`
	MonthlyOmset      float64 `json:"monthly_omset" example:"5000000"`
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
	ProspectID        string `json:"prospect_id" validate:"required" example:"EFM01426202106100001"`
	BranchID          string `json:"branch_id" validate:"min=2" example:"426"`
	ApplicationSource string `json:"application_source" validate:"required" example:"H"`
	Channel           string `json:"channel" validate:"channel" example:"OFF"`
	Lob               string `json:"lob" validate:"lob" example:"KMOB"`
	OrderAt           string `json:"order_at" validate:"required" example:"2021-07-15T11:44:05+07:00"`
	IncomingSource    string `json:"incoming_source" validate:"incoming" example:"SLY"`
}

type CustomerPersonal struct {
	IDType                     string   `json:"id_type" validate:"ktp" example:"KTP"`
	IDNumber                   string   `json:"id_number" validate:"required,id_number" example:"ENCRYPTED"`
	FullName                   string   `json:"full_name" validate:"required,allow_name" example:"ENCRYPTED"`
	LegalName                  string   `json:"legal_name" validate:"required,allow_name" example:"ENCRYPTED"`
	BirthPlace                 string   `json:"birth_place" validate:"min=3" example:"JAKARTA"`
	BirthDate                  string   `json:"birth_date" validate:"dateformat" example:"1991-01-12"`
	SurgateMotherName          string   `json:"surgate_mother_name" validate:"required,allow_name" example:"ENCRYPTED"`
	Gender                     string   `json:"gender" validate:"gender" example:"M"`
	MobilePhone                string   `json:"mobile_phone" validate:"min=9,max=14" example:"085689XXX01"`
	Email                      string   `json:"email" validate:"email,max=100" example:"jonathaxx@gmail.com"`
	StaySinceYear              string   `json:"stay_since_year" validate:"len=4" example:"2018"`
	StaySinceMonth             string   `json:"stay_since_month" validate:"len=2" example:"03"`
	HomeStatus                 string   `json:"home_status" validate:"home" example:"KL"`
	NPWP                       *string  `json:"npwp" example:"994646808XXX895"`
	Education                  string   `json:"education" validate:"education"  example:"S1"`
	MaritalStatus              string   `json:"marital_status" validate:"marital"  example:"M"`
	NumOfDependence            *int     `json:"num_of_dependence" validate:"required"  example:"1"`
	IDTypeIssueDate            *string  `json:"id_type_isssue_date"  example:"2021-07-29"`
	Religion                   string   `json:"religion" validate:"required,len=1"  example:"1"`
	PersonalCustomerType       string   `json:"personal_customer_type" validate:"required"  example:"M"`
	ExpiredDate                *string  `json:"expired_date"  example:"2021-07-29"`
	Nationality                string   `json:"nationality" validate:"required"  example:"WNI"`
	WNACountry                 string   `json:"wna_country" validate:"required"  example:"-"`
	HomeLocation               string   `json:"home_location" validate:"required"  example:"N"`
	CustomerGroup              string   `json:"customer_group" validate:"required"  example:"2"`
	KKNo                       string   `json:"kk_no" validate:"required"  example:"97846094XXX34346"`
	BankID                     string   `gorm:"column:BankID" json:"bank_id" example:"BCA"`
	AccountNo                  string   `gorm:"column:AccountNo" json:"account_no" example:"567XX021"`
	AccountName                string   `gorm:"column:AccountName" json:"account_name" example:"JONATHAN"`
	LivingCostAmount           *float64 `json:"living_cost_amount" validate:"required" example:"0"`
	Counterpart                int      `json:"counterpart" validate:"required" example:"169"`
	DebtBusinessScale          string   `json:"debt_business_scale" validate:"required" example:"01"`
	DebtGroup                  string   `json:"debt_group" validate:"required" example:"303"`
	IsAffiliateWithPP          string   `json:"is_affiliate_with_pp" validate:"required" example:"N"`
	AgreetoAcceptOtherOffering int      `json:"agree_to_accept_other_offering" validate:"required" example:"1"`
	DataType                   string   `json:"data_type" validate:"required" example:"G"`
	Status                     string   `json:"status" validate:"required" example:"F"`
}

type CustomerEmployment struct {
	ProfessionID          string   `json:"profession_id" validate:"profession" example:"WRST"`
	EmploymentSinceYear   string   `json:"employment_since_year" validate:"len=4" example:"2020"`
	EmploymentSinceMonth  string   `json:"employment_since_month" validate:"len=2" example:"02"`
	MonthlyFixedIncome    float64  `json:"monthly_fixed_income" validate:"gt=0" example:"5000000"`
	JobType               string   `json:"job_type" validate:"required" example:"008"`
	JobPosition           string   `json:"job_position" validate:"required" example:"S"`
	MonthlyVariableIncome *float64 `json:"monthly_variable_income" validate:"required" example:"3000000"`
	SpouseIncome          *float64 `json:"spouse_income" validate:"required" example:"6000000"`
	CompanyName           string   `json:"company_name" validate:"required" example:"PT.KIMIA FARMA"`
	IndustryTypeID        string   `json:"industry_type_id" validate:"required" example:"9990"`
	ExtCompanyPhone       *string  `json:"company_phone_ext" example:"442"`
	SourceOtherIncome     *string  `json:"source_other_income" example:"TOKO MAKMUR"`
}

type Address struct {
	Type      string `json:"type" validate:"address"  example:"RESIDENCE"`
	Address   string `json:"address" validate:"required"  example:"JL.PEGANGSAAN 1"`
	Rt        string `json:"rt" validate:"min=1,max=3"  example:"008"`
	Rw        string `json:"rw" validate:"min=1,max=3"  example:"017"`
	Kelurahan string `json:"kelurahan" validate:"required"  example:"TEGAL PARANG"`
	Kecamatan string `json:"kecamatan" validate:"required"  example:"MAMPANG PRAPATAN"`
	City      string `json:"city" validate:"required"  example:"JAKARTA SELATAN"`
	ZipCode   string `json:"zip_code" validate:"required"  example:"12790"`
	AreaPhone string `json:"area_phone" validate:"min=2,max=4"  example:"021"`
	Phone     string `json:"phone" validate:"required,max=10"  example:"84522"`
}

type CustomerPhoto struct {
	ID  string `json:"id" validate:"photo" example:"KTP"`
	Url string `json:"url" validate:"url" example:"https://dev-media.kreditplus.com/media/reference/20000/KPM-3677/ktp_EFM-3677.jpg"`
}

type CustomerEmcon struct {
	Name                   string `json:"name" validate:"min=2,allowcharsname" example:"MULYADI"`
	Relationship           string `json:"relationship" validate:"relationship" example:"FM"`
	MobilePhone            string `json:"mobile_phone" validate:"min=9,max=14" example:"0856789XXX1"`
	AreaPhone              string `json:"emergency_area_phone_office" validate:"min=2,max=4" example:"021"`
	Phone                  string `json:"emergency_phone_office" validate:"required" example:"567892"`
	VerificationWith       string `json:"verification_with" validate:"required" example:"JONO"`
	ApplicationEmconSesuai string `json:"application_emcon_sesuai" validate:"required" example:"1"`
	VerifyBy               string `json:"verify_by" validate:"required" example:"PHONE"`
	KnownCustomerAddress   string `json:"known_customer_address" validate:"required" example:"1"`
	KnownCustomerJob       string `json:"known_customer_job" validate:"required" example:"1"`
}

type CustomerSpouse struct {
	IDNumber             string  `json:"id_number" validate:"required,id_number" example:"177105550374XX01"`
	FullName             string  `json:"full_name" validate:"required,allow_name" example:"SUSI BUNGA"`
	LegalName            string  `json:"legal_name" validate:"required,allow_name" example:"SUSI BUNGA"`
	BirthDate            string  `json:"birth_date"  validate:"dateformat" example:"1991-01-29"`
	BirthPlace           string  `json:"birth_place" validate:"required" example:"JAKARTA"`
	Gender               string  `json:"gender" validate:"gender" example:"F"`
	SurgateMotherName    string  `json:"surgate_mother_name" validate:"required,allow_name" example:"TUTI"`
	CompanyPhone         *string `json:"company_phone" example:"865542"`
	CompanyName          *string `json:"company_name" example:"PT.BUMI KARYA"`
	MobilePhone          string  `json:"mobile_phone" validate:"min=9,max=14" example:"08772012XXX0"`
	ProfessionID         *string `json:"profession_id" example:"KRYSW"`
	EmploymentSinceYear  *string `json:"employment_since_year" example:"2020"`
	EmploymentSinceMonth *string `json:"employment_since_month" example:"02"`
	JobType              string  `json:"job_type" example:"001"`
	JobPosition          *string `json:"job_position" example:"S"`
	NPWP                 *string `json:"npwp" example:"994646808XXX895"`
	Education            string  `json:"education" example:"S1"`
	Email                string  `json:"email" example:"sulasxx@gmail.com"`
}

type Apk struct {
	OtherFee                    float64  `json:"other_fee" example:"0"`
	Tenor                       int      `json:"tenor" validate:"required" example:"36"`
	ProductOfferingID           string   `json:"product_offering_id" validate:"required" example:"NLMKKAPSEP"`
	ProductOfferingDesc         string   `json:"product_offering_desc"`
	ProductID                   string   `json:"product_id" validate:"required" example:"1SNLMK"`
	OTR                         float64  `json:"otr" validate:"required" example:"105000000"`
	DPAmount                    float64  `json:"down_payment_amount" validate:"required" example:"22000000"`
	NTF                         float64  `json:"ntf" validate:"required" example:"150528000"`
	AF                          float64  `json:"af" validate:"required" example:"84000000"`
	AoID                        string   `json:"aoid" validate:"required" example:"81088"`
	AdminFee                    *float64 `json:"admin_fee" validate:"required" example:"1500000"`
	InstallmentAmount           float64  `json:"installment_amount" validate:"required" example:"4181333"`
	PercentDP                   *float64 `json:"down_payment_rate" validate:"required" example:"20.95"`
	PremiumAmountToCustomer     float64  `json:"premium_amount_to_customer" example:"2184000"`
	FidusiaFee                  *float64 `json:"fidusia_fee" example:"0"`
	InterestRate                *float64 `json:"interest_rate" validate:"required" example:"2.2"`
	InsuranceRate               float64  `json:"insurance_rate" validate:"required" example:"3"`
	InterestAmount              *float64 `json:"interest_amount" validate:"required" example:"66528000"`
	InsuranceAmount             float64  `json:"insurance_amount" validate:"required" example:"3150000"`
	FirstPayment                float64  `json:"first_payment" validate:"required" example:"30831334"`
	FirstInstallment            string   `json:"first_installment" validate:"required" example:"AR"`
	FirstPaymentDate            string   `json:"first_payment_date" validate:"required,dateformat" example:"2021-08-16"`
	PaymentMethod               string   `json:"payment_method" validate:"required" example:"CA"`
	SurveyFee                   *float64 `json:"survey_fee" validate:"required" example:"0"`
	IsFidusiaCovered            string   `json:"is_fidusia_covered" validate:"required" example:"Y"`
	ProvisionFee                *float64 `json:"provision_fee" validate:"required" example:"2475000"`
	InsAssetPaidBy              string   `json:"ins_asset_paid_by" validate:"required" example:"CU"`
	InsAssetPeriod              string   `json:"ins_asset_period" validate:"required" example:"FT"`
	Discount                    *float64 `json:"discount" validate:"required" example:"0"`
	EffectiveRate               float64  `json:"effective_rate" validate:"required" example:"26.4"`
	SalesmanID                  string   `json:"salesman_id" validate:"required" example:"81088"`
	SupplierBankAccountID       string   `json:"supplier_bank_account_id" validate:"required" example:"1"`
	LifeInsuranceCoyBranchID    string   `json:"life_insurance_coy_branch_id" example:"426"`
	LifeInsuranceAmountCoverage float64  `json:"life_insurance_amount_coverage" example:"105000000"`
	CommisionSubsidy            float64  `json:"commision_subsidi" example:"0"`
	FinancePurpose              string   `json:"finance_purpose" validate:"required"`
	Dealer                      string   `json:"dealer"`
	LoanAmount                  float64  `json:"loan_amount" example:"105000000"`
}

type Item struct {
	SupplierID                   string  `json:"supplier_id" validate:"required" example:"42600342"`
	AssetCode                    string  `json:"asset_code" validate:"required" example:"SUZUKI,KMOBIL,GRAND VITARA.JLX 2,0 AT"`
	ManufactureYear              string  `json:"manufacture_year" validate:"len=4" example:"2020"`
	NoChassis                    string  `json:"chassis_number" validate:"required" example:"MHKV1AA2JBK107322"`
	NoEngine                     string  `json:"engine_number" validate:"required" example:"73218JAJK"`
	Qty                          int     `json:"qty"  validate:"required" example:"1"`
	POS                          string  `json:"pos" validate:"required" example:"426"`
	CC                           string  `json:"cc" validate:"required" example:"1500"`
	Condition                    string  `json:"condition" validate:"required" example:"U"`
	AssetUsage                   string  `json:"asset_usage" validate:"required" example:"N"`
	Region                       string  `json:"region" validate:"required" example:"0"`
	TaxDate                      string  `json:"tax_date" validate:"required" example:"2022-03-02"`
	STNKExpiredDate              string  `json:"stnk_expired_date" validate:"required" example:"2025-03-20"`
	CategoryID                   string  `json:"category_id" validate:"required" example:"SEDAN"`
	AssetDescription             string  `json:"asset_description" validate:"required" example:"SUZUKI.KMOBIL.GRAND VITARA.JLX 2,0 AT"`
	BPKBName                     string  `json:"bpkb_name" validate:"len=1" example:"B"`
	OwnerAsset                   string  `json:"owner_asset" validate:"required" example:"JONATHAN"`
	LicensePlate                 string  `json:"license_plate" validate:"required" example:"3006TBJ"`
	Color                        string  `json:"color" validate:"required" example:"HITAM"`
	AssetInsuranceAmountCoverage float64 `json:"asset_insurance_amount_coverage" validate:"required" example:"105000000"`
	InsAssetInsuredBy            string  `json:"ins_asset_insured_by" validate:"required" example:"CO"`
	InsuranceCoyBranchID         string  `json:"insurance_coy_branch_id" validate:"required" example:"426"`
	CoverageType                 string  `json:"coverage_type" validate:"required" example:"TLO"`
	OwnerKTP                     string  `json:"owner_ktp" validate:"len=16,number" example:"3172024508XXX002"`
	Brand                        string  `json:"brand" validate:"required" example:"TOYOTA"`
	PremiumAmountToCustomer      float64 `json:"premium_amount_to_customer" example:"2184000"`
}

type Agent struct {
	CmoRecom  string `json:"cmo_recom" validate:"recom" example:"1"`
	CmoName   string `json:"cmo_name" validate:"required" example:"SETO MULYA"`
	CmoNik    string `json:"cmo_nik" validate:"required" example:"93510"`
	RecomDate string `json:"recom_date" validate:"required" example:"2021-07-15"`
}

type Metadata struct {
	CustomerIp   string `json:"customer_ip" example:"202.147.198.222"`
	CustomerLat  string `json:"customer_lat" example:"-6.409235"`
	CustomerLong string `json:"customer_long" example:"106.974231"`
	CallbackUrl  string `json:"callback_url" example:"https://dev-sally-kmob-api.kreditplus.com/api/los/v1/los/status"`
}

type Filtering struct {
	ProspectID string           `json:"prospect_id" validate:"required,max=20" example:"SAL042600001"`
	BranchID   string           `json:"branch_id" validate:"required,branch_id" example:"426"`
	IDNumber   string           `json:"id_number" validate:"required,id_number" example:"ENCRYPTED NIK"`
	LegalName  string           `json:"legal_name" validate:"required,allow_name" example:"ENCRYPTED LEGAL NAME"`
	BirthDate  string           `json:"birth_date" validate:"required,dateformat" example:"YYYY-MM-DD"`
	Gender     string           `json:"gender" validate:"required,gender" example:"M"`
	MotherName string           `json:"surgate_mother_name" validate:"required,allow_name" example:"ENCRYPTED SURGATE MOTHER NAME"`
	BPKBName   string           `json:"bpkb_name" validate:"required,bpkbname" example:"K"`
	Spouse     *FilteringSpouse `json:"spouse" validate:"omitempty"`
}

type FilteringSpouse struct {
	IDNumber   string `json:"spouse_id_number" validate:"required,id_number"  example:"ENCRYPTED NIK"`
	LegalName  string `json:"spouse_legal_name" validate:"required,allow_name" example:"ENCRYPTED LEGAL NAME"`
	BirthDate  string `json:"spouse_birth_date" validate:"required,dateformat" example:"YYYY-MM-DD"`
	Gender     string `json:"spouse_gender" validate:"required,gender" example:"F"`
	MotherName string `json:"spouse_surgate_mother_name" validate:"required,allow_name" example:"ENCRYPTED SURGATE MOTHER NAME"`
}

type ElaborateLTV struct {
	ProspectID        string `json:"prospect_id" validate:"required,max=20"`
	Tenor             int    `json:"tenor" validate:"required"`
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
	ProspectID       string       `json:"prospect_id"`
	SupplierID       string       `json:"supplier_id"`
	CBFound          bool         `json:"cb_found"`
	StatusKonsumen   string       `json:"status_konsumen"`
	RequestorID      string       `json:"requestor_id"`
	Journey          string       `json:"journey"`
	PhoneNumber      string       `json:"phone_number"`
	ScoreGeneratorID string       `json:"score_generator_id"`
	Data             DataScorePro `json:"data"`
}

type DataScorePro struct {
	ZipCode         int     `json:"zip_code"`
	Education       string  `json:"education"`
	FirstFourOfCell string  `json:"first_four_of_cell"`
	Ltv             float64 `json:"ltv"`
	VehicleBrand    string  `json:"vehicle_brand"`
	LengthOfStay    int     `json:"length_of_stay"`
	NoKol1Active    int     `json:"no_kol1_active"`
	StnkStatus      string  `json:"stnk_status"`
	Nom0312MntAll   int     `json:"nom03_12mth_all"`
	HomeStatus      string  `json:"home_status"`
	Tenor           int     `json:"tenor"`
}

type ScoreProRoaoIntegrator struct {
	ProspectID       string           `json:"prospect_id"`
	SupplierID       string           `json:"supplier_id"`
	CBFound          bool             `json:"cb_found"`
	StatusKonsumen   string           `json:"status_konsumen"`
	RequestorID      string           `json:"requestor_id"`
	Journey          string           `json:"journey"`
	PhoneNumber      string           `json:"phone_number"`
	ScoreGeneratorID string           `json:"score_generator_id"`
	Data             DataScoreProROAO `json:"data"`
}

type DataScoreProROAO struct {
	StatusCustomer      string  `json:"status_customer"`
	MobFirst            int     `json:"mob_first"`
	NtfOtr              float64 `json:"ntf_otr"`
	LengthOfStay        int     `json:"length_of_stay"`
	ZipCode             int     `json:"zip_code"`
	ProfessionID        string  `json:"profession_id"`
	BPKBName            string  `json:"bpkb_name"`
	Tenor               int     `json:"tenor"`
	Gender              string  `json:"gender"`
	WorstOvd            int     `json:"worst_ovd"`
	TotBakiDebet3160Dpd int     `json:"tot_bakidebet_31_60dpd"`
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
