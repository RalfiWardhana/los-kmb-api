package dupcheck

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

type CustomerDomain struct {
	Code      string             `json:"code"`
	Message   string             `json:"message"`
	Data      CustomerDomainData `json:"data"`
	Errors    interface{}        `json:"errors"`
	RequestID string             `json:"request_id"`
	Timestamp string             `json:"timestamp"`
}

type CustomerDomainData struct {
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
