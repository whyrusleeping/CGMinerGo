package main

import (
	"net"
	"fmt"
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
	err error
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
	response := new(Response)
	con,err := net.Dial("tcp", "localhost:4028")
	if err != nil {
		response.err = err
		log.Println(err)
		return response
	}
	b,_ := json.Marshal(req)
	_,err = con.Write(b)
	if err != nil {
		response.err = err
		log.Println(err)
		return response
	}
	out,err := ioutil.ReadAll(con)
	if err != nil {
		response.err = err
		log.Println(err)
		return response
	}
	out = out[:len(out)-1]
	err = json.Unmarshal(out, response)
	if err != nil {
		response.err = err
		log.Println(err)
		return response
	}
	return response
}

//Reboot computer
func Reboot() {
	cmd := exec.Command("reboot")
	cmd.Run()
}

func SetLogger() *os.File {
	name := ""
	for i := 0; ; i++ {
		name = fmt.Sprintf("mining.log.%d",i)
		_,err := os.Stat(name)
		if err != nil {
			break
		}
	}
	fmt.Printf("Logging to %s\n", name)
	logfi,err := os.Create(name)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logfi)
	return logfi
}

var pollfreq = flag.Int("poll", 30, "Time in seconds to wait between polling")
var rebtime = flag.Int("rebt", 5, "Time to sleep after detecting a failure before rebooting")

func main() {
	defer SetLogger().Close()
	for {
		st := GetGPUStatus()
		if st == nil || len(st) == 0 {
			log.Println("Failed to retrieve GPU info... trying again in 30 seconds...")
			time.Sleep(time.Second * 30)
			st = GetGPUStatus()
			if st == nil || len(st) == 0 {
				log.Println("Second attempt to get info failed... rebooting!")
				time.Sleep(time.Second * 5)
				Reboot()
			}
			log.Println("Got info back! Success!")
		}
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
