package table

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func timestampToTime(timestamp float64) time.Time {
	sec := int64(timestamp)
	nsec := int64((timestamp - float64(sec)) * 1e9)
	return time.Unix(sec, nsec)
}

type Service int

const (
	K_ERROR       Service = -1
	K_ALIAS       Service = 0
	K_DOMAIN      Service = iota
	K_CREDENTIALS Service = iota
	K_NETADDR     Service = iota
	K_USERINFO    Service = iota
	K_SOURCE      Service = iota
	K_MAILADDR    Service = iota
	K_ADDRNAME    Service = iota
	K_MAILADDRMAP Service = iota
	K_AUTH        Service = iota
)

func (s Service) String() string {
	switch s {
	case K_ALIAS:
		return "alias"
	case K_DOMAIN:
		return "domain"
	case K_CREDENTIALS:
		return "credentials"
	case K_NETADDR:
		return "netaddr"
	case K_USERINFO:
		return "userinfo"
	case K_SOURCE:
		return "source"
	case K_MAILADDR:
		return "mailaddr"
	case K_ADDRNAME:
		return "addrname"
	case K_MAILADDRMAP:
		return "mailaddrmap"
	case K_AUTH:
		return "auth"
	default:
		log.Fatalf("Unknown service %d", s)
	}
	return ""
}

func serviceFromName(name string) Service {
	switch name {
	case "alias":
		return K_ALIAS
	case "domain":
		return K_DOMAIN
	case "credentials":
		return K_CREDENTIALS
	case "netaddr":
		return K_NETADDR
	case "userinfo":
		return K_USERINFO
	case "source":
		return K_SOURCE
	case "mailaddr":
		return K_MAILADDR
	case "addrname":
		return K_ADDRNAME
	case "mailaddrmap":
		return K_MAILADDRMAP
	case "auth":
		return K_AUTH
	default:
		log.Fatalf("Unknown service %s", name)
	}
	return K_ERROR
}

type onUpdateCb func(time.Time, string) error
type onCheckCb func(time.Time, string, string) (bool, error)
type onLookupCb func(time.Time, string, string) (string, error)
type onFetchCb func(time.Time, string) (string, error)

var onUpdate onUpdateCb
var onCheckMap map[Service]onCheckCb = make(map[Service]onCheckCb)
var onLookupMap map[Service]onLookupCb = make(map[Service]onLookupCb)
var onFetchMap map[Service]onFetchCb = make(map[Service]onFetchCb)

func Init() {
}

func OnUpdate(cb onUpdateCb) {
	onUpdate = cb
}

func OnCheck(service Service, cb onCheckCb) {
	if _, registered := onCheckMap[service]; registered {
		log.Fatalf("OnCheck already registered for service %s", service)
	} else {
		onCheckMap[service] = cb
	}
}

func OnLookup(service Service, cb onLookupCb) {
	if _, registered := onLookupMap[service]; registered {
		log.Fatalf("OnLookup already registered for service %s", service)
	} else {
		onLookupMap[service] = cb
	}
}

func OnFetch(service Service, cb onFetchCb) {
	if _, registered := onFetchMap[service]; registered {
		log.Fatalf("OnFetch already registered for service %s", service)
	} else {
		onFetchMap[service] = cb
	}
}

func Dispatch() {
	scanner := bufio.NewScanner(os.Stdin)

	protocolVersion := "0.1"

	// server configuration
	for {
		if !scanner.Scan() {
			log.Fatal("scanner.Scan() failed")
			break
		}
		line := scanner.Text()
		if line == "config|ready" {
			break
		}
	}

	// table registration
	services := make(map[string]struct{})
	for s := range onCheckMap {
		services[s.String()] = struct{}{}
	}
	for s := range onLookupMap {
		services[s.String()] = struct{}{}
	}
	for s := range onFetchMap {
		services[s.String()] = struct{}{}
	}
	for s := range services {
		fmt.Fprintf(os.Stdout, "register|%s\n", s)
	}
	fmt.Println("register|ready")

	for {
		if !scanner.Scan() {
			log.Fatal("scanner.Scan() failed")
			break
		}
		line := scanner.Text()

		atoms := strings.Split(line, "|")
		if len(atoms) < 5 {
			log.Fatalf("not enouh fields to be a valid line: %s", line)
		}

		if atoms[0] != "table" {
			log.Fatalf("Invalid command %s", atoms[0])
		}

		if atoms[1] != protocolVersion {
			log.Fatalf("Invalid protocol version %s", atoms[1])
		}

		timestamp, err := strconv.ParseFloat(atoms[2], 64)
		if err != nil {
			log.Fatalf("Failed to convert timestamp %s to float", atoms[2])
		}

		tablename := atoms[3]
		if tablename == "" {
			log.Fatalf("Empty tablename")
		}

		operation := atoms[4]
		atoms = atoms[5:]

		switch operation {
		case "update":
			if onUpdate != nil {
				if len(atoms) != 1 {
					log.Fatalf("Invalid number of arguments for update")
				}
				opaque := atoms[0]
				go func() {
					if err := onUpdate(timestampToTime(timestamp), tablename); err != nil {
						fmt.Fprintf(os.Stdout, "update-result|%s|ko\n", opaque)
					} else {
						fmt.Fprintf(os.Stdout, "update-result|%s|ok\n", opaque)
					}
				}()
			}

		case "check":
			if len(atoms) != 3 {
				log.Fatalf("Invalid number of arguments for check")
			}
			service := serviceFromName(atoms[0])
			opaque := atoms[1]
			key := atoms[2]

			if cb, ok := onCheckMap[service]; !ok {
				fmt.Fprintf(os.Stdout, "fetch-result|%s|error|no handler registered\n", opaque)
			} else {
				go func() {
					exists, err := cb(timestampToTime(timestamp), tablename, key)
					if err != nil {
						fmt.Fprintf(os.Stdout, "check-result|%s|%s|%s\n", opaque, "error", err)
					} else if !exists {
						fmt.Fprintf(os.Stdout, "check-result|%s|not-found\n", opaque)
					} else {
						fmt.Fprintf(os.Stdout, "check-result|%s|found\n", opaque)
					}
				}()
			}

		case "fetch":
			if len(atoms) != 2 {
				log.Fatalf("Invalid number of arguments for fetch")
			}
			service := serviceFromName(atoms[0])
			opaque := atoms[1]

			if cb, ok := onFetchMap[service]; !ok {
				fmt.Fprintf(os.Stdout, "fetch-result|%s|error|no handler registered\n", opaque)
			} else {
				go func() {
					result, err := cb(timestampToTime(timestamp), tablename)
					if err != nil {
						fmt.Fprintf(os.Stdout, "lookup-result|%s|%s|%s\n", opaque, "error", err)
					} else if result == "" {
						fmt.Fprintf(os.Stdout, "lookup-result|%s|not-found\n", opaque)
					} else {
						fmt.Fprintf(os.Stdout, "lookup-result|%s|found|%s\n", opaque, result)
					}
				}()
			}

		case "lookup":
			if len(atoms) != 3 {
				log.Fatalf("Invalid number of arguments for lookup")
			}
			service := serviceFromName(atoms[0])
			opaque := atoms[1]
			key := atoms[2]

			if cb, ok := onLookupMap[service]; !ok {
				fmt.Fprintf(os.Stdout, "fetch-result|%s|error|no handler registered\n", opaque)
			} else {
				go func() {
					result, err := cb(timestampToTime(timestamp), tablename, key)
					if err != nil {
						fmt.Fprintf(os.Stdout, "lookup-result|%s|%s|%s\n", opaque, "error", err)
					} else if result == "" {
						fmt.Fprintf(os.Stdout, "lookup-result|%s|not-found\n", opaque)
					} else {
						fmt.Fprintf(os.Stdout, "lookup-result|%s|found|%s\n", opaque, result)
					}
				}()
			}

		}

	}
}
