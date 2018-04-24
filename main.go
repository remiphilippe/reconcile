package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/golang/glog"

	goh4 "github.com/remiphilippe/go-h4"
)

func readCSV(path string) (map[string][]string, error) {
	csvFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	if _, err = reader.Read(); err != nil { //read header
		return nil, err
	}
	hostMap := map[string][]string{}
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			return nil, err
		}
		hostname := line[0]
		ip := line[1]
		hostMap[hostname] = append(hostMap[hostname], ip)
	}

	return hostMap, nil
}

func getSensors(config *Config, vrf int) (map[string][]string, error) {
	h4 := new(goh4.H4)
	h4.Secret = config.APISecret
	h4.Key = config.APIKey
	h4.Endpoint = config.APIEndpoint
	h4.Verify = config.APIVerify
	h4.Prefix = "/openapi/v1"

	results, err := h4.GetSWAgents()
	if err != nil {
		return nil, err
	}

	a := map[string][]string{}

	for _, r := range results {
		// if r.Hostname == "collectorDatamover-4" {
		// 	spew.Dump(r)
		// }
		for _, i := range r.Interfaces {
			// Are we in the right VRF?
			if i.VRFID == vrf {
				// Exclude localhost or this will mess up the matching
				if i.IP.String() != "127.0.0.1" && i.IP.String() != "::1" {
					hostname := strings.ToLower(r.Hostname)
					//hostname := r.Hostname
					a[hostname] = append(a[hostname], i.IP.String())
				}
			}
		}
	}

	return a, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func diff(desired map[string][]string, actual map[string][]string) map[string]map[string][]string {
	hostNotFound := map[string][]string{}
	missing := map[string][]string{}
	incorrectIP := map[string][]string{}
	incorrectHostname := map[string][]string{}
	allGood := map[string][]string{}

	var expectedHostname, actualHostname string

	// Cases:
	// Hostname Found and IP match for this Hostname
	// Hostname Found but no IP match
	// Hostname not Found and no IP match
	// Hostname not Found but IP match for another hostname

	for k, v := range desired {
		expectedHostname = strings.ToLower(k)
		if _, ok := actual[expectedHostname]; ok {
			// Hostname found
			if len(intersect(v, actual[expectedHostname]).([]interface{})) > 0 {
				// IP Match
				glog.V(1).Infof("Hostname and IP found: %s\n", k)
				allGood[k] = v
			} else {
				// IP not found
				glog.V(1).Infof("Hostname found but IP not found: %s\n", k)
				incorrectIP[k] = v
			}
		} else {
			// Hostname not found
			hostNotFound[k] = v
		}
	}

	// For missing items, let's see if they are there but with a different hostname
	found := false
	for e, m := range hostNotFound {
		expectedHostname = strings.ToLower(e)
		for h, s := range actual {
			actualHostname = strings.ToLower(h)
			i := intersect(m, s).([]interface{})
			if len(i) > 0 {
				glog.V(1).Infof("IP found but wrong hostname for %s: actual hostname: %s, expected hostname: %s\n", i[0], actualHostname, expectedHostname)
				incorrectHostname[e] = m
				found = true
				break
			}
		}
		if !found {
			// Didn't find IP
			glog.V(1).Infof("Missing: %s\n", e)
			missing[e] = m
		}
	}

	retval := map[string]map[string][]string{}
	retval["ok"] = allGood
	retval["missing"] = missing
	retval["wrong_ip"] = incorrectIP
	retval["wrong_hostname"] = incorrectHostname

	return retval
}

func main() {
	inputCSV := flag.String("input", "", "input CSV name")
	vrf := flag.Int("vrf", 1, "VRF ID (default 1)")
	flag.Parse()

	if *inputCSV == "" {
		flag.PrintDefaults()
		os.Exit(2)
	}

	config := NewConfig()

	hosts, err := readCSV(*inputCSV)
	if err != nil {
		glog.Errorf("%s", err.Error())
		os.Exit(2)
	}

	sensors, err := getSensors(config, *vrf)
	if err != nil {
		glog.Errorf("%s", err)
		os.Exit(2)
	}

	val := diff(hosts, sensors)

	fmt.Println("----")
	fmt.Println("Hosts OK:")
	for k, v := range val["ok"] {
		fmt.Printf("Hostname: %s, expectedIPs: %s, actualIPs: %s\n", k, strings.Join(v, ","), strings.Join(sensors[strings.ToLower(k)], ", "))
	}
	fmt.Println("----")

	fmt.Println("Hosts Missing:")
	for k, v := range val["missing"] {
		fmt.Printf("Hostname: %s, expectedIPs: %s\n", k, strings.Join(v, ", "))
	}
	fmt.Println("----")

	fmt.Println("Hostname exists but wrong IP:")
	for k, v := range val["wrong_ip"] {
		fmt.Printf("Hostname: %s, expectedIPs: %s, actualIPs: %s\n", k, strings.Join(v, ","), strings.Join(sensors[strings.ToLower(k)], ", "))
	}
	fmt.Println("----")

	fmt.Println("IP exists but wrong Hostname:")
	for k, v := range val["wrong_hostname"] {
		fmt.Printf("Hostname: %s, actualIPs: %s\n", k, strings.Join(v, ", "))
	}
	fmt.Println("----")

	glog.Flush()
}
