package utils

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
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
