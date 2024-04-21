package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
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

	_, stdio_pipe := fork_child(flag.Args())
	fp := os.NewFile(uintptr(stdio_pipe), "stdio_pipe")

	fmt.Fprintf(fp, "config|smtpd-version|%s\n", TABLE_SMTPD_VERSION)
	fmt.Fprintf(fp, "config|protocol|%s\n", TABLE_PROTOCOL_VERSION)
	fmt.Fprintf(fp, "config|ready\n")
	fp.Sync()

	scanner := bufio.NewScanner(fp)
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
		fmt.Fprintf(fp, "table|%s|%d|%s|fetch|%s|%s\n", TABLE_PROTOCOL_VERSION, time.Now().Unix(), opt_table, opt_service, "deadbeefabadf00d")
		fp.Sync()

		if !scanner.Scan() {
			log.Fatal("scanner.Scan() failed")
		}
		line := scanner.Text()
		fmt.Println(line)
	}

	if opt_lookup != "" {
		fmt.Fprintf(fp, "table|%s|%d|%s|lookup|%s|%s|%s\n", TABLE_PROTOCOL_VERSION, time.Now().Unix(), opt_table, opt_service, "deadbeefabadf00d", opt_lookup)
		fp.Sync()

		if !scanner.Scan() {
			log.Fatal("scanner.Scan() failed")
		}
		line := scanner.Text()
		fmt.Println(line)
	}

	if opt_check != "" {
		fmt.Fprintf(fp, "table|%s|%d|%s|check|%s|%s|%s\n", TABLE_PROTOCOL_VERSION, time.Now().Unix(), opt_table, opt_service, "deadbeefabadf00d", opt_lookup)
		fp.Sync()

		if !scanner.Scan() {
			log.Fatal("scanner.Scan() failed")
		}
		line := scanner.Text()
		fmt.Println(line)
	}

}

func fork_child(args []string) (int, int) {
	sp, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM, syscall.AF_UNSPEC)
	if err != nil {
		log.Fatal(err)
	}

	// XXX - not quite there yet
	//syscall.SetNonblock(sp[0], true)
	//syscall.SetNonblock(sp[1], true)

	procAttr := syscall.ProcAttr{}
	procAttr.Files = []uintptr{
		uintptr(sp[0]),
		uintptr(sp[0]),
		uintptr(syscall.Stderr),
	}

	var pid int

	pid, err = syscall.ForkExec(args[0], args, &procAttr)
	if err != nil {
		log.Fatal(err)
	}

	if syscall.Close(sp[0]) != nil {
		log.Fatal(err)
	}

	return pid, sp[1]
}
