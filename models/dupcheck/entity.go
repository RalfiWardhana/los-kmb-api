package dupcheck

import "time"

type AppConfig struct {
	GroupName string    `gorm:"type:varchar(50);column:group_name"`
	Lob       string    `gorm:"type:varchar(10);column:lob"`
	Key       string    `gorm:"type:varchar(50);column:key"`
	Value     string    `gorm:"type:varchar(255);column:value"`
	IsActive  int       `gorm:"column:is_active"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (c *AppConfig) TableName() string {
	return "app_config"
}

type NewDupcheck struct {
	ProspectID     string    `gorm:"column:ProspectID"`
	CustomerStatus string    `gorm:"column:customer_status"`
	CustomerType   string    `gorm:"column:customer_type"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (c *NewDupcheck) TableName() string {
	return "new_dupcheck"
}

type DummyCustomerDomain struct {
	IDNumber string `gorm:"type:varchar(50);column:id_number"`
	Response string `gorm:"type:text;column:response"`
	Note     string `gorm:"type:varchar(200);column:note"`
}

func (c *DummyCustomerDomain) TableName() string {
	return "dummy_cusomer_domain"
}

type DummyLatestPaidInstallment struct {
	IDNumber string `gorm:"type:varchar(50);column:id_number"`
	Response string `gorm:"type:text;column:response"`
	Note     string `gorm:"type:varchar(200);column:note"`
}

func (c *DummyLatestPaidInstallment) TableName() string {
	return "dummy_latest_paid_installment"
}

type ScanInstallmentAmount struct {
	IDNumber          string  `gorm:"column:IDNumber"`
	LegalName         string  `gorm:"column:LegalName"`
	BirthDate         string  `gorm:"column:BirthDate"`
	SurgateMotherName string  `gorm:"column:SurgateMotherName"`
	InstallmentAmount float64 `gorm:"column:InstallmentAmount"`
	NTF               float64 `gorm:"column:NTF"`
}

type Encrypted struct {
	LegalName         string `gorm:"column:LegalName"`
	FullName          string `gorm:"column:FullName"`
	SurgateMotherName string `gorm:"column:SurgateMotherName"`
	Email             string `gorm:"column:Email"`
	MobilePhone       string `gorm:"column:MobilePhone"`
	BirthPlace        string `gorm:"column:BirthPlace"`
	ResidenceAddress  string `gorm:"column:ResidenceAddress"`
	LegalAddress      string `gorm:"column:LegalAddress"`
	CompanyAddress    string `gorm:"column:CompanyAddress"`
	EmergencyAddress  string `gorm:"column:EmergencyAddress"`
	OwnerAddress      string `gorm:"column:OwnerAddress"`
	LocationAddress   string `gorm:"column:LocationAddress"`
	MailingAddress    string `gorm:"column:MailingAddress"`
	IDNumber          string `gorm:"column:IDNumber"`
}

type ConfigPMK struct {
	Data DataPMK `json:"data"`
}

type DataPMK struct {
	MinAgeMarried    int  `json:"min_age_married"`
	MinAgeSingle     int  `json:"min_age_single"`
	MaritalChecking  bool `json:"marital_checking"`
	MaxAgeLimit      int  `json:"max_age_limit"`
	LengthOfBusiness int  `json:"length_of_business"`
	LengthOfWork     int  `json:"length_of_work"`
	LengthOfStay     struct {
		RumahSendiri int `json:"sd"`
		RumahDinas   int `json:"rd"`
		RumahKontrak int `json:"rk"`
	} `json:"length_of_stay"`
	MinimalIncome     float64 `json:"minimal_income"`
	ManufacturingYear int     `json:"manufacturing_year"`
}

type DupcheckRejectionNokaNosin struct {
	Id                   string    `gorm:"column:id"`
	NoMesin              string    `gorm:"type:varchar(20);column:NoMesin"`
	NoRangka             string    `gorm:"type:varchar(20);column:NoRangka"`
	NumberOfRetry        int       `gorm:"column:NumberOfRetry"`
	IsBanned             int       `gorm:"column:IsBanned"`
	ProspectID           string    `gorm:"type:varchar(20);column:ProspectID"`
	IDNumber             string    `gorm:"type:varchar(16);column:IDNumber"`
	LegalName            string    `gorm:"type:varchar(100);column:LegalName"`
	BirthPlace           string    `gorm:"type:varchar(50);column:BirthPlace"`
	BirthDate            string    `gorm:"column:BirthDate"`
	MonthlyFixedIncome   float64   `gorm:"column:MonthlyFixedIncome"`
	EmploymentSinceYear  string    `gorm:"type:varchar(4);column:EmploymentSinceYear"`
	EmploymentSinceMonth string    `gorm:"type:varchar(2);column:EmploymentSinceMonth"`
	StaySinceYear        string    `gorm:"type:varchar(4);column:StaySinceYear"`
	StaySinceMonth       string    `gorm:"type:varchar(2);column:StaySinceMonth"`
	BPKBName             string    `gorm:"type:varchar(2);column:BPKBName"`
	Gender               string    `gorm:"type:varchar(1);column:Gender"`
	MaritalStatus        string    `gorm:"type:varchar(1);column:MaritalStatus"`
	NumOfDependence      int       `gorm:"column:NumOfDependence"`
	NTF                  float64   `gorm:"column:NTF"`
	OTRPrice             float64   `gorm:"column:OTRPrice"`
	LegalZipCode         string    `gorm:"type:varchar(5);column:LegalZipCode"`
	CompanyZipCode       string    `gorm:"type:varchar(5);column:CompanyZipCode"`
	Tenor                int       `gorm:"column:Tenor"`
	ManufacturingYear    string    `gorm:"type:varchar(4);column:ManufacturingYear"`
	ProfessionID         string    `gorm:"type:varchar(10);column:ProfessionID"`
	HomeStatus           string    `gorm:"type:varchar(2);column:HomeStatus"`
	CreatedAt            time.Time `gorm:"column:created_at"`
}

func (c *DupcheckRejectionNokaNosin) TableName() string {
	return "dupcheck_rejection_nokanosin"
}

type DupcheckRejectionPMK struct {
	RejectAttempt int    `gorm:"column:reject_attempt"`
	Date          string `gorm:"column:date"`
	RejectPMK     int    `gorm:"column:reject_pmk"`
	RejectDSR     int    `gorm:"column:reject_dsr"`
}

func (c *DupcheckRejectionPMK) TableName() string {
	return "dupcheck_rejection_pmk"
}

type TrxApiLog struct {
	ProspectID  string    `gorm:"type:varchar(20);column:ProspectID"`
	Request     string    `gorm:"type:text;column:request"`
	DtmRequest  time.Time `gorm:"column:dtm_request"`
	Response    string    `gorm:"type:text;column:response"`
	DtmResponse time.Time `gorm:"column:dtm_response"`
	Type        string    `gorm:"type:varchar(100);column:type"`
	Timestamps  time.Time `gorm:"column:timestamp"`
}

func (c *TrxApiLog) TableName() string {
	return "trx_api_logs"
}

type DummyAgreementChassisNumber struct {
	IDNumber string `gorm:"type:varchar(50);column:id_number"`
	Response string `gorm:"type:text;column:response"`
	Note     string `gorm:"type:varchar(200);column:note"`
}

func (c *DummyAgreementChassisNumber) TableName() string {
	return "dummy_agreement_chassis_number"
}
