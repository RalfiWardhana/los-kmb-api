package entity

import "time"

type SaveData struct {
	Keys       string `gorm:"type:varchar(20);column:keys"`
	StartValue *int   `grom:"type:int;column:start_value"`
	EndValue   *int   `gorm:"type:int;column:end_value"`
}

type ConfigPbk struct {
	Data struct {
		Url      string `json:"endpoint"`
		UserName string `json:"username"`
		Password string `json:"password"`
		Facility string `json:"facility"`
	}
}

type DupCheck struct {
	BirthDate         string `json:"birth_date"`
	IDNumber          string `json:"id_number"`
	LegalName         string `json:"legal_name"`
	SurgateMotherName string `json:"surgate_mother_name"`
	TransactionID     string `json:"transaction_id"`
}

type DummyColumn struct {
	NoKTP string `gorm:"type:varchar(20);column:NoKTP"`
	Value string `gorm:"column:Value"`
}

type DummyPBK struct {
	IDNumber string `gorm:"type:varchar(20);column:IDNumber"`
	Response string `gorm:"column:response"`
}

type ApiDupcheckKmb struct {
	ProspectID string      `gorm:"type:varchar(50);column:ProspectID"`
	RequestID  string      `gorm:"type:varchar(100);column:RequestID;primaryKey"`
	Request    string      `gorm:"type:text;column:Request"`
	Code       interface{} `gorm:"type:varchar(50);column:Code"`
	Decision   string      `gorm:"type:varchar(50);column:Decision"`
	Reason     string      `gorm:"type:varchar(200);column:Reason"`
	DtmRequest time.Time   `gorm:"column:DtmRequest"`
	Timestamp  time.Time   `gorm:"column:Timestamp"`
}

func (c *ApiDupcheckKmb) TableName() string {
	return "api_dupcheck_kmb"
}

type ApiDupcheckKmbUpdate struct {
	RequestID              string      `gorm:"type:varchar(100);column:RequestID;primaryKey"`
	ProspectID             string      `gorm:"type:varchar(100);column:ProspectID"`
	ResultDupcheckKonsumen interface{} `gorm:"type:text;column:ResultDupcheckKonsumen"`
	ResultDupcheckPasangan interface{} `gorm:"type:text;column:ResultDupcheckPasangan"`
	ResultKreditmu         interface{} `gorm:"type:text;column:ResultKreditmu"`
	ResultPefindo          interface{} `gorm:"type:text;column:ResultPefindo"`
	Response               interface{} `gorm:"type:text;column:Response"`
	CustomerType           interface{} `gorm:"type:text;column:CustomerType"`
	DtmResponse            time.Time   `gorm:"column:DtmResponse"`
	Code                   interface{} `gorm:"type:varchar(50);column:Code"`
	Decision               string      `gorm:"type:varchar(50);column:Decision"`
	Reason                 string      `gorm:"type:varchar(200);column:Reason"`
	Timestamp              time.Time   `gorm:"column:Timestamp"`
	PefindoID              interface{} `gorm:"column:PefindoID"`
	PefindoIDSpouse        interface{} `gorm:"column:PefindoIDSpouse"`
	PefindoScore           interface{} `gorm:"column:PefindoScore"`
}

func (c *ApiDupcheckKmbUpdate) TableName() string {
	return "api_dupcheck_kmb"
}

type ProfessionGroup struct {
	ID        string    `gorm:"type:varchar(40);column:id"`
	Prefix    string    `gorm:"type:varchar(50);column:prefix"`
	Name      string    `gorm:"type:varchar(50);column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

type RangeBranchDp struct {
	ID         string `gorm:"type:varchar(50);column:id"`
	Name       string `gorm:"type:varchar(20);column:name"`
	RangeStart int    `gorm:"column:range_start"`
	RangeEnd   int    `gorm:"column:range_end"`
	CreatedAt  string `gorm:"column:created_at"`
}

type BranchDp struct {
	Branch          string  `gorm:"type:varchar(10);column:branch"`
	CustomerStatus  string  `gorm:"type:varchar(10);column:customer_status"`
	ProfessionGroup string  `gorm:"type:varchar(20);column:profession_group"`
	MinimalDpName   string  `gorm:"type:varchar(10);column:minimal_dp_name"`
	MinimalDpValue  float64 `gorm:"column:minimal_dp_value"`
}

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

type ApiElaborateKmb struct {
	ProspectID       string    `gorm:"type:varchar(50);column:ProspectID"`
	RequestID        string    `gorm:"type:varchar(100);column:RequestID;primaryKey"`
	Request          string    `gorm:"type:text;column:Request"`
	Code             int       `gorm:"type:varchar(50);column:Code"`
	Decision         string    `gorm:"type:varchar(50);column:Decision"`
	Reason           string    `gorm:"type:varchar(200);column:Reason"`
	DtmRequest       time.Time `gorm:"column:DtmRequest"`
	IsMapping        int       `gorm:"column:IsMapping"`
	MappingParameter string    `gorm:"type:text;column:MappingParameter"`
	Timestamp        time.Time `gorm:"column:Timestamp"`
}

func (c *ApiElaborateKmb) TableName() string {
	return "api_elaborate_scheme"
}

type ApiElaborateKmbUpdate struct {
	ProspectID       string      `gorm:"type:varchar(50);column:ProspectID"`
	RequestID        string      `gorm:"type:varchar(100);column:RequestID;primaryKey"`
	Response         interface{} `gorm:"type:text;column:Response"`
	Code             int         `gorm:"type:varchar(50);column:Code"`
	Decision         string      `gorm:"type:varchar(50);column:Decision"`
	Reason           string      `gorm:"type:varchar(200);column:Reason"`
	DtmResponse      time.Time   `gorm:"column:DtmResponse"`
	IsMapping        int         `gorm:"column:IsMapping"`
	MappingParameter string      `gorm:"type:text;column:MappingParameter"`
	Timestamp        time.Time   `gorm:"column:Timestamp"`
}

func (c *ApiElaborateKmbUpdate) TableName() string {
	return "api_elaborate_scheme"
}

type ResultElaborate struct {
	BranchID       string `gorm:"type:varchar(10);column:branch_id"`
	CustomerStatus string `gorm:"type:varchar(10);column:customer_status"`
	BPKBNameType   int    `gorm:"column:bpkb_name_type"`
	Cluster        string `gorm:"type:varchar(20);column:cluster"`
	Decision       string `gorm:"type:varchar(10);column:decision"`
	LTV            int    `gorm:"type:int;column:ltv_start"`
}

type ClusterBranch struct {
	BranchID       string `gorm:"type:varchar(10);column:branch_id"`
	CustomerStatus string `gorm:"type:varchar(10);column:customer_status"`
	BPKBNameType   int    `gorm:"column:bpkb_name_type"`
	Cluster        string `gorm:"type:varchar(20);column:cluster"`
}

type MappingElaborateScheme struct {
	ResultPefindo  string  `gorm:"type:varchar(10);column:result_pefindo"`
	BranchID       string  `gorm:"type:varchar(10);column:branch_id"`
	CustomerStatus string  `gorm:"type:varchar(10);column:customer_status"`
	BPKBNameType   int     `gorm:"column:bpkb_name_type"`
	Cluster        string  `gorm:"type:varchar(20);column:cluster"`
	TotalBakiDebet int     `gorm:"column:total_baki_debet"`
	Tenor          int     `gorm:"column:tenor"`
	AgeVehicle     string  `gorm:"type:varchar(5);column:age_vehicle"`
	LTV            float64 `gorm:"column:ltv"`
	Decision       string  `gorm:"type:varchar(10);column:decision"`
}

func (c *MappingElaborateScheme) TableName() string {
	return "kmb_mapping_elaborate_scheme"
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

type VerificationFaceCompare struct {
	ID             string      `gorm:"column:id;primary_key:true"`
	CustomerID     int         `gorm:"column:customer_id"`
	ResultGetPhoto interface{} `gorm:"column:result_get_photo"`
	ResultFacePlus interface{} `gorm:"column:result_faceplus"`
	ResultASLIRI   interface{} `gorm:"column:result_asliri"`
	Decision       string      `gorm:"column:decision"`
	Result         interface{} `gorm:"column:result"`
	CreatedAt      time.Time   `gorm:"column:created_at;"`
	UpdatedAt      time.Time   `gorm:"column:updated_at"`
}

func (c *VerificationFaceCompare) TableName() string {
	return "verification_face_compare"
}

type DataInquiry struct {
	ProspectID    string    `gorm:"type:varchar(20);column:ProspectID"`
	IDNumber      string    `gorm:"type:varchar(50);column:IDNumber"`
	LegalName     string    `gorm:"type:varchar(100);column:LegalName"`
	FinalApproval int       `gorm:"column:final_approval"`
	DtmUpd        time.Time `gorm:"column:DtmUpd"`
	RejectDSR     int       `gorm:"column:reject_dsr"`
}

func (c *DataInquiry) TableName() string {
	return "data_inquiry"
}

type AsliriConfig struct {
	Data struct {
		KMB struct {
			AsliriActive bool `json:"asliri_service_active"`
			AsliriPhoto  int  `json:"asliri_photo_threshold"`
			AsliriName   int  `json:"asliri_name_threshold"`
			AsliriPDOB   int  `json:"asliri_pdob_threshold"`
		} `json:"kmb"`
	} `json:"data"`
}

type TrxMaster struct {
	ProspectID        string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	BranchID          string    `gorm:"type:varchar(5);column:BranchID"`
	TransactionType   *string   `gorm:"type:varchar(30);column:transaction_type"`
	ApplicationSource string    `gorm:"type:varchar(3);column:application_source"`
	MerchantID        *string   `gorm:"type:varchar(20);column:merchant_id"`
	MerchantName      *string   `gorm:"type:varchar(100);column:merchant_name"`
	Channel           string    `gorm:"type:varchar(5);column:channel"`
	Lob               string    `gorm:"type:varchar(5);column:lob"`
	OrderAt           string    `gorm:"type:varchar(30);column:order_at"`
	IncomingSource    string    `gorm:"type:varchar(10);column:incoming_source"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (c *TrxMaster) TableName() string {
	return "trx_master"
}

type CustomerAddress struct {
	ProspectID string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	Type       string    `gorm:"type:varchar(20);column:Type;primary_key:true"`
	Address    string    `gorm:"type:varchar(250);column:Address"`
	Rt         string    `gorm:"type:varchar(3);column:RT"`
	Rw         string    `gorm:"type:varchar(3);column:RW"`
	Kelurahan  string    `gorm:"type:varchar(30);column:Kelurahan"`
	Kecamatan  string    `gorm:"type:varchar(30);column:Kecamatan"`
	City       string    `gorm:"type:varchar(30);column:City"`
	ZipCode    string    `gorm:"type:varchar(5);column:ZipCode"`
	AreaPhone  string    `gorm:"type:varchar(5);column:AreaPhone"`
	Phone      string    `gorm:"type:varchar(20);column:Phone"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (c *CustomerAddress) TableName() string {
	return "trx_customer_address"
}

type CustomerPersonal struct {
	ProspectID                 string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true" json:"-"`
	IDType                     string      `gorm:"type:varchar(30);column:IDType" json:"-"`
	IDNumber                   string      `gorm:"type:varchar(100);column:IDNumber" json:"id_number"`
	IDTypeIssueDate            interface{} `gorm:"column:IDTypeIssuedDate" json:"-"`
	ExpiredDate                interface{} `gorm:"column:ExpiredDate" json:"-"`
	LegalName                  string      `gorm:"type:varchar(100);column:LegalName" json:"legal_name"`
	FullName                   string      `gorm:"type:varchar(100);column:FullName" json:"-"`
	BirthPlace                 string      `gorm:"type:varchar(100);column:BirthPlace" json:"birth_place"`
	BirthDate                  time.Time   `gorm:"column:BirthDate" json:"birth_date"`
	SurgateMotherName          string      `gorm:"type:varchar(100);column:SurgateMotherName" json:"surgate_mother_name"`
	Gender                     string      `gorm:"type:varchar(10);column:Gender" json:"gender"`
	PersonalNPWP               *string     `gorm:"type:varchar(255);column:PersonalNPWP" json:"-"`
	MobilePhone                string      `gorm:"type:varchar(14);column:MobilePhone" json:"mobile_phone"`
	Email                      string      `gorm:"type:varchar(100);column:Email" json:"email"`
	HomeStatus                 string      `gorm:"type:varchar(20);column:HomeStatus" json:"home_status"`
	StaySinceYear              string      `gorm:"type:varchar(10);column:StaySinceYear" json:"stay_since_year"`
	StaySinceMonth             string      `gorm:"type:varchar(10);column:StaySinceMonth" json:"stay_since_month"`
	Education                  string      `gorm:"type:varchar(50);column:Education" json:"education"`
	MaritalStatus              string      `gorm:"type:varchar(10);column:MaritalStatus" json:"marital_status"`
	NumOfDependence            int         `gorm:"column:NumOfDependence" json:"num_of_dependence"`
	LivingCostAmount           float64     `gorm:"column:LivingCostAmount" json:"-"`
	Religion                   string      `gorm:"type:varchar(30);column:Religion" json:"-"`
	CreatedAt                  time.Time   `gorm:"column:created_at" json:"-"`
	ExtCompanyPhone            *string     `gorm:"type:varchar(4);column:ExtCompanyPhone" json:"ext_company_phone"`
	SourceOtherIncome          *string     `gorm:"type:varchar(30);column:SourceOtherIncome" json:"source_other_income"`
	JobStatus                  string      `gorm:"type:varchar(10);column:job_status" json:"-"`
	EmergencyOfficeAreaPhone   string      `gorm:"type:varchar(4);column:EmergencyOfficeAreaPhone" json:"-"`
	EmergencyOfficePhone       string      `gorm:"type:varchar(20);column:EmergencyOfficePhone" json:"-"`
	PersonalCustomerType       string      `gorm:"type:varchar(20);column:PersonalCustomerType" json:"-"`
	Nationality                string      `gorm:"type:varchar(40);column:Nationality" json:"-"`
	WNACountry                 string      `gorm:"type:varchar(40);column:WNACountry" json:"-"`
	HomeLocation               string      `gorm:"type:varchar(10);column:HomeLocation" json:"-"`
	CustomerGroup              string      `gorm:"type:varchar(10);column:CustomerGroup" json:"-"`
	KKNo                       string      `gorm:"type:varchar(20);column:KKNo" json:"-"`
	BankID                     string      `gorm:"type:varchar(10);column:BankID" json:"-"`
	AccountNo                  string      `gorm:"type:varchar(20);column:AccountNo" json:"-"`
	AccountName                string      `gorm:"type:varchar(100);column:AccountName" json:"-"`
	Counterpart                int         `gorm:"column:Counterpart" json:"-"`
	DebtBusinessScale          string      `gorm:"type:varchar(50);column:DebtBusinessScale" json:"-"`
	DebtGroup                  string      `gorm:"type:varchar(50);column:DebtGroup" json:"-"`
	IsAffiliateWithPP          string      `gorm:"type:varchar(50);column:IsAffiliateWithPP" json:"-"`
	AgreetoAcceptOtherOffering int         `gorm:"column:AgreetoAcceptOtherOffering" json:"-"`
	DataType                   string      `gorm:"type:varchar(30);column:DataType" json:"-"`
	Status                     string      `gorm:"type:varchar(30);column:Status" json:"-"`
	IsPV                       *int        `gorm:"column:IsPV" json:"-"`
	IsRCA                      *int        `gorm:"column:IsRCA" json:"-"`
	CustomerID                 string      `gorm:"type:varchar(20);column:CustomerID" json:"-"`
	CustomerStatus             string      `gorm:"type:varchar(10);column:CustomerStatus" json:"customer_status"`
	SurveyResult               interface{} `gorm:"type:varchar(255);column:SurveyResult" json:"-"`
}

func (c *CustomerPersonal) TableName() string {
	return "trx_customer_personal"
}

type TrxMetadata struct {
	ProspectID   string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	CustomerIp   string    `gorm:"type:varchar(15);column:customer_ip"`
	CustomerLat  string    `gorm:"type:varchar(10);column:customer_lat"`
	CustomerLong string    `gorm:"type:varchar(10);column:customer_long"`
	CallbackUrl  string    `gorm:"type:varchar(250);column:callback_url"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (c *TrxMetadata) TableName() string {
	return "trx_metadata"
}

type CustomerPhoto struct {
	ProspectID string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true" json:"-"`
	PhotoID    string    `gorm:"type:varchar(50);column:photo_id;primary_key:true" json:"photo_id"`
	Url        string    `gorm:"type:varchar(250);column:url" json:"photo_url"`
	Width      string    `gorm:"type:varchar(10);column:width" json:"-"`
	Height     string    `gorm:"type:varchar(10);column:height" json:"-"`
	Position   string    `gorm:"type:varchar(3);column:position" json:"-"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"-"`
}

func (c *CustomerPhoto) TableName() string {
	return "trx_customer_photo"
}

type CustomerEmployment struct {
	ProspectID            string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true" json:"-"`
	ProfessionID          string    `gorm:"type:varchar(10);column:ProfessionID" json:"profession_id"`
	JobType               string    `gorm:"type:varchar(10);column:JobType" json:"job_type_id"`
	JobPosition           string    `gorm:"type:varchar(10);column:JobPosition" json:"job_position"`
	CompanyName           string    `gorm:"type:varchar(50);column:CompanyName" json:"company_name"`
	IndustryTypeID        string    `gorm:"type:varchar(10);column:IndustryTypeID" json:"industry"`
	EmploymentSinceYear   string    `gorm:"type:varchar(4);column:EmploymentSinceYear" json:"employment_since_year"`
	EmploymentSinceMonth  string    `gorm:"type:varchar(2);column:EmploymentSinceMonth" json:"employment_since_month"`
	MonthlyFixedIncome    float64   `gorm:"column:MonthlyFixedIncome" json:"monthly_fixed_income"`
	MonthlyVariableIncome float64   `gorm:"column:MonthlyVariableIncome" json:"monthly_variable_income"`
	SpouseIncome          float64   `gorm:"column:SpouseIncome" json:"spouse_income"`
	CreatedAt             time.Time `gorm:"column:created_at" json:"-"`
}

func (c *CustomerEmployment) TableName() string {
	return "trx_customer_employment"
}

type TrxApk struct {
	ProspectID                  string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	Tenor                       *int      `gorm:"column:Tenor"`
	ProductOfferingID           string    `gorm:"type:varchar(30);column:ProductOfferingID"`
	ProductID                   string    `gorm:"type:varchar(30);column:ProductID"`
	NTF                         float64   `gorm:"column:NTF"`
	AF                          float64   `gorm:"column:AF"`
	OTR                         float64   `gorm:"column:OTR"`
	DPAmount                    float64   `gorm:"column:DPAmount"`
	Discount                    float64   `gorm:"column:Discount"`
	InsuranceFee                float64   `gorm:"column:InsuranceFee"`
	InstallmentAmount           float64   `gorm:"column:InstallmentAmount"`
	FirstInstallment            string    `gorm:"column:FirstInstallment"`
	AdminFee                    float64   `gorm:"column:AdminFee"`
	AoID                        string    `gorm:"type:varchar(20);column:AOID"`
	CreatedAt                   time.Time `gorm:"column:created_at"`
	OtherFee                    float64   `gorm:"column:OtherFee"`
	PercentDP                   float64   `gorm:"column:percent_dp"`
	AssetInsuranceFee           float64   `gorm:"column:AssetInsuranceFee"`
	LifeInsuranceFee            float64   `gorm:"column:LifeInsuranceFee"`
	FidusiaFee                  float64   `gorm:"column:fidusia_fee"`
	InterestRate                float64   `gorm:"column:interest_rate"`
	InsuranceRate               float64   `gorm:"column:insurance_rate"`
	FirstPayment                float64   `gorm:"column:first_payment"`
	InsuranceAmount             float64   `gorm:"column:insurance_amount"`
	InterestAmount              float64   `gorm:"column:interest_amount"`
	FirstPaymentDate            time.Time `gorm:"column:first_payment_date"`
	PaymentMethod               string    `gorm:"type:varchar(10);column:payment_method"`
	SurveyFee                   float64   `gorm:"column:survey_fee"`
	IsFidusiaCovered            string    `gorm:"type:varchar(1);column:is_fidusia_covered"`
	ProvisionFee                float64   `gorm:"column:provision_fee"`
	InsAssetPaidBy              string    `gorm:"type:varchar(10);column:ins_asset_paid_by"`
	InsAssetPeriod              string    `gorm:"type:varchar(10);column:ins_asset_period"`
	EffectiveRate               float64   `gorm:"column:effective_rate"`
	SalesmanID                  string    `gorm:"type:varchar(20);column:salesman_id"`
	SupplierBankAccountID       string    `gorm:"type:varchar(20);column:supplier_bank_account_id"`
	LifeInsuranceCoyBranchID    string    `gorm:"type:varchar(20);column:life_insurance_coy_branch_id"`
	LifeInsuranceAmountCoverage float64   `gorm:"column:life_insurance_amount_coverage"`
	CommisionSubsidi            float64   `gorm:"column:commision_subsidi"`
	ProductOfferingDesc         string    `gorm:"column:product_offering_desc"`
	Dealer                      string    `gorm:"column:dealer"`
	LoanAmount                  float64   `gorm:"column:loan_amount"`
	FinancePurpose              string    `gorm:"type:varchar(30);column:FinancePurpose"`
	NTFAkumulasi                float64   `gorm:"column:NTFAkumulasi"`
	NTFOtherAmount              float64   `gorm:"column:NTFOtherAmount"`
	NTFOtherAmountSpouse        float64   `gorm:"column:NTFOtherAmountSpouse"`
	NTFOtherAmountDetail        string    `gorm:"column:NTFOtherDetail"`
	NTFConfinsAmount            float64   `gorm:"column:NTFConfinsAmount"`
	NTFConfins                  float64   `gorm:"column:NTFConfins"`
	NTFTopup                    float64   `gorm:"column:NTFTopup"`
	InstallmentOther            float64   `gorm:"column:InstallmentOther"`
	InstallmentOtherSpouse      float64   `gorm:"column:InstallmentOtherSpouse"`
	InstallmentOtherDetail      string    `gorm:"column:InstallmentOtherDetail"`
}

func (c *TrxApk) TableName() string {
	return "trx_apk"
}

type TrxSurveyor struct {
	ProspectID   string    `gorm:"type:varchar(20);column:ProspectID" json:"-"`
	Destination  string    `gorm:"type:varchar(10);column:destination" json:"destination"`
	RequestDate  time.Time `gorm:"column:request_date" json:"request_date"`
	RequestInfo  *string   `gorm:"type:varchar(255);column:request_info" json:"-"`
	AssignDate   time.Time `gorm:"column:assign_date" json:"assign_date"`
	SurveyorName string    `gorm:"type:varchar(100);column:surveyor_name" json:"surveyor_name"`
	ResultDate   time.Time `gorm:"column:result_date" json:"result_date"`
	Status       string    `gorm:"type:varchar(10);column:status" json:"status"`
	SurveyorNote *string   `gorm:"type:text;column:surveyor_note" json:"surveyor_note"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"-"`
}

func (c *TrxSurveyor) TableName() string {
	return "trx_surveyor"
}

type CustomerOmset struct {
	ProspectID        string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	SeqNo             int       `gorm:"type:smallint;column:SeqNo;primary_key:true"`
	MonthlyOmsetYear  string    `gorm:"type:varchar(10);column:MonthlyOmsetYear"`
	MonthlyOmsetMonth string    `gorm:"type:varchar(10);column:MonthlyOmsetMonth"`
	MonthlyOmset      float64   `gorm:"column:MonthlyOmset"`
	CreatedAt         time.Time `gorm:"column:created_at"`
}

func (c *CustomerOmset) TableName() string {
	return "trx_customer_omset"
}

type TrxStatus struct {
	ProspectID     string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	StatusProcess  string      `gorm:"type:varchar(3);column:status_process"`
	Activity       string      `gorm:"type:varchar(4);column:activity"`
	Decision       string      `gorm:"type:varchar(3);column:decision"`
	RuleCode       interface{} `gorm:"type:varchar(4);column:rule_code"`
	SourceDecision string      `gorm:"type:varchar(3);column:source_decision"`
	NextStep       interface{} `gorm:"type:varchar(3);column:next_step"`
	CreatedAt      time.Time   `gorm:"column:created_at"`
	Reason         string      `gorm:"type:varchar(255);column:reason"`
}

func (c *TrxStatus) TableName() string {
	return "trx_status"
}

type TrxItem struct {
	ProspectID                   string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	CategoryID                   string    `gorm:"type:varchar(100);column:category_id"`
	SupplierID                   string    `gorm:"type:varchar(100);column:supplier_id"`
	Qty                          int       `gorm:"column:qty"`
	AssetCode                    string    `gorm:"type:varchar(200);column:asset_code"`
	ManufactureYear              string    `gorm:"type:varchar(4);column:manufacture_year"`
	BPKBName                     string    `gorm:"type:varchar(2);column:bpkb_name"`
	OwnerAsset                   string    `gorm:"type:varchar(100);column:owner_asset"`
	LicensePlate                 string    `gorm:"type:varchar(20);column:license_plate"`
	Color                        string    `gorm:"type:varchar(50);column:color"`
	EngineNo                     string    `gorm:"type:varchar(30);column:engine_number"`
	ChassisNo                    string    `gorm:"type:varchar(30);column:chassis_number"`
	AssetDescription             string    `gorm:"type:varchar(255);column:asset_description"`
	Pos                          string    `gorm:"type:varchar(10);column:pos"`
	Cc                           string    `gorm:"type:varchar(10);column:cc"`
	Condition                    string    `gorm:"type:varchar(10);column:condition"`
	Region                       string    `gorm:"type:varchar(10);column:region"`
	TaxDate                      time.Time `gorm:"column:tax_date"`
	STNKExpiredDate              time.Time `gorm:"column:stnk_expired_date"`
	AssetInsuranceAmountCoverage float64   `gorm:"column:AssetInsuranceAmountCoverage"`
	InsAssetInsuredBy            string    `gorm:"type:varchar(20);column:InsAssetInsuredBy"`
	InsuranceCoyBranchID         string    `gorm:"type:varchar(10);column:InsuranceCoyBranchID"`
	CoverageType                 string    `gorm:"type:varchar(10);column:CoverageType"`
	OwnerKTP                     string    `gorm:"type:varchar(20);column:owner_ktp"`
	AssetUsage                   string    `gorm:"type:varchar(10);column:asset_usage"`
	Brand                        string    `gorm:"type:varchar(255);column:brand"`
	CreatedAt                    time.Time `gorm:"column:created_at"`
}

func (c *TrxItem) TableName() string {
	return "trx_item"
}

type TrxInfoAgent struct {
	ProspectID string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	NIK        string    `gorm:"type:varchar(20);column:nik"`
	Name       string    `gorm:"type:varchar(50);column:name"`
	Info       string    `gorm:"type:varchar(50);column:info"`
	RecomDate  string    `gorm:"type:varchar(50);column:recom_date"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (c *TrxInfoAgent) TableName() string {
	return "trx_info_agent"
}

type CustomerSpouse struct {
	ProspectID           string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true" json:"-"`
	IDNumber             string      `gorm:"type:varchar(40);column:IDNumber" json:"spouse_id_number"`
	FullName             string      `gorm:"type:varchar(200);column:FullName" json:"-"`
	LegalName            string      `gorm:"type:varchar(200);column:LegalName" json:"spouse_legal_name"`
	BirthPlace           string      `gorm:"type:varchar(30);column:BirthPlace" json:"-"`
	BirthDate            time.Time   `gorm:"column:BirthDate" json:"-"`
	SurgateMotherName    string      `gorm:"type:varchar(100);column:SurgateMotherName" json:"-"`
	Gender               string      `gorm:"type:varchar(1);column:Gender" json:"-"`
	CompanyPhone         interface{} `gorm:"type:varchar(20);column:CompanyPhone" json:"spouse_company_phone"`
	CompanyName          interface{} `gorm:"type:varchar(20);column:CompanyName" json:"spouse_company_name"`
	MobilePhone          string      `gorm:"type:varchar(20);column:MobilePhone" json:"spouse_mobile_phone"`
	EmploymentSinceYear  interface{} `gorm:"type:varchar(4);column:EmploymentSinceYear" json:"-"`
	EmploymentSinceMonth interface{} `gorm:"type:varchar(2);column:EmploymentSinceMonth" json:"-"`
	ProfessionID         interface{} `gorm:"type:varchar(10);column:ProfessionID" json:"spouse_profession"`
	JobType              string      `gorm:"type:varchar(10);column:JobType" json:"-"`
	JobPosition          interface{} `gorm:"type:varchar(10);column:JobPosition" json:"-"`
	Email                string      `gorm:"type:varchar(100);column:Email" json:"-"`
	PersonalNPWP         string      `gorm:"type:varchar(50);column:PersonalNPWP" json:"-"`
	Education            string      `gorm:"type:varchar(20);column:Education" json:"-"`
	CreatedAt            time.Time   `gorm:"column:created_at" json:"-"`
}

func (c *CustomerSpouse) TableName() string {
	return "trx_customer_spouse"
}

type CustomerEmcon struct {
	ProspectID           string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true" json:"-"`
	Name                 string    `gorm:"type:varchar(200);column:Name" json:"emcon_name"`
	Relationship         string    `gorm:"type:varchar(10);column:Relationship" json:"relationship"`
	MobilePhone          string    `gorm:"type:varchar(20);column:MobilePhone" json:"emcon_mobile_phone"`
	CreatedAt            time.Time `gorm:"column:created_at" json:"-"`
	EmconVerified        string    `gorm:"type:varchar(1);column:EmconVerified" json:"-"`
	VerifyBy             string    `gorm:"type:varchar(5);column:VerifyBy" json:"-"`
	KnownCustomerJob     string    `gorm:"type:varchar(1);column:KnownCustomerJob" json:"-"`
	KnownCustomerAddress string    `gorm:"type:varchar(1);column:KnownCustomerAddress" json:"-"`
	VerificationWith     string    `gorm:"type:varchar(100);column:VerificationWith" json:"-"`
}

func (c *CustomerEmcon) TableName() string {
	return "trx_customer_emcon"
}

type FilteringKMB struct {
	ProspectID                      string      `gorm:"column:prospect_id;type:varchar(20)" json:"prospect_id"`
	RequestID                       interface{} `gorm:"column:request_id;type:varchar(100)" json:"request_id"`
	BpkbName                        string      `gorm:"column:bpkb_name;type:varchar(2)" json:"bpkb_name"`
	BranchID                        string      `gorm:"column:branch_id;type:varchar(5)" json:"branch_id"`
	Decision                        string      `gorm:"column:decision;type:varchar(20)" json:"decision"`
	CustomerStatus                  interface{} `gorm:"column:customer_status;type:varchar(5)" json:"customer_status"`
	CustomerSegment                 interface{} `gorm:"column:customer_segment;type:varchar(20)" json:"customer_segment"`
	CustomerID                      interface{} `gorm:"column:customer_id;type:varchar(20)" json:"customer_id"`
	IsBlacklist                     int         `gorm:"column:is_blacklist" json:"is_blacklist"`
	NextProcess                     int         `gorm:"column:next_process" json:"next_process"`
	MaxOverdueBiro                  interface{} `gorm:"column:max_overdue_biro" json:"max_overdue_biro"`
	MaxOverdueLast12monthsBiro      interface{} `gorm:"column:max_overdue_last12months_biro" json:"max_overdue_last12months_biro"`
	IsWoContractBiro                interface{} `gorm:"column:is_wo_contract_biro" json:"is_wo_contract_biro"`
	IsWoWithCollateralBiro          interface{} `gorm:"column:is_wo_with_collateral_biro" json:"is_wo_with_collateral_biro"`
	TotalInstallmentAmountBiro      interface{} `gorm:"column:total_installment_amount_biro" json:"total_installment_amount_biro"`
	TotalBakiDebetNonCollateralBiro interface{} `gorm:"column:total_baki_debet_non_collateral_biro" json:"total_baki_debet_non_collateral_biro"`
	ScoreBiro                       interface{} `gorm:"column:score_biro;type:varchar(20)" json:"score_biro"`
	Cluster                         interface{} `gorm:"column:cluster;type:varchar(20)" json:"cluster"`
	Reason                          interface{} `gorm:"column:reason;type:varchar(250)" json:"reason"`
	CreatedAt                       time.Time   `gorm:"column:created_at" json:"created_at"`
}

func (c *FilteringKMB) TableName() string {
	return "trx_filtering"
}

type TrxDetail struct {
	ProspectID     string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	StatusProcess  string      `gorm:"type:varchar(3);column:status_process"`
	Activity       string      `gorm:"type:varchar(4);column:activity"`
	Decision       string      `gorm:"type:varchar(3);column:decision"`
	RuleCode       interface{} `gorm:"type:varchar(4);column:rule_code"`
	SourceDecision string      `gorm:"type:varchar(3);column:source_decision"`
	NextStep       interface{} `gorm:"type:varchar(3);column:next_step"`
	Type           interface{} `gorm:"type:varchar(3);column:type"`
	Info           interface{} `gorm:"type:text;column:info"`
	CreatedBy      string      `gorm:"type:varchar(100);column:created_by"`
	CreatedAt      time.Time   `gorm:"column:created_at"`
}

func (c *TrxDetail) TableName() string {
	return "trx_details"
}

type TrxDetailBiro struct {
	ProspectID             string      `gorm:"type:varchar(20);column:prospect_id;primary_key:true"`
	Subject                string      `gorm:"type:varchar(10);column:subject"`
	Source                 string      `gorm:"type:varchar(5);column:source"`
	BiroID                 string      `gorm:"type:varchar(20);column:biro_id"`
	Score                  string      `gorm:"type:varchar(20);column:score"`
	MaxOverdue             interface{} `gorm:"column:max_overdue"`
	MaxOverdueLast12months interface{} `gorm:"column:max_overdue_last12months"`
	InstallmentAmount      interface{} `gorm:"column:installment_amount"`
	WoContract             int         `gorm:"column:wo_contract"`
	WoWithCollateral       int         `gorm:"column:wo_with_collateral"`
	BakiDebetNonCollateral float64     `gorm:"column:baki_debet_non_collateral"`
	UrlPdfReport           string      `gorm:"type:varchar(200);column:url_pdf_report"`
	CreatedAt              time.Time   `gorm:"column:created_at"`
}

func (c *TrxDetailBiro) TableName() string {
	return "trx_detail_biro"
}

type MasterMappingCluster struct {
	BranchID       string    `gorm:"column:branch_id"`
	CustomerStatus string    `gorm:"column:customer_status"`
	BpkbNameType   int       `gorm:"column:bpkb_name_type"`
	Cluster        string    `gorm:"column:cluster"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (c *MasterMappingCluster) TableName() string {
	return "m_mapping_cluster"
}

type LogOrchestrator struct {
	ID           string    `gorm:"type:varchar(50);column:id;primary_key:true"`
	ProspectID   string    `gorm:"type:varchar(20);column:ProspectID"`
	Owner        string    `gorm:"type:varchar(10);column:owner"`
	Header       string    `gorm:"type:text;column:header"`
	Url          string    `gorm:"type:varchar(100);column:url"`
	Method       string    `gorm:"type:varchar(10);column:method"`
	RequestData  string    `gorm:"type:text;column:request_data"`
	ResponseData string    `gorm:"type:text;column:response_data"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (c *LogOrchestrator) TableName() string {
	return "log_orchestrators"
}

type TrxPrescreening struct {
	ProspectID string    `gorm:"column:ProspectID"`
	Decision   string    `gorm:"column:decision"`
	Reason     string    `gorm:"column:reason"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	CreatedBy  string    `gorm:"column:created_by"`
	DecisionBy string    `gorm:"column:decision_by"`
}

func (c *TrxPrescreening) TableName() string {
	return "trx_prescreening"
}

type TrxAkkk struct {
	ProspectID                       string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	SlikPlafond                      interface{} `gorm:"column:SlikPlafond"`
	SpouseSlikPlafond                interface{} `gorm:"column:SpouseSlikPlafond"`
	PefindoPlafond                   interface{} `gorm:"column:PefindoPlafond"`
	SpousePefindoPlafond             interface{} `gorm:"column:SpousePefindoPlafond"`
	SlikBakiDebet                    interface{} `gorm:"column:SlikBakiDebet"`
	SpouseSlikBakiDebet              interface{} `gorm:"column:SpouseSlikBakiDebet"`
	PefindoBakiDebet                 interface{} `gorm:"column:PefindoBakiDebet"`
	SpousePefindoBakiDebet           interface{} `gorm:"column:SpousePefindoBakiDebet"`
	SlikTotalFasilitasAktif          interface{} `gorm:"column:SlikTotalFasilitasAktif"`
	SpouseSlikTotalFasilitasAktif    interface{} `gorm:"column:SpouseSlikTotalFasilitasAktif"`
	WorstQualityStatus               interface{} `gorm:"type:varchar(50);column:WorstQualityStatus"`
	WorstQualityStatusMonth          interface{} `gorm:"type:varchar(50);column:WorstQualityStatusMonth"`
	LastQualityCredit                interface{} `gorm:"type:varchar(50);column:LastQualityCredit"`
	LastQualityCreditMonth           interface{} `gorm:"type:varchar(50);column:LastQualityCreditMonth"`
	SpouseWorstQualityStatus         interface{} `gorm:"type:varchar(50);column:SpouseWorstQualityStatus"`
	SpouseWorstQualityStatusMonth    interface{} `gorm:"type:varchar(50);column:SpouseWorstQualityStatusMonth"`
	SpouseLastQualityCredit          interface{} `gorm:"type:varchar(50);column:SpouseLastQualityCredit"`
	SpouseLastQualityCreditMonth     interface{} `gorm:"type:varchar(50);column:SpouseLastQualityCreditMonth"`
	PefindoTotalFasilitasAktif       interface{} `gorm:"column:PefindoTotalFasilitasAktif"`
	SpousePefindoTotalFasilitasAktif interface{} `gorm:"column:SpousePefindoTotalFasilitasAktif"`
	PefindoScore                     interface{} `gorm:"type:varchar(50);column:PefindoScore"`
	SpousePefindoScore               interface{} `gorm:"type:varchar(50);column:SpousePefindoScore"`
	LastMaxOVD                       interface{} `gorm:"column:LastMaxOVD"`
	SpouseLastMaxOVD                 interface{} `gorm:"column:SpouseLastMaxOVD"`
	AgreementStatus                  interface{} `gorm:"type:varchar(10);column:AgreementStatus"`
	TotalAgreementAktif              interface{} `gorm:"column:TotalAgreementAktif"`
	MaxOVDAgreementAktif             interface{} `gorm:"column:MaxOVDAgreementAktif"`
	LastMaxOVDAgreement              interface{} `gorm:"column:LastMaxOVDAgreement"`
	DSR                              interface{} `gorm:"column:DSR"`
	AsliriSimiliarity                interface{} `gorm:"column:AsliriSimiliarity"`
	AsliriReason                     interface{} `gorm:"type:varchar(30);column:AsliriReason"`
	FinancePurpose                   string      `gorm:"type:varchar(30);column:FinancePurpose"`
	CreatedAt                        time.Time   `gorm:"column:created_at"`
}

type MappingElaborateLTV struct {
	ID                  int    `gorm:"column:id"`
	ResultPefindo       string `gorm:"type:varchar(10);column:result_pefindo"`
	Cluster             string `gorm:"type:varchar(20);column:cluster"`
	TotalBakiDebetStart int    `gorm:"column:total_baki_debet_start"`
	TotalBakiDebetEnd   int    `gorm:"column:total_baki_debet_end"`
	TenorStart          int    `gorm:"column:tenor_start"`
	TenorEnd            int    `gorm:"column:tenor_end"`
	BPKBNameType        int    `gorm:"column:bpkb_name_type"`
	AgeVehicle          string `gorm:"type:varchar(5);column:age_vehicle"`
	LTV                 int    `gorm:"column:ltv"`
}

func (c *MappingElaborateLTV) TableName() string {
	return "m_mapping_elaborate_ltv"
}

type TrxElaborateLTV struct {
	ProspectID            string      `gorm:"column:prospect_id"`
	RequestID             interface{} `gorm:"column:request_id;type:varchar(100)" json:"request_id"`
	Tenor                 int         `gorm:"column:tenor"`
	ManufacturingYear     string      `gorm:"column:manufacturing_year"`
	MappingElaborateLTVID int         `gorm:"column:m_mapping_elaborate_ltv_id"`
	CreatedAt             time.Time   `gorm:"column:created_at"`
}

func (c *TrxElaborateLTV) TableName() string {
	return "trx_elaborate_ltv"
}

type SpIndustryTypeMaster struct {
	IndustryTypeID string `gorm:"column:IndustryTypeID"`
	Description    string `gorm:"column:Description"`
	IsActive       bool   `gorm:"column:IsActive"`
}

type InquiryPrescreening struct {
	CmoRecommendation int    `gorm:"column:CMORecommend"`
	Activity          string `gorm:"column:activity"`
	SourceDecision    string `gorm:"column:source_decision"`
	Decision          string `gorm:"column:decision"`
	Reason            string `gorm:"column:reason"`
	DecisionBy        string `gorm:"column:DecisionBy"`
	DecisionName      string `gorm:"column:DecisionName"`
	DecisionAt        string `gorm:"column:DecisionAt"`

	ProspectID     string `gorm:"column:ProspectID"`
	BranchName     string `gorm:"column:BranchName"`
	IncomingSource string `gorm:"column:incoming_source"`
	CreatedAt      string `gorm:"column:created_at"`
	OrderAt        string `gorm:"column:order_at"`

	CustomerStatus    string    `gorm:"column:customer_status"`
	IDNumber          string    `gorm:"column:IDNumber"`
	LegalName         string    `gorm:"column:LegalName"`
	BirthPlace        string    `gorm:"column:BirthPlace"`
	BirthDate         time.Time `gorm:"column:BirthDate"`
	SurgateMotherName string    `gorm:"column:SurgateMotherName"`
	Gender            string    `gorm:"column:Gender"`
	MobilePhone       string    `gorm:"column:MobilePhone"`
	Email             string    `gorm:"column:Email"`
	Education         string    `gorm:"column:Education"`
	MaritalStatus     string    `gorm:"column:MaritalStatus"`
	NumOfDependence   int       `gorm:"column:NumOfDependence"`
	HomeStatus        string    `gorm:"column:HomeStatus"`
	StaySinceMonth    string    `gorm:"column:StaySinceMonth"`
	StaySinceYear     string    `gorm:"column:StaySinceYear"`
	ExtCompanyPhone   *string   `gorm:"column:ExtCompanyPhone"`
	SourceOtherIncome *string   `gorm:"column:SourceOtherIncome"`

	Supplier           string  `gorm:"column:dealer"`
	ProductOfferingID  string  `gorm:"column:ProductOfferingID"`
	AssetType          string  `gorm:"column:AssetType"`
	AssetDescription   string  `gorm:"column:asset_description"`
	ManufacturingYear  string  `gorm:"column:manufacture_year"`
	Color              string  `gorm:"column:color"`
	ChassisNumber      string  `gorm:"column:chassis_number"`
	EngineNumber       string  `gorm:"column:engine_number"`
	InterestRate       float64 `gorm:"column:interest_rate"`
	InstallmentPeriod  int     `gorm:"column:InstallmentPeriod"`
	OTR                float64 `gorm:"column:OTR"`
	DPAmount           float64 `gorm:"column:DPAmount"`
	FinanceAmount      float64 `gorm:"column:FinanceAmount"`
	InterestAmount     float64 `gorm:"column:interest_amount"`
	InsuranceAmount    float64 `gorm:"column:insurance_amount"`
	AdminFee           float64 `gorm:"column:AdminFee"`
	ProvisionFee       float64 `gorm:"column:provision_fee"`
	NTF                float64 `gorm:"column:NTF"`
	Total              float64 `gorm:"column:Total"`
	MonthlyInstallment float64 `gorm:"column:MonthlyInstallment"`
	FirstInstallment   string  `gorm:"column:FirstInstallment"`

	ProfessionID          string  `gorm:"column:ProfessionID"`
	JobTypeID             string  `gorm:"column:JobType"`
	JobPosition           string  `gorm:"column:JobPosition"`
	CompanyName           string  `gorm:"column:CompanyName"`
	IndustryTypeID        string  `gorm:"column:IndustryTypeID"`
	EmploymentSinceYear   string  `gorm:"column:EmploymentSinceYear"`
	EmploymentSinceMonth  string  `gorm:"column:EmploymentSinceMonth"`
	MonthlyFixedIncome    float64 `gorm:"column:MonthlyFixedIncome"`
	MonthlyVariableIncome float64 `gorm:"column:MonthlyVariableIncome"`
	SpouseIncome          float64 `gorm:"column:SpouseIncome"`

	SpouseIDNumber     string `gorm:"column:SpouseIDNumber"`
	SpouseLegalName    string `gorm:"column:SpouseLegalName"`
	SpouseCompanyName  string `gorm:"column:SpouseCompanyName"`
	SpouseCompanyPhone string `gorm:"column:SpouseCompanyPhone"`
	SpouseMobilePhone  string `gorm:"column:SpouseMobilePhone"`
	SpouseProfession   string `gorm:"column:SpouseProfession"`

	EmconName        string `gorm:"column:EmconName"`
	Relationship     string `gorm:"column:Relationship"`
	EmconMobilePhone string `gorm:"column:EmconMobilePhone"`

	LegalAddress       string `gorm:"column:LegalAddress"`
	LegalRTRW          string `gorm:"column:LegalRTRW"`
	LegalKelurahan     string `gorm:"column:LegalKelurahan"`
	LegalKecamatan     string `gorm:"column:LegalKecamatan"`
	LegalZipCode       string `gorm:"column:LegalZipcode"`
	LegalCity          string `gorm:"column:LegalCity"`
	ResidenceAddress   string `gorm:"column:ResidenceAddress"`
	ResidenceRTRW      string `gorm:"column:ResidenceRTRW"`
	ResidenceKelurahan string `gorm:"column:ResidenceKelurahan"`
	ResidenceKecamatan string `gorm:"column:ResidenceKecamatan"`
	ResidenceZipCode   string `gorm:"column:ResidenceZipcode"`
	ResidenceCity      string `gorm:"column:ResidenceCity"`
	CompanyAddress     string `gorm:"column:CompanyAddress"`
	CompanyRTRW        string `gorm:"column:CompanyRTRW"`
	CompanyKelurahan   string `gorm:"column:CompanyKelurahan"`
	CompanyKecamatan   string `gorm:"column:CompanyKecamatan"`
	CompanyZipCode     string `gorm:"column:CompanyZipcode"`
	CompanyCity        string `gorm:"column:CompanyCity"`
	CompanyAreaPhone   string `gorm:"column:CompanyAreaPhone"`
	CompanyPhone       string `gorm:"column:CompanyPhone"`
	EmergencyAddress   string `gorm:"column:EmergencyAddress"`
	EmergencyRTRW      string `gorm:"column:EmergencyRTRW"`
	EmergencyKelurahan string `gorm:"column:EmergencyKelurahan"`
	EmergencyKecamatan string `gorm:"column:EmergencyKecamatan"`
	EmergencyZipcode   string `gorm:"column:EmergencyZipcode"`
	EmergencyCity      string `gorm:"column:EmergencyCity"`
	EmergencyAreaPhone string `gorm:"column:EmergencyAreaPhone"`
	EmergencyPhone     string `gorm:"column:EmergencyPhone"`
}

type ReasonMessage struct {
	ReasonID      string `gorm:"column:ReasonID" json:"reason_id"`
	Code          string `gorm:"column:Code" json:"code"`
	ReasonMessage string `gorm:"column:ReasonMessage" json:"reason_message"`
}

func (c *ReasonMessage) TableName() string {
	return "m_reason_message"
}

type InquiryData struct {
	Prescreening DataPrescreening   `json:"prescreening"`
	General      DataGeneral        `json:"general"`
	Personal     CustomerPersonal   `json:"personal"`
	Spouse       CustomerSpouse     `json:"spouse"`
	Employment   CustomerEmployment `json:"employment"`
	ItemApk      DataItemApk        `json:"item_apk"`
	Surveyor     []TrxSurveyor      `json:"surveyor"`
	Emcon        CustomerEmcon      `json:"emcon"`
	Address      DataAddress        `json:"address"`
	Photo        []CustomerPhoto    `json:"photo"`
}

type DataPrescreening struct {
	CmoRecommendation string `json:"cmo_recommendation"`
	ShowAction        bool   `json:"show_action"`
	Decision          string `gorm:"column:decision" json:"decision"`
	Reason            string `gorm:"column:reason" json:"reason"`
	DecisionBy        string `gorm:"column:DecisionBy" json:"decision_by"`
	DecisionName      string `gorm:"column:DecisionName" json:"decision_by_name"`
	DecisionAt        string `gorm:"column:DecisionAt" json:"decision_at"`
}

type DataGeneral struct {
	ProspectID     string `gorm:"column:ProspectID" json:"prospect_id"`
	BranchName     string `gorm:"column:BranchName" json:"branch_name"`
	IncomingSource string `gorm:"column:incoming_source" json:"incoming_source"`
	CreatedAt      string `gorm:"column:created_at" json:"created_at"`
	OrderAt        string `gorm:"column:order_at" json:"order_at"`
}

type DataItemApk struct {
	Supplier              string  `gorm:"column:dealer" json:"supplier"`
	ProductOfferingID     string  `gorm:"column:ProductOfferingID" json:"product_offering_id"`
	AssetDescription      string  `gorm:"column:asset_description" json:"asset_description"`
	AssetType             string  `gorm:"column:AssetType" json:"asset_type"`
	ManufacturingYear     string  `gorm:"column:manufacture_year" json:"manufacturing_year"`
	Color                 string  `gorm:"column:color" json:"color"`
	ChassisNumber         string  `gorm:"column:chassis_number" json:"chassis_number"`
	EngineNumber          string  `gorm:"column:engine_number" json:"engine_number"`
	InterestRate          float64 `gorm:"column:interest_rate" json:"interest_rate"`
	Tenor                 int     `gorm:"column:InstallmentPeriod" json:"installment_period"`
	OTR                   float64 `gorm:"column:OTR" json:"otr"`
	DPAmount              float64 `gorm:"column:DPAmount" json:"dp_amount"`
	AF                    float64 `gorm:"column:FinanceAmount" json:"finance_amount"`
	InterestAmount        float64 `gorm:"column:interest_amount" json:"interest_amount"`
	InsuranceAmount       float64 `gorm:"column:insurance_amount" json:"insurance_amount"`
	AdminFee              float64 `gorm:"column:AdminFee" json:"admin_fee"`
	ProvisionFee          float64 `gorm:"column:provision_fee" json:"provision_fee"`
	NTF                   float64 `gorm:"column:NTF" json:"ntf"`
	NTFPlusInterestAmount float64 `gorm:"column:Total" json:"total"`
	InstallmentAmount     float64 `gorm:"column:MonthlyInstallment" json:"monthly_installment"`
	FirstInstallment      string  `gorm:"column:FirstInstallment" json:"first_installment"`
}

type DataAddress struct {
	LegalAddress       string `gorm:"column:LegalAddress" json:"legal_address"`
	LegalRTRW          string `gorm:"column:LegalRTRW" json:"legal_rtrw"`
	LegalKelurahan     string `gorm:"column:LegalKelurahan" json:"legal_kelurahan"`
	LegalKecamatan     string `gorm:"column:LegalKecamatan" json:"legal_kecamatan"`
	LegalZipCode       string `gorm:"column:LegalZipcode" json:"legal_zipcode"`
	LegalCity          string `gorm:"column:LegalCity" json:"legal_city"`
	ResidenceAddress   string `gorm:"column:ResidenceAddress" json:"residence_address"`
	ResidenceRTRW      string `gorm:"column:ResidenceRTRW" json:"residence_rtrw"`
	ResidenceKelurahan string `gorm:"column:ResidenceKelurahan" json:"residence_kelurahan"`
	ResidenceKecamatan string `gorm:"column:ResidenceKecamatan" json:"residence_kecamatan"`
	ResidenceZipCode   string `gorm:"column:ResidenceZipcode" json:"residence_zipcode"`
	ResidenceCity      string `gorm:"column:ResidenceCity" json:"residence_city"`
	CompanyAddress     string `gorm:"column:CompanyAddress" json:"company_address"`
	CompanyRTRW        string `gorm:"column:CompanyRTRW" json:"company_rtrw"`
	CompanyKelurahan   string `gorm:"column:CompanyKelurahan" json:"company_kelurahan"`
	CompanyKecamatan   string `gorm:"column:CompanyKecamatan" json:"company_kecamatan"`
	CompanyZipCode     string `gorm:"column:CompanyZipcode" json:"company_zipcode"`
	CompanyCity        string `gorm:"column:CompanyCity" json:"company_city"`
	CompanyAreaPhone   string `gorm:"column:CompanyAreaPhone" json:"company_area_phone"`
	CompanyPhone       string `gorm:"column:CompanyPhone" json:"company_phone"`
	EmergencyAddress   string `gorm:"column:EmergencyAddress" json:"emergency_address"`
	EmergencyRTRW      string `gorm:"column:EmergencyRTRW" json:"emergency_rtrw"`
	EmergencyKelurahan string `gorm:"column:EmergencyKelurahan" json:"emergency_kelurahan"`
	EmergencyKecamatan string `gorm:"column:EmergencyKecamatan" json:"emergency_kecamatan"`
	EmergencyZipcode   string `gorm:"column:EmergencyZipcode" json:"emergency_zipcode"`
	EmergencyCity      string `gorm:"column:EmergencyCity" json:"emergency_city"`
	EmergencyAreaPhone string `gorm:"column:EmergencyAreaPhone" json:"emergency_area_phone"`
	EmergencyPhone     string `gorm:"column:EmergencyPhone" json:"emergency_phone"`
}

type TotalRow struct {
	Total int `gorm:"column:totalRow" json:"total"`
}
