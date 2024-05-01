package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"time"
)

const TABLE_SMTPD_VERSION = "7.4.0"
const TABLE_PROTOCOL_VERSION = "0.1"

func main() {
	var opt_table string
	var opt_service string
	var opt_fetch bool
	var opt_check string
	var opt_lookup string

	flag.StringVar(&opt_table, "table", "", "table service name")
	flag.StringVar(&opt_service, "service", "", "lookup service name")
	flag.BoolVar(&opt_fetch, "fetch", false, "fetch from service")
	flag.StringVar(&opt_check, "check", "", "check key from service")
	flag.StringVar(&opt_lookup, "lookup", "", "fetch key from service")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Please provide a table backend path")
		return
	}

	if opt_table == "" {
		fmt.Println("Please provide a table name")
		return
	}

	if opt_service == "" {
		fmt.Println("Please provide a service name")
		return
	}

	c := 0
	if opt_fetch {
		c++
	}
	if opt_check != "" {
		c++
	}
	if opt_lookup != "" {
		c++
	}
	if c != 1 {
		fmt.Println("Please provide one of -fetch, -check or -lookup")
		return
	}

	registeredServices := make(map[string]struct{})

	args := flag.Args()

	cmd := exec.Command(args[0], args[1:]...)
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(in, "config|smtpd-version|%s\n", TABLE_SMTPD_VERSION)
	fmt.Fprintf(in, "config|protocol|%s\n", TABLE_PROTOCOL_VERSION)
	fmt.Fprintf(in, "config|ready\n")

	scanner := bufio.NewScanner(out)
	for {
		if !scanner.Scan() {
			log.Fatal("scanner.Scan() failed")
			break
		}
		line := scanner.Text()
		if line == "register|ready" {
			break
		}
		serviceName := line[9:]
		registeredServices[serviceName] = struct{}{}
	}

	if _, ok := registeredServices[opt_service]; !ok {
		log.Fatalf("service %s not registered", opt_service)
		return
	}

	if opt_fetch {
		fmt.Fprintf(in, "table|%s|%d|%s|fetch|%s|%s\n", TABLE_PROTOCOL_VERSION, time.Now().Unix(), opt_table, opt_service, "deadbeefabadf00d")

		if !scanner.Scan() {
			log.Fatal("scanner.Scan() failed")
		}
		line := scanner.Text()
		fmt.Println(line)
	}

	if opt_lookup != "" {
		fmt.Fprintf(in, "table|%s|%d|%s|lookup|%s|%s|%s\n", TABLE_PROTOCOL_VERSION, time.Now().Unix(), opt_table, opt_service, "deadbeefabadf00d", opt_lookup)

		if !scanner.Scan() {
			log.Fatal("scanner.Scan() failed")
		}
		line := scanner.Text()
		fmt.Println(line)
	}

	if opt_check != "" {
		fmt.Fprintf(in, "table|%s|%d|%s|check|%s|%s|%s\n", TABLE_PROTOCOL_VERSION, time.Now().Unix(), opt_table, opt_service, "deadbeefabadf00d", opt_lookup)

		if !scanner.Scan() {
			log.Fatal("scanner.Scan() failed")
		}
		line := scanner.Text()
		fmt.Println(line)
	}

	in.Close()
	out.Close()
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
