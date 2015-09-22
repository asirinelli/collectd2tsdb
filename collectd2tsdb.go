package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Message struct {
	Values          []float64
	Dstypes         []string
	Dsnames         []string
	Time            float64
	Interval        float32
	Host            string
	Plugin          string
	Plugin_instance string
	Type            string
	Type_instance   string
}

type OpentsdbValue struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     float64           `json:"value"`
	TagList   map[string]string `json:"tags"`
}

type Config struct {
	Endpoint string
	User     string
	Password string
	Bind     string
}

var config = Config{}

func root(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	d := make([]Message, 0)
	out := make([]OpentsdbValue, 0)
	err := decoder.Decode(&d)
	if err != nil {
		log.Println(err)
	}

	for _, element := range d {
		for ii, value := range element.Values {
			fields := []string{element.Host}
			for _, f := range []string{element.Plugin,
				element.Plugin_instance,
				element.Type,
				element.Type_instance,
				element.Dsnames[ii]} {
				if f != "" && f != fields[len(fields)-1] {
					fields = append(fields, f)
				}
			}
			if fields[len(fields)-1] == "value" {
				fields = fields[:len(fields)-1]
			}
			name := strings.Join(fields, ".")
			export := OpentsdbValue{
				Metric:    name,
				Timestamp: int64(element.Time),
				Value:     value,
			}
			out = append(out, export)
		}
	}

	sendToOpentsdb(out)

}

func sendToOpentsdb(data []OpentsdbValue) {
	client := http.Client{}
	j, err := json.Marshal(data)
	if err != nil {
		log.Fatal("Json error:", err)
	}

	req, err := http.NewRequest("POST", config.Endpoint, bytes.NewReader(j))

	if err == nil {
		req.SetBasicAuth(config.User, config.Password)
		resp, err := client.Do(req)
		defer resp.Body.Close()

		if err != nil {
			log.Println("error:", err)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("error:", err)
			return
		}
		log.Println(len(data), " data points sent.")
		if resp.StatusCode != 204 {
			log.Println(fmt.Sprintf("Response from server: (%s) %s", resp.Status, string(body)))
		}
	}

}

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "collectd2tsdb.json",
		"Configuration file")
	flag.Parse()

	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	http.HandleFunc("/opentsdb", root)
	log.Println("Listening on", config.Bind, "and forwarding to", config.Endpoint)
	err = http.ListenAndServe(config.Bind, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
