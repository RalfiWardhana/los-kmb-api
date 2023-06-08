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
	Code                   string  `json:"code"`
	Decision               string  `json:"decision"`
	Reason                 string  `json:"reason"`
	StatusKonsumen         string  `json:"status_konsumen"`
	KategoriStatusKonsumen string  `json:"kategori_status_konsumen,omitempty"`
	IsBlacklist            int     `json:"is_blacklist"`
	NextProcess            int     `json:"next_process"`
	TotalBakiDebet         float64 `json:"total_baki_debet,omitempty"`
	PbkReport              string  `json:"pbk_report,omitempty"`
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
	OsInstallmentdue                 int     `json:"os_installmentdue"`
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
	SpouseIncome                     int     `json:"spouse_income"`
	SurgateMotherName                string  `json:"surgate_mother_name"`
	TotalInstallment                 int     `json:"total_installment"`
	TotalInstallmentNap              int     `json:"total_installment_nap"`
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

type ResposePefindo struct {
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
	Code     int    `json:"code"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
	LTV      int    `json:"ltv,omitempty"`
}
