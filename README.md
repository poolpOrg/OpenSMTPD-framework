# OpenSMTPD-framework

THIS IS A WIP, DO NOT USE UNLESS YOU KNOW WHAT YOU'RE DOING.


## Purpose

The OpenSMTPD-framework intends to be an unofficial implementation of a package to ease the writing of OpenSMTPD extensions,
and their debugging.


## Programming interfaces

### table
This is the entire set of table requests available at the moment,
it is not required to register callbacks for events that are not handled:

```go
package main

import (
	"github.com/poolpOrg/OpenSMTPD-framework/table"
)

func main() {
	table.Init()

	// called whenever `smtpctl update table` is called
	table.OnUpdate(func(timestamp time.Time, table string) error {
		return nil
	})

	// check that a value exists for a given key
	table.OnCheck(table.K_ALIAS, func(timestamp time.Time, table string, key string) (bool, error) {
		return true, nil
	})

	// returns the value associated to a given key or "" if it does not exist
	table.OnLookup(table.K_ALIAS, func(timestamp time.Time, table string, key string) (string, error) {
		return "", nil
	})

	// fetch the next key or "" if it does not exist
	table.OnFetch(table.K_ALIAS, func(timestamp time.Time, table string) (string, error) {
		return "", nil
	})

	table.Dispatch()
}
```

The following lookup services are exposed,
see the [table(5)](https://man.openbsd.org/table) man page for description:
```go
table.K_ALIAS
table.K_DOMAIN
table.K_CREDENTIALS
table.K_NETADDR
table.K_USERINFO
table.K_SOURCE
table.K_MAILADDR
table.K_ADDRNAME
table.K_MAILADDRMAP
```

### filter
This is the entire set of report events and filter requests available at the moment,
it is not required to register callbacks for events that are not handled:

```go
package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/poolpOrg/OpenSMTPD-framework/filter"
)

type SessionData struct {
}

func linkConnectCb(timestamp time.Time, session filter.Session, rdns string, fcrdns string, src net.Addr, dest net.Addr) {
	// if SessionAllocator is set below, this obtains a pointer to SessionData
	// which is concurrency-safe and usable to retain information between callbacks.
	// otherwise returns nil
	_ = session.Get()
	fmt.Fprintf(os.Stderr, "%s: %s: link-connect: %s|%s|%s|%s\n", timestamp, session, rdns, fcrdns, src, dest)
}

func linkDisconnectCb(timestamp time.Time, session filter.Session) {
	fmt.Fprintf(os.Stderr, "%s: %s: link-disconnect\n", timestamp, session)
}

func linkGreetingCb(timestamp time.Time, session filter.Session, hostname string) {
	fmt.Fprintf(os.Stderr, "%s: %s: link-greeting: %s\n", timestamp, session, hostname)
}

func linkIdentifyCb(timestamp time.Time, session filter.Session, method string, hostname string) {
	fmt.Fprintf(os.Stderr, "%s: %s: link-identify: %s|%s\n", timestamp, session, method, hostname)
}

func linkAuthCb(timestamp time.Time, session filter.Session, result string, username string) {
	fmt.Fprintf(os.Stderr, "%s: %s: link-auth: %s|%s\n", timestamp, session, result, username)
}

func linkTLSCb(timestamp time.Time, session filter.Session, tlsString string) {
	fmt.Fprintf(os.Stderr, "%s: %s: link-tls: %s\n", timestamp, session, tlsString)
}

func txResetCb(timestamp time.Time, session filter.Session, messageId string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-reset: %s\n", timestamp, session, messageId)
}

func txBeginCb(timestamp time.Time, session filter.Session, messageId string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-begin: %s\n", timestamp, session, messageId)
}

func txMailCb(timestamp time.Time, session filter.Session, messageId string, result string, from string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-mail: %s|%s|%s\n", timestamp, session, messageId, result, from)
}

func txRcptCb(timestamp time.Time, session filter.Session, messageId string, result string, to string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-rcpt: %s|%s|%s\n", timestamp, session, messageId, result, to)
}

func txEnvelopeCb(timestamp time.Time, session filter.Session, messageId string, envelopeId string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-envelope: %s|%s\n", timestamp, session, messageId, envelopeId)
}

func txDataCb(timestamp time.Time, session filter.Session, messageId string, result string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-data: %s|%s\n", timestamp, session, messageId, result)
}

func txCommmitCb(timestamp time.Time, session filter.Session, messageId string, messageSize int) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-commit: %s|%d\n", timestamp, session, messageId, messageSize)
}

func txRollbackCb(timestamp time.Time, session filter.Session, messageId string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-rollback: %s\n", timestamp, session, messageId)
}

func protocolClientCb(timestamp time.Time, session filter.Session, command string) {
	fmt.Fprintf(os.Stderr, "%s: %s: protocol-client: %s\n", timestamp, session, command)
}

func protocolServerCb(timestamp time.Time, session filter.Session, response string) {
	fmt.Fprintf(os.Stderr, "%s: %s: protocol-server: %s\n", timestamp, session, response)
}

func filterReportCb(timestamp time.Time, session filter.Session, filterKind string, name string, message string) {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-report: %s|%s|%s\n", timestamp, session, filterKind, name, message)
}

func filterResponseCb(timestamp time.Time, session filter.Session, phase string, response string, param ...string) {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-response: %s|%s|%v\n", timestamp, session, phase, response, param)
}

func timeoutCb(timestamp time.Time, session filter.Session) {
	fmt.Fprintf(os.Stderr, "%s: %s: timeout\n", timestamp, session)
}

func filterConnectCb(timestamp time.Time, session filter.Session, rdns string, src net.Addr) filter.Response {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-connect: %s|%s\n", timestamp, session, rdns, src)
	return filter.Proceed()
}

func filterHeloCb(timestamp time.Time, session filter.Session, helo string) filter.Response {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-helo: %s\n", timestamp, session, helo)
	return filter.Proceed()
}

func filterEhloCb(timestamp time.Time, session filter.Session, helo string) filter.Response {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-ehlo: %s\n", timestamp, session, helo)
	return filter.Proceed()
}

func filterStartTLSCb(timestamp time.Time, session filter.Session, tls string) filter.Response {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-starttls: %s\n", timestamp, session, tls)
	return filter.Proceed()
}

func filterAuthCb(timestamp time.Time, session filter.Session, mechanism string) filter.Response {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-auth: %s\n", timestamp, session, mechanism)
	return filter.Proceed()
}

func filterMailFromCb(timestamp time.Time, session filter.Session, from string) filter.Response {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-mail-from: %s\n", timestamp, session, from)
	return filter.Proceed()
}

func filterRcptToCb(timestamp time.Time, session filter.Session, to string) filter.Response {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-rcpt-to: %s\n", timestamp, session, to)
	return filter.Proceed()
}

func filterDataCb(timestamp time.Time, session filter.Session) filter.Response {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-data\n", timestamp, session)
	return filter.Proceed()
}

func filterCommitCb(timestamp time.Time, session filter.Session) filter.Response {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-commit\n", timestamp, session)
	return filter.Proceed()
}

func filterDataLineCb(timestamp time.Time, session filter.Session, line string) []string {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-data-line: %s\n", timestamp, session, line)
	return []string{line}
}

func main() {
	filter.Init()

	filter.SMTP_IN.SessionAllocator(func() filter.SessionData {
		return &SessionData{}
	})

	filter.SMTP_IN.OnLinkConnect(linkConnectCb)
	filter.SMTP_IN.OnLinkDisconnect(linkDisconnectCb)
	filter.SMTP_IN.OnLinkGreeting(linkGreetingCb)
	filter.SMTP_IN.OnLinkIdentify(linkIdentifyCb)
	filter.SMTP_IN.OnLinkAuth(linkAuthCb)
	filter.SMTP_IN.OnLinkTLS(linkTLSCb)

	filter.SMTP_IN.OnTxReset(txResetCb)
	filter.SMTP_IN.OnTxBegin(txBeginCb)
	filter.SMTP_IN.OnTxMail(txMailCb)
	filter.SMTP_IN.OnTxRcpt(txRcptCb)
	filter.SMTP_IN.OnTxEnvelope(txEnvelopeCb)
	filter.SMTP_IN.OnTxData(txDataCb)
	filter.SMTP_IN.OnTxCommit(txCommmitCb)
	filter.SMTP_IN.OnTxRollback(txRollbackCb)

	filter.SMTP_IN.OnProtocolClient(protocolClientCb)
	filter.SMTP_IN.OnProtocolServer(protocolServerCb)

	filter.SMTP_IN.OnFilterReport(filterReportCb)
	filter.SMTP_IN.OnFilterResponse(filterResponseCb)
	filter.SMTP_IN.OnTimeout(timeoutCb)

	filter.SMTP_IN.ConnectRequest(filterConnectCb)
	filter.SMTP_IN.HeloRequest(filterHeloCb)
	filter.SMTP_IN.EhloRequest(filterEhloCb)
	filter.SMTP_IN.StartTLSRequest(filterStartTLSCb)
	filter.SMTP_IN.AuthRequest(filterAuthCb)
	filter.SMTP_IN.MailFromRequest(filterMailFromCb)
	filter.SMTP_IN.RcptToRequest(filterRcptToCb)
	filter.SMTP_IN.DataRequest(filterDataCb)
	filter.SMTP_IN.DataLineRequest(filterDataLineCb)

	filter.SMTP_IN.CommitRequest(filterCommitCb)

	filter.Dispatch()
}
```

Filter requests support the following responses:
```go
// go on with the next filter
filter.Proceed()

// mark the mail as X-Spam: yes
filter.Junk()

// reject with SMTP status and message
filter.Reject("550 go away !")

// disconnect with SMTP status and message
filter.Disconnect("421 go away !")

// rewrite the parameter to new value
filter.Rewrite("gill3s@poolp.org")

// generate a report event with a rewritten parameter
filter.Report("gill3s@poolp.org")
```



## Utilities

### cmd/table

`cmd/table` is a utility to help test table backends during development.

```
$ table -table foobar -service userinfo -lookup gilles \
    /usr/local/libexec/smtpd/table-passwd /etc/passwd
lookup-result|deadbeefabadf00d|found|1000:1000:/home/gilles

$ table -table foobar -service userinfo -lookup gillou \
    /usr/local/libexec/smtpd/table-passwd /etc/passwd
lookup-result|deadbeefabadf00d|not-found
```



## Current state

- basic (working) implementation of current (0.7) [smtpd-filters(7)](https://man.openbsd.org/smtpd-filters) protocol
- basic (working) implementation of current (1.0) [smtpd-tables(7)](https://man.openbsd.org/smtpd-tables) protocol
- APIs are still subject to changes for improvement: they're not 100% stable and written in stone
