package dupcheck

type DupcheckApi struct {
	ProspectID            string             `json:"prospect_id" validate:"required"`
	Image1                string             `json:"ktp_image" validate:"required,url"`
	Image2                string             `json:"selfie_image" validate:"required,url"`
	MonthlyFixedIncome    float64            `json:"monthly_fixed_income" validate:"required"`
	HomeStatus            string             `json:"home_status" validate:"required,max=2"`
	MonthlyVariableIncome float64            `json:"monthly_variable_income"`
	SpouseIncome          float64            `json:"spouse_income"`
	JobPosition           string             `json:"job_position" validate:"required"`
	EmploymentSinceYear   string             `json:"employment_since_year" validate:"required,len=4"`
	EmploymentSinceMonth  string             `json:"employment_since_month" validate:"required,len=2"`
	StaySinceYear         string             `json:"stay_since_year" validate:"required,len=4"`
	StaySinceMonth        string             `json:"stay_since_month" validate:"required,len=2"`
	BirthDate             string             `json:"birth_date" validate:"required,dateformat"`
	Tenor                 int                `json:"tenor" validate:"required"`
	IDNumber              string             `json:"id_number" validate:"required,len=16,number"`
	LegalName             string             `json:"legal_name" validate:"required,allowcharsname"`
	MotherName            string             `json:"surgate_mother_name" validate:"required,allowcharsname"`
	Spouse                *DupcheckApiSpouse `json:"spouse" validate:"omitempty"`
	EngineNo              string             `json:"no_engine" validate:"required"`
	ManufactureYear       string             `json:"manufacture_year" validate:"required,len=4,number"`
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
