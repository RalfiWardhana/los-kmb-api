package request

type BodyRequest struct {
	ClientKey string `json:"client_key" validate:"required,key"`
	Data      Data   `json:"data" validate:"required"`
}

type BodyRequestKreditmu struct {
	ClientKey      string `json:"client_key" validate:"required,key"`
	Data           Data   `json:"data" validate:"required"`
	StatusKonsumen string `json:"status_konsumen"`
}

type Blacklist struct {
	IDNumber          string `json:"IDNumber" validate:"required"`
	LegalName         string `json:"LegalName" validate:"required,allowcharsname"`
	BirthPlace        string `json:"BirthPlace" validate:"required,allowcharsname"`
	BirthDate         string `json:"BirthDate" validate:"required,"`
	SurgateMotherName string `json:"SurgateMotherName" validate:"required,allowcharsname"`
	Gender            string `json:"Gender" validate:"required,"`
	MaritalStatus     string `json:"MaritalStatus" validate:"required,"`
	Spouse            Spouse `json:"Spouse" validate:"omitempty"`
}

type Data struct {
	BPKBName          string  `json:"BPKBName" validate:"required,bpkbname"`
	ProspectID        string  `json:"ProspectID" validate:"required"`
	BranchID          string  `json:"BranchID" validate:"required"`
	IDNumber          string  `json:"IDNumber" validate:"required,number"`
	LegalName         string  `json:"LegalName" validate:"required,allowcharsname"`
	BirthPlace        string  `json:"BirthPlace" validate:"required,allowcharsname"`
	BirthDate         string  `json:"BirthDate" validate:"required"`
	SurgateMotherName string  `json:"SurgateMotherName" validate:"required,allowcharsname"`
	Gender            string  `json:"Gender" validate:"required,gender"`
	MaritalStatus     string  `json:"MaritalStatus" validate:"required,marital"`
	ProfessionID      string  `json:"ProfessionID" validate:"required,profession"`
	Spouse            *Spouse `json:"Spouse" validate:"omitempty"`
	MobilePhone       string  `json:"MobilePhone" validate:"required,number"`
}

type Spouse struct {
	IDNumber          string `json:"Spouse_IDNumber"  validate:"required,number"`
	LegalName         string `json:"Spouse_LegalName"  validate:"required,allowcharsname"`
	BirthPlace        string `json:"Spouse_BirthPlace"  validate:"required,allowcharsname"`
	BirthDate         string `json:"Spouse_BirthDate"  validate:"required"`
	SurgateMotherName string `json:"Spouse_SurgateMotherName"  validate:"required,allowcharsname"`
	Gender            string `json:"Spouse_Gender"  validate:"required"`
}

type GenderCompare struct {
	Gender bool `json:"Spouse_Gender" validate:"spouse_gender"`
}
