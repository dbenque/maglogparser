package utils

import (
	"fmt"
	"strconv"
	"time"
)

const DateFormat = "2006/01/02 15:04:05.000000"

// ParseDate1 parse date using basic Go primitive
func ParseDate1(strdate string) (time.Time, error) {
	return time.Parse(DateFormat, strdate)
}

// ParseDate2 parse date using dedicated function
func ParseDate2(strdate string) (time.Time, error) {
	year, _ := strconv.Atoi(strdate[:4])
	month, _ := strconv.Atoi(strdate[5:7])
	day, _ := strconv.Atoi(strdate[8:10])
	hour, _ := strconv.Atoi(strdate[11:13])
	minute, _ := strconv.Atoi(strdate[14:16])
	second, _ := strconv.Atoi(strdate[17:19])
	us, _ := strconv.Atoi(strdate[20:26])

	return time.Date(year, time.Month(month), day, hour, minute, second, us, time.UTC), nil
}

// ParseDate3 parse date using basic Go primitive
func ParseDate3(date []byte) (time.Time, error) {
	year := ((((int(date[0])-'0')*100 + (int(date[1])-'0')*10) + (int(date[2]) - '0')) * 10) + (int(date[3]) - '0')
	month := time.Month(((int(date[5]) - '0') * 10) + (int(date[6]) - '0'))
	day := ((int(date[8]) - '0') * 10) + (int(date[9]) - '0')
	hour := ((int(date[11]) - '0') * 10) + (int(date[12]) - '0')
	minute := ((int(date[14]) - '0') * 10) + (int(date[15]) - '0')
	second := ((int(date[17]) - '0') * 10) + (int(date[18]) - '0')
	us := (int(date[20])-'0')*100000 + (int(date[21])-'0')*10000 + (int(date[22])-'0')*1000 + (int(date[23])-'0')*100 + (int(date[24])-'0')*10 + (int(date[25]) - '0')
	return time.Date(year, month, day, hour, minute, second, us*1000, time.UTC), nil
}

// ParseDate4 parse date optimized
func ParseDate4(date string) (time.Time, error) {
	year := ((((int(date[0])-'0')*100 + (int(date[1])-'0')*10) + (int(date[2]) - '0')) * 10) + (int(date[3]) - '0')
	month := time.Month(((int(date[5]) - '0') * 10) + (int(date[6]) - '0'))
	day := ((int(date[8]) - '0') * 10) + (int(date[9]) - '0')
	hour := ((int(date[11]) - '0') * 10) + (int(date[12]) - '0')
	minute := ((int(date[14]) - '0') * 10) + (int(date[15]) - '0')
	second := ((int(date[17]) - '0') * 10) + (int(date[18]) - '0')
	us := (int(date[20])-'0')*100000 + (int(date[21])-'0')*10000 + (int(date[22])-'0')*1000 + (int(date[23])-'0')*100 + (int(date[24])-'0')*10 + (int(date[25]) - '0')
	return time.Date(year, month, day, hour, minute, second, us*1000, time.UTC), nil
}

func ParseDate5(date string) (time.Time, error) {
	if date[4] != '/' || date[4] != date[7] || date[13] != ':' || date[13] != date[16] {
		return time.Unix(0, 0), fmt.Errorf("Bad date format")
	}
	year := ((((int(date[0])-'0')*100 + (int(date[1])-'0')*10) + (int(date[2]) - '0')) * 10) + (int(date[3]) - '0')
	month := time.Month(((int(date[5]) - '0') * 10) + (int(date[6]) - '0'))
	day := ((int(date[8]) - '0') * 10) + (int(date[9]) - '0')
	hour := ((int(date[11]) - '0') * 10) + (int(date[12]) - '0')
	minute := ((int(date[14]) - '0') * 10) + (int(date[15]) - '0')
	second := ((int(date[17]) - '0') * 10) + (int(date[18]) - '0')
	us := (int(date[20])-'0')*100000 + (int(date[21])-'0')*10000 + (int(date[22])-'0')*1000 + (int(date[23])-'0')*100 + (int(date[24])-'0')*10 + (int(date[25]) - '0')
	return time.Date(year, month, day, hour, minute, second, us*1000, time.UTC), nil
}
