package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

type ValCurs struct {
	Date    string   `xml:"Date,attr"`
	Valutes []Valute `xml:"Valute"`
}

type Valute struct {
	ID       string `xml:"ID,attr"`
	NumCode  string `xml:"NumCode"`
	CharCode string `xml:"CharCode"`
	Nominal  string `xml:"Nominal"`
	Name     string `xml:"Name"`
	Value    string `xml:"Value"`
}

type valuteStamp struct {
	value float64
	date  string
}

func iHateWindows(charset string, input io.Reader) (io.Reader, error) {
	switch charset {
	case "windows-1251":
		return charmap.Windows1251.NewDecoder().Reader(input), nil
	default:
		return nil, fmt.Errorf("unknown charset: %s", charset)
	}
}

func getRespondFromApi(url string) ValCurs {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var valCurs ValCurs
	d := xml.NewDecoder(resp.Body)
	d.CharsetReader = iHateWindows
	err = d.Decode(&valCurs)
	if err != nil {
		panic(err)
	}
	return valCurs
}

func getMin(data []valuteStamp) (float64, string) {
	minIndex := 0
	for i := 0; i < len(data); i++ {
		if data[i].value < data[minIndex].value {
			minIndex = i
		}
	}
	return data[minIndex].value, data[minIndex].date
}

func getMax(data []valuteStamp) (float64, string) {
	maxIndex := 0
	for i := 0; i < len(data); i++ {
		if data[i].value > data[maxIndex].value {
			maxIndex = i
		}
	}
	return data[maxIndex].value, data[maxIndex].date
}

func getAvg(data []valuteStamp) float64 {
	var sum float64
	for _, value := range data {
		sum += value.value
	}
	return sum / float64(len(data))
}

func main() {
	today := time.Now()
	countries := make(map[string][]valuteStamp)
	baseUrl := "http://www.cbr.ru/scripts/XML_daily_eng.asp?date_req="
	for i := 0; i < 90; i++ {
		neededDate := today.AddDate(0, 0, -90+i+1)
		url := baseUrl + neededDate.Format("02/01/2006")
		for _, valute := range getRespondFromApi(url).Valutes {
			_, ok := countries[valute.Name]
			if !ok {
				countries[valute.Name] = make([]valuteStamp, 0, 90)
			}
			value, err := strconv.ParseFloat(strings.ReplaceAll(valute.Value, ",", "."), 64)
			if err != nil {
				log.Println(err)
				continue
			}
			nomimal, err := strconv.Atoi(valute.Nominal)
			if err != nil {
				log.Println(err)
				continue
			}
			slice := countries[valute.Name]
			slice = append(slice, valuteStamp{value: value / float64(nomimal), date: neededDate.Format("02/01/2006")})
			countries[valute.Name] = slice
		}

	}
	fmt.Printf("%-30s%-8s%-20s%-8s%-20s%s\n", "Name", "Min", "Date", "Max", "Date", "Avg")
	fmt.Printf("-------------------------------------------------------------------------------------------\n")
	for key, value := range countries {
		minValue, minDate := getMin(value)
		maxValue, maxDate := getMax(value)
		avg := getAvg(value)
		minValueString := strconv.FormatFloat(minValue, 'f', 2, 64)
		maxValueString := strconv.FormatFloat(maxValue, 'f', 2, 64)
		avgString := strconv.FormatFloat(avg, 'f', 2, 64)
		fmt.Printf("%-30s%-8s%-20s%-8s%-20s%s\n", key, minValueString, minDate, maxValueString, maxDate, avgString)
	}

}
