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
	RequestID                      string      `gorm:"type:varchar(100);column:RequestID;primaryKey"`
	ProspectID                     string      `gorm:"type:varchar(100);column:ProspectID"`
	ResultDupcheckKonsumen         interface{} `gorm:"type:text;column:ResultDupcheckKonsumen"`
	ResultDupcheckPasangan         interface{} `gorm:"type:text;column:ResultDupcheckPasangan"`
	ResultKreditmu                 interface{} `gorm:"type:text;column:ResultKreditmu"`
	ResultPefindo                  interface{} `gorm:"type:text;column:ResultPefindo"`
	CategoryExcludeBNPL            float64     `gorm:"type:text;column:CategoryExcludeBNPL"`
	OverdueCurrentExcludeBNPL      float64     `gorm:"type:text;column:OverdueCurrentExcludeBNPL"`
	OverdueLast12MonthsExcludeBNPL float64     `gorm:"type:text;column:OverdueLast12MonthsExcludeBNPL"`
	ResultPefindoExcludeBNPL       string      `gorm:"type:text;column:ResultPefindoExcludeBNPL"`
	OverdueCurrentIncludeAll       float64     `gorm:"type:text;column:OverdueCurrentIncludeAll"`
	OverdueLast12MonthsIncludeAll  float64     `gorm:"type:text;column:OverdueLast12MonthsIncludeAll"`
	ResultPefindoIncludeAll        string      `gorm:"type:text;column:ResultPefindoIncludeAll"`
	Response                       interface{} `gorm:"type:text;column:Response"`
	CustomerStatus                 interface{} `gorm:"type:text;column:CustomerStatus"`
	CustomerType                   interface{} `gorm:"type:text;column:CustomerType"`
	DtmResponse                    time.Time   `gorm:"column:DtmResponse"`
	Code                           interface{} `gorm:"type:varchar(50);column:Code"`
	Decision                       string      `gorm:"type:varchar(50);column:Decision"`
	Reason                         string      `gorm:"type:varchar(200);column:Reason"`
	Timestamp                      time.Time   `gorm:"column:Timestamp"`
	PefindoID                      interface{} `gorm:"column:PefindoID"`
	PefindoIDSpouse                interface{} `gorm:"column:PefindoIDSpouse"`
	PefindoScore                   *string     `gorm:"column:PefindoScore"`
	MaxOverdue                     int         `gorm:"column:MaxOverdue"`
	MaxOverdueLast12Months         int         `gorm:"column:MaxOverdueLast12Months"`
	IsNullMaxOverdue               bool        `gorm:"column:IsNullMaxOverdue"`
	IsNullMaxOverdueLast12Months   bool        `gorm:"column:IsNullMaxOverdueLast12Months"`
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
	ResultPefindo       string `gorm:"type:varchar(10);column:result_pefindo"`
	Cluster             string `gorm:"type:varchar(50);column:cluster"`
	TotalBakiDebetStart int    `gorm:"column:total_baki_debet_start"`
	TotalBakiDebetEnd   int    `gorm:"column:total_baki_debet_end"`
	TenorStart          int    `gorm:"column:tenor_start"`
	TenorEnd            int    `gorm:"column:tenor_end"`
	BPKBNameType        int    `gorm:"column:bpkb_name_type"`
	AgeVehicle          string `gorm:"type:varchar(5);column:age_vehicle"`
	LTV                 string `gorm:"type:varchar(5);column:ltv"`
	LTVStart            int    `gorm:"column:ltv_start"`
	LTVEnd              int    `gorm:"column:ltv_end"`
	Decision            string `gorm:"type:varchar(10);column:decision"`
}

func (c *MappingElaborateScheme) TableName() string {
	return "kmb_mapping_elaborate_scheme"
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
type EncryptedString struct {
	MyString string `gorm:"column:my_string"`
}

type ConfigPMK struct {
	Data DataPMK `json:"data"`
}

type DataPMK struct {
	MinAgeMarried    int  `json:"min_age_marital_status_m"`
	MinAgeSingle     int  `json:"min_age_marital_status_s"`
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

type MappingIncomePMK struct {
	ID             string    `gorm:"column:id"`
	BranchID       string    `gorm:"column:branch_id"`
	StatusKonsumen string    `gorm:"column:status_konsumen"`
	Income         int       `gorm:"column:income"`
	Lob            string    `gorm:"column:lob"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (c *MappingIncomePMK) TableName() string {
	return "mapping_income_pmk"
}

type RejectChassisNumber struct {
	ProspectID           string  `gorm:"type:varchar(20);column:ProspectID"`
	IDNumber             string  `gorm:"type:varchar(100);column:IDNumber" json:"id_number"`
	LegalName            string  `gorm:"type:varchar(100);column:LegalName" json:"legal_name"`
	BirthPlace           string  `gorm:"type:varchar(100);column:BirthPlace" json:"birth_place"`
	BirthDate            string  `gorm:"column:BirthDate" json:"birth_date"`
	Gender               string  `gorm:"type:varchar(1);column:Gender"`
	MaritalStatus        string  `gorm:"type:varchar(10);column:MaritalStatus" json:"marital_status"`
	NumOfDependence      int     `gorm:"column:NumOfDependence" json:"num_of_dependence"`
	StaySinceYear        string  `gorm:"type:varchar(10);column:StaySinceYear" json:"stay_since_year"`
	StaySinceMonth       string  `gorm:"type:varchar(10);column:StaySinceMonth" json:"stay_since_month"`
	HomeStatus           string  `gorm:"type:varchar(20);column:HomeStatus" json:"home_status"`
	LegalZipCode         string  `gorm:"type:varchar(5);column:LegalZipCode"`
	CompanyZipCode       string  `gorm:"type:varchar(5);column:CompanyZipCode"`
	ProfessionID         string  `gorm:"type:varchar(10);column:ProfessionID" json:"profession_id"`
	MonthlyFixedIncome   float64 `gorm:"column:MonthlyFixedIncome"`
	EmploymentSinceYear  string  `gorm:"type:varchar(4);column:EmploymentSinceYear"`
	EmploymentSinceMonth string  `gorm:"type:varchar(2);column:EmploymentSinceMonth"`
	EngineNo             string  `gorm:"type:varchar(30);column:engine_number"`
	ChassisNo            string  `gorm:"type:varchar(30);column:chassis_number"`
	BPKBName             string  `gorm:"type:varchar(2);column:bpkb_name"`
	ManufactureYear      string  `gorm:"type:varchar(4);column:manufacture_year"`
	NTF                  float64 `gorm:"column:NTF"`
	OTR                  float64 `gorm:"column:OTR"`
	Tenor                int     `gorm:"column:Tenor"`
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

type SpDupcekChasisNo struct {
	InstallmentAmount float64 `gorm:"column:InstallmentAmount" json:"installment_amount"`
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
		AsliriActive bool `json:"asliri_service_active"`
		AsliriPhoto  int  `json:"asliri_threshold_selfie_photo"`
		AsliriName   int  `json:"asliri_threshold_name"`
		AsliriPDOB   int  `json:"asliri_threshold_pdob"`
	} `json:"data"`
}

type ConfigThresholdDukcapil struct {
	Data struct {
		VerifyData struct {
			Service    string     `json:"service_on"`
			VDIziData  VDIziData  `json:"izidata"`
			VDDukcapil VDDukcapil `json:"dukcapil"`
		} `json:"verify_data"`
		FaceRecognition struct {
			Service    string     `json:"service_on"`
			FRIziData  FRIziData  `json:"izidata"`
			FRDukcapil FRDukcapil `json:"dukcapil"`
		} `json:"face_recognition"`
	} `json:"data"`
}

type VDDukcapil struct {
	NamaLengkap float64 `json:"nama_lengkap"`
	Alamat      float64 `json:"alamat"`
}

type VDIziData struct {
	NamaLengkap float64 `json:"nama_lengkap"`
}

type FRDukcapil struct {
	Threshold float64 `json:"threshold"`
}

type FRIziData struct {
	Threshold float64 `json:"threshold"`
}

type MappingResultDukcapilVD struct {
	ID                     string      `gorm:"column:id"`
	ResultVD               string      `gorm:"column:result_vd"`
	StatusKonsumen         string      `gorm:"column:status_konsumen"`
	KategoriStatusKonsumen string      `gorm:"column:kategori_status_konsumen"`
	Decision               string      `gorm:"column:decision"`
	RuleCode               string      `gorm:"column:rule_code"`
	CreatedAt              time.Time   `gorm:"column:created_at"`
	IsValid                interface{} `gorm:"column:is_valid"`
}

func (c *MappingResultDukcapilVD) TableName() string {
	return "kmb_dukcapil_verify_result_v2"
}

type MappingResultDukcapil struct {
	ID                     string    `gorm:"column:id"`
	ResultVD               string    `gorm:"column:result_vd"`
	ResultFR               string    `gorm:"column:result_fr"`
	StatusKonsumen         string    `gorm:"column:status_konsumen"`
	KategoriStatusKonsumen string    `gorm:"column:kategori_status_konsumen"`
	Decision               string    `gorm:"column:decision"`
	RuleCode               string    `gorm:"column:rule_code"`
	CreatedAt              time.Time `gorm:"column:created_at"`
	UpdatedAt              time.Time `gorm:"column:updated_at"`
}

func (c *MappingResultDukcapil) TableName() string {
	return "kmb_dukcapil_mapping_result_v2"
}

type ScoreGenerator struct {
	Key               string `gorm:"type:varchar(100);column:key"`
	ScoreGeneratorsID string `gorm:"type:varchar(100);column:score_generators_id"`
}

type GetActiveLoanTypeLast6M struct {
	CustomerID           string `gorm:"column:CustomerID"`
	ActiveLoanTypeLast6M string `gorm:"column:active_loanType_last6m"`
}

type GetActiveLoanTypeLast24M struct {
	AgreementNo string `gorm:"column:AgreementNo"`
	MOB         string `gorm:"column:MOB"`
}

type GetMoblast struct {
	Moblast string `gorm:"column:moblast"`
}

type TrxMaster struct {
	ProspectID        string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	BranchID          string    `gorm:"type:varchar(5);column:BranchID"`
	ApplicationSource string    `gorm:"type:varchar(3);column:application_source"`
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
	PersonalNPWP               *string     `gorm:"type:varchar(25);column:PersonalNPWP" json:"-"`
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
	CustomerID                 string      `gorm:"type:varchar(20);column:CustomerID" json:"customer_id"`
	CustomerStatus             string      `gorm:"type:varchar(10);column:CustomerStatus" json:"customer_status"`
	SurveyResult               interface{} `gorm:"type:varchar(255);column:SurveyResult" json:"survey_result"`
	RentFinishDate             *string     `gorm:"column:RentFinishDate" json:"-"`
}

func (c *CustomerPersonal) TableName() string {
	return "trx_customer_personal"
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
	InsuranceAmount             float64   `gorm:"column:insurance_amount"`
	InterestAmount              float64   `gorm:"column:interest_amount"`
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
	WayOfPayment                string    `gorm:"type:varchar(20);column:WayOfPayment"`
	StampDutyFee                float64   `gorm:"column:stamp_duty_fee"`
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

type TrxBannedPMKDSR struct {
	ProspectID string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	IDNumber   string    `gorm:"type:varchar(40);column:IDNumber"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (c *TrxBannedPMKDSR) TableName() string {
	return "trx_banned_pmk_dsr"
}

type TrxLockSystem struct {
	ProspectID string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	IDNumber   string    `gorm:"type:varchar(40);column:IDNumber"`
	Reason     string    `gorm:"type:varchar(250);column:reason"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UnbanDate  time.Time `gorm:"column:unban_date"`
}

func (c *TrxLockSystem) TableName() string {
	return "trx_lock_system"
}

type TrxBannedChassisNumber struct {
	ProspectID string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	ChassisNo  string    `gorm:"type:varchar(30);column:chassis_number"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (c *TrxBannedChassisNumber) TableName() string {
	return "trx_banned_chassis_number"
}

type TrxReject struct {
	RejectPMKDSR int `gorm:"column:reject_pmk_dsr"`
	RejectNIK    int `gorm:"column:reject_nik"`
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
	ProspectID        string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true" json:"-"`
	IDNumber          string      `gorm:"type:varchar(40);column:IDNumber" json:"spouse_id_number"`
	FullName          string      `gorm:"type:varchar(200);column:FullName" json:"-"`
	LegalName         string      `gorm:"type:varchar(200);column:LegalName" json:"spouse_legal_name"`
	BirthPlace        string      `gorm:"type:varchar(30);column:BirthPlace" json:"-"`
	BirthDate         time.Time   `gorm:"column:BirthDate" json:"-"`
	SurgateMotherName string      `gorm:"type:varchar(100);column:SurgateMotherName" json:"-"`
	Gender            string      `gorm:"type:varchar(1);column:Gender" json:"-"`
	CompanyPhone      interface{} `gorm:"type:varchar(20);column:CompanyPhone" json:"spouse_company_phone"`
	CompanyName       interface{} `gorm:"type:varchar(20);column:CompanyName" json:"spouse_company_name"`
	MobilePhone       string      `gorm:"type:varchar(20);column:MobilePhone" json:"spouse_mobile_phone"`
	ProfessionID      interface{} `gorm:"type:varchar(10);column:ProfessionID" json:"spouse_profession"`
	CreatedAt         time.Time   `gorm:"column:created_at" json:"-"`
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
	CustomerStatusKMB               interface{} `gorm:"column:customer_status_kmb;type:varchar(5)" json:"customer_status_kmb"`
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
	CMOID                           interface{} `gorm:"column:cmo_id;type:varchar(20)" json:"cmo_id"`
	CMOJoinDate                     interface{} `gorm:"column:cmo_join_date" json:"cmo_join_date"`
	CMOCategory                     interface{} `gorm:"column:cmo_category;type:char(3)" json:"cmo_category"`
	CMOFPD                          interface{} `gorm:"column:cmo_fpd" json:"cmo_fpd"`
	CMOAccSales                     interface{} `gorm:"column:cmo_acc_sales" json:"cmo_acc_sales"`
	CMOCluster                      interface{} `gorm:"column:cmo_cluster;type:varchar(20)" json:"cmo_cluster"`
	Reason                          interface{} `gorm:"column:reason;type:varchar(250)" json:"reason"`
	Category                        interface{} `gorm:"column:category" json:"category"`
	MaxOverdueKORules               interface{} `gorm:"column:max_overdue_ko_rules" json:"max_overdue_ko_rules"`
	MaxOverdueLast12MonthsKORules   interface{} `gorm:"column:max_overdue_last12months_ko_rules" json:"max_overdue_last12months_ko_rules"`
	NewKoRules                      interface{} `gorm:"column:new_ko_rules" json:"new_ko_rules"`
	RrdDate                         interface{} `gorm:"column:rrd_date" json:"rrd_date"`
	IDNumber                        string      `gorm:"column:id_number;type:varchar(100)" json:"id_number"`
	CreatedAt                       time.Time   `gorm:"column:created_at" json:"created_at"`
}

func (c *FilteringKMB) TableName() string {
	return "trx_filtering"
}

type ResultFiltering struct {
	ProspectID                      string      `gorm:"column:prospect_id;type:varchar(20)" json:"prospect_id"`
	Decision                        string      `gorm:"column:decision;type:varchar(20)" json:"decision"`
	CustomerStatus                  interface{} `gorm:"column:customer_status;type:varchar(5)" json:"customer_status"`
	CustomerStatusKMB               interface{} `gorm:"column:customer_status_kmb;type:varchar(5)" json:"customer_status_kmb"`
	CustomerSegment                 interface{} `gorm:"column:customer_segment;type:varchar(20)" json:"customer_segment"`
	IsBlacklist                     bool        `gorm:"column:is_blacklist" json:"is_blacklist"`
	NextProcess                     bool        `gorm:"column:next_process" json:"next_process"`
	UrlPdfReport                    interface{} `gorm:"column:url_pdf_report" json:"url_pdf_report"`
	TotalBakiDebetNonCollateralBiro interface{} `gorm:"column:total_baki_debet_non_collateral_biro" json:"total_baki_debet_non_collateral_biro"`
	Reason                          string      `gorm:"column:reason;type:varchar(250)" json:"reason"`
	Subject                         string      `gorm:"type:varchar(10);column:subject" json:"subject"`
}

type TrxDetail struct {
	ProspectID     string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true" json:"-"`
	StatusProcess  string      `gorm:"type:varchar(3);column:status_process" json:"-"`
	Activity       string      `gorm:"type:varchar(4);column:activity" json:"-"`
	Decision       string      `gorm:"type:varchar(3);column:decision" json:"decision"`
	RuleCode       interface{} `gorm:"type:varchar(4);column:rule_code" json:"-"`
	SourceDecision string      `gorm:"type:varchar(3);column:source_decision" json:"source"`
	NextStep       interface{} `gorm:"type:varchar(3);column:next_step" json:"-"`
	Type           interface{} `gorm:"type:varchar(3);column:type" json:"-"`
	Info           interface{} `gorm:"type:text;column:info" json:"-"`
	Reason         interface{} `gorm:"type:varchar(200);column:reason" json:"reason"`
	CreatedBy      string      `gorm:"type:varchar(100);column:created_by" json:"-"`
	CreatedAt      time.Time   `gorm:"column:created_at" json:"created_at"`
}

func (c *TrxDetail) TableName() string {
	return "trx_details"
}

type TrxDetailBiro struct {
	ProspectID                             string      `gorm:"type:varchar(20);column:prospect_id;primary_key:true"`
	Subject                                string      `gorm:"type:varchar(10);column:subject"`
	Source                                 string      `gorm:"type:varchar(5);column:source"`
	BiroID                                 string      `gorm:"type:varchar(20);column:biro_id"`
	Score                                  string      `gorm:"type:varchar(20);column:score"`
	MaxOverdue                             interface{} `gorm:"column:max_overdue"`
	MaxOverdueLast12months                 interface{} `gorm:"column:max_overdue_last12months"`
	InstallmentAmount                      interface{} `gorm:"column:installment_amount"`
	WoContract                             int         `gorm:"column:wo_contract"`
	WoWithCollateral                       int         `gorm:"column:wo_with_collateral"`
	BakiDebetNonCollateral                 float64     `gorm:"column:baki_debet_non_collateral"`
	UrlPdfReport                           string      `gorm:"type:varchar(200);column:url_pdf_report"`
	CreatedAt                              time.Time   `gorm:"column:created_at"`
	Plafon                                 interface{} `gorm:"column:plafon"`
	FasilitasAktif                         interface{} `gorm:"column:fasilitas_aktif"`
	KualitasKreditTerburuk                 interface{} `gorm:"column:kualitas_kredit_terburuk"`
	BulanKualitasTerburuk                  interface{} `gorm:"column:bulan_kualitas_terburuk"`
	BakiDebetKualitasTerburuk              interface{} `gorm:"column:baki_debet_kualitas_terburuk"`
	KualitasKreditTerakhir                 interface{} `gorm:"column:kualitas_kredit_terakhir"`
	BulanKualitasKreditTerakhir            interface{} `gorm:"column:bulan_kualitas_kredit_terakhir"`
	OverdueLastKORules                     interface{} `gorm:"column:overdue_last_ko_rules"`
	OverdueLast12MonthsKORules             interface{} `gorm:"column:overdue_last_12month_ko_rules"`
	Category                               interface{} `gorm:"column:category"`
	MaxOverdueAgunanKORules                interface{} `gorm:"column:max_ovd_agunan_ko_rules"`
	MaxOverdueAgunanLast12MonthsKORules    interface{} `gorm:"column:max_ovd_agunan_last_12month_ko_rules"`
	MaxOverdueNonAgunanKORules             interface{} `gorm:"column:max_ovd_non_agunan_ko_rules"`
	MaxOverdueNonAgunanLast12MonthsKORules interface{} `gorm:"column:max_ovd_non_agunan_last_12month_ko_rules"`
}

func (c *TrxDetailBiro) TableName() string {
	return "trx_detail_biro"
}

type MasterMappingCluster struct {
	BranchID       string `gorm:"column:branch_id"`
	CustomerStatus string `gorm:"column:customer_status"`
	BpkbNameType   int    `gorm:"column:bpkb_name_type"`
	Cluster        string `gorm:"column:cluster"`
}

func (c *MasterMappingCluster) TableName() string {
	return "kmb_mapping_cluster_branch"
}

type MasterMappingMaxDSR struct {
	Cluster      string  `gorm:"column:cluster"`
	DSRThreshold float64 `gorm:"column:dsr_threshold"`
}

func (c *MasterMappingMaxDSR) TableName() string {
	return "kmb_mapping_cluster_dsr"
}

type TrxAgreement struct {
	ProspectID          string      `gorm:"type:varchar(20);column:ProspectID"`
	BranchID            interface{} `gorm:"type:varchar(5);column:BranchID"`
	CustomerID          interface{} `gorm:"type:varchar(50);column:CustomerID"`
	ApplicationID       interface{} `gorm:"type:varchar(50);column:ApplicationID"`
	AgreementNo         interface{} `gorm:"type:varchar(50);column:AgreementNo"`
	AgreementDate       interface{} `gorm:"column:AgreementDate"`
	NextInstallmentDate interface{} `gorm:"column:NextInstallmentDate"`
	MaturityDate        interface{} `gorm:"column:MaturityDate"`
	ContractStatus      string      `gorm:"type:varchar(10);column:ContractStatus"`
	NewApplicationDate  interface{} `gorm:"column:NewApplicationDate"`
	ApprovalDate        interface{} `gorm:"column:ApprovalDate"`
	PurchaseOrderDate   interface{} `gorm:"column:PurchaseOrderDate"`
	GoLiveDate          interface{} `gorm:"column:GoLiveDate"`
	ProductID           interface{} `gorm:"type:varchar(20);column:ProductID"`
	ProductOfferingID   interface{} `gorm:"type:varchar(20);column:ProductOfferingID"`
	TotalOTR            interface{} `gorm:"column:TotalOTR"`
	DownPayment         interface{} `gorm:"column:DownPayment"`
	NTF                 interface{} `gorm:"column:NTF"`
	PayToDealerAmount   interface{} `gorm:"column:PayToDealerAmount"`
	PayToDealerDate     interface{} `gorm:"column:PayToDealerDate"`
	CheckingStatus      string      `gorm:"type:varchar(5);column:checking_status"`
	LastCheckingAt      interface{} `gorm:"column:last_checking_at"`
	CreatedAt           time.Time   `gorm:"column:created_at"`
	UpdatedAt           time.Time   `gorm:"column:updated_at"`
	AF                  float64     `gorm:"column:AF"`
	MobilePhone         string      `gorm:"type:varchar(20);column:MobilePhone"`
	CustomerIDKreditmu  string      `gorm:"type:varchar(50);column:customer_id_kreditmu"`
}

func (c *TrxAgreement) TableName() string {
	return "trx_agreements"
}

type TrxWorker struct {
	ProspectID      string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	Activity        string      `gorm:"type:varchar(10);column:activity"`
	EndPointTarget  string      `gorm:"type:varchar(100);column:endpoint_target"`
	EndPointMethod  string      `gorm:"type:varchar(10);column:endpoint_method"`
	Payload         string      `gorm:"type:text;column:payload"`
	Header          string      `gorm:"type:text;column:header"`
	ResponseTimeout int         `gorm:"column:response_timeout"`
	APIType         string      `gorm:"type:varchar(3);column:api_type"`
	MaxRetry        int         `gorm:"max_retry"`
	CountRetry      int         `gorm:"count_retry"`
	CreatedAt       time.Time   `gorm:"column:created_at"`
	Category        string      `gorm:"type:varchar(30);column:category"`
	Action          string      `gorm:"type:varchar(50);column:action"`
	StatusCode      string      `gorm:"type:varchar(4);column:status_code"`
	Sequence        interface{} `gorm:"column:sequence"`
}

func (c *TrxWorker) TableName() string {
	return "trx_worker"
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

type TrxJourney struct {
	ProspectID string      `gorm:"type:varchar(20);column:ProspectID"`
	Request    string      `gorm:"type:varchar(8000);column:request"`
	Request2   interface{} `gorm:"type:varchar(8000);column:request2"`
	CreatedAt  time.Time   `gorm:"column:created_at"`
}

func (c *TrxJourney) TableName() string {
	return "trx_journey"
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
	ProspectID                   string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	ScsDate                      interface{} `gorm:"column:ScsDate"`
	ScsScore                     interface{} `gorm:"type:varchar(20);column:ScsScore"`
	ScsStatus                    interface{} `gorm:"type:varchar(10);column:ScsStatus"`
	CustomerType                 interface{} `gorm:"column:CustomerType"`
	SpouseType                   interface{} `gorm:"column:SpouseType"`
	AgreementStatus              interface{} `gorm:"type:varchar(10);column:AgreementStatus"`
	TotalAgreementAktif          interface{} `gorm:"column:TotalAgreementAktif"`
	MaxOVDAgreementAktif         interface{} `gorm:"column:MaxOVDAgreementAktif"`
	LastMaxOVDAgreement          interface{} `gorm:"column:LastMaxOVDAgreement"`
	DSRFMF                       interface{} `gorm:"column:DSRFMF"`
	DSRPBK                       interface{} `gorm:"column:DSRPBK"`
	TotalDSR                     interface{} `gorm:"column:TotalDSR"`
	EkycSource                   interface{} `gorm:"type:varchar(30);column:EkycSource"`
	EkycSimiliarity              interface{} `gorm:"column:EkycSimiliarity"`
	EkycReason                   interface{} `gorm:"type:varchar(30);column:EkycReason"`
	NumberOfPaidInstallment      interface{} `gorm:"column:NumberOfPaidInstallment"`
	OSInstallmentDue             interface{} `gorm:"column:OSInstallmentDue"`
	InstallmentAmountFMF         interface{} `gorm:"column:InstallmentAmountFMF"`
	InstallmentAmountSpouseFMF   interface{} `gorm:"column:InstallmentAmountSpouseFMF"`
	InstallmentAmountOther       interface{} `gorm:"column:InstallmentAmountOther"`
	InstallmentAmountOtherSpouse interface{} `gorm:"column:InstallmentAmountOtherSpouse"`
	InstallmentTopup             interface{} `gorm:"column:InstallmentTopup"`
	LatestInstallment            interface{} `gorm:"column:LatestInstallment"`
	UrlFormAkkk                  interface{} `gorm:"column:UrlFormAkkk"`
	CreatedAt                    time.Time   `gorm:"column:created_at"`
}

func (c *TrxAkkk) TableName() string {
	return "trx_akkk"
}

type Akkk struct {
	ProspectID              string      `gorm:"column:ProspectID" json:"prospect_id"`
	FinancePurpose          interface{} `gorm:"column:FinancePurpose" json:"finance_purpose"`
	LegalName               interface{} `gorm:"column:LegalName" json:"legal_name"`
	IDNumber                interface{} `gorm:"column:IDNumber" json:"id_number"`
	PersonalNPWP            interface{} `gorm:"column:PersonalNPWP" json:"personal_npwp"`
	SurgateMotherName       interface{} `gorm:"column:SurgateMotherName" json:"surgate_mother_name"`
	ProfessionID            interface{} `gorm:"column:ProfessionID" json:"profession_id"`
	CustomerStatus          interface{} `gorm:"column:CustomerStatus" json:"customer_status"`
	CustomerType            interface{} `gorm:"column:CustomerType" json:"customer_type"`
	Gender                  interface{} `gorm:"column:Gender" json:"gender"`
	BirthPlace              interface{} `gorm:"column:BirthPlace" json:"birth_place"`
	BirthDate               interface{} `gorm:"column:BirthDate" json:"birth_date"`
	Education               interface{} `gorm:"column:Education" json:"education"`
	MobilePhone             interface{} `gorm:"column:MobilePhone" json:"mobile_phone"`
	Email                   interface{} `gorm:"column:Email" json:"email"`
	SpouseLegalName         interface{} `gorm:"column:SpouseLegalName" json:"spouse_legal_name"`
	SpouseIDNumber          interface{} `gorm:"column:SpouseIDNumber" json:"spouse_id_number"`
	SpouseSurgateMotherName interface{} `gorm:"column:SpouseSurgateMotherName" json:"spouse_surgate_mother_name"`
	SpouseProfessionID      interface{} `gorm:"column:SpouseProfessionID" json:"spouse_profession_id"`
	SpouseType              interface{} `gorm:"column:SpouseType" json:"spouse_type"`
	SpouseGender            interface{} `gorm:"column:SpouseGender" json:"spouse_gender"`
	SpouseBirthPlace        interface{} `gorm:"column:SpouseBirthPlace" json:"spouse_birth_place"`
	SpouseBirthDate         interface{} `gorm:"column:SpouseBirthDate" json:"spouse_birth_date"`
	SpouseMobilePhone       interface{} `gorm:"column:SpouseMobilePhone" json:"spouse_mobile_phone"`
	VerificationWith        interface{} `gorm:"column:VerificationWith" json:"verification_with"`
	EmconRelationship       interface{} `gorm:"column:EmconRelationship" json:"emcon_relationship"`
	EmconVerified           interface{} `gorm:"column:EmconVerified" json:"emcon_verified"`
	Address                 interface{} `gorm:"column:Address" json:"address"`
	EmconMobilePhone        interface{} `gorm:"column:EmconMobilePhone" json:"emcon_mobile_phone"`
	VerifyBy                interface{} `gorm:"column:VerifyBy" json:"verify_by"`
	KnownCustomerAddress    interface{} `gorm:"column:KnownCustomerAddress" json:"known_customer_address"`
	StaySinceYear           interface{} `gorm:"column:StaySinceYear" json:"stay_since_year"`
	StaySinceMonth          interface{} `gorm:"column:StaySinceMonth" json:"stay_since_month"`
	KnownCustomerJob        interface{} `gorm:"column:KnownCustomerJob" json:"known_customer_job"`
	Job                     interface{} `gorm:"column:Job" json:"job"`
	EmploymentSinceYear     interface{} `gorm:"column:EmploymentSinceYear" json:"employment_since_year"`
	EmploymentSinceMonth    interface{} `gorm:"column:EmploymentSinceMonth" json:"employment_since_month"`
	IndustryType            interface{} `gorm:"column:IndustryType" json:"industry_type"`
	IndustryTypeID          interface{} `gorm:"column:IndustryTypeID" json:"industry_type_id"`
	MonthlyFixedIncome      interface{} `gorm:"column:MonthlyFixedIncome" json:"monthly_fixed_income"`
	MonthlyVariableIncome   interface{} `gorm:"column:MonthlyVariableIncome" json:"monthly_variable_income"`
	SpouseIncome            interface{} `gorm:"column:SpouseIncome" json:"spouse_income"`
	BpkbName                interface{} `gorm:"column:BpkbName" json:"bpkb_name"`
	Plafond                 interface{} `gorm:"column:Plafond" json:"plafond"`
	BakiDebet               interface{} `gorm:"column:BakiDebet" json:"baki_debet"`
	FasilitasAktif          interface{} `gorm:"column:FasilitasAktif" json:"fasilitas_aktif"`
	ColTerburuk             interface{} `gorm:"column:ColTerburuk" json:"col_terburuk"`
	BakiDebetTerburuk       interface{} `gorm:"column:BakiDebetTerburuk" json:"baki_debet_terburuk"`
	ColTerakhirAktif        interface{} `gorm:"column:ColTerakhirAktif" json:"col_terakhir_aktif"`
	SpousePlafond           interface{} `gorm:"column:SpousePlafond" json:"spouse_plafond"`
	SpouseBakiDebet         interface{} `gorm:"column:SpouseBakiDebet" json:"spouse_baki_debet"`
	SpouseFasilitasAktif    interface{} `gorm:"column:SpouseFasilitasAktif" json:"spouse_fasilitas_aktif"`
	SpouseColTerburuk       interface{} `gorm:"column:SpouseColTerburuk" json:"spouse_col_terburuk"`
	SpouseBakiDebetTerburuk interface{} `gorm:"column:SpouseBakiDebetTerburuk" json:"spouse_baki_debet_terburuk"`
	SpouseColTerakhirAktif  interface{} `gorm:"column:SpouseColTerakhirAktif" json:"spouse_col_terakhir_aktif"`
	ScsScore                interface{} `gorm:"column:ScsScore" json:"scs_score"`
	AgreementStatus         interface{} `gorm:"column:AgreementStatus" json:"agreement_status"`
	TotalAgreementAktif     interface{} `gorm:"column:TotalAgreementAktif" json:"total_agreement_aktif"`
	MaxOVDAgreementAktif    interface{} `gorm:"column:MaxOVDAgreementAktif" json:"max_ovd_agreement_aktif"`
	LastMaxOVDAgreement     interface{} `gorm:"column:LastMaxOVDAgreement" json:"last_max_ovd_agreement"`
	CustomerSegment         interface{} `gorm:"column:customer_segment" json:"customer_segment"`
	LatestInstallment       interface{} `gorm:"column:LatestInstallment" json:"latest_installment"`
	NTFAkumulasi            interface{} `gorm:"column:NTFAkumulasi" json:"ntf_akumulasi"`
	TotalInstallment        interface{} `gorm:"column:TotalInstallment" json:"total_installment"`
	TotalIncome             interface{} `gorm:"column:TotalIncome" json:"total_income"`
	TotalDSR                interface{} `gorm:"column:TotalDSR" json:"total_dsr"`
	EkycSource              interface{} `gorm:"column:EkycSource" json:"ekyc_source"`
	EkycSimiliarity         interface{} `gorm:"column:EkycSimiliarity" json:"ekyc_similarity"`
	CmoDecision             interface{} `gorm:"column:cmo_decision" json:"cmo_decision"`
	CmoName                 interface{} `gorm:"column:cmo_name" json:"cmo_name"`
	CmoDate                 interface{} `gorm:"column:cmo_date" json:"cmo_date"`
	CaDecision              interface{} `gorm:"column:ca_decision" json:"ca_decision"`
	CaNote                  interface{} `gorm:"column:ca_note" json:"ca_note"`
	CaName                  interface{} `gorm:"column:ca_name" json:"ca_name"`
	CaDate                  interface{} `gorm:"column:ca_date" json:"ca_date"`
	CbmDecision             interface{} `gorm:"column:cbm_decision" json:"cbm_decision"`
	CbmNote                 interface{} `gorm:"column:cbm_note" json:"cbm_note"`
	CbmName                 interface{} `gorm:"column:cbm_name" json:"cbm_name"`
	CbmDate                 interface{} `gorm:"column:cbm_date" json:"cbm_date"`
	DrmDecision             interface{} `gorm:"column:drm_decision" json:"drm_decision"`
	DrmNote                 interface{} `gorm:"column:drm_note" json:"drm_note"`
	DrmName                 interface{} `gorm:"column:drm_name" json:"drm_name"`
	DrmDate                 interface{} `gorm:"column:drm_date" json:"drm_date"`
	GmoDecision             interface{} `gorm:"column:gmo_decision" json:"gmo_decision"`
	GmoNote                 interface{} `gorm:"column:gmo_note" json:"gmo_note"`
	GmoName                 interface{} `gorm:"column:gmo_name" json:"gmo_name"`
	GmoDate                 interface{} `gorm:"column:gmo_date" json:"gmo_date"`
}

type TrxInternalRecord struct {
	ProspectID           string    `gorm:"column:ProspectID" json:"-"`
	CustomerID           string    `gorm:"column:CustomerID" json:"-"`
	ApplicationID        string    `gorm:"column:ApplicationID" json:"application_id"`
	ProductType          string    `gorm:"column:ProductType" json:"product_type"`
	AgreementDate        time.Time `gorm:"column:AgreementDate" json:"agreement_date"`
	AssetCode            string    `gorm:"column:AssetCode" json:"asset_code"`
	Tenor                int       `gorm:"column:Tenor" json:"tenor"`
	OutstandingPrincipal float64   `gorm:"column:OutstandingPrincipal" json:"outstanding_principal"`
	InstallmentAmount    float64   `gorm:"column:InstallmentAmount" json:"installment_amount"`
	ContractStatus       string    `gorm:"column:ContractStatus" json:"contract_status"`
	CurrentCondition     string    `gorm:"column:CurrentCondition" json:"current_condition"`
	CreatedAt            time.Time `gorm:"column:created_at" json:"-"`
}

func (c *TrxInternalRecord) TableName() string {
	return "trx_internal_record"
}

type MasterBranch struct {
	BranchCategory string `gorm:"column:branch_category"`
}

type MappingElaborateIncome struct {
	BranchCategory   string `gorm:"column:branch_category"`
	EstimationIncome string `gorm:"column:estimation_income"`
	StatusKonsumen   string `gorm:"column:status_konsumen"`
	BPKBNameType     int    `gorm:"column:bpkb_name_type"`
	Scoreband        string `gorm:"column:scoreband"`
	Worst24Mth       string `gorm:"column:worst_24mth"`
	Result           string `gorm:"column:result"`
}

func (c *MappingElaborateIncome) TableName() string {
	return "kmb_mapping_treatment_elaborated_income"
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

type TrxHistoryApprovalScheme struct {
	ID                    string      `gorm:"type:varchar(255);column:id;primary_key:true" json:"-"`
	ProspectID            string      `gorm:"type:varchar(20);column:ProspectID" json:"-"`
	Decision              string      `gorm:"type:varchar(3);column:decision" json:"decision"`
	Reason                string      `gorm:"type:varchar(100);column:reason" json:"-"`
	Note                  string      `gorm:"type:varchar(525);column:note" json:"-"`
	CreatedAt             time.Time   `gorm:"column:created_at" json:"approval_date"`
	CreatedBy             string      `gorm:"type:varchar(100);column:created_by" json:"-"`
	DecisionBy            string      `gorm:"type:varchar(250);column:decision_by" json:"pic_approval"`
	NeedEscalation        interface{} `gorm:"column:need_escalation" json:"need_escalation"`
	NextFinalApprovalFlag int         `gorm:"column:next_final_approval_flag" json:"next_final_approval_flag"`
	SourceDecision        string      `gorm:"type:varchar(3);column:source_decision" json:"source_decision"`
	NextStep              string      `gorm:"type:varchar(3);column:next_step" json:"next_step"`
}

func (c *TrxHistoryApprovalScheme) TableName() string {
	return "trx_history_approval_scheme"
}

type TrxDraftCaDecision struct {
	ProspectID  string      `gorm:"type:varchar(20);column:ProspectID" json:"-"`
	Decision    string      `gorm:"type:varchar(3);column:decision" json:"decision"`
	SlikResult  string      `gorm:"column:slik_result" json:"slik_result"`
	Note        interface{} `gorm:"type:varchar(525);column:note" json:"note"`
	CreatedAt   time.Time   `gorm:"column:created_at" json:"created_at"`
	CreatedBy   string      `gorm:"type:varchar(100);column:created_by" json:"created_by"`
	DecisionBy  string      `gorm:"type:varchar(250);column:decision_by" json:"decision_by"`
	Pernyataan1 interface{} `gorm:"column:pernyataan_1" json:"pernyataan_1"`
	Pernyataan2 interface{} `gorm:"column:pernyataan_2" json:"pernyataan_2"`
	Pernyataan3 interface{} `gorm:"column:pernyataan_3" json:"pernyataan_3"`
	Pernyataan4 interface{} `gorm:"column:pernyataan_4" json:"pernyataan_4"`
	Pernyataan5 interface{} `gorm:"column:pernyataan_5" json:"pernyataan_5"`
	Pernyataan6 interface{} `gorm:"column:pernyataan_6" json:"pernyataan_6"`
}

func (c *TrxDraftCaDecision) TableName() string {
	return "trx_draft_ca_decision"
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
	BPKBName           string  `gorm:"column:bpkb_name"`
	OwnerAsset         string  `gorm:"column:owner_asset"`
	LicensePlate       string  `gorm:"column:license_plate"`
	InterestRate       float64 `gorm:"column:interest_rate"`
	InstallmentPeriod  int     `gorm:"column:InstallmentPeriod"`
	OTR                float64 `gorm:"column:OTR"`
	DPAmount           float64 `gorm:"column:DPAmount"`
	FinanceAmount      float64 `gorm:"column:FinanceAmount"`
	InterestAmount     float64 `gorm:"column:interest_amount"`
	LifeInsuranceFee   float64 `gorm:"column:LifeInsuranceFee"`
	AssetInsuranceFee  float64 `gorm:"column:AssetInsuranceFee"`
	InsuranceAmount    float64 `gorm:"column:insurance_amount"`
	AdminFee           float64 `gorm:"column:AdminFee"`
	ProvisionFee       float64 `gorm:"column:provision_fee"`
	NTF                float64 `gorm:"column:NTF"`
	NTFAkumulasi       float64 `gorm:"column:NTFAkumulasi"`
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

type CancelReason struct {
	ReasonID string `gorm:"column:id_cancel_reason" json:"reason_id"`
	Show     string `gorm:"column:show" json:"show"`
	Reason   string `gorm:"column:reason" json:"reason_message"`
}

func (c *CancelReason) TableName() string {
	return "m_cancel_reason"
}

type InquiryDataNE struct {
	ProspectID      string `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	BranchID        string `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	IDNumber        string `gorm:"type:varchar(100);column:IDNumber" json:"id_number"`
	LegalName       string `gorm:"type:varchar(100);column:LegalName" json:"legal_name"`
	BirthDate       string `gorm:"column:BirthDate" json:"birth_date"`
	CreatedAt       string `gorm:"column:created_at" json:"created_at"`
	ResultFiltering string `gorm:"column:ResultFiltering" json:"result_filtering"`
	Reason          string `gorm:"column:Reason" json:"reason"`
}

type InquiryData struct {
	Prescreening DataPrescreening   `json:"prescreening"`
	General      DataGeneral        `json:"general"`
	Personal     DataPersonal       `json:"personal"`
	Spouse       CustomerSpouse     `json:"spouse"`
	Employment   CustomerEmployment `json:"employment"`
	ItemApk      DataItemApk        `json:"item_apk"`
	Surveyor     []TrxSurveyor      `json:"surveyor"`
	Emcon        CustomerEmcon      `json:"emcon"`
	Address      DataAddress        `json:"address"`
	Photo        []DataPhoto        `json:"photo"`
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
	BPKBName              string  `gorm:"column:bpkb_name" json:"bpkb_name"`
	OwnerAsset            string  `gorm:"column:owner_asset" json:"owner_asset"`
	LicensePlate          string  `gorm:"column:license_plate" json:"license_plate"`
	InterestRate          float64 `gorm:"column:interest_rate" json:"interest_rate"`
	Tenor                 int     `gorm:"column:InstallmentPeriod" json:"installment_period"`
	OTR                   float64 `gorm:"column:OTR" json:"otr"`
	DPAmount              float64 `gorm:"column:DPAmount" json:"dp_amount"`
	AF                    float64 `gorm:"column:FinanceAmount" json:"finance_amount"`
	InterestAmount        float64 `gorm:"column:interest_amount" json:"interest_amount"`
	LifeInsuranceFee      float64 `gorm:"column:LifeInsuranceFee" json:"life_insurance_fee"`
	AssetInsuranceFee     float64 `gorm:"column:AssetInsuranceFee" json:"asset_insurance_fee"`
	InsuranceAmount       float64 `gorm:"column:insurance_amount" json:"insurance_amount"`
	AdminFee              float64 `gorm:"column:AdminFee" json:"admin_fee"`
	ProvisionFee          float64 `gorm:"column:provision_fee" json:"provision_fee"`
	NTF                   float64 `gorm:"column:NTF" json:"ntf"`
	NTFAkumulasi          float64 `gorm:"column:NTFAkumulasi" json:"ntf_akumulasi"`
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

type DataPhoto struct {
	PhotoID string `gorm:"column:photo_id" json:"photo_id"`
	Label   string `gorm:"column:label" json:"photo_label"`
	Url     string `gorm:"column:url" json:"photo_url"`
}

type TotalRow struct {
	Total int `gorm:"column:totalRow" json:"total"`
}

type InquiryCa struct {
	ShowAction         bool    `gorm:"column:ShowAction"`
	ActionDate         string  `gorm:"column:ActionDate"`
	ActionFormAkk      bool    `gorm:"column:ActionFormAkk"`
	UrlFormAkkk        string  `gorm:"column:UrlFormAkkk"`
	ActionEditData     bool    `gorm:"column:ActionEditData"`
	AdditionalDP       float64 `gorm:"column:additional_dp"`
	Activity           string  `gorm:"column:activity"`
	SourceDecision     string  `gorm:"column:source_decision"`
	StatusDecision     string  `gorm:"column:decision"`
	StatusReason       string  `gorm:"column:reason"`
	CaDecision         string  `gorm:"column:ca_decision"`
	FinalApproval      string  `gorm:"column:final_approval"`
	CANote             string  `gorm:"column:ca_note"`
	ScsDate            string  `gorm:"column:ScsDate"`
	ScsScore           string  `gorm:"column:ScsScore"`
	ScsStatus          string  `gorm:"column:ScsStatus"`
	BiroCustomerResult string  `gorm:"column:BiroCustomerResult"`
	BiroSpouseResult   string  `gorm:"column:BiroSpouseResult"`
	IsLastApproval     bool    `gorm:"column:is_last_approval"`
	HasReturn          bool    `gorm:"column:HasReturn"`

	DraftDecision    string      `gorm:"column:draft_decision"`
	DraftSlikResult  string      `gorm:"column:draft_slik_result"`
	DraftNote        string      `gorm:"column:draft_note"`
	DraftCreatedAt   time.Time   `gorm:"column:draft_created_at"`
	DraftCreatedBy   string      `gorm:"column:draft_created_by"`
	DraftDecisionBy  string      `gorm:"column:draft_decision_by"`
	DraftPernyataan1 interface{} `gorm:"column:draft_pernyataan_1"`
	DraftPernyataan2 interface{} `gorm:"column:draft_pernyataan_2"`
	DraftPernyataan3 interface{} `gorm:"column:draft_pernyataan_3"`
	DraftPernyataan4 interface{} `gorm:"column:draft_pernyataan_4"`
	DraftPernyataan5 interface{} `gorm:"column:draft_pernyataan_5"`
	DraftPernyataan6 interface{} `gorm:"column:draft_pernyataan_6"`

	ProspectID     string `gorm:"column:ProspectID"`
	BranchName     string `gorm:"column:BranchName"`
	IncomingSource string `gorm:"column:incoming_source"`
	CreatedAt      string `gorm:"column:created_at"`
	OrderAt        string `gorm:"column:order_at"`

	CustomerID        string    `gorm:"column:CustomerID"`
	CustomerStatus    string    `gorm:"column:CustomerStatus"`
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
	SurveyResult      string    `gorm:"column:SurveyResult"`

	Supplier           string  `gorm:"column:dealer"`
	ProductOfferingID  string  `gorm:"column:ProductOfferingID"`
	AssetType          string  `gorm:"column:AssetType"`
	AssetDescription   string  `gorm:"column:asset_description"`
	ManufacturingYear  string  `gorm:"column:manufacture_year"`
	Color              string  `gorm:"column:color"`
	ChassisNumber      string  `gorm:"column:chassis_number"`
	EngineNumber       string  `gorm:"column:engine_number"`
	BPKBName           string  `gorm:"column:bpkb_name"`
	OwnerAsset         string  `gorm:"column:owner_asset"`
	LicensePlate       string  `gorm:"column:license_plate"`
	InterestRate       float64 `gorm:"column:interest_rate"`
	InstallmentPeriod  int     `gorm:"column:InstallmentPeriod"`
	OTR                float64 `gorm:"column:OTR"`
	DPAmount           float64 `gorm:"column:DPAmount"`
	FinanceAmount      float64 `gorm:"column:FinanceAmount"`
	InterestAmount     float64 `gorm:"column:interest_amount"`
	LifeInsuranceFee   float64 `gorm:"column:LifeInsuranceFee"`
	AssetInsuranceFee  float64 `gorm:"column:AssetInsuranceFee"`
	InsuranceAmount    float64 `gorm:"column:insurance_amount"`
	AdminFee           float64 `gorm:"column:AdminFee"`
	ProvisionFee       float64 `gorm:"column:provision_fee"`
	NTF                float64 `gorm:"column:NTF"`
	NTFAkumulasi       float64 `gorm:"column:NTFAkumulasi"`
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

	LegalAddress       string      `gorm:"column:LegalAddress"`
	LegalRTRW          string      `gorm:"column:LegalRTRW"`
	LegalKelurahan     string      `gorm:"column:LegalKelurahan"`
	LegalKecamatan     string      `gorm:"column:LegalKecamatan"`
	LegalZipCode       string      `gorm:"column:LegalZipcode"`
	LegalCity          string      `gorm:"column:LegalCity"`
	ResidenceAddress   string      `gorm:"column:ResidenceAddress"`
	ResidenceRTRW      string      `gorm:"column:ResidenceRTRW"`
	ResidenceKelurahan string      `gorm:"column:ResidenceKelurahan"`
	ResidenceKecamatan string      `gorm:"column:ResidenceKecamatan"`
	ResidenceZipCode   string      `gorm:"column:ResidenceZipcode"`
	ResidenceCity      string      `gorm:"column:ResidenceCity"`
	CompanyAddress     string      `gorm:"column:CompanyAddress"`
	CompanyRTRW        string      `gorm:"column:CompanyRTRW"`
	CompanyKelurahan   string      `gorm:"column:CompanyKelurahan"`
	CompanyKecamatan   string      `gorm:"column:CompanyKecamatan"`
	CompanyZipCode     string      `gorm:"column:CompanyZipcode"`
	CompanyCity        string      `gorm:"column:CompanyCity"`
	CompanyAreaPhone   string      `gorm:"column:CompanyAreaPhone"`
	CompanyPhone       string      `gorm:"column:CompanyPhone"`
	EmergencyAddress   string      `gorm:"column:EmergencyAddress"`
	EmergencyRTRW      string      `gorm:"column:EmergencyRTRW"`
	EmergencyKelurahan string      `gorm:"column:EmergencyKelurahan"`
	EmergencyKecamatan string      `gorm:"column:EmergencyKecamatan"`
	EmergencyZipcode   string      `gorm:"column:EmergencyZipcode"`
	EmergencyCity      string      `gorm:"column:EmergencyCity"`
	EmergencyAreaPhone string      `gorm:"column:EmergencyAreaPhone"`
	EmergencyPhone     string      `gorm:"column:EmergencyPhone"`
	DeviasiID          string      `gorm:"column:deviasi_id"`
	DeviasiDescription string      `gorm:"column:deviasi_description"`
	DeviasiDecision    string      `gorm:"column:deviasi_decision"`
	DeviasiReason      string      `gorm:"column:deviasi_reason"`
	IsEDD              bool        `gorm:"column:is_edd"`
	IsHighrisk         bool        `gorm:"column:is_highrisk"`
	Pernyataan1        interface{} `gorm:"column:pernyataan_1"`
	Pernyataan2        interface{} `gorm:"column:pernyataan_2"`
	Pernyataan3        interface{} `gorm:"column:pernyataan_3"`
	Pernyataan4        interface{} `gorm:"column:pernyataan_4"`
	Pernyataan5        interface{} `gorm:"column:pernyataan_5"`
	Pernyataan6        interface{} `gorm:"column:pernyataan_6"`
}

type InquiryDataCa struct {
	CA             DataCa              `json:"ca"`
	InternalRecord []TrxInternalRecord `json:"internal_record"`
	Approval       []HistoryApproval   `json:"approval"`
	Draft          TrxDraftCaDecision  `json:"draft"`
	General        DataGeneral         `json:"general"`
	Personal       DataPersonal        `json:"personal"`
	Spouse         CustomerSpouse      `json:"spouse"`
	Employment     CustomerEmployment  `json:"employment"`
	ItemApk        DataItemApk         `json:"item_apk"`
	Surveyor       []TrxSurveyor       `json:"surveyor"`
	Emcon          CustomerEmcon       `json:"emcon"`
	Address        DataAddress         `json:"address"`
	Photo          []DataPhoto         `json:"photo"`
	Deviasi        Deviasi             `json:"deviasi"`
	EDD            TrxEDD              `json:"edd"`
}

type DataCa struct {
	ShowAction         bool    `gorm:"column:ShowAction" json:"show_action"`
	ActionEditData     bool    `gorm:"column:ActionEditData" json:"show_edit_data"`
	AdditionalDP       float64 `gorm:"column:additional_dp" json:"additional_dp"`
	StatusDecision     string  `gorm:"column:decision" json:"status_decision"`
	StatusReason       string  `gorm:"column:reason" json:"status_reason"`
	CaDecision         string  `gorm:"column:ca_decision" json:"ca_decision"`
	CaNote             string  `gorm:"column:ca_note" json:"ca_note"`
	ActionDate         string  `gorm:"column:ActionDate" json:"action_date"`
	ScsDate            string  `gorm:"column:ScsDate" json:"scorepro_date"`
	ScsScore           string  `gorm:"column:ScsScore" json:"scorepro_score"`
	ScsStatus          string  `gorm:"column:ScsStatus" json:"scorepro_status"`
	BiroCustomerResult string  `gorm:"column:BiroCustomerResult" json:"biro_customer_result"`
	BiroSpouseResult   string  `gorm:"column:BiroSpouseResult" json:"biro_spouse_result"`
}

type TrxCaDecision struct {
	ProspectID    string      `gorm:"type:varchar(20);column:ProspectID" json:"-"`
	Decision      string      `gorm:"type:varchar(3);column:decision" json:"decision"`
	SlikResult    interface{} `gorm:"type:varchar(30);column:slik_result" json:"slik_result"`
	Note          string      `gorm:"type:varchar(525);column:note" json:"note"`
	CreatedAt     time.Time   `gorm:"column:created_at" json:"-"`
	CreatedBy     string      `gorm:"type:varchar(100);column:created_by" json:"-"`
	DecisionBy    string      `gorm:"type:varchar(250);column:decision_by" json:"-"`
	FinalApproval interface{} `gorm:"type:varchar(3);column:final_approval" json:"final_approval"`
}

func (c *TrxCaDecision) TableName() string {
	return "trx_ca_decision"
}

type MappingLimitApprovalScheme struct {
	ID               string    `gorm:"type:varchar(60);column:id"`
	Alias            string    `gorm:"type:varchar(3);column:alias"`
	Name             string    `gorm:"type:varchar(100);column:name"`
	CoverageNtfStart float64   `gorm:"column:coverage_ntf_start"`
	CoverageNtfEnd   float64   `gorm:"column:coverage_ntf_end"`
	Type             int       `gorm:"column:type"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (c *MappingLimitApprovalScheme) TableName() string {
	return "m_limit_approval_scheme"
}

type TrxFinalApproval struct {
	ProspectID string      `gorm:"type:varchar(20);column:ProspectID" json:"-"`
	Decision   string      `gorm:"type:varchar(3);column:decision" json:"decision"`
	Reason     string      `gorm:"type:varchar(100);column:reason" json:"reason"`
	Note       interface{} `gorm:"type:varchar(525);column:note" json:"note"`
	CreatedAt  time.Time   `gorm:"column:created_at" json:"-"`
	CreatedBy  string      `gorm:"type:varchar(100);column:created_by" json:"-"`
	DecisionBy string      `gorm:"type:varchar(250);column:decision_by" json:"-"`
}

func (c *TrxFinalApproval) TableName() string {
	return "trx_final_approval"
}

type InquirySearch struct {
	ActionReturn   bool   `gorm:"column:ActionReturn"`
	ActionCancel   bool   `gorm:"column:ActionCancel"`
	ActionFormAkk  bool   `gorm:"column:ActionFormAkk"`
	UrlFormAkkk    string `gorm:"column:UrlFormAkkk"`
	ProspectID     string `gorm:"column:ProspectID"`
	FinalStatus    string `gorm:"column:FinalStatus"`
	BranchName     string `gorm:"column:BranchName"`
	IncomingSource string `gorm:"column:incoming_source"`
	CreatedAt      string `gorm:"column:created_at"`
	OrderAt        string `gorm:"column:order_at"`

	CustomerID        string    `gorm:"column:CustomerID"`
	CustomerStatus    string    `gorm:"column:CustomerStatus"`
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
	BPKBName           string  `gorm:"column:bpkb_name"`
	OwnerAsset         string  `gorm:"column:owner_asset"`
	LicensePlate       string  `gorm:"column:license_plate"`
	InterestRate       float64 `gorm:"column:interest_rate"`
	InstallmentPeriod  int     `gorm:"column:InstallmentPeriod"`
	OTR                float64 `gorm:"column:OTR"`
	DPAmount           float64 `gorm:"column:DPAmount"`
	FinanceAmount      float64 `gorm:"column:FinanceAmount"`
	InterestAmount     float64 `gorm:"column:interest_amount"`
	LifeInsuranceFee   float64 `gorm:"column:LifeInsuranceFee"`
	AssetInsuranceFee  float64 `gorm:"column:AssetInsuranceFee"`
	InsuranceAmount    float64 `gorm:"column:insurance_amount"`
	AdminFee           float64 `gorm:"column:AdminFee"`
	ProvisionFee       float64 `gorm:"column:provision_fee"`
	NTF                float64 `gorm:"column:NTF"`
	NTFAkumulasi       float64 `gorm:"column:NTFAkumulasi"`
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

	LegalAddress       string      `gorm:"column:LegalAddress"`
	LegalRTRW          string      `gorm:"column:LegalRTRW"`
	LegalKelurahan     string      `gorm:"column:LegalKelurahan"`
	LegalKecamatan     string      `gorm:"column:LegalKecamatan"`
	LegalZipCode       string      `gorm:"column:LegalZipcode"`
	LegalCity          string      `gorm:"column:LegalCity"`
	ResidenceAddress   string      `gorm:"column:ResidenceAddress"`
	ResidenceRTRW      string      `gorm:"column:ResidenceRTRW"`
	ResidenceKelurahan string      `gorm:"column:ResidenceKelurahan"`
	ResidenceKecamatan string      `gorm:"column:ResidenceKecamatan"`
	ResidenceZipCode   string      `gorm:"column:ResidenceZipcode"`
	ResidenceCity      string      `gorm:"column:ResidenceCity"`
	CompanyAddress     string      `gorm:"column:CompanyAddress"`
	CompanyRTRW        string      `gorm:"column:CompanyRTRW"`
	CompanyKelurahan   string      `gorm:"column:CompanyKelurahan"`
	CompanyKecamatan   string      `gorm:"column:CompanyKecamatan"`
	CompanyZipCode     string      `gorm:"column:CompanyZipcode"`
	CompanyCity        string      `gorm:"column:CompanyCity"`
	CompanyAreaPhone   string      `gorm:"column:CompanyAreaPhone"`
	CompanyPhone       string      `gorm:"column:CompanyPhone"`
	EmergencyAddress   string      `gorm:"column:EmergencyAddress"`
	EmergencyRTRW      string      `gorm:"column:EmergencyRTRW"`
	EmergencyKelurahan string      `gorm:"column:EmergencyKelurahan"`
	EmergencyKecamatan string      `gorm:"column:EmergencyKecamatan"`
	EmergencyZipcode   string      `gorm:"column:EmergencyZipcode"`
	EmergencyCity      string      `gorm:"column:EmergencyCity"`
	EmergencyAreaPhone string      `gorm:"column:EmergencyAreaPhone"`
	EmergencyPhone     string      `gorm:"column:EmergencyPhone"`
	DeviasiID          string      `gorm:"column:deviasi_id"`
	DeviasiDescription string      `gorm:"column:deviasi_description"`
	DeviasiDecision    string      `gorm:"column:deviasi_decision"`
	DeviasiReason      string      `gorm:"column:deviasi_reason"`
	IsEDD              bool        `gorm:"column:is_edd"`
	IsHighrisk         bool        `gorm:"column:is_highrisk"`
	Pernyataan1        interface{} `gorm:"column:pernyataan_1"`
	Pernyataan2        interface{} `gorm:"column:pernyataan_2"`
	Pernyataan3        interface{} `gorm:"column:pernyataan_3"`
	Pernyataan4        interface{} `gorm:"column:pernyataan_4"`
	Pernyataan5        interface{} `gorm:"column:pernyataan_5"`
	Pernyataan6        interface{} `gorm:"column:pernyataan_6"`
}

type InquiryDataSearch struct {
	Action         ActionSearch       `json:"action"`
	HistoryProcess []HistoryProcess   `json:"history_process"`
	General        DataGeneral        `json:"general"`
	Personal       DataPersonal       `json:"personal"`
	Spouse         CustomerSpouse     `json:"spouse"`
	Employment     CustomerEmployment `json:"employment"`
	ItemApk        DataItemApk        `json:"item_apk"`
	Surveyor       []TrxSurveyor      `json:"surveyor"`
	Emcon          CustomerEmcon      `json:"emcon"`
	Address        DataAddress        `json:"address"`
	Photo          []DataPhoto        `json:"photo"`
	Deviasi        Deviasi            `json:"deviasi"`
	EDD            TrxEDD             `json:"edd"`
}

type DataPersonal struct {
	IDNumber          string      `gorm:"type:varchar(100);column:IDNumber" json:"id_number"`
	LegalName         string      `gorm:"type:varchar(100);column:LegalName" json:"legal_name"`
	BirthPlace        string      `gorm:"type:varchar(100);column:BirthPlace" json:"birth_place"`
	BirthDate         interface{} `gorm:"column:BirthDate" json:"birth_date"`
	SurgateMotherName string      `gorm:"type:varchar(100);column:SurgateMotherName" json:"surgate_mother_name"`
	Gender            string      `gorm:"type:varchar(10);column:Gender" json:"gender"`
	MobilePhone       string      `gorm:"type:varchar(14);column:MobilePhone" json:"mobile_phone"`
	Email             string      `gorm:"type:varchar(100);column:Email" json:"email"`
	HomeStatus        string      `gorm:"type:varchar(20);column:HomeStatus" json:"home_status"`
	StaySinceYear     string      `gorm:"type:varchar(10);column:StaySinceYear" json:"stay_since_year"`
	StaySinceMonth    string      `gorm:"type:varchar(10);column:StaySinceMonth" json:"stay_since_month"`
	Education         string      `gorm:"type:varchar(50);column:Education" json:"education"`
	MaritalStatus     string      `gorm:"type:varchar(10);column:MaritalStatus" json:"marital_status"`
	NumOfDependence   int         `gorm:"column:NumOfDependence" json:"num_of_dependence"`
	ExtCompanyPhone   *string     `gorm:"type:varchar(4);column:ExtCompanyPhone" json:"ext_company_phone"`
	SourceOtherIncome *string     `gorm:"type:varchar(30);column:SourceOtherIncome" json:"source_other_income"`
	CustomerID        string      `gorm:"type:varchar(20);column:CustomerID" json:"customer_id"`
	CustomerStatus    string      `gorm:"type:varchar(10);column:CustomerStatus" json:"customer_status"`
	SurveyResult      interface{} `gorm:"type:varchar(255);column:SurveyResult" json:"survey_result"`
}

type ActionSearch struct {
	FinalStatus   string `gorm:"column:FinalStatus" json:"final_status"`
	ActionReturn  bool   `gorm:"column:ActionReturn" json:"action_return"`
	ActionCancel  bool   `gorm:"column:ActionCancel" json:"action_cancel"`
	ActionFormAkk bool   `gorm:"column:ActionFormAkk" json:"action_form_akk"`
	UrlFormAkkk   string `gorm:"column:UrlFormAkkk" json:"url_form_akkk"`
}

type HistoryApproval struct {
	Decision              string      `gorm:"column:decision" json:"decision"`
	Note                  string      `gorm:"column:note" json:"note"`
	CreatedAt             time.Time   `gorm:"column:created_at" json:"approval_date"`
	DecisionBy            string      `gorm:"column:decision_by" json:"pic_approval"`
	NeedEscalation        interface{} `gorm:"column:need_escalation" json:"need_escalation"`
	NextFinalApprovalFlag int         `gorm:"column:next_final_approval_flag" json:"next_final_approval_flag"`
	SourceDecision        string      `gorm:"column:source_decision" json:"source_decision"`
	NextStep              string      `gorm:"column:next_step" json:"next_step"`
	SlikResult            string      `gorm:"column:slik_result" json:"slik_result"`
}

type ApprovalReason struct {
	ReasonID string `gorm:"column:id" json:"reason_id"`
	Value    string `gorm:"column:value" json:"value"`
	Type     string `gorm:"column:Type" json:"type"`
}

type InquiryDataApproval struct {
	CA             DataApproval        `json:"ca"`
	InternalRecord []TrxInternalRecord `json:"internal_record"`
	Approval       []HistoryApproval   `json:"approval"`
	General        DataGeneral         `json:"general"`
	Personal       DataPersonal        `json:"personal"`
	Spouse         CustomerSpouse      `json:"spouse"`
	Employment     CustomerEmployment  `json:"employment"`
	ItemApk        DataItemApk         `json:"item_apk"`
	Surveyor       []TrxSurveyor       `json:"surveyor"`
	Emcon          CustomerEmcon       `json:"emcon"`
	Address        DataAddress         `json:"address"`
	Photo          []DataPhoto         `json:"photo"`
	Deviasi        Deviasi             `json:"deviasi"`
	EDD            TrxEDD              `json:"edd"`
}

type DataApproval struct {
	ShowAction         bool   `gorm:"column:ShowAction" json:"show_action"`
	ActionFormAkk      bool   `gorm:"column:ActionFormAkk" json:"action_form_akk"`
	UrlFormAkkk        string `gorm:"column:UrlFormAkkk" json:"url_form_akkk"`
	IsLastApproval     bool   `gorm:"column:IsLastApproval" json:"is_last_approval"`
	HasReturn          bool   `gorm:"column:HasReturn" json:"has_return"`
	StatusDecision     string `gorm:"column:decision" json:"status_decision"`
	StatusReason       string `gorm:"column:reason" json:"status_reason"`
	FinalApproval      string `gorm:"column:final_approval" json:"final_approval"`
	CaDecision         string `gorm:"column:ca_decision" json:"ca_decision"`
	CaNote             string `gorm:"column:ca_note" json:"ca_note"`
	ActionDate         string `gorm:"column:ActionDate" json:"action_date"`
	ScsDate            string `gorm:"column:ScsDate" json:"scorepro_date"`
	ScsScore           string `gorm:"column:ScsScore" json:"scorepro_score"`
	ScsStatus          string `gorm:"column:ScsStatus" json:"scorepro_status"`
	BiroCustomerResult string `gorm:"column:BiroCustomerResult" json:"biro_customer_result"`
	BiroSpouseResult   string `gorm:"column:BiroSpouseResult" json:"biro_spouse_result"`
}

type RegionBranch struct {
	RegionName   string `gorm:"column:region_name"`
	BranchMember string `gorm:"column:branch_member"`
}

type AFMobilePhone struct {
	AFValue     float64 `gorm:"column:AF"`
	OTR         float64 `gorm:"column:OTR"`
	DPAmount    float64 `gorm:"column:DPAmount"`
	MobilePhone string  `gorm:"column:MobilePhone"`
}

type HistoryProcess struct {
	SourceDecision string `gorm:"column:source_decision" json:"source"`
	Alias          string `gorm:"column:alias" json:"-"`
	Decision       string `gorm:"column:decision" json:"decision"`
	Reason         string `gorm:"column:reason" json:"reason"`
	CreatedAt      string `gorm:"column:created_at" json:"created_at"`
	NextStep       string `gorm:"column:next_step" json:"-"`
}

type TrxRecalculate struct {
	ProspectID          string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	ProductOfferingID   string    `gorm:"type:varchar(30);column:ProductOfferingID"`
	ProductOfferingDesc string    `gorm:"column:product_offering_desc"`
	Tenor               *int      `gorm:"column:Tenor"`
	LoanAmount          float64   `gorm:"column:loan_amount"`
	AF                  float64   `gorm:"column:AF"`
	InstallmentAmount   float64   `gorm:"column:InstallmentAmount"`
	DPAmount            float64   `gorm:"column:DPAmount"`
	PercentDP           float64   `gorm:"column:percent_dp"`
	AdminFee            float64   `gorm:"column:AdminFee"`
	ProvisionFee        float64   `gorm:"column:provision_fee"`
	FidusiaFee          float64   `gorm:"column:fidusia_fee"`
	AssetInsuranceFee   float64   `gorm:"column:AssetInsuranceFee"`
	LifeInsuranceFee    float64   `gorm:"column:LifeInsuranceFee"`
	NTF                 float64   `gorm:"column:NTF"`
	NTFAkumulasi        float64   `gorm:"column:NTFAkumulasi"`
	InterestRate        float64   `gorm:"column:interest_rate"`
	InterestAmount      float64   `gorm:"column:interest_amount"`
	AdditionalDP        float64   `gorm:"column:additional_dp"`
	DSRFMF              float64   `gorm:"column:DSRFMF"`
	TotalDSR            float64   `gorm:"column:TotalDSR"`
	CreatedAt           time.Time `gorm:"column:created_at"`
}

func (c *TrxRecalculate) TableName() string {
	return "trx_recalculate"
}

type GetRecalculate struct {
	ProspectID          string  `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	ProductOfferingID   string  `gorm:"type:varchar(30);column:ProductOfferingID"`
	ProductOfferingDesc string  `gorm:"column:product_offering_desc"`
	Tenor               *int    `gorm:"column:Tenor"`
	LoanAmount          float64 `gorm:"column:loan_amount"`
	AF                  float64 `gorm:"column:AF"`
	InstallmentAmount   float64 `gorm:"column:InstallmentAmount"`
	DPAmount            float64 `gorm:"column:DPAmount"`
	PercentDP           float64 `gorm:"column:percent_dp"`
	AdminFee            float64 `gorm:"column:AdminFee"`
	ProvisionFee        float64 `gorm:"column:provision_fee"`
	FidusiaFee          float64 `gorm:"column:fidusia_fee"`
	AssetInsuranceFee   float64 `gorm:"column:AssetInsuranceFee"`
	LifeInsuranceFee    float64 `gorm:"column:LifeInsuranceFee"`
	NTF                 float64 `gorm:"column:NTF"`
	NTFAkumulasi        float64 `gorm:"column:NTFAkumulasi"`
	InterestRate        float64 `gorm:"column:interest_rate"`
	InterestAmount      float64 `gorm:"column:interest_amount"`
	AdditionalDP        float64 `gorm:"column:additional_dp"`
	TotalInstallmentFMF float64 `gorm:"column:TotalInstallmentFMF"`
	TotalIncome         float64 `gorm:"column:TotalIncome"`
	DSRFMF              float64 `gorm:"column:DSRFMF"`
	DSRPBK              float64 `gorm:"column:DSRPBK"`
	TotalDSR            float64 `gorm:"column:TotalDSR"`
}

// STAGING CONFINS
type STG_GEN_APP struct {
	BranchID              string    `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	ProspectID            string    `gorm:"type:varchar(20);column:ProspectId" json:"prospect_id"`
	ProductID             string    `gorm:"type:varchar(10);column:ProductID" json:"product_id"`
	ProductOfferingID     string    `gorm:"type:varchar(10);column:ProductOfferingID" json:"product_offering_id"`
	Tenor                 int       `gorm:"column:Tenor" json:"tenor"`
	NumOfAssetUnit        int       `gorm:"column:NumOfAssetUnit" json:"num_of_asset_unit"`
	POS                   string    `gorm:"type:varchar(3);column:POS" json:"pos"`
	GuarantorID           *string   `gorm:"type:varchar(50);column:GuarantorID" json:"guaranto_id"`
	GuarantorRelationship *string   `gorm:"type:varchar(10);column:GuarantorRelationship" json:"guarantor_relationship"`
	WayOfPayment          string    `gorm:"type:varchar(2);column:WayOfPayment" json:"way_of_payment"`
	ApplicationSource     string    `gorm:"type:varchar(10);column:ApplicationSource" json:"application_source"`
	AgreementDate         time.Time `gorm:"column:AgreementDate" json:"agreement_date"`
	InsAssetInsuredBy     string    `gorm:"type:varchar(10);column:InsAssetInsuredBy" json:"ins_asset_insured_by"`
	InsAssetPaidBy        string    `gorm:"type:varchar(10);column:InsAssetPaidBy" json:"ins_asset_paid_by"`
	InsAssetPeriod        string    `gorm:"type:varchar(2);column:InsAssetPeriod" json:"ins_asset_period"`
	IsLifeInsurance       string    `gorm:"type:varchar(1);column:IsLifeInsurance" json:"is_life_insurance"`
	AOID                  string    `gorm:"type:varchar(20);column:AOID" json:"aoid"`
	MailingAddress        string    `gorm:"type:varchar(100);column:MailingAddress" json:"mailing_address"`
	MailingRT             string    `gorm:"type:varchar(3);column:MailingRT" json:"mailing_rt"`
	MailingRW             string    `gorm:"type:varchar(3);column:MailingRW" json:"mailing_rw"`
	MailingKelurahan      string    `gorm:"type:varchar(30);column:MailingKelurahan" json:"mailing_kelurahan"`
	MailingKecamatan      string    `gorm:"type:varchar(30);column:MailingKecamatan" json:"mailing_kecamatan"`
	MailingCity           string    `gorm:"type:varchar(30);column:MailingCity" json:"mailing_city"`
	MailingZipCode        string    `gorm:"type:varchar(5);column:MailingZipCode" json:"mailing_zip_code"`
	MailingAreaPhone1     string    `gorm:"type:varchar(4);column:MailingAreaPhone1" json:"mailing_area_phone_1"`
	MailingPhone1         string    `gorm:"type:varchar(10);column:MailingPhone1" json:"mailing_phone_1"`
	MailingAreaFax        *string   `gorm:"type:varchar(4);column:MailingAreaFax" json:"mailing_area_fax"`
	MailingFax            *string   `gorm:"type:varchar(10);column:MailingFax" json:"mailing_fax"`
	UsrCrt                string    `gorm:"type:varchar(20);column:UsrCrt" json:"usr_crt"`
	DtmCrt                time.Time `gorm:"column:DtmCrt" json:"dtm_crt"`
	ApplicationPriority   string    `gorm:"type:varchar(20);column:ApplicationPriority" json:"application_priority"`
}

func (c *STG_GEN_APP) TableName() string {
	return "STG_GEN_APP"
}

type STG_GEN_ASD struct {
	BranchID          string      `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	ProspectID        string      `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	SuplierID         string      `gorm:"type:varchar(10);column:SuplierID" json:"suplier_id"`
	AssetCode         string      `gorm:"type:varchar(200);column:AssetCode" json:"asset_code"`
	OTRPrice          float64     `gorm:"column:OTRPrice" json:"otr_price"`
	DPAmount          float64     `gorm:"column:DPAmount" json:"dp_amount"`
	ManufacturingYear int         `gorm:"column:ManufacturingYear" json:"manufacturing_year"`
	NoRangka          string      `gorm:"type:varchar(50);column:NoRangka" json:"no_rangka"`
	NoMesin           string      `gorm:"type:varchar(50);column:NoMesin" json:"no_mesin"`
	UsedNew           string      `gorm:"type:varchar(10);column:UsedNew" json:"used_new"`
	AssetUsage        string      `gorm:"type:varchar(10);column:AssetUsage" json:"asset_usage"`
	CC                string      `gorm:"type:varchar(50);column:CC" json:"cc"`
	Color             string      `gorm:"type:varchar(50);column:Color" json:"color"`
	LicensePlate      string      `gorm:"type:varchar(50);column:LicensePlate" json:"license_plate"`
	OwnerAsset        string      `gorm:"type:varchar(50);column:OwnerAsset" json:"owner_asset"`
	OwnerKTP          string      `gorm:"type:varchar(50);column:OwnerKTP" json:"owner_ktp"`
	OwnerAddress      string      `gorm:"type:varchar(100);column:OwnerAddress" json:"owner_address"`
	OwnerRT           string      `gorm:"type:varchar(3);column:OwnerRT" json:"owner_rt"`
	OwnerRW           string      `gorm:"type:varchar(3);column:OwnerRW" json:"owner_rw"`
	OwnerKelurahan    string      `gorm:"type:varchar(50);column:OwnerKelurahan" json:"owner_kelurahan"`
	OwnerKecamatan    string      `gorm:"type:varchar(50);column:OwnerKecamatan" json:"owner_kecamatan"`
	OwnerCity         string      `gorm:"type:varchar(50);column:OwnerCity" json:"owner_city"`
	OwnerZipCode      string      `gorm:"type:varchar(5);column:OwnerZipCode" json:"owner_zip_code"`
	LocationAddress   string      `gorm:"type:varchar(100);column:LocationAddress" json:"location_address"`
	LocationKelurahan string      `gorm:"type:varchar(50);column:LocationKelurahan" json:"location_kelurahan"`
	LocationKecamatan string      `gorm:"type:varchar(50);column:LocationKecamatan" json:"location_kecamatan"`
	LocationCity      string      `gorm:"type:varchar(50);column:LocationCity" json:"location_city"`
	LocationZipCode   string      `gorm:"type:varchar(5);column:LocationZipCode" json:"location_zip_code"`
	Region            string      `gorm:"type:varchar(20);column:Region" json:"region"`
	SalesmanID        string      `gorm:"type:varchar(10);column:SalesmanID" json:"salesman_id"`
	UsrCrt            string      `gorm:"type:varchar(20);column:UsrCrt" json:"usr_crt"`
	DtmCrt            time.Time   `gorm:"column:DtmCrt" json:"dtm_crt"`
	PromoToCust       float64     `gorm:"column:PromoToCust" json:"promo_to_cust"`
	TaxDate           interface{} `gorm:"column:TaxDate" json:"tax_date"`
	STNKExpiredDate   interface{} `gorm:"column:STNKExpiredDate" json:"stnk_expired_date"`
}

func (c *STG_GEN_ASD) TableName() string {
	return "STG_GEN_ASD"
}

type STG_GEN_COM struct {
	BranchID                        string    `gorm:"type:varchar(3);column:BranchID" json:"BranchID"`
	ProspectID                      string    `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	SuplierInsuranceIncome          float64   `gorm:"column:SuplierInsuranceIncome" json:"SuplierInsuranceIncome"`
	SupplierUppingBunga             float64   `gorm:"column:SupplierUppingBunga" json:"SupplierUppingBunga"`
	SupplierOtherFee                float64   `gorm:"column:SupplierOtherFee" json:"SupplierOtherFee"`
	SupplierBankAccountID           string    `gorm:"type:varchar(10);column:SupplierBankAccountID" json:"SupplierBankAccountID"`
	SupplierEmployeeInsuranceIncome float64   `gorm:"column:SupplierEmployeeInsuranceIncome" json:"SupplierEmployeeInsuranceIncome"`
	SupplierEmployeeUppingBunga     float64   `gorm:"column:SupplierEmployeeUppingBunga" json:"SupplierEmployeeUppingBunga"`
	SupplierEmployeeOtherFee        float64   `gorm:"column:SupplierEmployeeOtherFee" json:"SupplierEmployeeOtherFee"`
	UsrCrt                          string    `gorm:"type:varchar(20);column:UsrCrt" json:"UsrCrt"`
	DtmCrt                          time.Time `gorm:"column:DtmCrt" json:"DtmCrt"`
}

func (c *STG_GEN_COM) TableName() string {
	return "STG_GEN_COM"
}

type STG_GEN_FIN struct {
	BranchID           string    `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	ProspectID         string    `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	FirstInstallment   string    `gorm:"type:varchar(2);column:FirstInstallment" json:"first_installment"`
	TDPAmount          float64   `gorm:"column:TDPAmount" json:"tdp_amount"`
	AdminFee           float64   `gorm:"column:AdminFee" json:"admin_fee"`
	AdditionalAdmin    float64   `gorm:"column:AdditionalAdmin" json:"additional_admin"`
	SurveyFee          float64   `gorm:"column:SurveyFee" json:"survey_fee"`
	OtherFee           float64   `gorm:"column:OtherFee" json:"other_fee"`
	CostOfSurvey       float64   `gorm:"column:CostOfSurvey" json:"cost_of_survey"`
	AdditionalOtherFee float64   `gorm:"column:AdditionalOtherFee" json:"additional_other_fee"`
	FiduciaFee         float64   `gorm:"column:FiduciaFee" json:"fiducia_fee"`
	IsFiduciaCovered   string    `gorm:"type:varchar(1);column:IsFiduciaCovered" json:"is_fiducia_covered"`
	ProvisionFee       float64   `gorm:"column:ProvisionFee" json:"provision_fee"`
	NotaryFee          float64   `gorm:"column:NotaryFee" json:"notary_fee"`
	InstallmentAmount  float64   `gorm:"column:InstallmentAmount" json:"installment_amount"`
	EffectiveRate      float64   `gorm:"column:EffectiveRate" json:"effective_rate"`
	CommisionSubsidy   float64   `gorm:"column:CommisionSubsidy" json:"commision_subsidy"`
	UsrCrt             string    `gorm:"type:varchar(20);column:UsrCrt" json:"usr_crt"`
	DtmCrt             time.Time `gorm:"column:DtmCrt" json:"dtm_crt"`
	TDPPaidAtCoy       float64   `gorm:"column:TDPPaidAtCoy" json:"tdp_paid_at_coy"`
	PrepaidAmount      float64   `gorm:"column:PrepaidAmount" json:"prepaid_amount"`
	StampDutyFee       float64   `gorm:"column:StampDutyFee" json:"stamp_duty_fee"`
}

func (c *STG_GEN_FIN) TableName() string {
	return "STG_GEN_FIN"
}

type STG_GEN_INS_D struct {
	BranchID      string    `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	ProspectID    string    `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	InsSequenceNo int       `gorm:"column:InsSequenceNo" json:"ins_sequence_no"`
	CoverageType  string    `gorm:"type:varchar(10);column:CoverageType" json:"coverage_type"`
	SRCC          string    `gorm:"type:varchar(1);column:SRCC" json:"srcc"`
	Flood         string    `gorm:"type:varchar(1);column:Flood" json:"flood"`
	EQVET         string    `gorm:"type:varchar(1);column:EQVET" json:"eqvet"`
	TPLAmount     float64   `gorm:"column:TPLAmount" json:"tpl_amount"`
	PAAmount      float64   `gorm:"column:PAAmount" json:"pa_amount"`
	UsrCrt        string    `gorm:"type:varchar(20);column:UsrCrt" json:"usr_crt"`
	DtmCrt        time.Time `gorm:"column:DtmCrt" json:"dtm_crt"`
}

func (c *STG_GEN_INS_D) TableName() string {
	return "STG_GEN_INS_D"
}

type STG_GEN_INS_H struct {
	BranchID                string      `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	ProspectID              string      `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	ApplicationType         string      `gorm:"type:varchar(10);column:ApplicationType" json:"application_type"`
	AmountCoverage          float64     `gorm:"column:AmountCoverage" json:"amount_coverage"`
	InsAssetInsuredBy       string      `gorm:"type:varchar(10);column:InsAssetInsuredBy" json:"ins_asset_insured_by"`
	InsuranceCoyBranchID    string      `gorm:"type:varchar(10);column:InsuranceCoyBranchID" json:"insurance_coy_branch_id"`
	InsLength               int         `gorm:"column:InsLength" json:"ins_length"`
	PremiumAmountToCustomer float64     `gorm:"column:PremiumAmountToCustomer" json:"premium_amount_to_customer"`
	CapitalizedAmount       float64     `gorm:"column:CapitalizedAmount" json:"capitalized_amount"`
	CoverageType            string      `gorm:"type:varchar(10);column:CoverageType" json:"coverage_type"`
	PolicyNo                *string     `gorm:"type:varchar(25);column:PolicyNo" json:"policy_no"`
	ExpiredDate             interface{} `gorm:"column:ExpiredDate" json:"expired_date"`
	InsuranceCompany        *string     `gorm:"type:varchar(50);column:InsuranceCompany" json:"insurance_company"`
	UsrCrt                  string      `gorm:"type:varchar(20);column:UsrCrt" json:"usr_rt"`
	DtmCrt                  time.Time   `gorm:"column:DtmCrt" json:"dtm_crt"`
}

func (c *STG_GEN_INS_H) TableName() string {
	return "STG_GEN_INS_H"
}

type STG_GEN_LFI struct {
	BranchID                 string    `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	ProspectID               string    `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	LifeInsuranceCoyBranchID string    `gorm:"type:varchar(10);column:LifeInsuranceCoyBranchID" json:"life_insurance_coy_branch_id"`
	AmountCoverage           float64   `gorm:"column:AmountCoverage" json:"amount_coverage"`
	PaymentMethod            string    `gorm:"type:varchar(2);column:PaymentMethod" json:"payment_method"`
	PremiumAmountToCustomer  float64   `gorm:"column:PremiumAmountToCustomer" json:"premium_amount_to_customer"`
	UsrCrt                   string    `gorm:"type:varchar(20);column:UsrCrt" json:"usr_crt"`
	DtmCrt                   time.Time `gorm:"column:DtmCrt" json:"dtm_crt"`
}

func (c *STG_GEN_LFI) TableName() string {
	return "STG_GEN_LFI"
}

type STG_MAIN struct {
	BranchID      string    `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	ProspectID    string    `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	IsRCA         int       `gorm:"column:IsRCA" json:"isRCA"`
	IsPV          int       `gorm:"column:Ispv" json:"ispv"`
	DataType      string    `gorm:"type:varchar(1);column:DataType" json:"data_type"`
	CreatedDate   time.Time `gorm:"column:CreatedDate" json:"created_date"`
	Status        string    `gorm:"type:varchar(1);column:Status" json:"status"`
	UpdatedDate   time.Time `gorm:"column:UpdatedDate" json:"updated_date"`
	CustomerID    *string   `gorm:"type:varchar(20);column:CustomerID" json:"customer_id"`
	ApplicationID *string   `gorm:"type:varchar(20);column:ApplicationID" json:"application_id"`
	UsrCrt        string    `gorm:"type:varchar(20);column:UsrCrt" json:"usr_crt"`
	DtmCrt        time.Time `gorm:"column:DtmCrt" json:"dtm_crt"`
	UsrUpd        string    `gorm:"type:varchar(20);column:UsrUpd" json:"usr_upd"`
	DtmUpd        time.Time `gorm:"column:DtmUpd" json:"dtm_upd"`
}

func (c *STG_MAIN) TableName() string {
	return "STG_MAIN"
}

type STG_CUST_H struct {
	BranchID             string    `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	ProspectID           string    `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	LegalName            string    `gorm:"type:varchar(50);column:LegalName" json:"legal_name"`
	FullName             string    `gorm:"type:varchar(50);column:FullName" json:"full_name"`
	PersonalCustomerType string    `gorm:"type:varchar(1);column:PersonalCustomerType" json:"personal_customer_type"`
	IDType               string    `gorm:"type:varchar(10);column:IDType" json:"id_type"`
	IDNumber             string    `gorm:"type:varchar(40);column:IDNumber" json:"id_number"`
	ExpiredDate          time.Time `gorm:"column:ExpiredDate" json:"expired_date"`
	Gender               string    `gorm:"type:varchar(1);column:Gender" json:"gender"`
	BirthPlace           string    `gorm:"type:varchar(100);column:BirthPlace" json:"birthplace"`
	BirthDate            time.Time `gorm:"column:BirthDate" json:"Birthdate"`
	PersonalNPWP         string    `gorm:"type:varchar(25);column:PersonalNPWP" json:"personalNPWP"`
	SurgateMotherName    string    `gorm:"type:varchar(50);column:SurgateMotherName" json:"surgate_mother_name"`
	UsrCrt               string    `gorm:"type:varchar(20);column:UsrCrt" json:"usr_crt"`
	DtmCrt               time.Time `gorm:"column:DtmCrt" json:"dtm_crt"`
}

func (c *STG_CUST_H) TableName() string {
	return "STG_CUST_H"
}

type STG_CUST_D struct {
	BranchID                         string      `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	ProspectID                       string      `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	IDTypeIssuedDate                 time.Time   `gorm:"column:IDTypeIssuedDate" json:"id_type_issued_date"`
	Education                        string      `gorm:"type:varchar(10);column:Education" json:"education"`
	Nationality                      string      `gorm:"type:varchar(3);column:Nationality" json:"nationality"`
	WNACountry                       string      `gorm:"type:varchar(50);column:WNACountry" json:"wna_country"`
	HomeStatus                       string      `gorm:"type:varchar(10);column:HomeStatus" json:"home_status"`
	RentFinishDate                   interface{} `gorm:"column:RentFinishDate" json:"rent_finish_date"`
	HomeLocation                     string      `gorm:"type:varchar(10);column:HomeLocation" json:"home_location"`
	StaySinceMonth                   int         `gorm:"column:StaySinceMonth" json:"stay_since_month"`
	StaySinceYear                    int         `gorm:"column:StaySinceYear" json:"stay_since_year"`
	Religion                         string      `gorm:"type:varchar(10);column:Religion" json:"religion"`
	MaritalStatus                    string      `gorm:"type:varchar(10);column:MaritalStatus" json:"marital_status"`
	NumOfDependence                  int         `gorm:"column:NumOfDependence" json:"num_of_dependence"`
	MobilePhone                      string      `gorm:"type:varchar(20);column:MobilePhone" json:"mobile_phone"`
	Email                            string      `gorm:"type:varchar(100);column:Email" json:"email"`
	CustomerGroup                    string      `gorm:"type:varchar(1);column:CustomerGroup" json:"customer_group"`
	KKNo                             string      `gorm:"type:varchar(30);column:KKNo" json:"kk_no"`
	LegalAddress                     string      `gorm:"type:varchar(100);column:LegalAddress" json:"legal_address"`
	LegalRT                          string      `gorm:"type:varchar(3);column:LegalRT" json:"legal_rt"`
	LegalRW                          string      `gorm:"type:varchar(3);column:LegalRW" json:"legal_rw"`
	LegalKelurahan                   string      `gorm:"type:varchar(30);column:LegalKelurahan" json:"legal_kelurahan"`
	LegalKecamatan                   string      `gorm:"type:varchar(30);column:LegalKecamatan" json:"legal_kecamatan"`
	LegalCity                        string      `gorm:"type:varchar(30);column:LegalCity" json:"legal_city"`
	LegalZipCode                     string      `gorm:"type:varchar(5);column:LegalZipCode" json:"legal_zip_code"`
	LegalAreaPhone1                  string      `gorm:"type:varchar(4);column:LegalAreaPhone1" json:"legal_area_phone_1"`
	LegalPhone1                      string      `gorm:"type:varchar(10);column:LegalPhone1" json:"legal_phone_1"`
	ResidenceAddress                 string      `gorm:"type:varchar(100);column:ResidenceAddress" json:"residence_address"`
	ResidenceRT                      string      `gorm:"type:varchar(3);column:ResidenceRT" json:"residence_rt"`
	ResidenceRW                      string      `gorm:"type:varchar(3);column:ResidenceRW" json:"residence_rw"`
	ResidenceKelurahan               string      `gorm:"type:varchar(30);column:ResidenceKelurahan" json:"residence_kelurahan"`
	ResidenceKecamatan               string      `gorm:"type:varchar(30);column:ResidenceKecamatan" json:"residence_kecamatan"`
	ResidenceCity                    string      `gorm:"type:varchar(30);column:ResidenceCity" json:"residence_city"`
	ResidenceZipCode                 string      `gorm:"type:varchar(5);column:ResidenceZipCode" json:"residence_zip_code"`
	ResidenceAreaPhone1              string      `gorm:"type:varchar(4);column:ResidenceAreaPhone1" json:"residence_area_phone_1"`
	ResidencePhone1                  string      `gorm:"type:varchar(10);column:ResidencePhone1" json:"residence_phone_1"`
	EmergencyContactName             string      `gorm:"type:varchar(50);column:EmergencyContactName" json:"emergency_contact_name"`
	EmergencyContactRelationship     string      `gorm:"type:varchar(10);column:EmergencyContactRelationship" json:"emergency_contact_relationship"`
	EmergencyContactAddress          string      `gorm:"type:varchar(100);column:EmergencyContactAddress" json:"emergency_contact_address"`
	EmergencyContactRT               string      `gorm:"type:varchar(3);column:EmergencyContactRT" json:"emergency_contact_rt"`
	EmergencyContactRW               string      `gorm:"type:varchar(3);column:EmergencyContactRW" json:"emergency_contact_rw"`
	EmergencyContactKelurahan        string      `gorm:"type:varchar(30);column:EmergencyContactKelurahan" json:"Emergency_contact_kelurahan"`
	EmergencyContactKecamatan        string      `gorm:"type:varchar(30);column:EmergencyContactKecamatan" json:"Emergency_contact_kecamatan"`
	EmergencyContactCity             string      `gorm:"type:varchar(30);column:EmergencyContactCity" json:"emergency_contact_city"`
	EmergencyContactZipCode          string      `gorm:"type:varchar(5);column:EmergencyContactZipCode" json:"emergency_contact_zip_code"`
	EmergencyContactHomePhoneArea1   string      `gorm:"type:varchar(4);column:EmergencyContactHomePhoneArea1" json:"emergency_contact_home_phone_area_1"`
	EmergencyContactHomePhone1       string      `gorm:"type:varchar(10);column:EmergencyContactHomePhone1" json:"emergency_contact_home_phone_1"`
	EmergencyContactOfficePhoneArea1 string      `gorm:"type:varchar(4);column:EmergencyContactOfficePhoneArea1" json:"emergency_contact_office_phone_area_1"`
	EmergencyContactOfficePhone1     string      `gorm:"type:varchar(10);column:EmergencyContactOfficePhone1" json:"emergency_contact_office_phone_1"`
	EmergencyContactMobilePhone      string      `gorm:"type:varchar(20);column:EmergencyContactMobilePhone" json:"emergency_contact_mobile_phone"`
	ProfessionID                     string      `gorm:"type:varchar(10);column:ProfessionID" json:"profession_id"`
	JobType                          string      `gorm:"type:varchar(10);column:JobType" json:"job_type"`
	JobPosition                      string      `gorm:"type:varchar(10);column:JobPosition" json:"job_position"`
	CompanyName                      string      `gorm:"type:varchar(50);column:CompanyName" json:"company_name"`
	IndustryTypeID                   string      `gorm:"type:varchar(10);column:IndustryTypeID" json:"industry_type_id"`
	CompanyAddress                   string      `gorm:"type:varchar(100);column:CompanyAddress" json:"company_address"`
	CompanyRT                        string      `gorm:"type:varchar(3);column:CompanyRT" json:"company_rt"`
	CompanyRW                        string      `gorm:"type:varchar(3);column:CompanyRW" json:"company_rw"`
	CompanyKelurahan                 string      `gorm:"type:varchar(30);column:CompanyKelurahan" json:"company_kelurahan"`
	CompanyKecamatan                 string      `gorm:"type:varchar(30);column:CompanyKecamatan" json:"company_kecamatan"`
	CompanyCity                      string      `gorm:"type:varchar(30);column:CompanyCity" json:"company_city"`
	CompanyZipCode                   string      `gorm:"type:varchar(5);column:CompanyZipCode" json:"company_zip_code"`
	CompanyAreaPhone1                string      `gorm:"type:varchar(4);column:CompanyAreaPhone1" json:"company_area_phone_1"`
	CompanyPhone1                    string      `gorm:"type:varchar(10);column:CompanyPhone1" json:"company_phone_1"`
	EmploymentSinceYear              int         `gorm:"column:EmploymentSinceYear" json:"employment_since_year"`
	MonthlyFixedIncome               float64     `gorm:"column:MonthlyFixedIncome" json:"monthly_fixed_income"`
	MonthlyVariableIncome            float64     `gorm:"column:MonthlyVariableIncome" json:"monthly_variable_income"`
	LivingCostAmount                 float64     `gorm:"column:LivingCostAmount" json:"living_cost_amount"`
	MonthlyOmset1Year                int         `gorm:"column:MonthlyOmset1_Year" json:"monthly_omset1_year"`
	MonthlyOmset1Month               int         `gorm:"column:MonthlyOmset1_Month" json:"monthly_omset1_month"`
	MonthlyOmset1Omset               float64     `gorm:"column:MonthlyOmset1_Omset" json:"monthly_omset1_omset"`
	MonthlyOmset2Year                int         `gorm:"column:MonthlyOmset2_Year" json:"monthly_omset2_year"`
	MonthlyOmset2Month               int         `gorm:"column:MonthlyOmset2_Month" json:"monthly_omset2_month"`
	MonthlyOmset2Omset               float64     `gorm:"column:MonthlyOmset2_Omset" json:"monthly_omset2_omset"`
	MonthlyOmset3Year                int         `gorm:"column:MonthlyOmset3_Year" json:"monthly_omset3_year"`
	MonthlyOmset3Month               int         `gorm:"column:MonthlyOmset3_Month" json:"monthly_omset3_month"`
	MonthlyOmset3Omset               float64     `gorm:"column:MonthlyOmset3_Omset" json:"monthly_omset3_omset"`
	Counterpart                      int         `gorm:"column:Counterpart" json:"counterpart"`
	DebtBusinessScale                string      `gorm:"type:varchar(10);column:DebtBusinessScale" json:"debt_business_scale"`
	DebtGroup                        string      `gorm:"type:varchar(10);column:DebtGroup" json:"debt_group"`
	IsAffiliateWithPP                string      `gorm:"type:varchar(1);column:IsAffiliateWithPP" json:"is_affiliate_with_pp"`
	AgreetoAcceptOtherOffering       int         `gorm:"column:AgreetoAcceptOtherOffering" json:"agree_to_accept_other_offering"`
	UsrCrt                           string      `gorm:"type:varchar(20);column:UsrCrt" json:"usr_crt"`
	DtmCrt                           time.Time   `gorm:"column:DtmCrt" json:"dtm_crt"`
	SpouseIncome                     float64     `gorm:"column:SpouseIncome" json:"spouse_income"`
	BankID                           string      `gorm:"type:varchar(5);column:BankID" json:"bank_id"`
	AccountNo                        string      `gorm:"type:varchar(20);column:AccountNo" json:"account_no"`
	AccountName                      string      `gorm:"type:varchar(50);column:AccountName" json:"account_name"`
}

func (c *STG_CUST_D) TableName() string {
	return "STG_CUST_D"
}

type STG_CUST_FAM struct {
	BranchID       string    `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	ProspectID     string    `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	SeqNo          int       `gorm:"column:SeqNo" json:"SeqNo"`
	Name           string    `gorm:"type:varchar(50);column:Name" json:"Name"`
	IDNumber       string    `gorm:"type:varchar(40);column:IDNumber" json:"IDNumber"`
	BirthDate      time.Time `gorm:"column:BirthDate" json:"BirthDate"`
	FamilyRelation string    `gorm:"type:varchar(10);column:FamilyRelation" json:"FamilyRelation"`
	UsrCrt         string    `gorm:"type:varchar(20);column:UsrCrt" json:"usr_crt"`
	DtmCrt         time.Time `gorm:"column:DtmCrt" json:"dtm_crt"`
}

func (c *STG_CUST_FAM) TableName() string {
	return "STG_CUST_FAM"
}

type NewEntry struct {
	ProspectID       string    `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	BranchID         string    `gorm:"type:varchar(3);column:BranchID" json:"branch_id"`
	IDNumber         string    `gorm:"type:varchar(100);column:IDNumber" json:"id_number"`
	LegalName        string    `gorm:"type:varchar(100);column:LegalName" json:"legal_name"`
	BirthDate        string    `gorm:"column:BirthDate" json:"birth_date"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	CreatedByID      string    `gorm:"type:varchar(100);column:created_by_id" json:"created_by_id"`
	CreatedByName    string    `gorm:"type:varchar(200);column:created_by_name" json:"created_by_name"`
	PayloadNE        string    `gorm:"type:varchar(5000);column:payload_ne" json:"payload_ne"`
	PayloadFiltering string    `gorm:"type:varchar(1000);column:payload_filtering" json:"payload_filtering"`
	PayloadLTV       string    `gorm:"type:varchar(200);column:payload_ltv" json:"payload_ltv"`
	PayloadJourney   string    `gorm:"type:varchar(5000);column:payload_journey" json:"payload_journey"`
}

func (c *NewEntry) TableName() string {
	return "trx_new_entry"
}

type InquirySettingQuotaDeviasi struct {
	BranchID       string  `gorm:"column:BranchID" json:"branch_id"`
	BranchName     string  `gorm:"column:branch_name" json:"branch_name"`
	QuotaAmount    float64 `gorm:"column:quota_amount" json:"quota_amount"`
	QuotaAccount   int     `gorm:"column:quota_account" json:"quota_account"`
	BookingAmount  float64 `gorm:"column:booking_amount" json:"booking_amount"`
	BookingAccount int     `gorm:"column:booking_account" json:"booking_account"`
	BalanceAmount  float64 `gorm:"column:balance_amount" json:"balance_amount"`
	BalanceAccount int     `gorm:"column:balance_account" json:"balance_account"`
	IsActive       bool    `gorm:"column:is_active" json:"is_active"`
	UpdatedBy      string  `gorm:"column:updated_by" json:"updated_by"`
	UpdatedAt      string  `gorm:"column:updated_at" json:"updated_at"`
}

type DataQuotaDeviasiBranch struct {
	QuotaAmount    float64   `gorm:"column:quota_amount" json:"quota_amount"`
	QuotaAccount   int       `gorm:"column:quota_account" json:"quota_account"`
	BookingAmount  float64   `gorm:"column:booking_amount" json:"booking_amount"`
	BookingAccount int       `gorm:"column:booking_account" json:"booking_account"`
	BalanceAmount  float64   `gorm:"column:balance_amount" json:"balance_amount"`
	BalanceAccount int       `gorm:"column:balance_account" json:"balance_account"`
	IsActive       bool      `gorm:"column:is_active" json:"is_active"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy      string    `gorm:"column:updated_by" json:"updated_by"`
}

type ResultCheckDeviation struct {
	BranchID       string      `gorm:"column:BranchID"`
	NTF            float64     `gorm:"column:NTF"`
	CustomerStatus string      `gorm:"column:customer_status"`
	Decision       interface{} `gorm:"column:decision"`
}

type InquiryMappingCluster struct {
	BranchID       string `gorm:"column:branch_id" json:"branch_id"`
	BranchName     string `gorm:"column:branch_name" json:"branch_name"`
	CustomerStatus string `gorm:"column:customer_status" json:"customer_status"`
	BpkbNameType   int    `gorm:"column:bpkb_name_type" json:"bpkb_name_type"`
	Cluster        string `gorm:"column:cluster" json:"cluster"`
}

type HistoryConfigChanges struct {
	ID         string    `gorm:"type:varchar(50);column:id"`
	ConfigID   string    `gorm:"type:varchar(50);column:config_id"`
	ObjectName string    `gorm:"type:varchar(50);column:object_name"`
	Action     string    `gorm:"type:varchar(10);column:action"`
	DataBefore string    `gorm:"type:text;column:data_before"`
	DataAfter  string    `gorm:"type:text;column:data_after"`
	CreatedBy  string    `gorm:"type:varchar(20);column:created_by"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (c *HistoryConfigChanges) TableName() string {
	return "history_config_changes"
}

type ConfinsBranch struct {
	BranchID   string `gorm:"type:varchar(10);column:BranchID" json:"branch_id"`
	BranchName string `gorm:"type:varchar(200);column:BranchName" json:"branch_name"`
}

func (c *ConfinsBranch) TableName() string {
	return "confins_branch"
}

type MappingClusterChangeLog struct {
	ID         string `json:"id"`
	DataBefore string `json:"data_before"`
	DataAfter  string `json:"data_after"`
	UserName   string `json:"user_name"`
	CreatedAt  string `json:"created_at"`
}

type MasterMappingFpdCluster struct {
	Cluster     string    `gorm:"column:cluster"`
	FpdStartHte float64   `gorm:"column:fpd_start_hte"`
	FpdEndLt    float64   `gorm:"column:fpd_end_lt"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (c *MasterMappingFpdCluster) TableName() string {
	return "m_mapping_fpd_cluster"
}

type TrxCmoNoFPD struct {
	ProspectID              string    `gorm:"column:prospect_id;type:varchar(20)" json:"prospect_id"`
	CMOID                   string    `gorm:"column:cmo_id;type:varchar(20)" json:"cmo_id"`
	BPKBName                string    `gorm:"column:bpkb_name;type:varchar(9)" json:"bpkb_name"`
	CmoCategory             string    `gorm:"column:cmo_category;char(3)" json:"cmo_category"`
	CmoJoinDate             string    `gorm:"column:cmo_join_date" json:"cmo_join_date"`
	DefaultCluster          string    `gorm:"column:default_cluster;type:varchar(20)" json:"default_cluster"`
	DefaultClusterStartDate string    `gorm:"column:default_cluster_start_date" json:"default_cluster_start_date"`
	DefaultClusterEndDate   string    `gorm:"column:default_cluster_end_date" json:"default_cluster_end_date"`
	CreatedAt               time.Time `gorm:"column:created_at" json:"created_at"`
}

func (c *TrxCmoNoFPD) TableName() string {
	return "trx_cmo_no_fpd"
}

type MappingNegativeCustomer struct {
	IsActive    int    `gorm:"is_active"`
	IsBlacklist int    `gorm:"is_blacklist"`
	IsHighrisk  int    `gorm:"is_highrisk"`
	BadType     string `gorm:"bad_type"`
	Reason      string `gorm:"reason"`
	Decision    string `gorm:"decision"`
}

func (c *MappingNegativeCustomer) TableName() string {
	return "m_mapping_negative_customer"
}

type TrxEDD struct {
	ProspectID  string      `gorm:"type:varchar(20);column:ProspectID;primary_key:true" json:"-"`
	IsEDD       bool        `gorm:"-" json:"is_edd"`
	IsHighrisk  bool        `gorm:"column:is_highrisk;type:int" json:"is_highrisk"`
	Pernyataan1 interface{} `gorm:"column:pernyataan_1" json:"pernyataan_1"`
	Pernyataan2 interface{} `gorm:"column:pernyataan_2" json:"pernyataan_2"`
	Pernyataan3 interface{} `gorm:"column:pernyataan_3" json:"pernyataan_3"`
	Pernyataan4 interface{} `gorm:"column:pernyataan_4" json:"pernyataan_4"`
	Pernyataan5 interface{} `gorm:"column:pernyataan_5" json:"pernyataan_5"`
	Pernyataan6 interface{} `gorm:"column:pernyataan_6" json:"pernyataan_6"`
	CreatedAt   time.Time   `gorm:"column:created_at" json:"-"`
}

func (c *TrxEDD) TableName() string {
	return "trx_edd"
}

type InquiryDataListOrder struct {
	OrderAt        time.Time   `gorm:"column:OrderAt" json:"order_at"`
	BranchName     string      `gorm:"column:BranchName" json:"branch_name"`
	ProspectID     string      `gorm:"type:varchar(20);column:ProspectID" json:"prospect_id"`
	LegalName      string      `gorm:"column:LegalName" json:"legal_name"`
	IDNumber       string      `gorm:"column:IDNumber" json:"id_number"`
	BirthDate      time.Time   `gorm:"column:BirthDate" json:"birth_date"`
	Profession     string      `gorm:"column:Profession" json:"profession"`
	JobType        string      `gorm:"column:JobType" json:"job_type"`
	JobPosition    string      `gorm:"column:JobPosition" json:"job_position"`
	IsHighRisk     interface{} `gorm:"column:IsHighRisk;" json:"is_highrisk"`
	Pernyataan1    interface{} `gorm:"column:Pernyataan1" json:"pernyataan_1"`
	Pernyataan2    interface{} `gorm:"column:Pernyataan2" json:"pernyataan_2"`
	Pernyataan3    interface{} `gorm:"column:Pernyataan3" json:"pernyataan_3"`
	Pernyataan4    interface{} `gorm:"column:Pernyataan4" json:"pernyataan_4"`
	Pernyataan5    interface{} `gorm:"column:Pernyataan5" json:"pernyataan_5"`
	Pernyataan6    interface{} `gorm:"column:Pernyataan6" json:"pernyataan_6"`
	UrlFormAkkk    string      `gorm:"column:UrlFormAkkk" json:"url_form_akkk"`
	Decision       string      `gorm:"column:Decision" json:"decision"`
	SourceDecision string      `gorm:"column:SourceDecision" json:"source_decision"`
	RuleCode       string      `gorm:"column:RuleCode" json:"rule_code"`
	Reason         string      `gorm:"column:Reason" json:"reason"`
	DecisionBy     string      `gorm:"column:DecisionBy" json:"decision_by"`
	DecisionAt     time.Time   `gorm:"column:DecisionAt" json:"decision_at"`
}

type EncryptString struct {
	Encrypt string `json:"encrypt"`
}

type ConfirmDeviasi struct {
	NTF            float64 `gorm:"column:NTF" json:"NTF"`
	BranchID       string  `gorm:"type:varchar(10);column:BranchID" json:"branch_id"`
	FinalApproval  string  `gorm:"type:varchar(3);column:final_approval"`
	QuotaAmount    float64 `gorm:"column:quota_amount" json:"quota_amount"`
	QuotaAccount   int     `gorm:"column:quota_account" json:"quota_account"`
	BookingAmount  float64 `gorm:"column:booking_amount" json:"booking_amount"`
	BookingAccount int     `gorm:"column:booking_account" json:"booking_account"`
	BalanceAmount  float64 `gorm:"column:balance_amount" json:"balance_amount"`
	BalanceAccount int     `gorm:"column:balance_account" json:"balance_account"`
	IsActive       bool    `gorm:"column:is_active" json:"is_active"`
	Deviasi        bool    `gorm:"column:deviasi" json:"deviasi"`
}

type Deviasi struct {
	DeviasiID          string `json:"deviasi_id"`
	DeviasiDescription string `json:"deviasi_description"`
	DeviasiDecision    string `json:"deviasi_decision"`
	DeviasiReason      string `json:"deviasi_reason"`
}

type TrxDeviasi struct {
	ProspectID string    `gorm:"type:varchar(20);column:ProspectID;primary_key:true"`
	DeviasiID  string    `gorm:"type:varchar(20);column:deviasi_id"`
	Reason     string    `gorm:"type:varchar(255);column:reason"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"-"`
}

func (c *TrxDeviasi) TableName() string {
	return "trx_deviasi"
}

type MappingKodeDeviasi struct {
	DeviasiID string `gorm:"type:varchar(20);column:deviasi_id"`
	Deskripsi string `gorm:"type:varchar(255);column:deskripsi"`
}

func (c *MappingKodeDeviasi) TableName() string {
	return "m_kode_deviasi"
}

type MappingBranchDeviasi struct {
	BranchID       string    `gorm:"type:varchar(10);column:BranchID" json:"branch_id"`
	FinalApproval  string    `gorm:"type:varchar(3);column:final_approval" json:"final_approval"`
	QuotaAmount    float64   `gorm:"column:quota_amount" json:"quota_amount"`
	QuotaAccount   int       `gorm:"column:quota_account" json:"quota_account"`
	BookingAmount  float64   `gorm:"column:booking_amount" json:"booking_amount"`
	BookingAccount int       `gorm:"column:booking_account" json:"booking_account"`
	BalanceAmount  float64   `gorm:"column:balance_amount" json:"balance_amount"`
	BalanceAccount int       `gorm:"column:balance_account" json:"balance_account"`
	IsActive       bool      `gorm:"column:is_active" json:"is_active"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy      string    `gorm:"type:varchar(255);column:updated_by" json:"updated_by"`
}

func (c *MappingBranchDeviasi) TableName() string {
	return "m_branch_deviasi"
}

type MasterMappingDeviasiDSR struct {
	TotalIncomeStart float64 `gorm:"column:total_income_start"`
	TotalIncomeEnd   float64 `gorm:"column:total_income_end"`
	DSRThreshold     float64 `gorm:"column:dsr_threshold"`
}

func (c *MasterMappingDeviasiDSR) TableName() string {
	return "m_mapping_deviasi_dsr"
}

type MappingVehicleAge struct {
	VehicleAgeStart int         `gorm:"column:vehicle_age_start"`
	VehicleAgeEnd   int         `gorm:"column:vehicle_age_end"`
	Cluster         string      `gorm:"type:varchar(50);column:cluster"`
	BPKBNameType    int         `gorm:"column:bpkb_name_type"`
	TenorStart      int         `gorm:"column:tenor_start"`
	TenorEnd        int         `gorm:"column:tenor_end"`
	Decision        string      `gorm:"type:varchar(20);column:decision"`
	CreatedAt       time.Time   `gorm:"type:datetime2(2);column:created_at"`
	Info            interface{} `gorm:"column:info"`
}

func (c *MappingVehicleAge) TableName() string {
	return "m_mapping_vehicle_age"
}

type MasterMappingIncomeMaxDSR struct {
	TotalIncomeStart float64 `gorm:"column:total_income_start"`
	TotalIncomeEnd   float64 `gorm:"column:total_income_end"`
	DSRThreshold     float64 `gorm:"column:dsr_threshold"`
}

func (c *MasterMappingIncomeMaxDSR) TableName() string {
	return "kmb_mapping_income_dsr"
}

type DraftPrinciple struct {
	ProspectID         string      `gorm:"column:type:varchar(20);ProspectID"`
	IDNumber           string      `gorm:"column:type:varchar(16);IDNumber"`
	SpouseIDNumber     interface{} `gorm:"column:type:varchar(16);SpouseIDNumber"`
	ManufactureYear    string      `gorm:"column:type:varchar(4);ManufactureYear"`
	NoChassis          string      `gorm:"column:type:varchar(30);ChassisNumber"`
	NoEngine           string      `gorm:"column:type:varchar(30);EngineNumber"`
	BranchID           string      `gorm:"column:type:varchar(10);BranchID"`
	CMOID              string      `gorm:"column:type:varchar(20);CmoID"`
	CMOName            string      `gorm:"column:type:varchar(50);CmoName"`
	CC                 string      `gorm:"column:type:varchar(10);CC"`
	TaxDate            time.Time   `gorm:"column:TaxDate"`
	STNKExpiredDate    time.Time   `gorm:"column:STNKExpiredDate"`
	OwnerAsset         string      `gorm:"column:type:varchar(50);OwnerAsset"`
	LicensePlate       string      `gorm:"column:type:varchar(50);LicensePlate"`
	Color              string      `gorm:"column:type:varchar(50);Color"`
	Brand              string      `gorm:"column:type:varchar(255);Brand"`
	ResidenceAddress   string      `gorm:"column:type:varchar(100);ResidenceAddress"`
	ResidenceRT        string      `gorm:"column:type:varchar(3);ResidenceRT"`
	ResidenceRW        string      `gorm:"column:type:varchar(3);ResidenceRW"`
	ResidenceProvice   string      `gorm:"column:type:varchar(50);ResidenceProvice"`
	ResidenceCity      string      `gorm:"column:type:varchar(30);ResidenceCity"`
	ResidenceKecamatan string      `gorm:"column:type:varchar(30);ResidenceKecamatan"`
	ResidenceKelurahan string      `gorm:"column:type:varchar(30);ResidenceKelurahan"`
	ResidenceZipCode   string      `gorm:"column:type:varchar(5);ResidenceZipCode"`
	ResidenceAreaPhone string      `gorm:"column:type:varchar(4);ResidenceAreaPhone"`
	ResidencePhone     string      `gorm:"column:type:varchar(10);ResidencePhone"`
	HomeStatus         string      `gorm:"column:type:varchar(2);HomeStatus"`
	StaySinceYear      int         `gorm:"column:type:varchar(4);StaySinceYear"`
	StaySinceMonth     int         `gorm:"column:type:varchar(2);StaySinceMonth"`
	Decision           string      `gorm:"column:type:varchar(20);Decision"`
	Reason             string      `gorm:"column:type:varchar(255);Reason"`
	BPKBName           string      `gorm:"column:type:varchar(2);BPKBName"`
	CreatedAt          time.Time   `gorm:"column:created_at"`
}

type TrxPrincipleStepOne struct {
	ProspectID         string      `gorm:"column:ProspectID;type:varchar(20);"`
	IDNumber           string      `gorm:"column:IDNumber;type:varchar(16)"`
	SpouseIDNumber     interface{} `gorm:"column:SpouseIDNumber;type:varchar(16)"`
	ManufactureYear    string      `gorm:"column:ManufactureYear;type:varchar(4)"`
	NoChassis          string      `gorm:"column:ChassisNumber;type:varchar(30)"`
	NoEngine           string      `gorm:"column:EngineNumber;type:varchar(30)"`
	BranchID           string      `gorm:"column:BranchID;type:varchar(10)"`
	CMOID              string      `gorm:"column:CmoID;type:varchar(20)"`
	CMOName            string      `gorm:"column:CmoName;type:varchar(50)"`
	CC                 string      `gorm:"column:CC;type:varchar(10)"`
	TaxDate            time.Time   `gorm:"column:TaxDate"`
	STNKExpiredDate    time.Time   `gorm:"column:STNKExpiredDate"`
	OwnerAsset         string      `gorm:"column:OwnerAsset;type:varchar(50)"`
	LicensePlate       string      `gorm:"column:LicensePlate;type:varchar(50)"`
	Color              string      `gorm:"column:Color;type:varchar(50);"`
	Brand              string      `gorm:"column:Brand;type:varchar(255)"`
	ResidenceAddress   string      `gorm:"column:ResidenceAddress;type:varchar(100);"`
	ResidenceRT        string      `gorm:"column:ResidenceRT;type:varchar(3)"`
	ResidenceRW        string      `gorm:"column:ResidenceRW;type:varchar(3)"`
	ResidenceProvince  string      `gorm:"column:ResidenceProvince;type:varchar(50)"`
	ResidenceCity      string      `gorm:"column:ResidenceCity;type:varchar(30)"`
	ResidenceKecamatan string      `gorm:"column:ResidenceKecamatan;type:varchar(30)"`
	ResidenceKelurahan string      `gorm:"column:ResidenceKelurahan;type:varchar(30)"`
	ResidenceZipCode   string      `gorm:"column:ResidenceZipCode;type:varchar(5)"`
	ResidenceAreaPhone string      `gorm:"column:ResidenceAreaPhone;type:varchar(4)"`
	ResidencePhone     string      `gorm:"column:ResidencePhone;type:varchar(10)"`
	HomeStatus         string      `gorm:"column:HomeStatus;type:varchar(2)"`
	StaySinceYear      int         `gorm:"column:StaySinceYear;type:varchar(4)"`
	StaySinceMonth     int         `gorm:"column:StaySinceMonth;type:varchar(2)"`
	Decision           string      `gorm:"column:Decision;type:varchar(20)"`
	RuleCode           string      `gorm:"column:RuleCode;type:varchar(10);"`
	Reason             string      `gorm:"column:Reason;type:varchar(255)"`
	BPKBName           string      `gorm:"column:BPKBName;type:varchar(2)"`
	AssetCode          string      `gorm:"column:AssetCode;type:varchar(200)"`
	STNKPhoto          interface{} `gorm:"column:STNKPhoto;type:varchar(250);"`
	KPMID              int         `gorm:"column:KPMID;"`
	CreatedAt          time.Time   `gorm:"column:created_at"`
}

func (c *TrxPrincipleStepOne) TableName() string {
	return "trx_principle_step_one"
}

type TrxPrincipleStepTwo struct {
	ProspectID              string      `gorm:"column:ProspectID;type:varchar(20);"`
	IDNumber                string      `gorm:"column:IDNumber;type:varchar(16)"`
	LegalName               string      `gorm:"column:LegalName;type:varchar(100);"`
	MobilePhone             string      `gorm:"column:MobilePhone;type:varchar(20);"`
	Email                   string      `gorm:"column:Email;type:varchar(100);"`
	FullName                string      `gorm:"column:FullName;type:varchar(100);"`
	BirthDate               time.Time   `gorm:"column:BirthDate"`
	BirthPlace              string      `gorm:"column:BirthPlace;type:varchar(100);"`
	SurgateMotherName       string      `gorm:"column:SurgateMotherName;type:varchar(100);"`
	Gender                  string      `gorm:"column:Gender;type:varchar(10);"`
	Religion                string      `gorm:"column:Religion;type:varchar(10);"`
	SpouseIDNumber          interface{} `gorm:"column:SpouseIDNumber;type:varchar(16)"`
	LegalAddress            string      `gorm:"column:LegalAddress;type:varchar(100);"`
	LegalRT                 string      `gorm:"column:LegalRT;type:varchar(3);"`
	LegalRW                 string      `gorm:"column:LegalRW;type:varchar(3);"`
	LegalProvince           string      `gorm:"column:LegalProvince;type:varchar(50)"`
	LegalCity               string      `gorm:"column:LegalCity;type:varchar(30);"`
	LegalKecamatan          string      `gorm:"column:LegalKecamatan;type:varchar(30);"`
	LegalKelurahan          string      `gorm:"column:LegalKelurahan;type:varchar(30);"`
	LegalZipCode            string      `gorm:"column:LegalZipCode;type:varchar(5)"`
	LegalAreaPhone          string      `gorm:"column:LegalAreaPhone;type:varchar(4)"`
	LegalPhone              string      `gorm:"column:LegalPhone;type:varchar(10)"`
	CompanyName             string      `gorm:"column:CompanyName;type:varchar(50);"`
	CompanyAddress          string      `gorm:"column:CompanyAddress;type:varchar(100);"`
	CompanyRT               string      `gorm:"column:CompanyRT;type:varchar(3);"`
	CompanyRW               string      `gorm:"column:CompanyRW;type:varchar(3);"`
	CompanyProvince         string      `gorm:"column:CompanyProvince;type:varchar(50)"`
	CompanyCity             string      `gorm:"column:CompanyCity;type:varchar(30);"`
	CompanyKecamatan        string      `gorm:"column:CompanyKecamatan;type:varchar(30);"`
	CompanyKelurahan        string      `gorm:"column:CompanyKelurahan;type:varchar(30);"`
	CompanyZipCode          string      `gorm:"column:CompanyZipCode;type:varchar(5)"`
	CompanyAreaPhone        string      `gorm:"column:CompanyAreaPhone;type:varchar(4)"`
	CompanyPhone            string      `gorm:"column:CompanyPhone;type:varchar(10)"`
	MonthlyFixedIncome      float64     `gorm:"column:MonthlyFixedIncome"`
	MaritalStatus           string      `gorm:"column:MaritalStatus;type:varchar(10);"`
	SpouseIncome            interface{} `gorm:"column:SpouseIncome"`
	SelfiePhoto             interface{} `gorm:"column:SelfiePhoto;type:varchar(250);"`
	KtpPhoto                interface{} `gorm:"column:KtpPhoto;type:varchar(250);"`
	SpouseFullName          interface{} `gorm:"column:SpouseFullName;type:varchar(100);"`
	SpouseBirthDate         interface{} `gorm:"column:SpouseBirthDate"`
	SpouseBirthPlace        interface{} `gorm:"column:SpouseBirthPlace;type:varchar(100);"`
	SpouseGender            interface{} `gorm:"column:SpouseGender;type:varchar(10);"`
	SpouseLegalName         interface{} `gorm:"column:SpouseLegalName;type:varchar(100);"`
	SpouseMobilePhone       interface{} `gorm:"column:SpouseMobilePhone;type:varchar(20);"`
	SpouseSurgateMotherName interface{} `gorm:"column:SpouseSurgateMotherName;type:varchar(100);"`
	EconomySectorID         string      `gorm:"column:EconomySectorID;type:varchar(10)"`
	Education               string      `gorm:"column:Education;type:varchar(50);"`
	EmploymentSinceMonth    int         `gorm:"column:EmploymentSinceMonth;type:varchar(2);"`
	EmploymentSinceYear     int         `gorm:"column:EmploymentSinceYear;type:varchar(4);"`
	IndustryTypeID          string      `gorm:"column:IndustryTypeID;type:varchar(10);"`
	JobPosition             string      `gorm:"column:JobPosition;type:varchar(10);"`
	JobType                 string      `gorm:"column:JobType;type:varchar(10);"`
	ProfessionID            string      `gorm:"column:ProfessionID;type:varchar(10);"`
	CheckBannedPMKDSRResult interface{} `gorm:"column:CheckBannedPMKDSRResult;type:varchar(50);"`
	CheckBannedPMKDSRCode   interface{} `gorm:"column:CheckBannedPMKDSRCode;type:varchar(50);"`
	CheckBannedPMKDSRReason interface{} `gorm:"column:CheckBannedPMKDSRReason;type:varchar(200);"`
	CheckRejectionResult    interface{} `gorm:"column:CheckRejectionResult;type:varchar(50);"`
	CheckRejectionCode      interface{} `gorm:"column:CheckRejectionCode;type:varchar(50);"`
	CheckRejectionReason    interface{} `gorm:"column:CheckRejectionReason;type:varchar(200);"`
	CheckBlacklistResult    interface{} `gorm:"column:CheckBlacklistResult;type:varchar(50);"`
	CheckBlacklistCode      interface{} `gorm:"column:CheckBlacklistCode;type:varchar(50);"`
	CheckBlacklistReason    interface{} `gorm:"column:CheckBlacklistReason;type:varchar(200);"`
	CheckPMKResult          interface{} `gorm:"column:CheckPMKResult;type:varchar(50);"`
	CheckPMKCode            interface{} `gorm:"column:CheckPMKCode;type:varchar(50);"`
	CheckPMKReason          interface{} `gorm:"column:CheckPMKReason;type:varchar(200);"`
	CheckEkycResult         interface{} `gorm:"column:CheckEkycResult;type:varchar(50);"`
	CheckEkycCode           interface{} `gorm:"column:CheckEkycCode;type:varchar(50);"`
	CheckEkycReason         interface{} `gorm:"column:CheckEkycReason;type:varchar(200);"`
	CheckEkycSource         interface{} `gorm:"column:CheckEkycSource;type:varchar(5);"`
	CheckEkycInfo           interface{} `gorm:"column:CheckEkycInfo;type:text;"`
	CheckEkycSimiliarity    interface{} `gorm:"column:CheckEkycSimiliarity;type:float;"`
	FilteringResult         interface{} `gorm:"column:FilteringResult;type:varchar(50);"`
	FilteringCode           interface{} `gorm:"column:FilteringCode;type:varchar(50);"`
	FilteringReason         interface{} `gorm:"column:FilteringReason;type:varchar(200);"`
	Decision                string      `gorm:"column:Decision;type:varchar(20)"`
	Reason                  string      `gorm:"column:Reason;type:varchar(255)"`
	RuleCode                string      `gorm:"column:RuleCode;type:varchar(10);"`
	CreatedAt               time.Time   `gorm:"column:created_at"`
	DupcheckData            string      `gorm:"column:DupcheckData;type:text"`
}

func (c *TrxPrincipleStepTwo) TableName() string {
	return "trx_principle_step_two"
}

type TrxPrincipleStepThree struct {
	ProspectID               string      `gorm:"column:ProspectID;type:varchar(20);"`
	IDNumber                 string      `gorm:"column:IDNumber;type:varchar(16)"`
	Tenor                    int         `gorm:"column:Tenor"`
	AF                       float64     `gorm:"column:AF"`
	NTF                      float64     `gorm:"column:NTF"`
	OTR                      float64     `gorm:"column:OTR"`
	DPAmount                 float64     `gorm:"column:DPAmount"`
	AdminFee                 float64     `gorm:"column:AdminFee"`
	InstallmentAmount        float64     `gorm:"column:InstallmentAmount"`
	Dealer                   string      `gorm:"column:Dealer;type:varchar(50);"`
	MonthlyVariableIncome    float64     `gorm:"column:MonthlyVariableIncome"`
	AssetCategoryID          string      `gorm:"column:AssetCategoryID;type:varchar(100);"`
	FinancePurpose           string      `gorm:"column:FinancePurpose;type:varchar(100);"`
	TipeUsaha                string      `gorm:"column:TipeUsaha;type:varchar(100);"`
	CheckVehicleResult       interface{} `gorm:"column:CheckVehicleResult;type:varchar(50);"`
	CheckVehicleCode         interface{} `gorm:"column:CheckVehicleCode;type:varchar(50);"`
	CheckVehicleReason       interface{} `gorm:"column:CheckVehicleReason;type:varchar(200);"`
	CheckVehicleInfo         interface{} `gorm:"column:CheckVehicleInfo;type:text;"`
	CheckRejectTenor36Result interface{} `gorm:"column:CheckRejectTenor36Result;type:varchar(50);"`
	CheckRejectTenor36Code   interface{} `gorm:"column:CheckRejectTenor36Code;type:varchar(50);"`
	CheckRejectTenor36Reason interface{} `gorm:"column:CheckRejectTenor36Reason;type:varchar(200);"`
	ScoreProResult           interface{} `gorm:"column:ScoreProResult;type:varchar(50);"`
	ScoreProCode             interface{} `gorm:"column:ScoreProCode;type:varchar(50);"`
	ScoreProReason           interface{} `gorm:"column:ScoreProReason;type:varchar(200);"`
	ScoreProInfo             interface{} `gorm:"column:ScoreProInfo;type:text;"`
	ScoreProScoreResult      interface{} `gorm:"column:ScoreProScoreResult;type:varchar(20);"`
	CheckDSRResult           interface{} `gorm:"column:CheckDSRResult;type:varchar(50);"`
	CheckDSRCode             interface{} `gorm:"column:CheckDSRCode;type:varchar(50);"`
	CheckDSRReason           interface{} `gorm:"column:CheckDSRReason;type:varchar(200);"`
	CheckDSRFMFPBKResult     interface{} `gorm:"column:CheckDSRFMFPBKResult;type:varchar(50);"`
	CheckDSRFMFPBKCode       interface{} `gorm:"column:CheckDSRFMFPBKCode;type:varchar(50);"`
	CheckDSRFMFPBKReason     interface{} `gorm:"column:CheckDSRFMFPBKReason;type:varchar(200);"`
	CheckDSRFMFPBKInfo       interface{} `gorm:"column:CheckDSRFMFPBKInfo;type:text;"`
	Decision                 string      `gorm:"column:Decision;type:varchar(20)"`
	Reason                   string      `gorm:"column:Reason;type:varchar(255)"`
	RuleCode                 string      `gorm:"column:RuleCode;type:varchar(10);"`
	CreatedAt                time.Time   `gorm:"column:created_at"`
}

func (c *TrxPrincipleStepThree) TableName() string {
	return "trx_principle_step_three"
}

type TrxPrincipleStatus struct {
	ProspectID string    `gorm:"column:ProspectID;type:varchar(20)"`
	IDNumber   string    `gorm:"column:IDNumber;type:varchar(16);"`
	Step       int       `gorm:"column:Step"`
	Decision   string    `gorm:"column:Decision;type:varchar(20);"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (c *TrxPrincipleStatus) TableName() string {
	return "trx_principle_status"
}

type TrxPrincipleEmergencyContact struct {
	ProspectID   string    `gorm:"column:ProspectID;type:varchar(20);"`
	Name         string    `gorm:"column:Name;type:varchar(200);"`
	Relationship string    `gorm:"column:Relationship;type:varchar(10);"`
	MobilePhone  string    `gorm:"column:MobilePhone;type:varchar(20);"`
	Address      string    `gorm:"column:Address;type:varchar(255);"`
	Rt           string    `gorm:"column:RT;type:varchar(3);"`
	Rw           string    `gorm:"column:RW;type:varchar(3);"`
	Kelurahan    string    `gorm:"column:Kelurahan;type:varchar(30);"`
	Kecamatan    string    `gorm:"column:Kecamatan;type:varchar(30);"`
	City         string    `gorm:"column:City;type:varchar(30);"`
	Province     string    `gorm:"column:Province;type:varchar(30);"`
	ZipCode      string    `gorm:"column:ZipCode;type:varchar(5);"`
	AreaPhone    string    `gorm:"column:AreaPhone;type:varchar(5);"`
	Phone        string    `gorm:"column:Phone;type:varchar(20);"`
	CustomerID   int       `gorm:"column:CustomerID;"`
	KPMID        int       `gorm:"column:KPMID;"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (c *TrxPrincipleEmergencyContact) TableName() string {
	return "trx_principle_emergency_contact"
}

type TrxPrincipleMarketingProgram struct {
	ProspectID                 string    `gorm:"column:ProspectID;type:varchar(20);"`
	ProgramID                  string    `gorm:"column:ProgramID;type:varchar(50);"`
	ProgramName                string    `gorm:"column:ProgramName;type:varchar(200);"`
	ProductOfferingID          string    `gorm:"column:ProductOfferingID;type:varchar(50);"`
	ProductOfferingDescription string    `gorm:"column:ProductOfferingDescription;type:varchar(255);"`
	LoanAmount                 float64   `gorm:"column:LoanAmount"`
	LoanAmountMaximum          float64   `gorm:"column:LoanAmountMaximum"`
	AdminFee                   float64   `gorm:"column:AdminFee"`
	ProvisionFee               float64   `gorm:"column:ProvisionFee"`
	DPAmount                   float64   `gorm:"column:DPAmount"`
	FinanceAmount              float64   `gorm:"column:FinanceAmount"`
	InstallmentAmount          float64   `gorm:"column:InstallmentAmount"`
	NTF                        float64   `gorm:"column:NTF"`
	OTR                        float64   `gorm:"column:OTR"`
	Dealer                     string    `gorm:"column:Dealer;type:varchar(50);"`
	AssetCategoryID            string    `gorm:"column:AssetCategoryID;type:varchar(100);"`
	CreatedAt                  time.Time `gorm:"column:created_at"`
}

func (c *TrxPrincipleMarketingProgram) TableName() string {
	return "trx_principle_marketing_program"
}

type AutoCancel struct {
	ProspectID string `gorm:"column:ProspectID;type:varchar(20)"`
	AssetCode  string `gorm:"column:AssetCode;type:varchar(200)"`
	KPMID      int    `gorm:"column:KPMID;"`
	BranchID   string `gorm:"column:BranchID;type:varchar(10)"`
}

type TrxPrincipleError struct {
	ProspectID string    `gorm:"column:ProspectID;type:varchar(20)"`
	KpmId      int       `gorm:"column:KpmID;type:varchar(20)"`
	Step       int       `gorm:"column:Step"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (c *TrxPrincipleError) TableName() string {
	return "trx_principle_error"
}

type TrxKPM struct {
	ID                          string      `gorm:"column:id;type:varchar(50)"`
	ProspectID                  string      `gorm:"column:ProspectID;type:varchar(20)"`
	IDNumber                    string      `gorm:"column:IDNumber;type:varchar(200)"`
	LegalName                   string      `gorm:"column:LegalName;type:varchar(200);"`
	MobilePhone                 string      `gorm:"column:MobilePhone;type:varchar(200);"`
	Email                       string      `gorm:"column:Email;type:varchar(200);"`
	BirthPlace                  string      `gorm:"column:BirthPlace;type:varchar(200);"`
	BirthDate                   time.Time   `gorm:"column:BirthDate"`
	SurgateMotherName           string      `gorm:"column:SurgateMotherName;type:varchar(200);"`
	Gender                      string      `gorm:"column:Gender;type:varchar(10);"`
	ResidenceAddress            string      `gorm:"column:ResidenceAddress;type:varchar(200);"`
	ResidenceRT                 string      `gorm:"column:ResidenceRT;type:varchar(3)"`
	ResidenceRW                 string      `gorm:"column:ResidenceRW;type:varchar(3)"`
	ResidenceProvince           string      `gorm:"column:ResidenceProvince;type:varchar(50)"`
	ResidenceCity               string      `gorm:"column:ResidenceCity;type:varchar(30)"`
	ResidenceKecamatan          string      `gorm:"column:ResidenceKecamatan;type:varchar(30)"`
	ResidenceKelurahan          string      `gorm:"column:ResidenceKelurahan;type:varchar(30)"`
	ResidenceZipCode            string      `gorm:"column:ResidenceZipCode;type:varchar(5)"`
	BranchID                    string      `gorm:"column:BranchID;type:varchar(10)"`
	AssetCode                   string      `gorm:"column:AssetCode;type:varchar(200)"`
	ManufactureYear             string      `gorm:"column:ManufactureYear;type:varchar(4)"`
	LicensePlate                string      `gorm:"column:LicensePlate;type:varchar(50)"`
	AssetUsageTypeCode          string      `gorm:"column:AssetUsageTypeCode;type:varchar(2)"`
	BPKBName                    string      `gorm:"column:BPKBName;type:varchar(2)"`
	OwnerAsset                  string      `gorm:"column:OwnerAsset;type:varchar(50)"`
	LoanAmount                  float64     `gorm:"column:LoanAmount"`
	MaxLoanAmount               float64     `gorm:"column:MaxLoanAmount"`
	Tenor                       int         `gorm:"column:Tenor"`
	InstallmentAmount           float64     `gorm:"column:InstallmentAmount"`
	NumOfDependence             int         `gorm:"column:NumOfDependence"`
	MaritalStatus               string      `gorm:"column:MaritalStatus;type:varchar(10);"`
	SpouseIDNumber              interface{} `gorm:"column:SpouseIDNumber;type:varchar(16)"`
	SpouseLegalName             interface{} `gorm:"column:SpouseLegalName;type:varchar(100);"`
	SpouseBirthDate             interface{} `gorm:"column:SpouseBirthDate"`
	SpouseBirthPlace            interface{} `gorm:"column:SpouseBirthPlace;type:varchar(100);"`
	SpouseSurgateMotherName     interface{} `gorm:"column:SpouseSurgateMotherName;type:varchar(100);"`
	SpouseMobilePhone           interface{} `gorm:"column:SpouseMobilePhone;type:varchar(20);"`
	Education                   string      `gorm:"column:Education;type:varchar(50);"`
	ProfessionID                string      `gorm:"column:ProfessionID;type:varchar(10);"`
	JobType                     string      `gorm:"column:JobType;type:varchar(10);"`
	JobPosition                 string      `gorm:"column:JobPosition;type:varchar(10);"`
	EmploymentSinceMonth        int         `gorm:"column:EmploymentSinceMonth;type:varchar(2);"`
	EmploymentSinceYear         int         `gorm:"column:EmploymentSinceYear;type:varchar(4);"`
	MonthlyFixedIncome          float64     `gorm:"column:MonthlyFixedIncome"`
	SpouseIncome                interface{} `gorm:"column:SpouseIncome"`
	NoChassis                   string      `gorm:"column:ChassisNumber;type:varchar(30)"`
	HomeStatus                  string      `gorm:"column:HomeStatus;type:varchar(2)"`
	StaySinceYear               int         `gorm:"column:StaySinceYear;type:varchar(4)"`
	StaySinceMonth              int         `gorm:"column:StaySinceMonth;type:varchar(2)"`
	KtpPhoto                    interface{} `gorm:"column:KtpPhoto;type:varchar(250);"`
	SelfiePhoto                 interface{} `gorm:"column:SelfiePhoto;type:varchar(250);"`
	AF                          float64     `gorm:"column:AF"`
	NTF                         float64     `gorm:"column:NTF"`
	OTR                         float64     `gorm:"column:OTR"`
	DPAmount                    float64     `gorm:"column:DPAmount"`
	AdminFee                    float64     `gorm:"column:AdminFee"`
	Dealer                      string      `gorm:"column:Dealer;type:varchar(50);"`
	AssetCategoryID             string      `gorm:"column:AssetCategoryID;type:varchar(100);"`
	KPMID                       int         `gorm:"column:KPMID;"`
	RentFinishDate              interface{} `gorm:"column:RentFinishDate"`
	ReferralCode                interface{} `gorm:"column:ReferralCode"`
	CheckNokaNosinResult        interface{} `gorm:"column:CheckNokaNosinResult;type:varchar(50);"`
	CheckNokaNosinCode          interface{} `gorm:"column:CheckNokaNosinCode;type:varchar(50);"`
	CheckNokaNosinReason        interface{} `gorm:"column:CheckNokaNosinReason;type:varchar(200);"`
	CheckNegativeCustomerResult interface{} `gorm:"column:CheckNegativeCustomerResult;type:varchar(50);"`
	CheckNegativeCustomerCode   interface{} `gorm:"column:CheckNegativeCustomerCode;type:varchar(50);"`
	CheckNegativeCustomerReason interface{} `gorm:"column:CheckNegativeCustomerReason;type:varchar(200);"`
	CheckNegativeCustomerInfo   interface{} `gorm:"column:CheckNegativeCustomerInfo;type:text;"`
	CheckMobilePhoneFMFResult   interface{} `gorm:"column:CheckMobilePhoneFMFResult;type:varchar(50);"`
	CheckMobilePhoneFMFCode     interface{} `gorm:"column:CheckMobilePhoneFMFCode;type:varchar(50);"`
	CheckMobilePhoneFMFReason   interface{} `gorm:"column:CheckMobilePhoneFMFReason;type:varchar(200);"`
	CheckMobilePhoneFMFInfo     interface{} `gorm:"column:CheckMobilePhoneFMFInfo;type:text;"`
	CheckBlacklistResult        interface{} `gorm:"column:CheckBlacklistResult;type:varchar(50);"`
	CheckBlacklistCode          interface{} `gorm:"column:CheckBlacklistCode;type:varchar(50);"`
	CheckBlacklistReason        interface{} `gorm:"column:CheckBlacklistReason;type:varchar(200);"`
	CheckPMKResult              interface{} `gorm:"column:CheckPMKResult;type:varchar(50);"`
	CheckPMKCode                interface{} `gorm:"column:CheckPMKCode;type:varchar(50);"`
	CheckPMKReason              interface{} `gorm:"column:CheckPMKReason;type:varchar(200);"`
	CheckEkycResult             interface{} `gorm:"column:CheckEkycResult;type:varchar(50);"`
	CheckEkycCode               interface{} `gorm:"column:CheckEkycCode;type:varchar(50);"`
	CheckEkycReason             interface{} `gorm:"column:CheckEkycReason;type:varchar(200);"`
	CheckEkycSource             interface{} `gorm:"column:CheckEkycSource;type:varchar(5);"`
	CheckEkycInfo               interface{} `gorm:"column:CheckEkycInfo;type:text;"`
	CheckEkycSimiliarity        interface{} `gorm:"column:CheckEkycSimiliarity;type:float;"`
	FilteringResult             interface{} `gorm:"column:FilteringResult;type:varchar(50);"`
	FilteringCode               interface{} `gorm:"column:FilteringCode;type:varchar(50);"`
	FilteringReason             interface{} `gorm:"column:FilteringReason;type:varchar(200);"`
	CheckVehicleResult          interface{} `gorm:"column:CheckVehicleResult;type:varchar(50);"`
	CheckVehicleCode            interface{} `gorm:"column:CheckVehicleCode;type:varchar(50);"`
	CheckVehicleReason          interface{} `gorm:"column:CheckVehicleReason;type:varchar(200);"`
	CheckVehicleInfo            interface{} `gorm:"column:CheckVehicleInfo;type:text;"`
	ScoreProResult              interface{} `gorm:"column:ScoreProResult;type:varchar(50);"`
	ScoreProCode                interface{} `gorm:"column:ScoreProCode;type:varchar(50);"`
	ScoreProReason              interface{} `gorm:"column:ScoreProReason;type:varchar(200);"`
	ScoreProInfo                interface{} `gorm:"column:ScoreProInfo;type:text;"`
	ScoreProScoreResult         interface{} `gorm:"column:ScoreProScoreResult;type:varchar(20);"`
	CheckDSRResult              interface{} `gorm:"column:CheckDSRResult;type:varchar(50);"`
	CheckDSRCode                interface{} `gorm:"column:CheckDSRCode;type:varchar(50);"`
	CheckDSRReason              interface{} `gorm:"column:CheckDSRReason;type:varchar(200);"`
	CheckDSRFMFPBKResult        interface{} `gorm:"column:CheckDSRFMFPBKResult;type:varchar(50);"`
	CheckDSRFMFPBKCode          interface{} `gorm:"column:CheckDSRFMFPBKCode;type:varchar(50);"`
	CheckDSRFMFPBKReason        interface{} `gorm:"column:CheckDSRFMFPBKReason;type:varchar(200);"`
	CheckDSRFMFPBKInfo          interface{} `gorm:"column:CheckDSRFMFPBKInfo;type:text;"`
	DupcheckData                string      `gorm:"column:DupcheckData;type:text"`
	NegativeCustomerData        string      `gorm:"column:NegativeCustomerData;type:text"`
	ResultPefindo               string      `gorm:"column:ResultPefindo;type:varchar(20)"`
	BakiDebet                   float64     `gorm:"column:BakiDebet"`
	Decision                    string      `gorm:"column:Decision;type:varchar(20)"`
	Reason                      string      `gorm:"column:Reason;type:varchar(255)"`
	RuleCode                    string      `gorm:"column:RuleCode;type:varchar(10);"`
	ReadjustContext             string      `gorm:"column:ReadjustContext;type:varchar(50);"`
	CreatedAt                   time.Time   `gorm:"column:created_at"`
	CreatedBy                   string      `gorm:"column:created_by;type:varchar(100);"`
	UpdatedAt                   time.Time   `gorm:"column:updated_at"`
	UpdatedBy                   string      `gorm:"column:updated_by;type:varchar(100);"`
	DeletedAt                   *time.Time  `gorm:"column:deleted_at;type:datetime2"`
	DeletedBy                   string      `gorm:"column:deleted_by;type:varchar(100);"`
}

func (c *TrxKPM) TableName() string {
	return "trx_kpm"
}

type TrxKPMStatus struct {
	ID         string     `gorm:"column:id;type:varchar(50)"`
	ProspectID string     `gorm:"column:ProspectID;type:varchar(20)"`
	Decision   string     `gorm:"column:Decision;type:varchar(20);"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	CreatedBy  string     `gorm:"column:created_by;type:varchar(100);"`
	UpdatedAt  time.Time  `gorm:"column:updated_at"`
	UpdatedBy  string     `gorm:"column:updated_by;type:varchar(100);"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;type:datetime2"`
	DeletedBy  string     `gorm:"column:deleted_by;type:varchar(100);"`
}

func (c *TrxKPMStatus) TableName() string {
	return "trx_kpm_status"
}

type TrxKPMError struct {
	ProspectID string    `gorm:"column:ProspectID;type:varchar(20)"`
	KpmId      int       `gorm:"column:KpmID;type:varchar(20)"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (c *TrxKPMError) TableName() string {
	return "trx_kpm_error"
}

type TrxKPMStatusHistory struct {
	ID         string    `gorm:"column:id;type:varchar(50)"`
	ProspectID string    `gorm:"column:ProspectID;type:varchar(20)"`
	Decision   string    `gorm:"column:Decision;type:varchar(20);"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}
