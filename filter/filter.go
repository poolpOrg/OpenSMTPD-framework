package filter

import (
	"bufio"
	"fmt"
	"log"
	"net"
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

func parseAddress(addr string) (net.Addr, error) {
	if strings.Contains(addr, "/") {
		// Unix domain socket
		return net.ResolveUnixAddr("unix", addr)
	}

	// Check if the address includes a port
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// No port provided, use default port 80
		addr = net.JoinHostPort(addr, "0")
	} else {
		// Use the original address as it includes a port
		addr = net.JoinHostPort(host, port)
	}

	// Try to parse as a TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err == nil {
		return tcpAddr, nil
	}

	return nil, fmt.Errorf("failed to parse as any known address type: %s", err)
}

type Response interface {
	_x()
}

type proceed struct{}
type junk struct{}
type reject struct{ errorMsg string }
type disconnect struct{ errorMsg string }
type rewrite struct{ parameter string }
type report struct{ parameter string }

func (f proceed) _x()    {}
func (f junk) _x()       {}
func (f reject) _x()     {}
func (f disconnect) _x() {}
func (f rewrite) _x()    {}
func (f report) _x()     {}

func Proceed() Response {
	return proceed{}
}

func Junk() Response {
	return junk{}
}

func Reject(errorMsg string) Response {
	return reject{errorMsg: errorMsg}
}

func Disconnect(errorMsg string) Response {
	return disconnect{errorMsg: errorMsg}
}

func Rewrite(parameter string) Response {
	return rewrite{parameter: parameter}
}

func Report(parameter string) Response {
	return report{parameter: parameter}
}

type LinkConnectCb func(timestamp time.Time, sessionId string, rdns string, fcrdns string, src net.Addr, dest net.Addr)
type LinkGreetingCb func(timestamp time.Time, sessionId string, hostname string)
type LinkIdentifyCb func(timestamp time.Time, sessionId string, method string, hostname string)
type LinkTLSCb func(timestamp time.Time, sessionId string, tlsString string)
type LinkAuthCb func(timestamp time.Time, sessionId string, result string, username string)
type LinkDisconnectCb func(timestamp time.Time, sessionId string)

type TxResetCb func(timestamp time.Time, sessionId string, messageId string)
type TxBeginCb func(timestamp time.Time, sessionId string, messageId string)
type TxMailCb func(timestamp time.Time, sessionId string, messageId string, result string, from string)
type TxRcptCb func(timestamp time.Time, sessionId string, messageId string, result string, to string)
type TxEnvelopeCb func(timestamp time.Time, sessionId string, messageId string, envelopeId string)
type TxDataCb func(timestamp time.Time, sessionId string, messageId string, result string)
type TxCommitCb func(timestamp time.Time, sessionId string, messageId string, messageSize int)
type TxRollbackCb func(timestamp time.Time, sessionId string, messageId string)

type ProtocolClientCb func(timestamp time.Time, sessionId string, command string)
type ProtocolServerCb func(timestamp time.Time, sessionId string, response string)

type FilterReportCb func(timestamp time.Time, sessionId string, filterKind string, name string, message string)
type FilterResponseCb func(timestamp time.Time, sessionId string, phase string, response string, param ...string)

type TimeoutCb func(timestamp time.Time, sessionId string)

type ConnectRequestCb func(timestamp time.Time, sessionId string, rdns string, src net.Addr) Response
type HeloRequestCb func(timestamp time.Time, sessionId string, helo string) Response
type EhloRequestCb func(timestamp time.Time, sessionId string, ehlo string) Response
type StartTLSRequestCb func(timestamp time.Time, sessionId string, tlsString string) Response
type AuthRequestCb func(timestamp time.Time, sessionId string, method string) Response
type MailFromRequestCb func(timestamp time.Time, sessionId string, from string) Response
type RcptToRequestCb func(timestamp time.Time, sessionId string, to string) Response
type DataRequestCb func(timestamp time.Time, sessionId string) Response
type DataLineRequestCb func(timestamp time.Time, sessionId string, line string) []string
type CommitRequestCb func(timestamp time.Time, sessionId string) Response

type reporting struct {
	linkConnect    LinkConnectCb
	linkGreeting   LinkGreetingCb
	linkIdentify   LinkIdentifyCb
	linkTLS        LinkTLSCb
	linkAuth       LinkAuthCb
	linkDisconnect LinkDisconnectCb

	txReset    TxResetCb
	txBegin    TxBeginCb
	txMail     TxMailCb
	txRcpt     TxRcptCb
	txEnvelope TxEnvelopeCb
	txData     TxDataCb
	txCommit   TxCommitCb
	txRollback TxRollbackCb

	protocolClient ProtocolClientCb
	protocolServer ProtocolServerCb

	filterReport   FilterReportCb
	filterResponse FilterResponseCb

	timeout TimeoutCb
}

func (r *reporting) reportEvents() []string {
	ret := make([]string, 0)
	if r.linkConnect != nil {
		ret = append(ret, "link-connect")
	}
	if r.linkGreeting != nil {
		ret = append(ret, "link-greeting")
	}
	if r.linkIdentify != nil {
		ret = append(ret, "link-identify")
	}
	if r.linkTLS != nil {
		ret = append(ret, "link-tls")
	}
	if r.linkAuth != nil {
		ret = append(ret, "link-auth")
	}
	if r.linkDisconnect != nil {
		ret = append(ret, "link-disconnect")
	}
	if r.txReset != nil {
		ret = append(ret, "tx-reset")
	}
	if r.txBegin != nil {
		ret = append(ret, "tx-begin")
	}
	if r.txMail != nil {
		ret = append(ret, "tx-mail")
	}
	if r.txRcpt != nil {
		ret = append(ret, "tx-rcpt")
	}
	if r.txEnvelope != nil {
		ret = append(ret, "tx-envelope")
	}
	if r.txData != nil {
		ret = append(ret, "tx-data")
	}
	if r.txCommit != nil {
		ret = append(ret, "tx-commit")
	}
	if r.txRollback != nil {
		ret = append(ret, "tx-rollback")
	}
	if r.protocolClient != nil {
		ret = append(ret, "protocol-client")
	}
	if r.protocolServer != nil {
		ret = append(ret, "protocol-server")
	}
	if r.filterReport != nil {
		ret = append(ret, "filter-report")
	}
	if r.filterResponse != nil {
		ret = append(ret, "filter-response")
	}
	if r.timeout != nil {
		ret = append(ret, "timeout")
	}
	return ret
}

type filtering struct {
	filterConnect  ConnectRequestCb
	filterHelo     HeloRequestCb
	filterEhlo     EhloRequestCb
	filterStartTLS StartTLSRequestCb
	filterAuth     AuthRequestCb
	filterMailFrom MailFromRequestCb
	filterRcptTo   RcptToRequestCb
	filterData     DataRequestCb
	filterDataLine DataLineRequestCb
	filterCommit   CommitRequestCb
}

func (f *filtering) filterEvents() []string {
	ret := make([]string, 0)
	if f.filterConnect != nil {
		ret = append(ret, "connect")
	}
	if f.filterHelo != nil {
		ret = append(ret, "helo")
	}
	if f.filterEhlo != nil {
		ret = append(ret, "ehlo")
	}
	if f.filterStartTLS != nil {
		ret = append(ret, "starttls")
	}
	if f.filterAuth != nil {
		ret = append(ret, "auth")
	}
	if f.filterMailFrom != nil {
		ret = append(ret, "mail-from")
	}
	if f.filterRcptTo != nil {
		ret = append(ret, "rcpt-to")
	}
	if f.filterData != nil {
		ret = append(ret, "data")
	}
	if f.filterDataLine != nil {
		ret = append(ret, "data-line")
	}
	if f.filterCommit != nil {
		ret = append(ret, "commit")
	}
	return ret
}

type smtpIn struct {
	reporting
	filtering
}

type smtpOut struct {
	reporting
}

type direction struct {
	linkConnect    LinkConnectCb
	linkGreeting   LinkGreetingCb
	linkIdentify   LinkIdentifyCb
	linkTLS        LinkTLSCb
	linkAuth       LinkAuthCb
	linkDisconnect LinkDisconnectCb

	txReset    TxResetCb
	txBegin    TxBeginCb
	txMail     TxMailCb
	txRcpt     TxRcptCb
	txEnvelope TxEnvelopeCb
	txData     TxDataCb
	txCommit   TxCommitCb
	txRollback TxRollbackCb

	protocolClient ProtocolClientCb
	protocolServer ProtocolServerCb

	filterReport   FilterReportCb
	filterResponse FilterResponseCb

	timeout TimeoutCb

	// SMTP_IN ONLY FOR NOW
	filterConnect  ConnectRequestCb
	filterHelo     HeloRequestCb
	filterEhlo     EhloRequestCb
	filterStartTLS StartTLSRequestCb
	filterAuth     AuthRequestCb
	filterMailFrom MailFromRequestCb
	filterRcptTo   RcptToRequestCb
	filterData     DataRequestCb
	filterDataLine DataLineRequestCb
	filterCommit   CommitRequestCb
}

var SMTP_IN = &smtpIn{}
var SMTP_OUT = &smtpOut{}

func Init() {
}

func (r *reporting) OnLinkConnect(cb LinkConnectCb) {
	r.linkConnect = cb
}

func (r *reporting) OnLinkDisconnect(cb LinkDisconnectCb) {
	r.linkDisconnect = cb
}

func (r *reporting) OnLinkGreeting(cb LinkGreetingCb) {
	r.linkGreeting = cb
}

func (r *reporting) OnLinkIdentify(cb LinkIdentifyCb) {
	r.linkIdentify = cb
}

func (r *reporting) OnLinkAuth(cb LinkAuthCb) {
	r.linkAuth = cb
}

func (r *reporting) OnLinkTLS(cb LinkTLSCb) {
	r.linkTLS = cb
}

func (r *reporting) OnTxReset(cb TxResetCb) {
	r.txReset = cb
}

func (r *reporting) OnTxBegin(cb TxBeginCb) {
	r.txBegin = cb
}

func (r *reporting) OnTxMail(cb TxMailCb) {
	r.txMail = cb
}

func (r *reporting) OnTxRcpt(cb TxRcptCb) {
	r.txRcpt = cb
}

func (r *reporting) OnTxEnvelope(cb TxEnvelopeCb) {
	r.txEnvelope = cb
}

func (r *reporting) OnTxData(cb TxDataCb) {
	r.txData = cb
}

func (r *reporting) OnTxCommit(cb TxCommitCb) {
	r.txCommit = cb
}

func (r *reporting) OnTxRollback(cb TxRollbackCb) {
	r.txRollback = cb
}

func (r *reporting) OnProtocolClient(cb ProtocolClientCb) {
	r.protocolClient = cb
}

func (r *reporting) OnProtocolServer(cb ProtocolServerCb) {
	r.protocolServer = cb
}

func (r *reporting) OnFilterReport(cb FilterReportCb) {
	r.filterReport = cb
}

func (r *reporting) OnFilterResponse(cb FilterResponseCb) {
	r.filterResponse = cb
}

func (r *reporting) OnTimeout(cb TimeoutCb) {
	r.timeout = cb
}

func (f *filtering) ConnectRequest(cb ConnectRequestCb) {
	f.filterConnect = cb
}

func (f *filtering) HeloRequest(cb HeloRequestCb) {
	f.filterHelo = cb
}

func (f *filtering) EhloRequest(cb EhloRequestCb) {
	f.filterEhlo = cb
}

func (f *filtering) StartTLSRequest(cb StartTLSRequestCb) {
	f.filterStartTLS = cb
}

func (f *filtering) AuthRequest(cb AuthRequestCb) {
	f.filterAuth = cb
}

func (f *filtering) MailFromRequest(cb MailFromRequestCb) {
	f.filterMailFrom = cb
}

func (f *filtering) RcptToRequest(cb RcptToRequestCb) {
	f.filterRcptTo = cb
}

func (f *filtering) DataRequest(cb DataRequestCb) {
	f.filterData = cb
}

func (f *filtering) DataLineRequest(cb DataLineRequestCb) {
	f.filterDataLine = cb
}

func (f *filtering) CommitRequest(cb CommitRequestCb) {
	f.filterCommit = cb
}

func handleReport(timestamp time.Time, event string, dir *reporting, sessionId string, atoms []string) {

	// XXX - need to ensure atoms is properly parsed (last field may be split multiple times)

	switch event {
	case "link-connect":
		if dir.linkConnect == nil {
			return
		}
		if len(atoms) != 4 {
			log.Fatalf("Invalid input, not enough fields: %s", atoms)
		}
		if srcAddr, err := parseAddress(atoms[2]); err != nil {
			log.Fatalf("Failed to parse source address %s", atoms[2])
		} else if destAddr, err := parseAddress(atoms[3]); err != nil {
			log.Fatalf("Failed to parse destination address %s", atoms[3])
		} else {
			dir.linkConnect(timestamp, sessionId, atoms[0], atoms[1], srcAddr, destAddr)
		}

	case "link-disconnect":
		if dir.linkDisconnect == nil {
			return
		}
		if len(atoms) != 0 {
			log.Fatalf("Invalid input, too many fields: %s", atoms)
		}
		dir.linkDisconnect(timestamp, sessionId)

	case "link-greeting":
		if dir.linkGreeting == nil {
			return
		}
		if len(atoms) != 1 {
			log.Fatalf("Invalid input, expects only one field: %s", atoms)
		}
		dir.linkGreeting(timestamp, sessionId, atoms[0])

	case "link-identify":
		if dir.linkIdentify == nil {
			return
		}
		dir.linkIdentify(timestamp, sessionId, atoms[0], atoms[1])

	case "link-auth":
		if dir.linkAuth == nil {
			return
		}
		dir.linkAuth(timestamp, sessionId, atoms[0], atoms[1])

	case "link-tls":
		if dir.linkTLS == nil {
			return
		}
		dir.linkTLS(timestamp, sessionId, atoms[0])

	case "tx-reset":
		if dir.txReset == nil {
			return
		}
		dir.txReset(timestamp, sessionId, atoms[0])

	case "tx-begin":
		if dir.txBegin == nil {
			return
		}
		dir.txBegin(timestamp, sessionId, atoms[0])

	case "tx-mail":
		if dir.txMail == nil {
			return
		}
		dir.txMail(timestamp, sessionId, atoms[0], atoms[1], atoms[2])

	case "tx-rcpt":
		if dir.txRcpt == nil {
			return
		}
		dir.txRcpt(timestamp, sessionId, atoms[0], atoms[1], atoms[2])

	case "tx-envelope":
		if dir.txEnvelope == nil {
			return
		}
		dir.txEnvelope(timestamp, sessionId, atoms[0], atoms[1])

	case "tx-data":
		if dir.txData == nil {
			return
		}
		dir.txData(timestamp, sessionId, atoms[0], atoms[1])

	case "tx-commit":
		if dir.txCommit == nil {
			return
		}

		if size, err := strconv.Atoi(atoms[1]); err != nil {
			log.Fatalf("Failed to convert size %s to int", atoms[1])
		} else {
			dir.txCommit(timestamp, sessionId, atoms[0], size)
		}

	case "tx-rollback":
		if dir.txRollback == nil {
			return
		}
		dir.txRollback(timestamp, sessionId, atoms[0])

	case "protocol-client":
		if dir.protocolClient == nil {
			return
		}
		dir.protocolClient(timestamp, sessionId, strings.Join(atoms, "|"))

	case "protocol-server":
		if dir.protocolServer == nil {
			return
		}
		dir.protocolServer(timestamp, sessionId, strings.Join(atoms, "|"))

	case "filter-report":
		if dir.filterReport == nil {
			return
		}
		dir.filterReport(timestamp, sessionId, atoms[0], atoms[1], atoms[2])

	case "filter-response":
		if dir.filterResponse == nil {
			return
		}
		dir.filterResponse(timestamp, sessionId, atoms[0], atoms[1], atoms[2:]...)

	case "timeout":
		if dir.timeout == nil {
			return
		}
		dir.timeout(timestamp, sessionId)

	default:
		log.Fatalf("Unknown event %s", event)
	}
}

func handleFilter(timestamp time.Time, event string, dir *filtering, sessionId string, atoms []string) {
	var res Response

	// XXX - need to ensure atoms is properly parsed (last field may be split multiple times)
	opaqueValue, atoms := atoms[0], atoms[1:]

	switch event {
	case "connect":
		if dir.filterConnect == nil {
			return
		}
		if srcAddr, err := parseAddress(atoms[1]); err != nil {
			log.Fatalf("Failed to parse source address %s", atoms[1])
		} else {
			res = dir.filterConnect(timestamp, sessionId, atoms[0], srcAddr)
		}

	case "helo":
		if dir.filterHelo == nil {
			return
		}
		res = dir.filterHelo(timestamp, sessionId, atoms[0])

	case "ehlo":
		if dir.filterHelo == nil {
			return
		}
		res = dir.filterEhlo(timestamp, sessionId, atoms[0])

	case "starttls":
		if dir.filterStartTLS == nil {
			return
		}
		res = dir.filterStartTLS(timestamp, sessionId, atoms[0])

	case "auth":
		if dir.filterAuth == nil {
			return
		}
		res = dir.filterAuth(timestamp, sessionId, atoms[0])

	case "mail-from":
		if dir.filterMailFrom == nil {
			return
		}
		res = dir.filterMailFrom(timestamp, sessionId, atoms[0])

	case "rcpt-to":
		if dir.filterRcptTo == nil {
			return
		}
		res = dir.filterRcptTo(timestamp, sessionId, atoms[0])

	case "data":
		if dir.filterData == nil {
			return
		}
		res = dir.filterData(timestamp, sessionId)

	case "data-line":
		if dir.filterDataLine == nil {
			return
		}
		// data line has special handling
		lines := dir.filterDataLine(timestamp, sessionId, strings.Join(atoms, "|"))
		for _, line := range lines {
			fmt.Fprintf(os.Stdout, "filter-dataline|%s|%s|%s\n", sessionId, opaqueValue, line)
		}
		return

	case "commit":
		if dir.filterCommit == nil {
			return
		}
		res = dir.filterCommit(timestamp, sessionId)

	default:
		log.Fatalf("Unknown event %s", event)
	}

	switch res := res.(type) {
	case proceed:
		fmt.Fprintf(os.Stdout, "filter-result|%s|%s|proceed\n", sessionId, opaqueValue)
	case junk:
		fmt.Fprintf(os.Stdout, "filter-result|%s|%s|junk\n", sessionId, opaqueValue)
	case reject:
		fmt.Fprintf(os.Stdout, "filter-result|%s|%s|reject|%s\n", sessionId, opaqueValue, res.errorMsg)
	case disconnect:
		fmt.Fprintf(os.Stdout, "filter-result|%s|%s|disconnect|%s\n", sessionId, opaqueValue, res.errorMsg)
	case rewrite:
		fmt.Fprintf(os.Stdout, "filter-result|%s|%s|rewrite|%s\n", sessionId, opaqueValue, res.parameter)
	case report:
		fmt.Fprintf(os.Stdout, "filter-result|%s|%s|report|%s\n", sessionId, opaqueValue, res.parameter)
	}
}

func Dispatch() {
	scanner := bufio.NewScanner(os.Stdin)

	protocolVersion := "0.7"
	_ = protocolVersion

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
	for _, event := range SMTP_IN.reportEvents() {
		fmt.Fprintf(os.Stdout, "register|report|smtp-in|%s\n", event)
	}
	for _, event := range SMTP_OUT.reportEvents() {
		fmt.Fprintf(os.Stdout, "register|report|smtp-out|%s\n", event)
	}
	for _, event := range SMTP_IN.filterEvents() {
		fmt.Fprintf(os.Stdout, "register|filter|smtp-in|%s\n", event)
	}
	fmt.Println("register|ready")

	for {
		if !scanner.Scan() {
			log.Fatal("scanner.Scan() failed")
			break
		}
		line := scanner.Text()
		atoms := strings.Split(line, "|")

		if len(atoms) < 6 {
			log.Fatalf("Invalid input, not enough fields: %s", line)
		}

		// checked below
		eventType := atoms[0]

		eventVersion := atoms[1]
		if eventVersion != protocolVersion {
			log.Fatalf("Unsupported protocol version %s", eventVersion)
		}

		eventTimestamp := atoms[2]
		timestamp, err := strconv.ParseFloat(eventTimestamp, 64)
		if err != nil {
			log.Fatalf("Failed to convert timestamp %s to float", eventTimestamp)
		}

		eventDirection := atoms[3]
		if eventDirection != "smtp-in" && eventDirection != "smtp-out" {
			log.Fatalf("Unknown direction %s", eventDirection)
		}

		// these are validated in the handleReport function
		eventKind := atoms[4]

		eventSessionId := atoms[5]
		_, err = strconv.ParseUint(eventSessionId, 16, 64)
		if err != nil {
			log.Fatalf("Failed to convert session id %s to uint64", eventSessionId)
		}

		atoms = atoms[6:]

		if eventType == "report" {
			var direction *reporting
			if eventDirection == "smtp-in" {
				direction = &SMTP_IN.reporting
			} else if eventDirection == "smtp-out" {
				direction = &SMTP_OUT.reporting
			}
			handleReport(timestampToTime(timestamp), eventKind, direction, eventSessionId, atoms)
		} else if eventType == "filter" {
			var direction *filtering
			if eventDirection != "smtp-in" {
				log.Fatalf("Unknown direction %s", eventDirection)
			}
			direction = &SMTP_IN.filtering
			//fmt.Fprintf(os.Stderr, "atoms: %s\n", atoms)
			handleFilter(timestampToTime(timestamp), eventKind, direction, eventSessionId, atoms)
		} else {
			log.Fatalf("Unknown command %s", eventType)
		}

	}

}

//report|0.7|1576146008.006099|smtp-in|link-connect|7641df9771b4ed00|mail.openbsd.org|pass|199.185.178.25:33174|45.77.67.80:25
//filter|0.7|1576146008.006099|smtp-in|connect|7641df9771b4ed00|1ef1c203cc576e5d|mail.openbsd.org|pass|199.185.178.25:33174|45.77.67.80:25
/*
filter|0.7|1576146008.006099|smtp-in|data-line|7641df9771b4ed00|1ef1c203cc576e5d|line 1
filter|0.7|1576146008.006103|smtp-in|data-line|7641df9771b4ed00|1ef1c203cc576e5d|line 2
filter|0.7|1576146008.006105|smtp-in|data-line|7641df9771b4ed00|1ef1c203cc576e5d|.
*/
