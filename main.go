/*
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2025-02-16 22:38:51
 * @LastEditTime: 2025-02-17 21:57:31
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /wol-http/main.go
 */
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"slices"

	"gopkg.in/ini.v1"
)

type Entry struct {
	Name    string
	Alias   string
	MacAddr string
	IpAddr  string
	Note    string
}

type Config struct {
	Addr    string
	Cert    string
	Key     string
	Token   string
	WolBin  string
	Entries []Entry
}

type FastFetch struct {
	ByName    map[string]*Entry
	ByAlias   map[string]*Entry
	ByMacAddr map[string]*Entry
	ByIpAddr  map[string]*Entry
}

const (
	VER             string = "0.1.0"
	CODENAME        string = "SatenRuiko"
	DEFAULT_WOL_BIN string = "/usr/bin/wakeonlan"
)

var (
	Conf Config    // Do not modify after init.
	FF   FastFetch // Do not modify after init.
)

func ErrHandler(err error) {
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
}

func FileExists(path string) bool {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) || stat.IsDir() {
		return false
	}
	return true
}

func getArg() string {
	if len(os.Args) <= 1 || len(os.Args) > 2 {
		fmt.Println("Usage: wol-http <ConfFile>")
		os.Exit(1)
	}
	return os.Args[1]
}

func hello() {
	fmt.Printf("WOL-HTTP Server [ Version: %s (%s) ]\n", VER, CODENAME)
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token != Conf.Token {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("403 wrong token provided"))
		return
	}

	var target *Entry = nil
	key := r.PathValue("key")
	switch r.PathValue("by") {
	case "name":
		target = FF.ByName[key]
	case "alias":
		target = FF.ByAlias[key]
	case "ip":
		target = FF.ByIpAddr[key]
	case "mac":
		target = FF.ByMacAddr[key]
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 bad request"))
		return
	}

	if target == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 entry not found"))
		return
	}

	switch r.PathValue("action") {
	case "info":
		w.WriteHeader(http.StatusOK)

		// Construct info string //
		infoStr := fmt.Sprintf("Name: %s\nAlias: %s\nIP: %s\nMAC: %s\nNote: %s\n",
			target.Name, target.Alias, target.IpAddr, target.MacAddr, target.Note)

		w.Write([]byte(infoStr))
	case "wake":
		cmd := exec.Command(Conf.WolBin, target.MacAddr)
		output, err := cmd.CombinedOutput()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(slices.Concat([]byte("500 internal server error\n"), output))
			return
		}
		w.Write(output)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 bad request"))
		return
	}
}

func main() {
	hello()
	conf_file, err := ini.Load(getArg())
	ErrHandler(err)

	if !conf_file.HasSection("GENERAL") {
		fmt.Println("Error: General config section not found")
		os.Exit(1)
	}

	// Get General Config //
	sec_general := conf_file.Section("GENERAL")
	Conf.Addr = sec_general.Key("Addr").String()
	if len(Conf.Addr) == 0 {
		ErrHandler(fmt.Errorf("no correct listening address specified"))
	}
	Conf.Cert = sec_general.Key("Cert").String()
	Conf.Key = sec_general.Key("Key").String()
	Conf.Token = sec_general.Key("Token").String()
	if len(Conf.Token) < 8 {
		ErrHandler(fmt.Errorf("no correct access token specified or token is too short"))
	}
	Conf.WolBin = sec_general.Key("WolBin").String()
	if len(Conf.WolBin) == 0 {
		Conf.WolBin = DEFAULT_WOL_BIN
	}

	if !FileExists(Conf.WolBin) {
		ErrHandler(fmt.Errorf("wakeonlan binary not exists"))
	}

	// Get Entries //
	Conf.Entries = make([]Entry, 0)
	FF.ByName = make(map[string]*Entry)
	FF.ByAlias = make(map[string]*Entry)
	FF.ByIpAddr = make(map[string]*Entry)
	FF.ByMacAddr = make(map[string]*Entry)
	for _, sec := range conf_file.Sections() {
		if sec.Name() == "DEFAULT" || sec.Name() == "GENERAL" {
			continue
		}
		newEntry := Entry{}
		newEntry.Name = sec.Name()
		newEntry.Alias = sec.Key("Alias").String()
		newEntry.IpAddr = sec.Key("IP").String()
		newEntry.MacAddr = sec.Key("MAC").String()
		newEntry.Note = sec.Key("Note").String()
		Conf.Entries = append(Conf.Entries, newEntry)
		last := &Conf.Entries[len(Conf.Entries)-1]
		FF.ByName[newEntry.Name] = last
		FF.ByAlias[newEntry.Alias] = last
		FF.ByIpAddr[newEntry.IpAddr] = last
		FF.ByMacAddr[newEntry.MacAddr] = last
	}

	// HTTP(S) Service Part //
	http.HandleFunc("GET /{token}/{action}/{by}/{key}", httpHandler)
	if len(Conf.Cert) != 0 && len(Conf.Key) != 0 {
		err = http.ListenAndServeTLS(Conf.Addr, Conf.Cert, Conf.Key, nil)
	} else {
		err = http.ListenAndServe(Conf.Addr, nil)
	}
	ErrHandler(err)
}
