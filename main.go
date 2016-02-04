package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	endpoint   string
	tsuruToken string
	tsuruHost  string
	apps       map[string]string = map[string]string{}
)

type document struct {
	Client string `json:"client"`
	Metric string `json:"metric"`
	Count  int    `json:"count"`
	App    string `json:"app"`
	Value  int    `json:"value"`
}

type app struct {
	Name  string   `json:"name"`
	Ip    string   `json:"ip"`
	Cname []string `json:"cname"`
}

func getApps() error {
	url := fmt.Sprintf("%s/apps", tsuruHost)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "b "+tsuruToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error trying to get apps info: HTTP %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error trying to get apps info: %s", err)
	}
	appList := []app{}
	err = json.Unmarshal(body, &appList)
	if err != nil {
		return err
	}
	for _, a := range appList {
		apps[a.Ip] = a.Name
		for _, cname := range a.Cname {
			apps[cname] = a.Name
		}
	}
	return nil
}

func sendDocument(doc *document) error {
	addr, err := net.ResolveUDPAddr("udp", endpoint)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	b, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	_, err = conn.Write(b)
	return err
}

func appFromAddr(addr string) string {
	return apps[addr]
}

func parseAddr(addr string) string {
	return strings.Replace(addr, "_", ".", -1)
}

func handle(data []byte) (*document, error) {
	r := regexp.MustCompile(`galeb\.(?P<addr>[\w-_]+)\..*.requestTime:(?P<value>\d+)|ms.*`)
	d := r.FindStringSubmatch(string(data))
	value, err := strconv.Atoi(d[2])
	if err != nil {
		return nil, err
	}
	app := appFromAddr(parseAddr(d[1]))
	doc := &document{
		Client: "tsuru",
		Metric: "response_time",
		Count:  1,
		App:    app,
		Value:  value,
	}
	return doc, nil
}

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":8125")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	for {
		buf := make([]byte, 1600)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Print(err)
		}
		document, err := handle(buf[0:n])
		err = sendDocument(document)
		if err != nil {
			log.Print(err)
		}
	}
}
