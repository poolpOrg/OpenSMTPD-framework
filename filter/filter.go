package filter

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

type LinkConnectCb func(timestamp time.Time, sessionId string, rdns string, fcrdns string, src string, dest string)
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

type ConnectRequestCb func(timestamp time.Time, sessionId string, rdns string, fcrdns string, src string, dest string) Response
type HeloRequestCb func(timestamp time.Time, sessionId string, helo string) Response
type EhloRequestCb func(timestamp time.Time, sessionId string, ehlo string) Response
type StartTLSRequestCb func(timestamp time.Time, sessionId string, tlsString string) Response
type AuthRequestCb func(timestamp time.Time, sessionId string, method string) Response
type MailFromRequestCb func(timestamp time.Time, sessionId string, from string) Response
type RcptToRequestCb func(timestamp time.Time, sessionId string, to string) Response
type DataRequestCb func(timestamp time.Time, sessionId string) Response
type DataLineRequestCb func(timestamp time.Time, sessionId string, line string) []string
type CommitRequestCb func(timestamp time.Time, sessionId string) Response

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

func (d *direction) registeredReportEvents() []string {
	ret := make([]string, 0)
	if d.linkConnect != nil {
		ret = append(ret, "link-connect")
	}
	if d.linkGreeting != nil {
		ret = append(ret, "link-greeting")
	}
	if d.linkIdentify != nil {
		ret = append(ret, "link-identify")
	}
	if d.linkTLS != nil {
		ret = append(ret, "link-tls")
	}
	if d.linkAuth != nil {
		ret = append(ret, "link-auth")
	}
	if d.linkDisconnect != nil {
		ret = append(ret, "link-disconnect")
	}
	if d.txReset != nil {
		ret = append(ret, "tx-reset")
	}
	if d.txBegin != nil {
		ret = append(ret, "tx-begin")
	}
	if d.txMail != nil {
		ret = append(ret, "tx-mail")
	}
	if d.txRcpt != nil {
		ret = append(ret, "tx-rcpt")
	}
	if d.txEnvelope != nil {
		ret = append(ret, "tx-envelope")
	}
	if d.txData != nil {
		ret = append(ret, "tx-data")
	}
	if d.txCommit != nil {
		ret = append(ret, "tx-commit")
	}
	if d.txRollback != nil {
		ret = append(ret, "tx-rollback")
	}
	if d.protocolClient != nil {
		ret = append(ret, "protocol-client")
	}
	if d.protocolServer != nil {
		ret = append(ret, "protocol-server")
	}
	if d.filterReport != nil {
		ret = append(ret, "filter-report")
	}
	if d.filterResponse != nil {
		ret = append(ret, "filter-response")
	}
	if d.timeout != nil {
		ret = append(ret, "timeout")
	}
	return ret
}

var SMTP_IN = &direction{}
var SMTP_OUT = &direction{}

func Init() {
}

func (d *direction) OnLinkConnect(cb LinkConnectCb) {
	d.linkConnect = cb
}

func (d *direction) OnLinkDisconnect(cb LinkDisconnectCb) {
	d.linkDisconnect = cb
}

func (d *direction) OnLinkGreeting(cb LinkGreetingCb) {
	d.linkGreeting = cb
}

func (d *direction) OnLinkIdentify(cb LinkIdentifyCb) {
	d.linkIdentify = cb
}

func (d *direction) OnLinkAuth(cb LinkAuthCb) {
	d.linkAuth = cb
}

func (d *direction) OnLinkTLS(cb LinkTLSCb) {
	d.linkTLS = cb
}

func (d *direction) OnTxReset(cb TxResetCb) {
	d.txReset = cb
}

func (d *direction) OnTxBegin(cb TxBeginCb) {
	d.txBegin = cb
}

func (d *direction) OnTxMail(cb TxMailCb) {
	d.txMail = cb
}

func (d *direction) OnTxRcpt(cb TxRcptCb) {
	d.txRcpt = cb
}

func (d *direction) OnTxEnvelope(cb TxEnvelopeCb) {
	d.txEnvelope = cb
}

func (d *direction) OnTxData(cb TxDataCb) {
	d.txData = cb
}

func (d *direction) OnTxCommit(cb TxCommitCb) {
	d.txCommit = cb
}

func (d *direction) OnTxRollback(cb TxRollbackCb) {
	d.txRollback = cb
}

func (d *direction) OnProtocolClient(cb ProtocolClientCb) {
	d.protocolClient = cb
}

func (d *direction) OnProtocolServer(cb ProtocolServerCb) {
	d.protocolServer = cb
}

func (d *direction) OnFilterReport(cb FilterReportCb) {
	d.filterReport = cb
}

func (d *direction) OnFilterResponse(cb FilterResponseCb) {
	d.filterResponse = cb
}

func (d *direction) OnTimeout(cb TimeoutCb) {
	d.timeout = cb
}

func (d *direction) ConnectRequest(cb ConnectRequestCb) {
	d.filterConnect = cb
}

func (d *direction) HeloRequest(cb HeloRequestCb) {
	d.filterHelo = cb
}

func (d *direction) EhloRequest(cb EhloRequestCb) {
	d.filterEhlo = cb
}

func (d *direction) StartTLSRequest(cb StartTLSRequestCb) {
	d.filterStartTLS = cb
}

func (d *direction) AuthRequest(cb AuthRequestCb) {
	d.filterAuth = cb
}

func (d *direction) MailFromRequest(cb MailFromRequestCb) {
	d.filterMailFrom = cb
}

func (d *direction) RcptToRequest(cb RcptToRequestCb) {
	d.filterRcptTo = cb
}

func (d *direction) DataRequest(cb DataRequestCb) {
	d.filterData = cb
}

func (d *direction) DataLineRequest(cb DataLineRequestCb) {
	d.filterDataLine = cb
}

func (d *direction) CommitRequest(cb CommitRequestCb) {
	d.filterCommit = cb
}

func handleReport(timestamp time.Time, event string, dir *direction, sessionId string, atoms []string) {
	switch event {
	case "link-connect":
		if dir.linkConnect == nil {
			return
		}
		dir.linkConnect(timestamp, sessionId, atoms[0], atoms[1], atoms[2], atoms[3])

	case "link-disconnect":
		if dir.linkDisconnect == nil {
			return
		}
		dir.linkDisconnect(timestamp, sessionId)

	case "link-greeting":
		if dir.linkGreeting == nil {
			return
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
		dir.protocolClient(timestamp, sessionId, atoms[0])

	case "protocol-server":
		if dir.protocolServer == nil {
			return
		}
		dir.protocolServer(timestamp, sessionId, atoms[0])

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

func handleFilter(timestamp time.Time, event string, dir *direction, sessionId string, atoms []string) {
	var res Response

	opaqueValue := atoms[0]

	atoms = atoms[1:]
	switch event {
	case "connect":
		if dir.filterConnect == nil {
			return
		}
		res = dir.filterConnect(timestamp, sessionId, atoms[0], atoms[1], atoms[2], atoms[3])

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
		lines := dir.filterDataLine(timestamp, sessionId, atoms[0])
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
	for _, event := range SMTP_IN.registeredReportEvents() {
		fmt.Fprintf(os.Stdout, "register|report|smtp-in|%s\n", event)
	}
	for _, event := range SMTP_OUT.registeredReportEvents() {
		fmt.Fprintf(os.Stdout, "register|report|smtp-out|%s\n", event)
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
		var eventDirectionPtr *direction
		if eventDirection == "smtp-in" {
			eventDirectionPtr = SMTP_IN
		} else if eventDirection == "smtp-out" {
			eventDirectionPtr = SMTP_OUT
		} else {
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
			handleReport(timestampToTime(timestamp), eventKind, eventDirectionPtr, eventSessionId, atoms)
		} else if eventType == "filter" {
			handleFilter(timestampToTime(timestamp), eventKind, eventDirectionPtr, eventSessionId, atoms)
		} else {
			log.Fatalf("Unknown command %s", eventType)
		}

	}

}

//report|0.7|1576146008.006099|smtp-in|link-connect|7641df9771b4ed00|mail.openbsd.org|pass|199.185.178.25:33174|45.77.67.80:25
//filter|0.7|1576146008.006099|smtp-in|connect|7641df9771b4ed00|1ef1c203cc576e5d|mail.openbsd.org|pass|199.185.178.25:33174|45.77.67.80:25
