package response

import "time"

type ApiResponse struct {
	Message    string      `json:"messages"`
	Errors     interface{} `json:"errors"`
	Data       interface{} `json:"data"`
	ServerTime string      `json:"server_time"`
	RequestID  string      `json:"request_id"`
}

type ErrorValidation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type DupcheckResult struct {
	Code                   interface{} `json:"code"`
	Decision               string      `json:"decision"`
	Reason                 string      `json:"reason"`
	StatusKonsumen         string      `json:"status_konsumen"`
	KategoriStatusKonsumen string      `json:"kategori_status_konsumen,omitempty"`
	IsBlacklist            int         `json:"is_blacklist"`
	NextProcess            int         `json:"next_process"`
	TotalBakiDebet         float64     `json:"total_baki_debet,omitempty"`
	PbkReport              string      `json:"pbk_report,omitempty"`
}

type Filtering struct {
	ProspectID        string      `json:"prospect_id"`
	Code              interface{} `json:"code"`
	Decision          string      `json:"decision"`
	Reason            string      `json:"reason"`
	CustomerStatus    interface{} `json:"customer_status"`
	CustomerSegment   interface{} `json:"customer_segment"`
	IsBlacklist       bool        `json:"is_blacklist"`
	NextProcess       bool        `json:"next_process"`
	PbkReportCustomer interface{} `json:"pbk_report_customer"`
	PbkReportSpouse   interface{} `json:"pbk_report_spouse"`
}

type CustomerDomain struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Data      DataCustomer `json:"data"`
	Errors    interface{}  `json:"errors"`
	RequestID string       `json:"request_id"`
	Timestamp string       `json:"timestamp"`
}

type DataCustomer struct {
	CustomerSegmentation []CustomerSegmentation `json:"customer_segmentation"`
}

type CustomerSegmentation struct {
	LobID                      int    `json:"lob_id"`
	SegmentID                  int    `json:"segment_id"`
	SegmentName                string `json:"segment_name"`
	TransactionTypeID          int    `json:"transaction_type_id"`
	TransactionTypeName        string `json:"transaction_type_name"`
	TransactionTypeDescription string `json:"transaction_type_description"`
}

type DupCheckData struct {
	Data       DataDupcheck `json:"data"`
	Messages   string       `json:"messages"`
	ServerTime string       `json:"server_time"`
}

type DataDupcheck struct {
	BadType                          string  `json:"bad_type"`
	BirthDate                        string  `json:"birth_date"`
	BirthPlace                       string  `json:"birth_place"`
	CompanyAddress                   string  `json:"company_address"`
	CompanyCity                      string  `json:"company_city"`
	CompanyKecamatan                 string  `json:"company_kecamatan"`
	CompanyKelurahan                 string  `json:"company_kelurahan"`
	CompanyZipcode                   string  `json:"company_zipcode"`
	CustomerID                       string  `json:"customer_id"`
	Education                        string  `json:"education"`
	EmergencyContactAddress          string  `json:"emergency_contact_address"`
	FullName                         string  `json:"full_name"`
	Gender                           string  `json:"gender"`
	HomeStatus                       string  `json:"home_status"`
	IDNumber                         string  `json:"id_number"`
	IsRestructure                    int     `json:"is_restructure"`
	IsSimiliar                       int     `json:"is_similiar"`
	JobPos                           string  `json:"job_pos"`
	JobTypeID                        string  `json:"job_type_id"`
	LagalZipcode                     string  `json:"lagal_zipcode"`
	LegalAddress                     string  `json:"legal_address"`
	LegalCity                        string  `json:"legal_city"`
	LegalKecamatan                   string  `json:"legal_kecamatan"`
	LegalKelurahan                   string  `json:"legal_kelurahan"`
	MaritalStatus                    string  `json:"marital_status"`
	MaxOverduedays                   int     `json:"max_overduedays"`
	MaxOverduedaysForActiveAgreement int     `json:"max_overduedays_for_active_agreement"`
	MaxOverduedaysForPrevEom         int     `json:"max_overduedays_for_prev_eom"`
	MaxOverduedaysRoao               int     `json:"max_overduedays_roao"`
	MonthlyFixedIncome               int     `json:"monthly_fixed_income"`
	MonthlyVariableIncome            int     `json:"monthly_variable_income"`
	NumOfAssetInventoried            int     `json:"num_of_asset_inventoried"`
	NumOfDependence                  int     `json:"num_of_dependence"`
	NumberOfAgreement                int     `json:"number_of_agreement"`
	OsInstallmentdue                 float64 `json:"os_installmentdue"`
	OutstandingPrincipal             float64 `json:"outstanding_principal"`
	OverduedaysAging                 int     `json:"overduedays_aging"`
	PersonalNpwp                     string  `json:"personal_npwp"`
	ProfessionID                     string  `json:"profession_id"`
	ResidenceAddress                 string  `json:"residence_address"`
	ResidenceCity                    string  `json:"residence_city"`
	ResidenceKecamatan               string  `json:"residence_kecamatan"`
	ResidenceKelurahan               string  `json:"residence_kelurahan"`
	ResidenceZipcode                 string  `json:"residence_zipcode"`
	RrdDate                          string  `json:"rrd_date"`
	SisaJumlahAngsuran               int     `json:"sisa_jumlah_angsuran"`
	SpouseIncome                     float64 `json:"spouse_income"`
	SurgateMotherName                string  `json:"surgate_mother_name"`
	TotalInstallment                 float64 `json:"total_installment"`
	TotalInstallmentNap              float64 `json:"total_installment_nap"`
	WorkSinceYear                    string  `json:"work_since_year"`
	InstallmentAmount_ChassisNo      string  `json:"installment_amount_chassis_no"`
}

type KreditMuResponse struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Data      DataKreditmu `json:"data"`
	Errors    interface{}  `json:"errors"`
	RequestID string       `json:"request_id"`
	Timestamp string       `json:"timestamp"`
}

type DataKreditmu struct {
	CustomerStatus        string           `json:"customer_status"`
	ID                    int              `json:"id"`
	IsAllowedUpgradeLimit bool             `json:"is_allowed_upgrade_limit"`
	Limit                 int              `json:"limit"`
	LimitAvailable        []LimitAvailable `json:"limit_available"`
	LimitStatus           string           `json:"limit_status"`
}

type LimitAvailable struct {
	CurrentLimit int   `json:"current_limit"`
	GrossLimit   int   `json:"gross_limit"`
	Tenor        []int `json:"tenor"`
	TenorLimit   int   `json:"tenor_limit"`
}

type ResponsePefindo struct {
	Code         string                `json:"code"`
	Status       string                `json:"status"`
	Result       interface{}           `json:"result"`
	Konsumen     PefindoResultKonsumen `json:"konsumen"`
	Pasangan     PefindoResultPasangan `json:"pasangan"`
	ServerTime   time.Time             `json:"server_time"`
	DurationTime string                `json:"duration_time"`
}

type PefindoResult struct {
	SearchID                string      `json:"search_id"`
	PefindoID               string      `json:"pefindo_id"`
	Score                   string      `json:"score"`
	MaxOverdue              interface{} `json:"max_overdue"`
	MaxOverdueLast12Months  interface{} `json:"max_overdue_last12months"`
	AngsuranAktifPbk        float64     `json:"angsuran_aktif_pbk"`
	WoContract              bool        `json:"wo_contract"`
	WoAdaAgunan             bool        `json:"wo_ada_agunan"`
	TotalBakiDebetNonAgunan float64     `json:"total_baki_debet_non_agunan"`
	DetailReport            string      `json:"detail_report"`
}

type PefindoResultKonsumen struct {
	SearchID               string      `json:"search_id"`
	PefindoID              string      `json:"pefindo_id"`
	Score                  string      `json:"score"`
	MaxOverdue             interface{} `json:"max_overdue"`
	MaxOverdueLast12Months interface{} `json:"max_overdue_last12months"`
	AngsuranAktifPbk       float64     `json:"angsuran_aktif_pbk"`
	WoContract             int         `json:"wo_contract"`
	WoAdaAgunan            int         `json:"wo_ada_agunan"`
	BakiDebetNonAgunan     float64     `json:"baki_debet_non_agunan"`
	DetailReport           string      `json:"detail_report"`
}

type PefindoResultPasangan struct {
	SearchID               string      `json:"search_id"`
	PefindoID              string      `json:"pefindo_id"`
	Score                  string      `json:"score"`
	MaxOverdue             interface{} `json:"max_overdue"`
	MaxOverdueLast12Months interface{} `json:"max_overdue_last12months"`
	AngsuranAktifPbk       float64     `json:"angsuran_aktif_pbk"`
	WoContract             int         `json:"wo_contract"`
	WoAdaAgunan            int         `json:"wo_ada_agunan"`
	BakiDebetNonAgunan     float64     `json:"baki_debet_non_agunan"`
	DetailReport           string      `json:"detail_report"`
}

type ElaborateResult struct {
	Code           int     `json:"code"`
	Decision       string  `json:"decision"`
	Reason         string  `json:"reason"`
	LTV            int     `json:"ltv,omitempty"`
	ResultPefindo  string  `json:"result_pefindo,omitempty"`
	BPKBNameType   int     `json:"bpkb_name_type,omitempty"`
	Cluster        string  `json:"cluster,omitempty"`
	AgeVehicle     string  `json:"age_vehicle,omitempty"`
	LTVOrigin      float64 `json:"ltv_origin,omitempty"`
	TotalBakiDebet float64 `json:"total_balki_debet,omitempty"`
}

type ResponseMappingElaborateScheme struct {
	ResultPefindo  string  `json:"result_pefindo"`
	BranchID       string  `json:"branch_id"`
	BranchIDMask   string  `json:"branch_id_masking,omitempty"`
	CustomerStatus string  `json:"customer_status"`
	BPKBNameType   int     `json:"bpkb_name_type"`
	Cluster        string  `json:"cluster"`
	TotalBakiDebet int     `json:"total_baki_debet"`
	Tenor          int     `json:"tenor"`
	AgeVehicle     string  `json:"age_vehicle"`
	LTV            float64 `json:"ltv"`
	Decision       string  `json:"decision"`
}

type SpDupcheckMap struct {
	MaxOverdueDaysROAO               int         `json:"max_overduedays_roao"`
	NumberOfPaidInstallment          int         `json:"number_of_paid_installment"`
	MaxOverdueDaysforActiveAgreement int         `json:"max_overduedays_for_active_agreement"`
	OSInstallmentDue                 float64     `json:"os_installmentdue"`
	Reason                           string      `json:"reason"`
	CustomerID                       interface{} `json:"customer_id"`
	CustomerType                     interface{} `json:"customer_type"`
	SpouseType                       interface{} `json:"spouse_type"`
	InstallmentAmountFMF             float64     `json:"installment_amount_fmf"`
	InstallmentAmountSpouseFMF       float64     `json:"installment_amout_spouse_fmf"`
	InstallmentAmountOther           float64     `json:"installment_amount_other"`
	InstallmentAmountOtherSpouse     float64     `json:"installment_amount_other_spouse"`
	NumberofAgreement                int         `json:"number_of_agreement"`
	AgreementStatus                  string      `json:"agreement_status"`
	Dsr                              float64     `json:"dsr"`
	InstallmentTopup                 float64     `json:"installment_topup"`
	StatusKonsumen                   string      `json:"status_konsumen"`
}

type SpDupCekCustomerByID struct {
	CustomerID                       interface{} `json:"customer_id"`
	IDNumber                         string      `json:"id_number"`
	FullName                         string      `json:"full_name"`
	BirthDate                        string      `json:"birth_date"`
	SurgateMotherName                string      `json:"surgate_mother_name"`
	BirthPlace                       string      `json:"birth_place"`
	Gender                           string      `json:"gender"`
	EmergencyContactAddress          string      `json:"emergency_contact_address"`
	LegalAddress                     string      `json:"legal_address"`
	LegalKelurahan                   string      `json:"legal_kelurahan"`
	LegalKecamatan                   string      `json:"legal_kecamatan"`
	LegalCity                        string      `json:"legal_city"`
	LegalZipCode                     string      `json:"lagal_zipcode"`
	ResidenceAddress                 string      `json:"residence_address"`
	ResidenceKelurahan               string      `json:"residence_kelurahan"`
	ResidenceKecamatan               string      `json:"residence_kecamatan"`
	ResidenceCity                    string      `json:"residence_city"`
	ResidenceZipCode                 string      `json:"residence_zipcode"`
	CompanyAddress                   string      `json:"company_address"`
	CompanyKelurahan                 string      `json:"company_kelurahan"`
	CompanyKecamatan                 string      `json:"company_kecamatan"`
	CompanyCity                      string      `json:"company_city"`
	CompanyZipCode                   string      `json:"company_zipcode"`
	PersonalNPWP                     string      `json:"personal_npwp"`
	Education                        string      `json:"education"`
	MaritalStatus                    string      `json:"marital_status"`
	NumOfDependence                  int         `json:"num_of_dependence"`
	HomeStatus                       string      `json:"home_status"`
	ProfessionID                     string      `json:"profession_id"`
	JobTypeID                        string      `json:"job_type_id"`
	JobPos                           interface{} `json:"job_pos"`
	MonthlyFixedIncome               float64     `json:"monthly_fixed_income"`
	SpouseIncome                     *float64    `json:"spouse_income"`
	MonthlyVariableIncome            float64     `json:"monthly_variable_income"`
	TotalInstallment                 float64     `json:"total_installment"`
	TotalInstallmentNAP              float64     `json:"total_installment_nap"`
	BadType                          interface{} `json:"bad_type"`
	MaxOverdueDays                   int         `json:"max_overduedays"`
	MaxOverdueDaysROAO               *int        `json:"max_overduedays_roao"`
	NumOfAssetInventoried            int         `json:"num_of_asset_inventoried"`
	OverdueDaysAging                 *int        `json:"overduedays_aging"`
	MaxOverdueDaysforActiveAgreement *int        `json:"max_overduedays_for_active_agreement"`
	MaxOverdueDaysforPrevEOM         *int        `json:"max_overduedays_for_prev_eom"`
	NumberOfPaidInstallment          *int        `json:"sisa_jumlah_angsuran"`
	RRDDate                          interface{} `json:"rrd_date"`
	NumberofAgreement                int         `json:"number_of_agreement"`
	WorkSinceYear                    interface{} `json:"work_since_year"`
	OutstandingPrincipal             float64     `json:"outstanding_principal"`
	OSInstallmentDue                 float64     `json:"os_installmentdue"`
	IsRestructure                    int         `json:"is_restructure"`
	IsSimiliar                       int         `json:"is_similiar"`
	CustomerSegment                  string      `json:"customer_segment"`
	CustomerStatus                   string      `json:"customer_status"`
}

type Dsr struct {
	Result  string      `json:"result"`
	Details interface{} `json:"details"`
	Dsr     float64     `json:"dsr"`
	Code    string      `json:"code"`
	Reason  string      `json:"reason"`
}

type UsecaseApi struct {
	Code           string  `json:"code"`
	Result         string  `json:"result"`
	Reason         string  `json:"reason"`
	StatusKonsumen string  `json:"status_konsumen,omitempty"`
	Dsr            float64 `json:"dsr,omitempty"`
	Confidence     string  `json:"confidence,omitempty"`
}

type DupcheckConfig struct {
	Data DataDupcheckConfig `json:"data"`
}

type DataDupcheckConfig struct {
	VehicleAge       int     `json:"vehicle_age"`
	MinOvd           int     `json:"min_ovd"`
	MaxOvd           int     `json:"max_ovd"`
	MaxDsr           float64 `json:"max_dsr"`
	AngsuranBerjalan int     `json:"angsuran_berjalan"`
}

type LatestPaidInstallment struct {
	Code      string                    `json:"code"`
	Message   string                    `json:"message"`
	Data      LatestPaidInstallmentData `json:"data"`
	Errors    interface{}               `json:"errors"`
	RequestID string                    `json:"request_id"`
	Timestamp string                    `json:"timestamp"`
}

type LatestPaidInstallmentData struct {
	CustomerID           string  `json:"customer_id"`
	ApplicationID        string  `json:"application_id"`
	AgreementNo          string  `json:"agreement_no"`
	InstallmentAmount    float64 `json:"installment_amount"`
	ContractStatus       string  `json:"contract_status"`
	OutstandingPrinsiple float64 `json:"outstanding_principal"`
	RRDDate              string  `json:"rrd_date"`
}

type InstallmentOther struct {
	InstallmentAmountWgOff   float64 `json:"installment_wg_off"`
	InstallmentAmountKmbOff  float64 `json:"installment_kmb_off"`
	InstallmentAmountKmobOff float64 `json:"installment_kmob_off"`
	InstallmentAmountWgOnl   float64 `json:"installment_wg_onl"`
}

type DsrDetails struct {
	Customer interface{} `json:"customer"`
	Spouse   interface{} `json:"spouse"`
}

type SpDupcekChasisNo struct {
	ApplicationID     interface{} `json:"application_id"`
	InstallmentAmount interface{} `json:"installment_amount"`
	DownPayment       interface{} `json:"down_payment"`
	TotalOTR          interface{} `json:"total_otr"`
}

type RejectionNoka struct {
	Code                  string  `json:"code"`
	Result                string  `json:"result"`
	Reason                string  `json:"reason"`
	NumberOfRetry         int     `json:"NumberOfRetry"`
	IsBanned              int     `json:"IsBanned"`
	ProspectID            string  `json:"ProspectID"`
	IDNumber              string  `json:"IDNumber"`
	LegalName             string  `json:"LegalName"`
	BirthPlace            string  `json:"BirthPlace"`
	BirthDate             string  `json:"BirthDate"`
	MonthlyFixedIncome    float64 `json:"MonthlyFixedIncome"`
	EmploymentSinceYear   string  `json:"EmploymentSinceYear"`
	EmploymentSinceMonth  string  `json:"EmploymentSinceMonth"`
	StaySinceYear         string  `json:"StaySinceYear"`
	StaySinceMonth        string  `json:"StaySinceMonth"`
	BPKBName              string  `json:"BPKBName"`
	Gender                string  `json:"Gender"`
	MaritalStatus         string  `json:"MaritalStatus"`
	NumOfDependence       int     `json:"NumOfDependence"`
	NTF                   float64 `json:"NTF"`
	OTRPrice              float64 `json:"OTRPrice"`
	LegalZipCode          string  `json:"LegalZipCode"`
	CompanyZipCode        string  `json:"CompanyZipCode"`
	Tenor                 int     `json:"Tenor"`
	ManufacturingYear     string  `json:"ManufacturingYear"`
	ProfessionID          string  `json:"ProfessionID"`
	HomeStatus            string  `json:"HomeStatus"`
	IsBannedActive        bool    `json:"IsBannedActive"`
	CurrentBannedNotEmpty bool    `json:"CurrentBannedEmpty"`
}

type ResAgreementChassisNumber struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Data      AgreementChassisNumber `json:"data"`
	Errors    interface{}            `json:"errors"`
	RequestID string                 `json:"request_id"`
	Timestamp string                 `json:"timestamp"`
}

type AgreementChassisNumber struct {
	IsRegistered bool   `json:"is_registered"`
	IsActive     bool   `json:"is_active"`
	LegalName    string `json:"legal_name"`
	IDNumber     string `json:"id_number"`
	Status       string `json:"status"`
	GoLiveDate   string `json:"go_live_date"`
}

type FaceCompareResponse struct {
	ProspectID string `json:"prospect_id" validate:"required"`
	RequestID  string `json:"request_id"`
	Result     string `json:"result"`
	Reason     string `json:"reason"`
	Info       interface{}
}

type DetectImageResponse struct {
	Meta struct {
		Code          int    `json:"code"`
		Status        string `json:"status"`
		Message       string `json:"message"`
		Error         string `json:"error"`
		ExecutionTime string `json:"executionTime"`
	} `json:"meta"`
	Data struct {
		Facetoken string  `json:"face_token"`
		BlurValue float64 `json:"blur_value"`
	} `json:"data"`
}

type CompareResponse struct {
	Meta struct {
		Code          int    `json:"code"`
		Status        string `json:"status"`
		Message       string `json:"message"`
		Error         string `json:"error"`
		ExecutionTime string `json:"executionTime"`
	} `json:"meta"`
	Data struct {
		Confidence string `json:"confidence"`
		Facetoken1 string `json:"face_token_1"`
		Facetoken2 string `json:"face_token_2"`
	} `json:"data"`
}

type ImageDecodeResponse struct {
	Messages string `json:"messages"`
	Data     struct {
		Encode string `json:"encode"`
	} `json:"data"`
	Errors interface{} `json:"errors"`
	Code   string      `json:"code"`
}

type Config struct {
	Data struct {
		WG struct {
			Online  int `json:"online"`
			Offline int `json:"offline"`
		} `json:"wg"`
		Kmb  int `json:"kmb"`
		Kmob int `json:"kmob"`
		Blur int `json:"blur"`
	} `json:"data"`
}

type InfoEkyc struct {
	Vd     interface{} `json:"vd"`
	Fr     interface{} `json:"fr"`
	Asliri interface{} `json:"asliri"`
	Ktp    interface{} `json:"ktp"`
}

type Ekyc struct {
	Result      string      `json:"result"`
	Code        string      `json:"code"`
	Reason      string      `json:"reason"`
	Source      string      `json:"source"`
	Info        interface{} `json:"info"`
	Similiarity interface{} `json:"similiarity"`
}

type AsliriIntegrator struct {
	Name        interface{} `json:"name"`
	PDOB        interface{} `json:"pdob"`
	SelfiePhoto interface{} `json:"selfie_photo"`
	NotFound    bool        `json:"not_found"`
	RefID       string      `json:"ref_id"`
}

type KtpValidator struct {
	Code   string `json:"code"`
	Result string `json:"result"`
	Reason string `json:"reason"`
}

type VerifyDataIntegratorResponse struct {
	TransactionID string  `json:"transaction_id"`
	Threshold     string  `json:"threshold"`
	RefID         string  `json:"ref_id"`
	IsValid       bool    `json:"is_valid"`
	Reason        *string `json:"reason,omitempty"`
	VerifyDataDetailIntegratorResponse
}

type VerifyDataDetailIntegratorResponse struct {
	NoKk        string `json:"no_kk,omitempty"`
	NamaLgkp    string `json:"nama_lgkp,omitempty"`
	TmptLhr     string `json:"tmpt_lhr,omitempty"`
	TglLhr      string `json:"tgl_lhr,omitempty"`
	PropName    string `json:"prop_name,omitempty"`
	KabName     string `json:"kab_name,omitempty"`
	KecName     string `json:"kec_name,omitempty"`
	KelName     string `json:"kel_name,omitempty"`
	NoRt        string `json:"no_rt,omitempty"`
	NoRw        string `json:"no_rw,omitempty"`
	Alamat      string `json:"alamat,omitempty"`
	NamaLgkpIbu string `json:"nama_lgkp_ibu,omitempty"`
	StatusKawin string `json:"status_kawin,omitempty"`
	JenisPkrjn  string `json:"jenis_pkrjn,omitempty"`
	JenisKlmin  string `json:"jenis_klmin,omitempty"`
	NoProp      string `json:"no_prop,omitempty"`
	NoKab       string `json:"no_kab,omitempty"`
	NoKec       string `json:"no_kec,omitempty"`
	NoKel       string `json:"no_kel,omitempty"`
	Nik         string `json:"nik,omitempty"`
}

type FaceRecognitionIntegratorData struct {
	TransactionID string `json:"transaction_id"`
	RuleCode      string `json:"rule_code"`
	Reason        string `json:"reason"`
	Threshold     string `json:"threshold"`
	RefID         string `json:"ref_id"`
}

type Additional struct {
	DupcheckData            SpDupcheckMap `json:"dupcheck_data"`
	CustomerStatus          string        `json:"customer_status"`
	ScsDecision             ScsDecision   `json:"scs_decision"`
	PDFCustomer             interface{}   `json:"pdf_customer"`
	PDFSpouse               interface{}   `json:"pdf_spouse"`
	Reprocess               bool          `json:"reprocess"`
	CustomerType            interface{}   `json:"customer_type"`
	SpouseType              interface{}   `json:"spouse_type"`
	AsliriSimiliarity       interface{}   `json:"asliri_similiarity"`
	AsliriReason            interface{}   `json:"asliri_reason"`
	DSR                     interface{}   `json:"dsr"`
	DataAkkk                BiroAkkk      `json:"biro_akkk"`
	AngsuranAktifBiro       float64       `json:"angsuran_aktif_biro"`
	AngsuranAktifSpouseBiro float64       `json:"angsuran_aktif_biro_spouse"`
}

type ScsDecision struct {
	ScsDate   interface{} `json:"scs_date"`
	ScsScore  interface{} `json:"scs_score"`
	ScsStatus interface{} `json:"scs_status"`
}

type BiroAkkk struct {
	PefindoPlafon                    interface{} `json:"pefindo_plafon,omitempty"`
	PefindoBakiDebet                 interface{} `json:"pefindo_baki_debet,omitempty"`
	PefindoTotalFasilitasAktif       interface{} `json:"pefindo_total_fasilitas_aktif,omitempty"`
	PefindoSpousePlafon              interface{} `json:"pefindo_spouse_plafon,omitempty"`
	PefindoSpouseBakiDebet           interface{} `json:"pefindo_spouse_baki_debet,omitempty"`
	PefindoSpouseTotalFasilitasAktif interface{} `json:"pefindo_spouse_total_fasilitas_aktif,omitempty"`
	Score                            interface{} `json:"score"`
	MaxOvd                           interface{} `json:"max_overdue"`
	WorstQualityStatus               interface{} `json:"worst_quality_status"`
	WorstQualityStatusMonth          interface{} `json:"worst_quality_status_month"`
	LastQualityCredit                interface{} `json:"last_quality_credit"`
	LastQualityCreditMonth           interface{} `json:"last_quality_credit_month"`
	SpouseScore                      interface{} `json:"spouse_score"`
	SpouseMaxOvd                     interface{} `json:"spouse_max_overdue"`
	SpouseWorstQualityStatus         interface{} `json:"spouse_worst_quality_status"`
	SpouseWorstQualityStatusMonth    interface{} `json:"spouse_worst_quality_status_month"`
	SpouseLastQualityCredit          interface{} `json:"spouse_last_quality_credit"`
	SpouseLastQualityCreditMonth     interface{} `json:"spouse_last_quality_credit_month"`
	PefindoDetailPdf                 interface{} `json:"pefindo_detail_pdf"`
	PefindoSpouseDetailPdf           interface{} `json:"pefindo_spouse_detail_pdf"`
	PefindoBakiDebetWo               interface{} `json:"pefindo_baki_debet_wo"`
	PefindoSpouseBakiDebetWo         interface{} `json:"pefindo_spouse_baki_debet_wo"`
	PefindoAgunan                    interface{} `json:"pefindo_agunan"`
	PefindoSpouseAgunan              interface{} `json:"pefindo_spouse_agunan"`
}
