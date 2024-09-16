package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/csv"
	"encoding/pem"
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	quic "github.com/quic-go/quic-go"
	"golang.org/x/sync/semaphore"
)

var mutex sync.Mutex

const (
	VersionDoQ00 = "doq-i00"
	VersionDoQ01 = "doq-i01"
	VersionDoQ02 = "doq-i02"
	VersionDoQ03 = "doq-i03"
	VersionDoQ04 = "doq-i04"
	VersionDoQ05 = "doq-i05"
	VersionDoQ06 = "doq-i06"
	VersionDoQ07 = "doq-i07"
	VersionDoQ08 = "doq-i08"
	VersionDoQ09 = "doq-i09"
	VersionDoQ10 = "doq-i10"
	VersionDoQ11 = "doq-i11"
	VersionDoQ12 = "doq-i12"
	VersionDoQ   = "doq" //RFC9250
)

var port = flag.String("port", "853", "the port number to verify on")
var useSNI = flag.Bool("sni", false, "Use SNI for queries. Expects input format as 'ip, sni' if true.")

var DefaultDoQVersions = []string{VersionDoQ, VersionDoQ12, VersionDoQ11, VersionDoQ10, VersionDoQ09, VersionDoQ08, VersionDoQ07, VersionDoQ06, VersionDoQ05, VersionDoQ04, VersionDoQ03, VersionDoQ02, VersionDoQ01, VersionDoQ00}

var DefaultQUICVersions = []quic.VersionNumber{
	quic.Version1,
	quic.Version2,
}

func establishConnection(ip net.IP, port, domain string, certWriter, errorWriter *csv.Writer) bool {
	
	log.Printf("Starting connection attempt for IP: %s with domain: %s", ip, domain)
	defer log.Printf("Ending connection attempt for IP: %s", ip)
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         DefaultDoQVersions,
	}

	if *useSNI && domain != "" {
		tlsConf.ServerName = domain
	}

	quicConf := &quic.Config{
		HandshakeIdleTimeout: time.Second * 20,
		Versions:             DefaultQUICVersions,
	}

	//var ports = []string{*port}
	reachable := make(chan bool)
	go func() {
		//for _, port := range ports {
		log.Println("port:%s", port)
		ipStr := ip.String()+":"+port

		if ip.To4() == nil { // This is an IPv6 address
			ipStr = "[" + ip.String() + "]:"+port
		
		}
		session, err := quic.DialAddr(context.Background(), ipStr , tlsConf, quicConf)

		if err != nil {
			log.Printf("Failed to establish QUIC connection to %s:%s: %v", ip, port, err)
			mutex.Lock()
			if err := errorWriter.Write([]string{ip.String(), port, domain, "", "", err.Error()}); err != nil {
				log.Printf("Failed to write error to CSV: %v", err)
			}
			errorWriter.Flush()
			mutex.Unlock()
			//continue
			//这里将continue修改了，因为并不会尝试多个端口
			reachable <- false
			return
		}

		// Access TLS connection state
		tlsState := session.ConnectionState().TLS
		doqVersion := "unknown"
		tlsVersion := "unknown"

		if len(tlsState.NegotiatedProtocol) > 0 {
			doqVersion = tlsState.NegotiatedProtocol
		}

		switch tlsState.Version {
		case tls.VersionTLS13:
			tlsVersion = "TLS 1.3"
		case tls.VersionTLS12:
			tlsVersion = "TLS 1.2"
		case tls.VersionTLS11:
			tlsVersion = "TLS 1.1"
		case tls.VersionTLS10:
			tlsVersion = "TLS 1.0"
		default:
			tlsVersion = "unknown"
		}

		if len(tlsState.PeerCertificates) > 0 {
			for _, cert := range tlsState.PeerCertificates {
				pemBlock := &pem.Block{
					Type:  "CERTIFICATE",
					Bytes: cert.Raw,
				}
				pemBytes := pem.EncodeToMemory(pemBlock)
				if pemBytes == nil {
					log.Println("Failed to encode certificate to PEM")
					continue
				}
				mutex.Lock()
				if err := certWriter.Write([]string{ip.String(), port, domain, doqVersion, tlsVersion, string(pemBytes)}); err != nil {
					log.Printf("Failed to write to CSV: %v", err)
				}
				certWriter.Flush()
				mutex.Unlock()
			}
		} else {
			log.Printf("No peer certificates found for IP: %s", ip)
		}

		reachable <- true
		session.CloseWithError(0, "")
		//return
		//}
		//reachable <- false
	}()
	return <-reachable
}

func main() {
	parallelLimit := flag.Int("parallel", 30, "sets the limit for parallel processes")
	flag.Parse()

	args := flag.Args()
	if len(args) != 4 {
		println("need 4 arguments: [in file] [out file] [cert file] [error file]")
		println("now only get ", len(args))
		os.Exit(1)
	}

	inFile, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	outFile, err := os.OpenFile(args[1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	certFile, err := os.OpenFile(args[2], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer certFile.Close()

	errorFile, err := os.OpenFile(args[3], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer errorFile.Close()

	//create a csv writer
	certWriter := csv.NewWriter(certFile)
	errorWriter := csv.NewWriter(errorFile)

	// Write CSV headers
	if err := certWriter.Write([]string{"IP", "Domain", "DoQ Version", "TLS Version", "Certificate"}); err != nil {
		log.Fatalf("Failed to write cert CSV header: %v", err)
	}
	if err := errorWriter.Write([]string{"IP", "Domain", "ErrorMessage"}); err != nil {
		log.Fatalf("Failed to write error CSV header: %v", err)
	}
	certWriter.Flush()
	errorWriter.Flush()

	var sem = semaphore.NewWeighted(int64(*parallelLimit))

	var wg sync.WaitGroup

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		var ip net.IP
		var port, domain string
		
		parts := strings.Split(line, ",")
		ip = net.ParseIP(parts[0])
		port = parts[1]

		if *useSNI {
			parts := strings.Split(line, ",")
			if len(parts) != 3 {
				log.Println("Skipping invalid line:", line)
				continue
			}
			domain = parts[2]
		} else {
			parts := strings.Split(line, ",")
			if len(parts) != 2 {
				log.Println("Skipping invalid line:", line)
				continue
			}
			domain = ""
		}


		if ip == nil {
			log.Println("Invalid IP address:", line)
			continue
		}
		sem.Acquire(context.Background(), 1)
		wg.Add(1)
		go func(ip net.IP, port, domain string) {
			reachable := establishConnection(ip, port, domain, certWriter, errorWriter)
			if reachable {
				if _, err := outFile.WriteString(ip.String() + "," + port + "," + domain + "\n"); err != nil {
					log.Println(err)
				}
			}
			sem.Release(1)
			wg.Done()
		}(ip, port, domain)
	}

	wg.Wait()
}
