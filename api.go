package main

import (
	"net"
	"time"
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"flag"
)

type MSS map[string]string
type MI map[string]interface{}

type GPU struct {
	Status string
}

type Summary struct {
	MHSav float64 `json:"MHS av"`
	MHS5s float64 `json:"MHS 5s"`
}

type Response struct {
	SUMMARY []*Summary
	DEVS []*GPU
}

func GetGPUStatus() []string {
	var out []string
	devinf := MakeReqAlt(MSS{"command":"devs"})
	for _,gpu := range devinf.DEVS {
		status := gpu.Status
		out = append(out, status)
	}
	return out
}

func GetCurrentHashRate() (float64, float64) {
	resp := MakeReqAlt(MSS{"command":"summary"})
	return resp.SUMMARY[0].MHSav,resp.SUMMARY[0].MHS5s
}

func MakeReqAlt(req MSS) *Response {
	con,err := net.Dial("tcp", "localhost:4028")
	if err != nil {
		panic(err)
	}
	b,_ := json.Marshal(req)
	con.Write(b)
	out,_ := ioutil.ReadAll(con)
	out = out[:len(out)-1]
	response := new(Response)
	err = json.Unmarshal(out, response)
	if err != nil {
		panic(err)
	}
	return response
}

//Reboot computer
func Reboot() {
	cmd := exec.Command("reboot")
	cmd.Run()
}

var pollfreq = flag.Int("poll", 30, "Time in seconds to wait between polling")
var rebtime = flag.Int("rebt", 5, "Time to sleep after detecting a failure before rebooting")

func main() {
	for {
		st := GetGPUStatus()
		for _,v := range st {
			if v != "Alive" {
				log.Printf("Rebooting in %d seconds...\n", *rebtime)
				time.Sleep(time.Second * time.Duration(*rebtime))
				Reboot()
			}
		}
		av,rec := GetCurrentHashRate()
		log.Printf("All GPU's healthy! Hashrate: [%fMhs,%fMhs] Sleeping %ds", rec,av,*pollfreq)
		time.Sleep(time.Second * time.Duration(*pollfreq))
	}
}
