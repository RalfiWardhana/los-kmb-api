package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"math"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
)

func DiffTwoDate(date time.Time) time.Duration {
	//fetching current time
	loc, _ := time.LoadLocation("Asia/Jakarta")
	currentTime := time.Now().In(loc)
	//differnce between pastdate and current date
	diff := currentTime.Sub(date)
	// fmt.Printf("time difference is %v or %v in minutes\n", diff, diff.Minutes())
	return diff
}

func GenerateTimeNow() string {
	//fetching current time
	loc, _ := time.LoadLocation("Asia/Jakarta")
	currentTime := time.Now().In(loc).Format(time.RFC3339)
	//differnce between pastdate and current date
	return currentTime
}

func GenerateUnixTimeNow() int64 {
	//fetching current time
	currentTime := time.Now().Local().Unix()

	return currentTime
}

func ItemExists(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

func StrConvInt(plainText string) (newText int) {
	newText, _ = strconv.Atoi(plainText)
	return newText
}

func StrConvFloat64(plainText string) (newText float64) {
	newText, _ = strconv.ParseFloat(plainText, 64)
	return newText
}

func IntConvStr(newText int) (plainText string) {
	plainText = strconv.Itoa(newText)
	return plainText
}

func AizuArrayInt(A string, N string) []int {
	a := strings.Split(A, ",")
	n, _ := strconv.Atoi(N) // int 32bit
	b := make([]int, n)
	for i, v := range a {
		b[i], _ = strconv.Atoi(v)

	}
	return b
}

func AizuArrayString(A string) []string {
	s := strings.Split(A, `,`)
	return s
}

func GenerateID() (s string, err error) {
	b := make([]byte, 4)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	s = fmt.Sprintf("Req%x", b)

	return
}

func CheckVriable(x interface{}) string {
	xt := reflect.TypeOf(x).Kind()
	xs := ""
	if xt == reflect.String {
		xs = "string"
		return xs
	} else if xt == reflect.Float64 {
		xs = "float64"
		return xs
	} else if xt == reflect.Float32 {
		xs = "float32"
		return xs
	} else if xt == reflect.Int {
		xs = "int"
		return xs
	} else if xt == reflect.Int8 {
		xs = "int8"
		return xs
	} else if xt == reflect.Int16 {
		xs = "int16"
		return xs
	} else if xt == reflect.Int32 {
		xs = "int32"
		return xs
	} else if xt == reflect.Int64 {
		xs = "int64"
		return xs
	} else if xt == reflect.Bool {
		xs = "bool"
		return xs
	}

	return xs
}

// Convert struct to map string interface
func StructToMap(item interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	if item == nil {
		return res
	}
	v := reflect.TypeOf(item)
	if v.Kind() != reflect.Struct && v.Kind() != reflect.Ptr {
		res["data"] = item
		return res
	}

	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("json")
		if strings.Contains(tag, ",omitempty") {
			tag = strings.TrimSuffix(tag, ",omitempty")
		}
		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = StructToMap(field)
			} else {
				res[tag] = field
			}
		} else if tag == "" && field != nil &&
			v.Field(i).Type.Kind() == reflect.Struct { // Support embedded struct
			tempMap := StructToMap(field)
			for k, v := range tempMap {
				res[k] = v
			}
		}
	}
	return res
}

func GenerateTimeInMilisecond() int64 {
	return time.Now().Local().UnixNano() / int64(time.Millisecond)
}

func GenerateTimeWithFormat(format string) string {

	currentTime := time.Now().Local().Format(format)
	return currentTime
}

func Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(math.Ceil(num*output)) / output
}

func GenerateUUID() string {
	id := uuid.Must(uuid.NewRandom())
	return id.String()
}

func HumanAgeCalculator(birthdate, today time.Time) int {
	today = today.In(birthdate.Location())
	ty, tm, td := today.Date()
	today = time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC)
	by, bm, bd := birthdate.Date()
	birthdate = time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC)
	if today.Before(birthdate) {
		return 0
	}
	age := ty - by
	anniversary := birthdate.AddDate(age, 0, 0)
	if anniversary.After(today) {
		age--
	}
	return age
}

func UniqueID(length int) string {
	rand.Seed(time.Now().UnixNano())
	randomID := fmt.Sprintf("%s%d", UniqueIDFromTime(), rand.Intn(1000))
	return randomID[:length]
}

func UniqueIDFromTime() string {
	timestamp := time.Now().UnixNano()
	uniqueID := fmt.Sprintf("%s%d", MD5Hash(fmt.Sprintf("%s%d", UniqueIDFromUniqid(), timestamp)), rand.Intn(1000))
	return uniqueID
}

func UniqueIDFromUniqid() string {
	return fmt.Sprintf("%s%d", Uniqid(), rand.Intn(1000))
}

func Uniqid() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func Contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func SafeEncoding(arrByte []byte) []byte {

	arrByte = bytes.ReplaceAll(arrByte, []byte("\\u0026"), []byte("&"))
	arrByte = bytes.ReplaceAll(arrByte, []byte("\\u003c"), []byte("<"))
	arrByte = bytes.ReplaceAll(arrByte, []byte("\\u003e"), []byte(">"))

	return arrByte
}

func SafeDecoding(s string) string {
	s = strings.ReplaceAll(s, "&", "\\u0026")
	s = strings.ReplaceAll(s, "<", "\\u003c")
	s = strings.ReplaceAll(s, ">", "\\u003e")
	return s
}

func SafeJsonReplacer(myString string) string {
	r := strings.NewReplacer(
		"\\u0026", "&",
		"\\u003c", "<",
		"\\u003e", ">",
		"\\n", "",
		"\r", "",
		"\t", "",
		"\"{", "{",
		"}\"", "}",
		"\\", "",
	)
	myString = r.Replace(myString)
	return myString
}

var floatType = reflect.TypeOf(float64(0))
var stringType = reflect.TypeOf("")

func GetFloat(unk interface{}) (float64, error) {
	if unk == nil {
		unk = 0
	}
	switch i := unk.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		return strconv.ParseFloat(i, 64)
	default:
		v := reflect.ValueOf(unk)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Float(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			return strconv.ParseFloat(s, 64)
		} else {
			return math.NaN(), fmt.Errorf("Can't convert %v to float64", v.Type())
		}
	}
}

func GenerateBranchFilter(branchId string) string {
	if branchId == "" || branchId == "999" {
		return ""
	}

	arrBranch := strings.Split(branchId, ",")
	var branch string

	if len(arrBranch) == 1 && arrBranch[0] != "999" {
		branch = fmt.Sprintf("'%s'", branchId)
	} else {
		for i, val := range arrBranch {
			branch += fmt.Sprintf("'%s'", val)
			if i < len(arrBranch)-1 {
				branch += ","
			}
		}
	}

	return fmt.Sprintf("WHERE tm.BranchID IN (%s)", branch)
}

func GenerateFilter(search, encrypted, filterBranch, rangeDays, inquiryType string) string {
	var filter, filterIdNumber, filterLegalName string
	var regexpPpid = regexp.MustCompile(`SAL-|NE-`)
	var regexpIDNumber = regexp.MustCompile(`^[0-9]*$`)
	var regexpLegalName = regexp.MustCompile("^[a-zA-Z.,'` ]*$")

	switch inquiryType {
	case "NE":
		filterIdNumber = fmt.Sprintf("(tm.IDNumber = '%s')", encrypted)
		filterLegalName = fmt.Sprintf("(tm.LegalName = '%s')", encrypted)
	default:
		filterIdNumber = fmt.Sprintf("(tcp.IDNumber = '%s')", encrypted)
		filterLegalName = fmt.Sprintf("(tcp.LegalName = '%s')", encrypted)
	}

	if search != "" {
		if filterBranch != "" {
			if regexpPpid.MatchString(search) {
				filter = filterBranch + fmt.Sprintf(" AND (tm.ProspectID = '%s')", search)
			} else if regexpIDNumber.MatchString(search) {
				filter = filterBranch + fmt.Sprintf(" AND %s", filterIdNumber)
			} else if regexpLegalName.MatchString(search) {
				filter = filterBranch + fmt.Sprintf(" AND %s", filterLegalName)
			} else {
				filter = filterBranch + fmt.Sprintf(" AND (tm.ProspectID = '%s') OR %s OR %s", search, filterIdNumber, filterLegalName)
			}
		} else {
			if regexpPpid.MatchString(search) {
				filter = fmt.Sprintf("WHERE (tm.ProspectID = '%s')", search)
			} else if regexpIDNumber.MatchString(search) {
				filter = fmt.Sprintf("WHERE %s", filterIdNumber)
			} else if regexpLegalName.MatchString(search) {
				filter = fmt.Sprintf("WHERE %s", filterLegalName)
			} else {
				filter = fmt.Sprintf("WHERE (tm.ProspectID = '%s') OR %s OR %s", search, filterIdNumber, filterLegalName)
			}
		}
	} else {
		if filterBranch != "" {
			filter = filterBranch + fmt.Sprintf(" AND CAST(tm.created_at AS date) >= DATEADD(day, %s, CAST(GETDATE() AS date))", rangeDays)
		} else {
			filter = fmt.Sprintf("WHERE CAST(tm.created_at AS date) >= DATEADD(day, %s, CAST(GETDATE() AS date))", rangeDays)
		}
	}

	return filter
}

func ApprovalScheme(req request.ReqSubmitApproval) (result response.RespApprovalScheme, err error) {
	// get master limit
	limit := []entity.MappingLimitApprovalScheme{
		{
			Alias: "CBM",
			Name:  "Branch Manager",
		},
		{
			Alias: "DRM",
			Name:  "Regional Manager",
		},
		{
			Alias: "GMO",
			Name:  "GM Bisnis Operational",
		},
		{
			Alias: "COM",
			Name:  "Credit Operation Manager",
		},
		{
			Alias: "GMC",
			Name:  "GM Credit",
		},
		{
			Alias: "UCC",
			Name:  "UCC",
		},
	}

	for i, v := range limit {
		if req.Alias == v.Alias {
			result.Name = v.Name
			// add next
			if req.Alias != req.FinalApproval {
				result.NextStep = limit[i+1].Alias
				break
			} else {
				if req.NeedEscalation {
					result.NextStep = limit[i+1].Alias
					result.IsEscalation = true
				} else {
					result.IsFinal = true
				}
				break
			}
		}
	}
	return
}

func ValidateDiffMonthYear(given, today string) error {
	// Fungsi untuk memvalidasi bulan+tahun yang diberikan tidak lebih besar dari bulan+tahun hari ini

	givenDate, err := time.Parse("2006-01-02", given)
	if err != nil {
		return err
	}
	todayDate, err := time.Parse("2006-01-02", today)
	if err != nil {
		return err
	}

	if givenDate.Year() > todayDate.Year() || (givenDate.Year() == todayDate.Year() && givenDate.Month() > todayDate.Month()) {
		return errors.New("given monthYear greater than today monthYear")
	}

	return nil
}

func DiffInMonths(t1, t2 time.Time) int {
	// Fungsi untuk menghitung selisih bulan antara dua tanggal

	y1, m1, _ := t1.Date()
	y2, m2, _ := t2.Date()

	// Konversi tanggal ke bulan
	months := (y1-y2)*12 + int(m1-m2)

	return months + 1 // Tambahkan 1 karena perhitungan dimulai dari bulan berikutnya setelah t1
	// Contoh :
	// t1 08-05-2024
	// t2 10-03-2024
	// maka selisih dari 2 tanggal ini adalah 3 bulan
}

// Function to calculate the difference in months between two dates considering days as well
func PreciseMonthsDifference(date1, date2 time.Time) (int, error) {
	year1, month1, day1 := date1.Date()
	year2, month2, day2 := date2.Date()

	// Calculate the initial difference in months
	months := (year2-year1)*12 + int(month2-month1)

	// Adjust the difference if day2 is earlier in the month than day1
	if day2 < day1 {
		months--
	}

	// Check if there are additional days beyond the full months
	if day2 > day1 {
		months++
	}

	// Return an error if the difference is negative
	if months < 0 {
		return 0, errors.New("upstream_service_error - Difference of months rrd_date and current_date is negative (-)")
	}

	return months, nil
}

func CheckNullCategory(Category interface{}) float64 {
	var category float64

	if CheckVriable(Category) == reflect.String.String() {
		category = StrConvFloat64(Category.(string))
	} else {
		category = Category.(float64)
	}

	return category
}

// function to map reason category values to Roman numerals
func GetReasonCategoryRoman(category interface{}) string {
	switch category.(float64) {
	case 1:
		return "(I)"
	case 2:
		return "(II)"
	case 3:
		return "(III)"
	default:
		return ""
	}
}

func CheckNullMaxOverdueLast12Months(MaxOverdueLast12Months interface{}) float64 {
	var max_overdue_last12_months float64

	if CheckVriable(MaxOverdueLast12Months) == reflect.String.String() {
		max_overdue_last12_months = StrConvFloat64(MaxOverdueLast12Months.(string))
	} else {
		max_overdue_last12_months = MaxOverdueLast12Months.(float64)
	}

	return max_overdue_last12_months
}

func CheckNullMaxOverdue(MaxOverdueLast interface{}) float64 {
	var max_overdue_months float64

	if CheckVriable(MaxOverdueLast) == reflect.String.String() {
		max_overdue_months = StrConvFloat64(MaxOverdueLast.(string))
	} else {
		max_overdue_months = MaxOverdueLast.(float64)
	}

	return max_overdue_months
}

func CheckEmptyString(data string) interface{} {

	if data != "" {
		return data
	}

	return nil
}

func CapitalizeEachWord(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes[0]) + strings.ToLower(string(runes[1:]))
		}
	}
	return strings.Join(words, " ")
}

func GetLicensePlateCode(licensePlate string) string {
	re := regexp.MustCompile(`^[A-Z]+`)

	match := re.FindString(licensePlate)

	return match
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func Sanitize(sanitizer *bluemonday.Policy, data interface{}) {
	switch d := data.(type) {
	case map[string]interface{}:
		for k, v := range d {
			switch tv := v.(type) {
			case string:
				d[k] = sanitizer.Sanitize(tv)
			case map[string]interface{}:
				Sanitize(sanitizer, tv)
			case []interface{}:
				Sanitize(sanitizer, tv)
			case nil:
				delete(d, k)
			}
		}
	case []interface{}:
		if len(d) > 0 {
			switch d[0].(type) {
			case string:
				for i, s := range d {
					d[i] = sanitizer.Sanitize(s.(string))
				}
			case map[string]interface{}:
				for _, t := range d {
					Sanitize(sanitizer, t)
				}
			case []interface{}:
				for _, t := range d {
					Sanitize(sanitizer, t)
				}
			}
		}
	}
}
