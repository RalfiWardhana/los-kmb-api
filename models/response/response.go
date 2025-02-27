package response

import (
	"los-kmb-api/models/entity"
	"time"
)

type ApiResponseV2 struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Errors     interface{} `json:"errors"`
	Data       interface{} `json:"data"`
	ServerTime string      `json:"server_time"`
	RequestID  string      `json:"request_id"`
}

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
	CustomerStatusKMB interface{} `json:"customer_status_kmb"`
	CustomerSegment   interface{} `json:"customer_segment"`
	IsBlacklist       bool        `json:"is_blacklist"`
	NextProcess       bool        `json:"next_process"`
	PbkReportCustomer interface{} `json:"pbk_report_customer"`
	PbkReportSpouse   interface{} `json:"pbk_report_spouse"`
	TotalBakiDebet    interface{} `json:"total_baki_debet"`
	Cluster           interface{} `json:"-"`
	ClusterCMO        interface{} `json:"-"`
}

type PefindoIDX struct {
	ProspectID              string  `json:"prospect_id"`
	OldestMobPl             int     `json:"oldestmob_pl"`
	FinalNom6012Mth         int     `json:"final_nom60_12mth"`
	TotBakiDebetBanksActive int     `json:"tot_bakidebet_banks_active"`
	TotBakiDebet3160Dpd     int     `json:"tot_bakidebet_31_60dpd"`
	Worst24Mth              int     `json:"worst_24mth"`
	MaxLimitOth             float64 `json:"max_limit_oth"`

	Nom036MonthAll int     `json:"nom03_6mth_all"`
	MaxLimitPl     float64 `json:"max_limit_pl"`
	TotBakiDebet4  int     `json:"tot_bakidebet4"`
	Worst12MthAuto int     `json:"worst_12mth_auto"`

	Nom0312MntAll  int `json:"nom03_12mth_all"`
	Worst24MthAuto int `json:"worst_24mth_auto"`
}

type IntegratorScorePro struct {
	ProspectID  string      `json:"prospect_id"`
	Score       interface{} `json:"score"`
	Result      string      `json:"result"`
	ScoreBand   string      `json:"score_band"`
	ScoreResult string      `json:"score_result"`
	Status      string      `json:"status"`
	Segmen      string      `json:"segmen"`
	IsTsi       bool        `json:"is_tsi"`
	ScoreBin    interface{} `json:"score_bin"`
	Deviasi     interface{} `json:"deviasi"`
}

type ScorePro struct {
	Result    string `json:"result"`
	Code      string `json:"code"`
	Reason    string `json:"reason"`
	Source    string `json:"source"`
	Info      string `json:"info"`
	IsDeviasi bool   `json:"-"`
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

type ElaborateLTV struct {
	LTV         int    `json:"ltv"`
	AdjustTenor bool   `json:"adjut_tenor"`
	MaxTenor    int    `json:"max_tenor"`
	Reason      string `json:"reason"`
}

type Recalculate struct {
	ProspectID string `json:"prospect_id"`
}

type InsertStaging struct {
	ProspectID string `json:"prospect_id"`
}

type RespApprovalScheme struct {
	Name         string
	NextStep     string
	IsFinal      bool
	IsEscalation bool
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

type ResultNewKoRules struct {
	CategoryPBK       string `json:"category_pbk"`
	ContractCode      string `json:"contract_code"`
	ContractStatus    string `json:"contract_status"`
	CreditorType      string `json:"creditor_type"`
	Creditor          string `json:"creditor"`
	ConditionDate     string `json:"condition_date"`
	RestructuringDate string `json:"restructuring_date"`
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
	SearchID                      string           `json:"search_id"`
	PefindoID                     string           `json:"pefindo_id"`
	Score                         string           `json:"score"`
	MaxOverdue                    interface{}      `json:"max_overdue"`
	MaxOverdueLast12Months        interface{}      `json:"max_overdue_last12months"`
	AngsuranAktifPbk              float64          `json:"angsuran_aktif_pbk"`
	WoContract                    bool             `json:"wo_contract"`
	WoAdaAgunan                   bool             `json:"wo_ada_agunan"`
	TotalBakiDebetNonAgunan       float64          `json:"total_baki_debet_non_agunan"`
	DetailReport                  string           `json:"detail_report"`
	Category                      interface{}      `json:"category"`
	MaxOverdueKORules             interface{}      `json:"max_overdue_ko_rules"`
	MaxOverdueLast12MonthsKORules interface{}      `json:"max_overdue_last12months_ko_rules"`
	NewKoRules                    ResultNewKoRules `json:"new_ko_rules"`
}

type PefindoResultKonsumen struct {
	SearchID                               string           `json:"search_id"`
	PefindoID                              string           `json:"pefindo_id"`
	Score                                  string           `json:"score"`
	MaxOverdue                             interface{}      `json:"max_overdue"`
	MaxOverdueLast12Months                 interface{}      `json:"max_overdue_last12months"`
	AngsuranAktifPbk                       float64          `json:"angsuran_aktif_pbk"`
	WoContract                             int              `json:"wo_contract"`
	WoAdaAgunan                            int              `json:"wo_ada_agunan"`
	BakiDebetNonAgunan                     float64          `json:"baki_debet_non_agunan"`
	DetailReport                           string           `json:"detail_report"`
	Plafon                                 float64          `json:"plafon"`
	FasilitasAktif                         int              `json:"fasilitas_aktif"`
	KualitasKreditTerburuk                 string           `json:"kualitas_kredit_terburuk"`
	BulanKualitasTerburuk                  string           `json:"bulan_kualitas_terburuk"`
	BakiDebetKualitasTerburuk              float64          `json:"baki_debet_kualitas_terburuk"`
	KualitasKreditTerakhir                 string           `json:"kualitas_kredit_terakhir"`
	BulanKualitasKreditTerakhir            string           `json:"bulan_kualitas_kredit_terakhir"`
	OverdueLastKORules                     interface{}      `json:"overdue_last_ko_rules"`
	OverdueLast12MonthsKORules             interface{}      `json:"overdue_last_12month_ko_rules"`
	Category                               interface{}      `json:"category"`
	MaxOverdueAgunanKORules                interface{}      `json:"max_ovd_agunan_ko_rules"`
	MaxOverdueAgunanLast12MonthsKORules    interface{}      `json:"max_ovd_agunan_last_12month_ko_rules"`
	MaxOverdueNonAgunanKORules             interface{}      `json:"max_ovd_non_agunan_ko_rules"`
	MaxOverdueNonAgunanLast12MonthsKORules interface{}      `json:"max_ovd_non_agunan_last_12month_ko_rules"`
	NewKoRules                             ResultNewKoRules `json:"new_ko_rules"`
}

type PefindoResultPasangan struct {
	SearchID                               string      `json:"search_id"`
	PefindoID                              string      `json:"pefindo_id"`
	Score                                  string      `json:"score"`
	MaxOverdue                             interface{} `json:"max_overdue"`
	MaxOverdueLast12Months                 interface{} `json:"max_overdue_last12months"`
	AngsuranAktifPbk                       float64     `json:"angsuran_aktif_pbk"`
	WoContract                             int         `json:"wo_contract"`
	WoAdaAgunan                            int         `json:"wo_ada_agunan"`
	BakiDebetNonAgunan                     float64     `json:"baki_debet_non_agunan"`
	DetailReport                           string      `json:"detail_report"`
	Plafon                                 float64     `json:"plafon"`
	FasilitasAktif                         int         `json:"fasilitas_aktif"`
	KualitasKreditTerburuk                 string      `json:"kualitas_kredit_terburuk"`
	BulanKualitasTerburuk                  string      `json:"bulan_kualitas_terburuk"`
	BakiDebetKualitasTerburuk              float64     `json:"baki_debet_kualitas_terburuk"`
	KualitasKreditTerakhir                 string      `json:"kualitas_kredit_terakhir"`
	BulanKualitasKreditTerakhir            string      `json:"bulan_kualitas_kredit_terakhir"`
	OverdueLastKORules                     interface{} `json:"overdue_last_ko_rules"`
	OverdueLast12MonthsKORules             interface{} `json:"overdue_last_12month_ko_rules"`
	Category                               interface{} `json:"category"`
	MaxOverdueAgunanKORules                interface{} `json:"max_ovd_agunan_ko_rules"`
	MaxOverdueAgunanLast12MonthsKORules    interface{} `json:"max_ovd_agunan_last_12month_ko_rules"`
	MaxOverdueNonAgunanKORules             interface{} `json:"max_ovd_non_agunan_ko_rules"`
	MaxOverdueNonAgunanLast12MonthsKORules interface{} `json:"max_ovd_non_agunan_last_12month_ko_rules"`
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
	IsMappingOvd   bool    `json:"is_mapping_ovd,omitempty"`
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
	IsMappingOvd   string  `json:"is_mapping_ovd,omitempty"`
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
	RRDDate                          interface{} `json:"rrd_date"`
	DetailsDSR                       interface{} `json:"details_dsr"`
	ConfigMaxDSR                     float64     `json:"config_max_dsr"`
	Cluster                          interface{} `json:"cluster"`
	AgreementSettledExist            bool        `json:"agreement_settled_exist"`
	NegativeCustomer                 interface{} `json:"negative_customer"`
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
	CustomerStatusKMB                string      `json:"customer_status_lob"`
}

type Dsr struct {
	Result  string      `json:"result"`
	Details interface{} `json:"details"`
	Dsr     float64     `json:"dsr"`
	Code    string      `json:"code"`
	Reason  string      `json:"reason"`
}

type DsrDetails struct {
	Customer    interface{} `json:"customer"`
	Spouse      interface{} `json:"spouse"`
	DetailTopUP interface{} `json:"detail_topup"`
}

type DetailTopUP struct {
	Pencairan              interface{} `json:"pencairan"`
	AgreementChassisNumber interface{} `json:"agreement_chassis_number"`
	TotalOutstanding       interface{} `json:"total_outstanding"`
	MinimumPencairan       interface{} `json:"minimum_pencairan"`
}

type UsecaseApi struct {
	Code           string      `json:"code"`
	Result         string      `json:"result"`
	Reason         string      `json:"reason"`
	StatusKonsumen string      `json:"status_konsumen,omitempty"`
	Dsr            float64     `json:"dsr,omitempty"`
	Confidence     string      `json:"confidence,omitempty"`
	SourceDecision string      `json:"source_decision,omitempty"`
	Info           interface{} `json:"info,omitempty"`
	IsDeviasi      bool        `json:"-"`
}

type NegativeCustomer struct {
	IsActive    int    `json:"is_active"`
	IsBlacklist int    `json:"is_blacklist"`
	IsHighrisk  int    `json:"is_highrisk"`
	BadType     string `json:"bad_type"`
	Result      string `json:"result"`
	Decision    string `json:"decision"`
}

type LowIncome struct {
	NoApplication string  `json:"no_application"`
	Income        float64 `json:"income"`
	Range         string  `json:"range"`
}

type LockSystem struct {
	IsBanned  bool   `json:"is_banned"`
	Reason    string `json:"reason"`
	UnbanDate string `json:"unban_date"`
}

type DupcheckConfig struct {
	Data DataDupcheckConfig `json:"data"`
}

type DataDupcheckConfig struct {
	VehicleAge              int     `json:"vehicle_age"`
	MaxOvd                  int     `json:"max_ovd"`
	MaxOvdAOPrimePriority   int     `json:"max_ovd_ao_prime_priority"`
	MaxOvdAORegular         int     `json:"max_ovd_ao_regular"`
	MaxDsr                  float64 `json:"max_dsr"`
	AngsuranBerjalan        int     `json:"angsuran_berjalan"`
	AttemptPMKDSR           int     `json:"attempt_pmk_dsr"`
	AttemptNIK              int     `json:"attempt_nik"`
	AttemptChassisNumber    int     `json:"attempt_chassis_number"`
	MinimumPencairanROTopUp struct {
		Prime    float64 `json:"prime"`
		Priority float64 `json:"priority"`
		Regular  float64 `json:"regular"`
	} `json:"minimum_pencairan_ro_top_up"`
}

type LockSystemConfig struct {
	Data DataLockSystemConfig `json:"data"`
}

type DataLockSystemConfig struct {
	LockRejectAttempt int    `json:"lock_reject_attempt"`
	LockRejectBan     int    `json:"lock_reject_ban"`
	LockRejectCheck   int    `json:"lock_reject_check"`
	LockCancelAttempt int    `json:"lock_cancel_attempt"`
	LockCancelBan     int    `json:"lock_cancel_ban"`
	LockCancelCheck   int    `json:"lock_cancel_check"`
	LockStartDate     string `json:"lock_start_date"`
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

type NTFOther struct {
	NTFAmountWgOff   float64 `json:"wg_offline"`
	NTFAmountKmbOff  float64 `json:"kmb"`
	NTFAmountKmobOff float64 `json:"kmob"`
	NTFAmountWgOnl   float64 `json:"new_wg"`
	NTFAmountUC      float64 `json:"uc"`
	NTFAmountNewKmb  float64 `json:"new_kmb"`
	TotalOutstanding float64 `json:"outstanding"`
}

type InstallmentOther struct {
	InstallmentAmountWgOff   float64 `json:"wg_offline"`
	InstallmentAmountKmbOff  float64 `json:"kmb"`
	InstallmentAmountKmobOff float64 `json:"kmob"`
	InstallmentAmountWgOnl   float64 `json:"new_wg"`
	InstallmentAmountUC      float64 `json:"uc"`
	InstallmentAmountNewKmb  float64 `json:"new_kmb"`
}

type IntegratorAgreementChassisNumber struct {
	GoLiveDate           string  `json:"go_live_date"`
	IDNumber             string  `json:"id_number"`
	IsActive             bool    `json:"is_active"`
	IsRegistered         bool    `json:"is_registered"`
	LCInstallment        float64 `json:"lc_installment"`
	LegalName            string  `json:"legal_name"`
	OutstandingInterest  float64 `json:"outstanding_interest"`
	OutstandingPrincipal float64 `json:"outstanding_principal"`
	Status               string  `json:"status"`
}

type NTFDetails struct {
	Customer interface{} `json:"customer"`
	Spouse   interface{} `json:"spouse"`
}

type OutstandingConfins struct {
	CustomerID       string  `json:"customer_id"`
	TotalOutstanding float64 `json:"total_outstanding"`
}
type TrxFMF struct {
	NTFAkumulasi            float64
	NTFOtherAmount          float64
	NTFOtherAmountSpouse    float64
	NTFOtherAmountDetail    string
	NTFConfinsAmount        float64
	NTFConfins              float64
	NTFTopup                float64
	DupcheckData            SpDupcheckMap `json:"dupcheck_data"`
	CustomerStatus          string        `json:"customer_status"`
	ScsDecision             ScsDecision   `json:"scs_decision"`
	CustomerType            interface{}   `json:"customer_type"`
	SpouseType              interface{}   `json:"spouse_type"`
	DSRFMF                  interface{}   `json:"dsr_fmf"`
	DSRPBK                  interface{}   `json:"dsr_pbk"`
	TotalDSR                interface{}   `json:"total_dsr"`
	TrxBannedPMKDSR         entity.TrxBannedPMKDSR
	TrxBannedChassisNumber  entity.TrxBannedChassisNumber
	AgreementCONFINS        []AgreementCONFINS
	InstallmentThreshold    float64 `json:"installment_threshold"`
	LatestInstallmentAmount float64 `json:"latest_installment_amount"`
	TrxCaDecision           entity.TrxCaDecision
	EkycSource              interface{} `json:"ekyc_source"`
	EkycSimiliarity         interface{} `json:"ekyc_similiarity"`
	EkycReason              interface{} `json:"ekyc_reason"`
	TrxDeviasi              entity.TrxDeviasi
	TrxEDD                  entity.TrxEDD
}

type RoaoAkkk struct {
	MaxOverdueDaysROAO               interface{} `json:"max_overduedays_roao"`
	MaxOverdueDaysforActiveAgreement interface{} `json:"max_overduedays_for_active_agreement"`
	NumberofAgreement                interface{} `json:"number_of_agreement"`
	AgreementStatus                  interface{} `json:"agreement_status"`
	NumberOfPaidInstallment          interface{} `json:"NumberOfPaidInstallment"`
	OSInstallmentDue                 interface{} `json:"os_installmentdue"`
	InstallmentAmountFMF             interface{} `json:"installment_amount_fmf"`
	InstallmentAmountSpouseFMF       interface{} `json:"installment_amount_spouse_fmf"`
	InstallmentAmountOther           interface{} `json:"installment_amount_other"`
	InstallmentAmountOtherSpouse     interface{} `json:"installment_amount_other_spouse"`
	InstallmentTopup                 interface{} `json:"installment_topup"`
	LatestInstallment                interface{} `json:"latest_installment"`
}

type Metrics struct {
	ProspectID     string      `json:"prospect_id"`
	Code           interface{} `json:"code"`
	Decision       string      `json:"decision"`
	DecisionReason string      `json:"decision_reason"`
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
	GoLiveDate           string  `json:"go_live_date"`
	IDNumber             string  `json:"id_number"`
	InstallmentAmount    float64 `json:"installment_amount"`
	IsActive             bool    `json:"is_active"`
	IsRegistered         bool    `json:"is_registered"`
	LcInstallment        float64 `json:"lc_installment"`
	LegalName            string  `json:"legal_name"`
	OutstandingInterest  float64 `json:"outstanding_interest"`
	OutstandingPrincipal float64 `json:"outstanding_principal"`
	Status               string  `json:"status"`
}

type AgreementCONFINS struct {
	ApplicationID        string  `json:"application_id"`
	ProductType          string  `json:"product_type"`
	AgreementDate        string  `json:"agreement_date"`
	AssetCode            string  `json:"asset_code"`
	Tenor                int     `json:"period"`
	OutstandingPrincipal float64 `json:"outstanding_principal"`
	InstallmentAmount    float64 `json:"installment_amount"`
	ContractStatus       string  `json:"contract_status"`
	CurrentCondition     string  `json:"current_condition"`
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
	Vd        interface{} `json:"vd"`
	VdService interface{} `json:"vd_service"`
	VdError   interface{} `json:"vd_error"`
	Fr        interface{} `json:"fr"`
	FrService interface{} `json:"fr_service"`
	FrError   interface{} `json:"fr_error"`
	Asliri    interface{} `json:"asliri"`
	Ktp       interface{} `json:"ktp"`
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
	NamaLgkp    int    `json:"nama_lgkp,omitempty"`
	TmptLhr     int    `json:"tmpt_lhr,omitempty"`
	TglLhr      string `json:"tgl_lhr,omitempty"`
	PropName    string `json:"prop_name,omitempty"`
	KabName     string `json:"kab_name,omitempty"`
	KecName     string `json:"kec_name,omitempty"`
	KelName     string `json:"kel_name,omitempty"`
	NoRt        string `json:"no_rt,omitempty"`
	NoRw        string `json:"no_rw,omitempty"`
	Alamat      int    `json:"alamat,omitempty"`
	NamaLgkpIbu int    `json:"nama_lgkp_ibu,omitempty"`
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
	MatchScore    string `json:"matchScore"`
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

type ResponseGenerateFormAKKK struct {
	MediaUrl string `json:"media_url"`
	Path     string `json:"path"`
}

type InquiryRow struct {
	Inquiry        interface{} `json:"inquiry"`
	RecordFiltered int         `json:"recordsFiltered"`
	RecordTotal    int         `json:"recordsTotal"`
}

type ReasonMessageRow struct {
	Reason         interface{} `json:"reason"`
	RecordFiltered int         `json:"recordsFiltered"`
	RecordTotal    int         `json:"recordsTotal"`
}

type ReviewPrescreening struct {
	ProspectID string      `json:"prospect_id"`
	Code       interface{} `json:"code"`
	Decision   string      `json:"decision"`
	Reason     string      `json:"reason"`
}

type CAResponse struct {
	ProspectID string `json:"prospect_id"`
	Decision   string `json:"decision"`
	SlikResult string `json:"slik_result"`
	Note       string `json:"note"`
}

type CancelResponse struct {
	ProspectID string `json:"prospect_id"`
	Reason     string `json:"reason"`
	Status     string `json:"status"`
}

type ReturnResponse struct {
	ProspectID string `json:"prospect_id"`
	Status     string `json:"status"`
}

type RecalculateResponse struct {
	ProspectID string  `json:"prospect_id"`
	DPAmount   float64 `json:"dp_amount"`
	Status     string  `json:"status"`
}

type ApprovalResponse struct {
	ProspectID     string `json:"prospect_id"`
	Decision       string `json:"decision"`
	Reason         string `json:"reason"`
	Code           string `json:"code"`
	Note           string `json:"note"`
	IsFinal        bool   `json:"is_final"`
	NeedEscalation bool   `json:"need_escalation"`
}

type SubmitRecalculateResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Errors     interface{} `json:"errors"`
	RequestID  string      `json:"request_id"`
	ServerTime string      `json:"timestamp"`
}

type UpdateQuotaDeviasiBranchResponse struct {
	Status           string                        `json:"status"`
	Message          string                        `json:"message"`
	BranchID         string                        `json:"branch_id"`
	DataBeforeUpdate entity.DataQuotaDeviasiBranch `json:"data_before_update,omitempty"`
	DataAfterUpdate  entity.DataQuotaDeviasiBranch `json:"data_after_update,omitempty"`
}

type UploadQuotaDeviasiBranchResponse struct {
	Status           string                        `json:"status"`
	Message          string                        `json:"message"`
	DataBeforeUpdate []entity.MappingBranchDeviasi `json:"data_before_update,omitempty"`
	DataAfterUpdate  []entity.MappingBranchDeviasi `json:"data_after_update,omitempty"`
}

type EmployeeCMOResponse struct {
	EmployeeID         string      `json:"employee_id"`
	EmployeeName       string      `json:"employee_name"`
	EmployeeIDWithName string      `json:"employee_id_with_name"`
	JoinDate           string      `json:"join_date"`
	PositionGroupCode  string      `json:"position_group_code"`
	PositionGroupName  string      `json:"position_group_name"`
	CMOCategory        string      `json:"cmo_category"`
	IsCmoSpv           interface{} `json:"is_cmo_spv"`
}

type EmployeeCareerHistory struct {
	EmployeeID        string `json:"employee_id"`
	EmployeeName      string `json:"name"`
	RealCareerDate    string `json:"real_career_date"`
	RegistrationDate  string `json:"registration_date"`
	RegistrationNo    string `json:"registration_no"`
	PositionGroupCode string `json:"position_group_code"`
	PositionGroupName string `json:"position_group_name"`
	PositionCodeOld   string `json:"position_code_old"`
	PositionNameOld   string `json:"position_name_old"`
	PositionCodeNew   string `json:"position_code_new"`
	PositionNameNew   string `json:"position_name_new"`
	IsResign          bool   `json:"is_resign"`
}

type GetEmployeeByID struct {
	Error   interface{}             `json:"error"`
	Message string                  `json:"message"`
	Data    []EmployeeCareerHistory `json:"data"`
}

type HrisListEmployee struct {
	EmployeeID  string      `json:"employee_id"`
	IDNumber    interface{} `json:"id_number"`
	PhoneNumber interface{} `json:"phone_number"`
	IsResign    bool        `json:"is_resign"`
}

type FpdCMOResponse struct {
	FpdExist    bool    `json:"fpd_exist"`
	CmoFpd      float64 `json:"cmo_fpd"`
	CmoAccSales int     `json:"cmo_acc_sales"`
}

type FpdData struct {
	CmoID        string  `json:"cmo_id"`
	BpkbNameType string  `json:"bpkb_name_type"`
	Fpd          float64 `json:"fpd"`
	AccSales     int     `json:"acc_sales"`
}

type GetFPDCmoByID struct {
	Code     string      `json:"code"`
	Message  string      `json:"message"`
	Data     []FpdData   `json:"data"`
	Errors   interface{} `json:"errors"`
	Metadata interface{} `json:"metadata"`
}

type AgreementData struct {
	BranchID              string    `json:"branch_id"`
	CustomerID            string    `json:"customer_id"`
	ApplicationID         string    `json:"application_id"`
	AgreementNo           string    `json:"agreement_no"`
	LegalName             string    `json:"legal_name"`
	InstallmentAmount     float64   `json:"installment_amount"`
	DownPayment           float64   `json:"down_payment"`
	Tenor                 int       `json:"tenor"`
	GoLiveDate            time.Time `json:"go_live_date"`
	OutstandingPrincipal  float64   `json:"outstanding_principal"`
	ContractStatus        string    `json:"contract_status"`
	NextInstallmentNumber int       `json:"next_installment_number"`
	NextInstallmentDate   time.Time `json:"next_installment_date"`
	LicensePlate          string    `json:"license_plate"`
	AssetTypeID           string    `json:"asset_type_id"`
	AssetCode             string    `json:"asset_code"`
	ManufacturingYear     int       `json:"manufacturing_year"`
	RrdDate               time.Time `json:"rrd_date"`
	Bpkb                  string    `json:"bpkb"`
	SerialNo1             string    `json:"serial_no_1"`
	SerialNo2             string    `json:"serial_no_2"`
	TotalOtr              float64   `json:"total_otr"`
	DiscountOtr           float64   `json:"discount_otr"`
}

type ChassisNumberOfLicensePlateResponse struct {
	ChassisNumber string `json:"chassis_number"`
	EngineNumber  string `json:"engine_number"`
}

type ConfinsAgreementCustomer struct {
	Code     string           `json:"code"`
	Message  string           `json:"message"`
	Data     *[]AgreementData `json:"data"`
	Errors   interface{}      `json:"errors"`
	Metadata interface{}      `json:"metadata"`
}

type ExpiredContractConfig struct {
	Data ConfigExpiredContract `json:"data"`
}

type ConfigExpiredContract struct {
	ExpiredContractCheckEnabled bool `json:"expired_contract_check_enabled"`
	ExpiredContractMaxMonths    int  `json:"expired_contract_max_months"`
}

type StepPrinciple struct {
	ProspectID string `json:"prospect_id"`
	ColorCode  string `json:"color_code"`
	Status     string `json:"status"`
	UpdatedAt  string `json:"updated_at"`
}

type CustomerDomainValidate struct {
	Code      string                     `json:"code"`
	Message   string                     `json:"message"`
	Data      CustomerDomainValidateData `json:"data"`
	Errors    interface{}                `json:"errors"`
	RequestID string                     `json:"request_id"`
	Timestamp string                     `json:"timestamp"`
}

type CustomerDomainValidateData struct {
	CustomerID int `json:"customer_id"`
	KPMID      int `json:"kpm_id"`
}

type CustomerDomainInsert struct {
	Code      string                   `json:"code"`
	Message   string                   `json:"message"`
	Data      CustomerDomainInsertData `json:"data"`
	Errors    interface{}              `json:"errors"`
	RequestID string                   `json:"request_id"`
	Timestamp string                   `json:"timestamp"`
}

type CustomerDomainInsertData struct {
	CustomerID    int  `json:"customer_id"`
	IsNewCustomer bool `json:"is_new_customer"`
}

type CustomerDomainUpdateCustomerTransaction struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Errors    interface{} `json:"errors"`
	RequestID string      `json:"request_id"`
	Timestamp string      `json:"timestamp"`
}

type MarsevLoanAmountResponse struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    MarsevLoanAmountData `json:"data"`
	Errors  interface{}          `json:"errors"`
}

type MarsevLoanAmountData struct {
	LoanAmountMaximum  float64 `json:"loan_amount_maximum"`
	AmountOfFinance    float64 `json:"amount_of_finance"`
	DpAmount           float64 `json:"dp_amount"`
	DpPercentFinal     float64 `json:"dp_percent_final"`
	LtvPercentFinal    float64 `json:"ltv_percent_final"`
	AdminFeeAmount     float64 `json:"admin_fee_amount"`
	ProvisionFeeAmount float64 `json:"provision_fee_amount"`
	LoanAmountFinal    float64 `json:"loan_amount_final"`
	IsPsa              bool    `json:"is_psa"`
}

type MarsevFilterProgramResponse struct {
	Code     int                       `json:"code"`
	Message  string                    `json:"message"`
	Data     []MarsevFilterProgramData `json:"data"`
	PageInfo interface{}               `json:"page_info"`
	Errors   interface{}               `json:"errors"`
}

type MarsevFilterProgramData struct {
	ID                         string      `json:"id"`
	ProgramName                string      `json:"program_name"`
	MINumber                   int         `json:"mi_number"`
	PeriodStart                string      `json:"period_start"`
	PeriodEnd                  string      `json:"period_end"`
	Priority                   int         `json:"priority"`
	Description                string      `json:"description"`
	ProductID                  string      `json:"product_id"`
	ProductOfferingID          string      `json:"product_offering_id"`
	ProductOfferingDescription string      `json:"product_offering_description"`
	Tenors                     interface{} `json:"tenors"`
}

type MarsevCalculateInstallmentResponse struct {
	Code    int                              `json:"code"`
	Message string                           `json:"message"`
	Data    []MarsevCalculateInstallmentData `json:"data"`
	Errors  interface{}                      `json:"errors"`
}

type MarsevCalculateInstallmentData struct {
	InstallmentTypeCode        string  `json:"installment_type_code"`
	IsPSA                      bool    `json:"is_psa"`
	Tenor                      int     `json:"tenor"`
	AdminFee                   float64 `json:"admin_fee"`
	AdminFeePSA                float64 `json:"admin_fee_psa"`
	ProvisionFee               float64 `json:"provision_fee"`
	AmountOfFinance            float64 `json:"amount_of_finance"`
	DPAmount                   float64 `json:"dp_amount"`
	DPPercent                  float64 `json:"dp_percent"`
	AdditionalRate             float64 `json:"additional_rate"`
	EffectiveRate              float64 `json:"effective_rate"`
	LifeInsurance              float64 `json:"life_insurance"`
	AssetInsurance             float64 `json:"asset_insurance"`
	TotalInsurance             float64 `json:"total_insurance"`
	FiduciaFee                 float64 `json:"fiducia_fee"`
	NTF                        float64 `json:"ntf"`
	MonthlyInstallment         float64 `json:"monthly_installment"`
	MonthlyInstallmentMin      float64 `json:"monthly_installment_min"`
	MonthlyInstallmentMax      float64 `json:"monthly_installment_max"`
	TotalLoan                  float64 `json:"total_loan"`
	AmountOfInterest           float64 `json:"amount_of_interest"`
	FlatRateYearlyPercent      float64 `json:"flat_rate_yearly_percent"`
	FlatRateMonthlyPercent     float64 `json:"flat_rate_monthly_percent"`
	ProductID                  string  `json:"product_id"`
	ProductOfferingID          string  `json:"product_offering_id"`
	ProductOfferingDescription string  `json:"product_offering_description"`
	SubsidyAmountScheme        float64 `json:"subsidy_amount_scheme"`
	FineAmount                 float64 `json:"fine_amount"`
	FineAmountFormula          string  `json:"fine_amount_formula"`
	FineAmountDetail           string  `json:"fine_amount_detail"`
	NTFFormula                 string  `json:"ntf_formula"`
	NTFDetail                  string  `json:"ntf_detail"`
	AmountOfInterestFormula    string  `json:"amount_of_interest_formula"`
	AmountOfInterestDetail     string  `json:"amount_of_interest_detail"`
	WanprestasiFreightFee      float64 `json:"wanprestasi_freight_fee"`
	ExternalFreightFee         float64 `json:"external_freight_fee"`
	WanprestasiFreightFormula  string  `json:"wanprestasi_freight_formula"`
	WanprestasiFreightDetail   string  `json:"wanprestasi_freight_detail"`
	ExternalFreightFormula     string  `json:"external_freight_formula"`
	ExternalFreightDetail      string  `json:"external_freight_detail"`
	IsStampDutyAsLoan          *bool   `json:"is_stamp_duty_as_loan"`
	StampDutyFee               float64 `json:"stamp_duty_fee"`
}

type MDMAgreementByLicensePlateResponse struct {
	Code     string      `json:"code"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data"`
	Errors   interface{} `json:"errors"`
	Metadata interface{} `json:"metadata,omitempty"`
}

type MDMMasterMappingLicensePlateResponse struct {
	Code      string                           `json:"code"`
	Message   string                           `json:"message"`
	Data      MDMMasterMappingLicensePlateData `json:"data"`
	Errors    interface{}                      `json:"errors"`
	RequestID string                           `json:"request_id"`
	Timestamp string                           `json:"timestamp"`
}

type MDMMasterMappingLicensePlateData struct {
	Records     []MDMMasterMappingLicensePlateRecord `json:"records"`
	MaxPage     int                                  `json:"max_page"`
	Total       int                                  `json:"total"`
	PageSize    int                                  `json:"page_size"`
	CurrentPage int                                  `json:"current_page"`
}

type MDMMasterMappingLicensePlateRecord struct {
	PlateAreaID     int     `json:"plate_area_id"`
	PlateID         int     `json:"plate_id"`
	PlateCode       string  `json:"plate_code"`
	AreaID          string  `json:"area_id"`
	AreaDescription string  `json:"area_description"`
	LobID           int     `json:"lob_id"`
	CreatedAt       string  `json:"created_at"`
	CreatedBy       string  `json:"created_by"`
	UpdatedAt       *string `json:"updated_at,omitempty"`
	UpdatedBy       *string `json:"updated_by,omitempty"`
	DeletedAt       *string `json:"deleted_at,omitempty"`
	DeletedBy       *string `json:"deleted_by,omitempty"`
}

type MDMMasterDetailBranchResponse struct {
	Code      string                    `json:"code"`
	Message   string                    `json:"message"`
	Data      MDMMasterDetailBranchData `json:"data"`
	Errors    interface{}               `json:"errors"`
	RequestID string                    `json:"request_id"`
	Timestamp string                    `json:"timestamp"`
}

type MDMMasterDetailBranchData struct {
	BranchID      string  `json:"branch_id"`
	BranchName    string  `json:"branch_name"`
	CreatedAt     string  `json:"created_at"`
	CreatedBy     string  `json:"created_by"`
	UpdatedAt     *string `json:"updated_at,omitempty"`
	UpdatedBy     *string `json:"updated_by,omitempty"`
	IsActive      bool    `json:"is_active"`
	BranchAddress string  `json:"branch_address"`
}

type SallySubmit2wPrincipleResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Errors    interface{} `json:"errors"`
	RequestID string      `json:"request_id"`
	Timestamp string      `json:"timestamp"`
}

type MDMMasterMappingBranchEmployeeResponse struct {
	Code     string                                 `json:"code"`
	Message  string                                 `json:"message"`
	Data     []MDMMasterMappingBranchEmployeeRecord `json:"data"`
	Errors   interface{}                            `json:"errors"`
	Metadata MDMMasterMappingBranchEmployeeMetadata `json:"metadata"`
}

type MDMMasterMappingBranchEmployeeRecord struct {
	ID         int     `json:"id"`
	BranchID   string  `json:"branch_id"`
	BranchName string  `json:"branch_name"`
	CMOID      string  `json:"cmo_id"`
	CMOName    string  `json:"cmo_name"`
	LobID      int     `json:"lob_id"`
	CreatedAt  string  `json:"created_at"`
	CreatedBy  string  `json:"created_by"`
	UpdatedAt  *string `json:"updated_at,omitempty"`
	UpdatedBy  *string `json:"updated_by,omitempty"`
	DeletedAt  *string `json:"deleted_at,omitempty"`
	DeletedBy  *string `json:"deleted_by,omitempty"`
}

type MDMMasterMappingBranchEmployeeMetadata struct {
	Pagination MDMMasterMappingBranchEmployeePagination `json:"pagination"`
}

type MDMMasterMappingBranchEmployeePagination struct {
	Limit    int  `json:"limit"`
	NextPage bool `json:"next_page"`
	Page     int  `json:"page"`
	PrevPage bool `json:"prev_page"`
	Total    int  `json:"total"`
}

type AssetYearList struct {
	Records []struct {
		AssetCode        string `json:"asset_code"`
		BranchID         string `json:"branch_id"`
		Brand            string `json:"brand"`
		ManufactureYear  int    `json:"manufacturing_year"`
		MarketPriceValue int    `json:"market_price_value"`
	} `json:"records"`
}

type PrincipleElaborateLTV struct {
	LTV               int         `json:"ltv"`
	AdjustTenor       bool        `json:"adjust_tenor"`
	MaxTenor          int         `json:"max_tenor"`
	Reason            string      `json:"reason"`
	LoanAmountMaximum float64     `json:"loan_amount_maximum"`
	IsPsa             interface{} `json:"is_psa,omitempty"`
	Dealer            interface{} `json:"dealer,omitempty"`
	InstallmentAmount interface{} `json:"installment_amount,omitempty"`
	AF                interface{} `json:"af,omitempty"`
	AdminFee          interface{} `json:"admin_fee,omitempty"`
	NTF               interface{} `json:"ntf,omitempty"`
	AssetCategoryID   interface{} `json:"asset_category_id,omitempty"`
	Otr               interface{} `json:"otr,omitempty"`
}

type AssetList struct {
	Records []struct {
		AssetCode           string `json:"asset_code"`
		AssetDescription    string `json:"asset_description"`
		AssetDisplay        string `json:"asset_display"`
		AssetTypeID         string `json:"asset_type_id"`
		BranchID            string `json:"branch_id"`
		Brand               string `json:"brand"`
		CategoryID          string `json:"category_id"`
		CategoryDescription string `json:"category_description"`
		IsElectric          bool   `json:"is_electric"`
		Model               string `json:"model"`
	} `json:"records"`
}
