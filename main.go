package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
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
	VersionDoQ08 = "doq-i07"
	VersionDoQ09 = "doq-i07"
	VersionDoQ10 = "doq-i07"
	VersionDoQ11 = "doq-i07"
	VersionDoQ12 = "doq-i07"
	VersionDoQ   = "doq" //RFC9250

)

var port = flag.String("port", "853", "the port number to verify on")
var useSNI = flag.Bool("sni", false, "Use SNI for queries. Expects input format as 'ip, sni' if true.")

var DefaultDoQVersions = []string{VersionDoQ, VersionDoQ12, VersionDoQ11, VersionDoQ10, VersionDoQ09, VersionDoQ08, VersionDoQ07, VersionDoQ06, VersionDoQ05, VersionDoQ04, VersionDoQ03, VersionDoQ02, VersionDoQ01, VersionDoQ00}

var DefaultQUICVersions = []quic.VersionNumber{
	quic.Version1,
	quic.Version2,
}

func establishConnection(ip net.IP, domain string, csvWriter *csv.Writer) bool {
	log.Printf("Starting connection attempt for IP: %s with domain: %s", ip, domain)
	defer log.Printf("Ending connection attempt for IP: %s", ip)
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         DefaultDoQVersions,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			for _, rawCert := range rawCerts {
				cert, err := x509.ParseCertificate(rawCert)
				if err != nil {
					log.Printf("Failed to parse certificate: %v", err)
					continue
				}
				pemBlock := &pem.Block{
					Type:  "CERTIFICATE",
					Bytes: rawCert,
				}
				pemBytes := pem.EncodeToMemory(pemBlock)
				if pemBytes == nil {
					log.Println("Failed to encode certificate to PEM")
					continue
				}
				caName := cert.Issuer.CommonName
				mutex.Lock()
				if err := csvWriter.Write([]string{ip.String(), string(pemBytes), caName}); err != nil {
					log.Printf("Failed to write to CSV: %v", err)
				}
				csvWriter.Flush()
				mutex.Unlock()
			}
			return nil
		},
	}

	if *useSNI && domain != "" {
		tlsConf.ServerName = domain
	}

	quicConf := &quic.Config{
		HandshakeIdleTimeout: time.Second * 2,
		Versions:             DefaultQUICVersions,
	}
	var ports = []string{*port}
	reachable := make(chan bool)
	go func() {
		for _, port := range ports {
			session, err := quic.DialAddr(context.Background(), ip.String()+":"+port, tlsConf, quicConf)
			if err != nil {
				continue
			}
			reachable <- true
			session.CloseWithError(0, "")
			return
		}
		reachable <- false
	}()
	return <-reachable
}

func main() {
	//设置平行limit
	parallelLimit := flag.Int("parallel", 30, "sets the limit for parallel processes")
	//解析参数
	flag.Parse()

	args := flag.Args()
	if len(args) != 3 {
		println("need 3 arguments: [in file] [out file] [cert file]")
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
	//create a csv writer
	certWriter := csv.NewWriter(certFile)

	var sem = semaphore.NewWeighted(int64(*parallelLimit))

	var wg sync.WaitGroup

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		var ip net.IP
		var domain string

		if *useSNI {
			parts := strings.Split(line, ",")
			if len(parts) != 2 {
				log.Println("Skipping invalid line:", line)
				continue
			}

			ip = net.ParseIP(parts[0])
			domain = parts[1]
		} else {
			ip = net.ParseIP(line)
			domain = ""
		}

		if ip == nil {
			log.Println("Invalid IP address:", line)
			continue
		}
		sem.Acquire(context.Background(), 1)
		wg.Add(1)
		go func(ip net.IP, domain string) {
			reachable := establishConnection(ip, domain, certWriter)
			if reachable {
				if _, err := outFile.WriteString(ip.String() + " " + domain + "\n"); err != nil {
					log.Println(err)
				}
			}
			sem.Release(1)
			wg.Done()
		}(ip, domain)
	}

	wg.Wait()
}
