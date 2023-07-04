package dupcheck

type DupcheckApi struct {
	CustomerID            int                `json:"customer_id" validate:"required"`
	ProspectID            string             `json:"prospect_id" validate:"required"`
	ImageSelfie1          string             `json:"image_selfie_1" validate:"required"`
	ImageSelfie2          string             `json:"image_selfie_2" validate:"required"`
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

type GenderCompare struct {
	Gender bool `json:"spouse_gender" validate:"spouse_gender"`
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
	CustomerID   int     `json:"customer_id" validate:"required"`
	ImageSelfie1 string  `json:"image_selfie_1" validate:"required"`
	ImageSelfie2 string  `json:"image_selfie_2" validate:"required"`
	ImageKtp     string  `json:"ktp_url" validate:"required"`
	Lob          string  `json:"lob"  validate:"required"`
	IDNumber     string  `json:"id_number" validate:"required,len=16,number"`
	BirthDate    string  `json:"birth_date" validate:"required,dateformat"`
	BirthPlace   string  `json:"birth_place" validate:"required"`
	LegalName    string  `json:"legal_name" validate:"min=2,allowcharsname" example:"JONATHAN"`
	FaceType     *string `json:"type" validate:"omitempty,oneof=null PIN DEVICE"`
}

type ReqRejectTenor struct {
	ProspectID string `json:"prospect_id" validate:"required"`
	IDNumber   string `json:"id_number" validate:"required,len=16,number" example:"1234XXXXXXXX0001"`
}
