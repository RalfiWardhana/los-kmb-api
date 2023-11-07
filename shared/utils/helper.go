package utils

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"math"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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

func GetIsMedia(urlImage string) bool {

	urlMedia := strings.Split(os.Getenv("URL_MEDIA"), ",")

	for _, url := range urlMedia {
		if strings.Contains(urlImage, url) {
			return true
		}
	}

	return false
}

func DecodeNonMedia(url string) (base64Image string, err error) {

	image, err := http.Get(url)

	if err != nil {
		return
	}

	reader := bufio.NewReader(image.Body)
	ioutil, err := ioutil.ReadAll(reader)

	if err != nil {
		return
	}

	base64Image = base64.StdEncoding.EncodeToString(ioutil)

	return
}

func SafeEncoding(arrByte []byte) []byte {

	arrByte = bytes.ReplaceAll(arrByte, []byte("\\u0026"), []byte("&"))
	arrByte = bytes.ReplaceAll(arrByte, []byte("\\u003c"), []byte("<"))
	arrByte = bytes.ReplaceAll(arrByte, []byte("\\u003e"), []byte(">"))

	return arrByte
}

func SafeJsonReplacer(myString string) string {
	r := strings.NewReplacer(
		"\\u0026", "&",
		"\\u003c", "<",
		"\\u003e", ">",
		"\\n", "",
		"  ", "",
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

	return fmt.Sprintf("WHERE tt.BranchID IN (%s)", branch)
}

func GenerateFilter(search, filterBranch, rangeDays string) string {
	var filter string

	if search != "" {
		if filterBranch != "" {
			filter = filterBranch + fmt.Sprintf(" AND (tt.ProspectID LIKE '%%%s%%' OR tt.IDNumber LIKE '%%%s%%' OR tt.LegalName LIKE '%%%s%%')", search, search, search)
		} else {
			filter = fmt.Sprintf("WHERE (tt.ProspectID LIKE '%%%s%%' OR tt.IDNumber LIKE '%%%s%%' OR tt.LegalName LIKE '%%%s%%')", search, search, search)
		}
	} else {
		if filterBranch != "" {
			filter = filterBranch + fmt.Sprintf(" AND CAST(tt.created_at AS date) >= DATEADD(day, %s, CAST(GETDATE() AS date))", rangeDays)
		} else {
			filter = fmt.Sprintf("WHERE CAST(tt.created_at AS date) >= DATEADD(day, %s, CAST(GETDATE() AS date))", rangeDays)
		}
	}

	return filter
}

func ApprovalScheme(req request.ReqSubmitApproval) (result response.RespApprovalScheme, err error) {
	// get master limit
	limit := []entity.MappingLimitApprovalScheme{
		{
			Alias: "CBM",
		},
		{
			Alias: "DRM",
		},
		{
			Alias: "GMO",
		},
		{
			Alias: "COM",
		},
		{
			Alias: "GMC",
		},
		{
			Alias: "UCC",
		},
	}

	for i, v := range limit {
		if req.Alias == v.Alias {
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
