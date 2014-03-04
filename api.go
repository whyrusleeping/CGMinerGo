package main

import (
	"net"
	"encoding/json"
	"io/ioutil"
	"fmt"
)

type MSS map[string]string
type MI map[string]interface{}

func GetGPUStatus() []string {
	var out []string
	devinf := MakeReq(MSS{"command":"devs"})
	devarr := devinf["DEVS"].([]interface{})
	for _,gpui := range devarr {
		gpu := gpui.(map[string]interface{})
		status := gpu["Status"].(string)
		out = append(out, status)
	}
	return out
}

func GetCurrentHashRate() (float64, float64) {
	resp := MakeReq(MSS{"command":"summary"})
	summarr := resp["SUMMARY"].([]interface{})
	summ := summarr[0].(map[string]interface{})
	avg := summ["MHS av"].(float64)
	lfs := summ["MHS 5s"].(float64)

	return avg,lfs
}

func MakeReq(req MSS) MI {
	con,err := net.Dial("tcp", "localhost:4028")
	if err != nil {
		panic(err)
	}
	b,_ := json.Marshal(req)
	con.Write(b)
	out,_ := ioutil.ReadAll(con)
	out = out[:len(out)-1]
	omap := MI{}
	err = json.Unmarshal(out, &omap)
	if err != nil {
		panic(err)
	}
	return omap
}

func main() {
	fmt.Println(GetGPUStatus())
	fmt.Println(GetCurrentHashRate())
}
