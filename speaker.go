package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/araddon/dateparse"
)

type Plate struct {
	Text          string `json:"text"`
	Termin        string `json:"termin"`
	TerminName    string `json:"terminname"`
	Readall       string `json:"readall"`
	Interval      string `json:"interval"`
	IntervalValue string `json:"intervalvalue"`
}

type Termine struct {
	datum         string
	name          string
	interval      string
	intervalvalue string
}

type Delete struct {
	Idd int `json:"did"`
}

var temptermine []Termine
var terminedone []Termine

func main() {

	speakerfunction("Hallo, ich bin Kilian")

	//go routine who checks every 30 sec
	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				checkIfSomethingIsGoingOn()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	http.HandleFunc("/main", controller)
	http.HandleFunc("/all", allstuff)
	http.HandleFunc("/del", deleterfu)
	http.ListenAndServe(":3000", nil)

}

//main
func controller(rw http.ResponseWriter, req *http.Request) {
	(rw).Header().Set("Access-Control-Allow-Origin", "*")
	(rw).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(rw).Header().Set("Access-Control-Allow-Headers", "*")
	if req.Method == "OPTIONS" {
		//fmt.Println("mach nix")
		return
	}
	fmt.Println("main func")

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	p := Plate{}
	json.Unmarshal(b, &p)

	end, err := json.Marshal(p)
	if err != nil {
		return
	}

	if p.Readall == "true" {
		if len(temptermine) == 0 {
			speakerfunction("Es sind momentan keine Termine verfügbar, zumindest keine von denen ich weiß")
		} else {
			for i := 0; i < len(temptermine); i++ {
				terminformatetmomentan, err := dateparse.ParseAny(temptermine[i].datum)
				if err != nil {
					fmt.Println(err)
				}
				speakerfunction(temptermine[i].name + " ,am, " + terminformatetmomentan.Format("02-01-2006 15:04"))
			}
		}

	} else if p.Readall == "done" {
		if len(terminedone) == 0 {
			speakerfunction("Es sind keine vergangenen Termine verfügbar")
		} else {
			for i := 0; i < len(terminedone); i++ {
				terminformatetdone, err := dateparse.ParseAny(terminedone[i].datum)
				if err != nil {
					fmt.Println(err)
				}
				speakerfunction(terminedone[i].name + " ,am, " + terminformatetdone.Format("02-01-2006 15:04"))
			}
		}
	} else {
		if p.Text != "" {
			fmt.Println(p.Text)
			speakerfunction(p.Text)
		}
		if p.Termin != "" && p.TerminName != "" {
			terminformatet, err := dateparse.ParseAny(p.Termin)
			if err != nil {
				fmt.Println(err)
			}
			if p.Interval != "none" && p.IntervalValue != "" {
				speakerfunction("termin, " + p.TerminName + " am, " + terminformatet.Format("02-01-2006 15:04") + " angelegt, mit einem Interval von " + p.IntervalValue + " " + p.Interval)
				var endurancerunner Termine
				endurancerunner.datum = p.Termin
				endurancerunner.name = p.TerminName
				endurancerunner.interval = p.Interval
				endurancerunner.intervalvalue = p.IntervalValue
				temptermine = append(temptermine, endurancerunner)
			} else {
				speakerfunction("termin, " + p.TerminName + " am, " + terminformatet.Format("02-01-2006 15:04") + " angelegt, ich werde dich daran erinnern")
				var box Termine
				box.datum = p.Termin
				box.name = p.TerminName
				temptermine = append(temptermine, box)
			}
		}
	}

	//fmt.Println(temptermine)

	rw.Write(end)
}

//speaker function triggers test.py
func speakerfunction(sentence string) {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app := "python"
	arg0 := pwd + "/" + "test.py"
	arg1 := "--text"
	arg2 := "\" " + sentence + "\""

	fmt.Println(app + " " + arg0 + " " + arg1 + " " + arg2)

	cmd := exec.Command(app, arg0, arg1, arg2)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println("fail error")
		fmt.Println(err.Error())
		return
	}
	fmt.Print(string(stdout))
}

// go routine
func checkIfSomethingIsGoingOn() {
	for i := 0; i < len(temptermine); i++ {
		t, err := dateparse.ParseAny(temptermine[i].datum)
		if err != nil {
			fmt.Println(err)
		}

		actualtime := time.Now()
		t2, err2 := dateparse.ParseAny(actualtime.Format("2006-01-02 15:04:05"))

		if err2 != nil {
			fmt.Println(err2)
		}

		if t2.After(t) {
			if temptermine[i].interval != "" {
				speakerfunction("Ja moin was geht, ein Termin ist fällig, " + temptermine[i].name)
				newtime, err := dateparse.ParseAny(temptermine[i].datum)
				if err != nil {
					fmt.Println(err)
					return
				}
				timeval, err := strconv.Atoi(temptermine[i].intervalvalue)
				newtimefinish := time.Now()

				switch temptermine[i].interval {
				case "min":
					newtimefinish = newtime.Add(time.Duration(timeval) * time.Minute)
				case "std":
					newtimefinish = newtime.Add(time.Duration(timeval) * time.Hour)
				case "tag":
					newtimefinish = newtime.AddDate(0, 0, timeval)
				case "woche":
					newtimefinish = newtime.AddDate(0, 0, timeval*7)
				case "monat":
					newtimefinish = newtime.AddDate(0, timeval, 0)
				case "jahr":
					newtimefinish = newtime.AddDate(timeval, 0, 0)
				default:
					speakerfunction("ACHTUNG! da ist was gewaltig schiefgelaufen")
				}

				temptermine[i].datum = newtimefinish.String()
				speakerfunction("Ich erinner dich in einem Interval von, " + temptermine[i].intervalvalue + " " + temptermine[i].interval + " wieder daran")

			} else {
				var done Termine
				done.datum = temptermine[i].datum
				done.name = temptermine[i].name
				terminedone = append(terminedone, done)
				speakerfunction("Ja moin was geht, ein Termin ist fällig, " + temptermine[i].name)
				temptermine = append(temptermine[:i], temptermine[i+1:]...)
			}
		}
	}
}

// all
func allstuff(rw http.ResponseWriter, req *http.Request) {
	(rw).Header().Set("Access-Control-Allow-Origin", "*")
	(rw).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(rw).Header().Set("Access-Control-Allow-Headers", "*")
	if req.Method == "OPTIONS" {
		//fmt.Println("mach nix")
		return
	}

	fmt.Println("all func")

	var alles []interface{}
	if len(temptermine) > 0 {
		for i := 0; i < len(temptermine); i++ {
			t3, err2 := dateparse.ParseAny(temptermine[i].datum)
			if err2 != nil {
				fmt.Println(err2)
			}
			alles = append(alles, []interface{}{i, temptermine[i].name, t3.Format("02-01-2006 15:04:05")})
		}
	}
	all, err := json.Marshal(alles)
	if err != nil {
		return
	}

	rw.Write(all)
}

// del
func deleterfu(rw2 http.ResponseWriter, req2 *http.Request) {
	(rw2).Header().Set("Access-Control-Allow-Origin", "*")
	(rw2).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(rw2).Header().Set("Access-Control-Allow-Headers", "*")
	if req2.Method == "OPTIONS" {
		//fmt.Println("mach nix")
		return
	}

	b, err := ioutil.ReadAll(req2.Body)
	if err != nil {
		panic(err)
	}
	p := Delete{}
	json.Unmarshal(b, &p)

	speakerfunction("ich entferne " + temptermine[p.Idd].name)
	temptermine = append(temptermine[:p.Idd], temptermine[p.Idd+1:]...)

	end, err := json.Marshal(p)
	if err != nil {
		return
	}

	fmt.Println(p.Idd)

	rw2.Write(end)
}
