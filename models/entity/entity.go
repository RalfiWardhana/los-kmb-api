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
	ProspectID                 string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	IDType                     string      `gorm:"type:varchar(30);column:IDType"`
	IDNumber                   string      `gorm:"type:varchar(100);column:IDNumber"`
	IDTypeIssueDate            interface{} `gorm:"column:IDTypeIssuedDate"`
	ExpiredDate                interface{} `gorm:"column:ExpiredDate"`
	LegalName                  string      `gorm:"type:varchar(100);column:LegalName"`
	FullName                   string      `gorm:"type:varchar(100);column:FullName"`
	BirthPlace                 string      `gorm:"type:varchar(100);column:BirthPlace"`
	BirthDate                  time.Time   `gorm:"column:BirthDate"`
	SurgateMotherName          string      `gorm:"type:varchar(100);column:SurgateMotherName"`
	Gender                     string      `gorm:"type:varchar(10);column:Gender"`
	PersonalNPWP               *string     `gorm:"type:varchar(255);column:PersonalNPWP"`
	MobilePhone                string      `gorm:"type:varchar(14);column:MobilePhone"`
	Email                      string      `gorm:"type:varchar(100);column:Email"`
	HomeStatus                 string      `gorm:"type:varchar(20);column:HomeStatus"`
	StaySinceYear              string      `gorm:"type:varchar(10);column:StaySinceYear"`
	StaySinceMonth             string      `gorm:"type:varchar(10);column:StaySinceMonth"`
	Education                  string      `gorm:"type:varchar(50);column:Education"`
	MaritalStatus              string      `gorm:"type:varchar(10);column:MaritalStatus"`
	NumOfDependence            int         `gorm:"column:NumOfDependence"`
	LivingCostAmount           float64     `gorm:"column:LivingCostAmount"`
	Religion                   string      `gorm:"type:varchar(30);column:Religion"`
	CreatedAt                  time.Time   `gorm:"column:created_at"`
	ExtCompanyPhone            *string     `gorm:"type:varchar(4);column:ExtCompanyPhone"`
	SourceOtherIncome          *string     `gorm:"type:varchar(30);column:SourceOtherIncome"`
	JobStatus                  string      `gorm:"type:varchar(10);column:job_status"`
	EmergencyOfficeAreaPhone   string      `gorm:"type:varchar(4);column:EmergencyOfficeAreaPhone"`
	EmergencyOfficePhone       string      `gorm:"type:varchar(20);column:EmergencyOfficePhone"`
	PersonalCustomerType       string      `gorm:"type:varchar(20);column:PersonalCustomerType"`
	Nationality                string      `gorm:"type:varchar(40);column:Nationality"`
	WNACountry                 string      `gorm:"type:varchar(40);column:WNACountry"`
	HomeLocation               string      `gorm:"type:varchar(10);column:HomeLocation"`
	CustomerGroup              string      `gorm:"type:varchar(10);column:CustomerGroup"`
	KKNo                       string      `gorm:"type:varchar(20);column:KKNo"`
	BankID                     string      `gorm:"type:varchar(10);column:BankID"`
	AccountNo                  string      `gorm:"type:varchar(20);column:AccountNo"`
	AccountName                string      `gorm:"type:varchar(100);column:AccountName"`
	Counterpart                int         `gorm:"column:Counterpart"`
	DebtBusinessScale          string      `gorm:"type:varchar(50);column:DebtBusinessScale"`
	DebtGroup                  string      `gorm:"type:varchar(50);column:DebtGroup"`
	IsAffiliateWithPP          string      `gorm:"type:varchar(50);column:IsAffiliateWithPP"`
	AgreetoAcceptOtherOffering int         `gorm:"column:AgreetoAcceptOtherOffering"`
	DataType                   string      `gorm:"type:varchar(30);column:DataType"`
	Status                     string      `gorm:"type:varchar(30);column:Status"`
	IsPV                       *int        `gorm:"column:IsPV"`
	IsRCA                      *int        `gorm:"column:IsRCA"`
	SurveyResult               interface{} `gorm:"type:varchar(255);column:SurveyResult"`
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
	ProspectID string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	PhotoID    string    `gorm:"type:varchar(50);column:photo_id;primary_key:true"`
	Url        string    `gorm:"type:varchar(250);column:url"`
	Width      string    `gorm:"type:varchar(10);column:width"`
	Height     string    `gorm:"type:varchar(10);column:height"`
	Position   string    `gorm:"type:varchar(3);column:position"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (c *CustomerPhoto) TableName() string {
	return "trx_customer_photo"
}

type CustomerEmployment struct {
	ProspectID            string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	ProfessionID          string    `gorm:"type:varchar(10);column:ProfessionID"`
	JobType               string    `gorm:"type:varchar(10);column:JobType"`
	JobPosition           string    `gorm:"type:varchar(10);column:JobPosition"`
	CompanyName           string    `gorm:"type:varchar(50);column:CompanyName"`
	IndustryTypeID        string    `gorm:"type:varchar(10);column:IndustryTypeID"`
	EmploymentSinceYear   string    `gorm:"type:varchar(4);column:EmploymentSinceYear"`
	EmploymentSinceMonth  string    `gorm:"type:varchar(2);column:EmploymentSinceMonth"`
	MonthlyFixedIncome    float64   `gorm:"column:MonthlyFixedIncome"`
	MonthlyVariableIncome float64   `gorm:"column:MonthlyVariableIncome"`
	SpouseIncome          float64   `gorm:"column:SpouseIncome"`
	CreatedAt             time.Time `gorm:"column:created_at"`
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
	ProspectID   string    `gorm:"type:varchar(20);column:ProspectID"`
	Destination  string    `gorm:"type:varchar(10);column:destination"`
	RequestDate  time.Time `gorm:"column:request_date"`
	RequestInfo  *string   `gorm:"type:varchar(255);column:request_info"`
	AssignDate   time.Time `gorm:"column:assign_date"`
	SurveyorName string    `gorm:"type:varchar(100);column:surveyor_name"`
	ResultDate   time.Time `gorm:"column:result_date"`
	Status       string    `gorm:"type:varchar(10);column:status"`
	SurveyorNote *string   `gorm:"type:text;column:surveyor_note"`
	CreatedAt    time.Time `gorm:"column:created_at"`
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
	ProspectID           string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	IDNumber             string      `gorm:"type:varchar(40);column:IDNumber"`
	FullName             string      `gorm:"type:varchar(200);column:FullName"`
	LegalName            string      `gorm:"type:varchar(200);column:LegalName"`
	BirthPlace           string      `gorm:"type:varchar(30);column:BirthPlace"`
	BirthDate            time.Time   `gorm:"column:BirthDate"`
	SurgateMotherName    string      `gorm:"type:varchar(100);column:SurgateMotherName"`
	Gender               string      `gorm:"type:varchar(1);column:Gender"`
	CompanyPhone         interface{} `gorm:"type:varchar(20);column:CompanyPhone"`
	CompanyName          interface{} `gorm:"type:varchar(20);column:CompanyName"`
	MobilePhone          string      `gorm:"type:varchar(20);column:MobilePhone"`
	EmploymentSinceYear  interface{} `gorm:"type:varchar(4);column:EmploymentSinceYear"`
	EmploymentSinceMonth interface{} `gorm:"type:varchar(2);column:EmploymentSinceMonth"`
	ProfessionID         interface{} `gorm:"type:varchar(10);column:ProfessionID"`
	JobType              string      `gorm:"type:varchar(10);column:JobType"`
	JobPosition          interface{} `gorm:"type:varchar(10);column:JobPosition"`
	Email                string      `gorm:"type:varchar(100);column:Email"`
	PersonalNPWP         string      `gorm:"type:varchar(50);column:PersonalNPWP"`
	Education            string      `gorm:"type:varchar(20);column:Education"`
	CreatedAt            time.Time   `gorm:"column:created_at"`
}

func (c *CustomerSpouse) TableName() string {
	return "trx_customer_spouse"
}

type CustomerEmcon struct {
	ProspectID           string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	Name                 string    `gorm:"type:varchar(200);column:Name"`
	Relationship         string    `gorm:"type:varchar(10);column:Relationship"`
	MobilePhone          string    `gorm:"type:varchar(20);column:MobilePhone"`
	CreatedAt            time.Time `gorm:"column:created_at"`
	EmconVerified        string    `gorm:"type:varchar(1);column:EmconVerified"`
	VerifyBy             string    `gorm:"type:varchar(5);column:VerifyBy"`
	KnownCustomerJob     string    `gorm:"type:varchar(1);column:KnownCustomerJob"`
	KnownCustomerAddress string    `gorm:"type:varchar(1);column:KnownCustomerAddress"`
	VerificationWith     string    `gorm:"type:varchar(100);column:VerificationWith"`
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
