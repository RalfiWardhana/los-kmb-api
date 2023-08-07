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

type FilteringKMB struct {
	ProspectID                      string      `gorm:"column:prospect_id;type:varchar(20)" json:"prospect_id"`
	RequestID                       interface{} `gorm:"column:request_id;type:varchar(100)" json:"request_id"`
	BpkbName                        string      `gorm:"column:bpkb_name;type:varchar(2)" json:"bpkb_name"`
	BranchID                        string      `gorm:"column:branch_id;type:varchar(5)" json:"branch_id"`
	Decision                        string      `gorm:"column:decision;type:varchar(20)" json:"decision"`
	CustomerStatus                  interface{} `gorm:"column:customer_status;type:varchar(5)" json:"customer_status"`
	CustomerSegment                 interface{} `gorm:"column:customer_segment;type:varchar(20)" json:"customer_segment"`
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
