package main

import (
	"net"
	"time"
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"flag"
	"os"
)

type MSS map[string]string
type MI map[string]interface{}

type GPU struct {
	Status string
	Temperature float64
	Hashrate float64 `json:"MHS av"`
}

type Summary struct {
	MHSav float64 `json:"MHS av"`
	MHS5s float64 `json:"MHS 5s"`
}

type Response struct {
	SUMMARY []*Summary
	DEVS []*GPU
}

func GetGPUStatus() []*GPU {
	devinf := MakeReqAlt(MSS{"command":"devs"})
	return devinf.DEVS
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
	logfi,err := os.Create("mining.log")
	if err != nil {
		panic(err)
	}
	log.SetOutput(logfi)
	defer logfi.Close()
	for {
		st := GetGPUStatus()
		for i,gpu := range st {
			if gpu.Status != "Alive" {
				log.Printf("GPU %d is %d.\n", i, gpu.Status)
				log.Printf("Rebooting in %d seconds...\n", *rebtime)
				time.Sleep(time.Second * time.Duration(*rebtime))
				Reboot()
			}
			log.Printf("GPU %d: Temp %f Hashrate: %fMh/s\n", i, gpu.Temperature, gpu.Hashrate)
		}
		av,rec := GetCurrentHashRate()
		log.Printf("All GPU's healthy! Hashrate: [%fMhs,%fMhs] Sleeping %ds", rec,av,*pollfreq)
		time.Sleep(time.Second * time.Duration(*pollfreq))
	}
}
