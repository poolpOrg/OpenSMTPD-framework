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

type Hook int

const (
	HOOK_LINK_CONNECT Hook = 0
)

func (h Hook) String() string {
	switch h {
	case HOOK_LINK_CONNECT:
		return "link-connect"
	default:
		log.Fatalf("Unknown hook %d", h)
	}
	return ""
}

func timestampToTime(timestamp float64) time.Time {
	sec := int64(timestamp)
	nsec := int64((timestamp - float64(sec)) * 1e9)
	return time.Unix(sec, nsec)
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

func handleReport(timestamp time.Time, event string, dir *direction, sessionId string, atoms []string) {
	switch event {
	case "link-connect":
		dir.linkConnect(timestamp, sessionId, atoms[0], atoms[1], atoms[2], atoms[3])
	case "link-disconnect":
		dir.linkDisconnect(timestamp, sessionId)
	case "link-greeting":
		dir.linkGreeting(timestamp, sessionId, atoms[0])
	case "link-identify":
		dir.linkIdentify(timestamp, sessionId, atoms[0], atoms[1])
	case "link-auth":
		dir.linkAuth(timestamp, sessionId, atoms[0], atoms[1])
	case "link-tls":
		dir.linkTLS(timestamp, sessionId, atoms[0])
	case "tx-reset":
		dir.txReset(timestamp, sessionId, atoms[0])
	case "tx-begin":
		dir.txBegin(timestamp, sessionId, atoms[0])
	case "tx-mail":
		dir.txMail(timestamp, sessionId, atoms[0], atoms[1], atoms[2])
	case "tx-rcpt":
		dir.txRcpt(timestamp, sessionId, atoms[0], atoms[1], atoms[2])
	case "tx-envelope":
		dir.txEnvelope(timestamp, sessionId, atoms[0], atoms[1])
	case "tx-data":
		dir.txData(timestamp, sessionId, atoms[0], atoms[1])
	case "tx-commit":
		if size, err := strconv.Atoi(atoms[1]); err != nil {
			log.Fatalf("Failed to convert size %s to int", atoms[1])
		} else {
			dir.txCommit(timestamp, sessionId, atoms[0], size)
		}
	case "tx-rollback":
		dir.txRollback(timestamp, sessionId, atoms[0])
	case "protocol-client":
		dir.protocolClient(timestamp, sessionId, atoms[0])
	case "protocol-server":
		dir.protocolServer(timestamp, sessionId, atoms[0])
	case "filter-report":
		dir.filterReport(timestamp, sessionId, atoms[0], atoms[1], atoms[2])
	case "filter-response":
		dir.filterResponse(timestamp, sessionId, atoms[0], atoms[1], atoms[2:]...)
	case "timeout":
		dir.timeout(timestamp, sessionId)
	default:
		log.Fatalf("Unknown event %s", event)
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

		eventType := atoms[0]
		if eventType != "report" {
			log.Fatalf("Invalid event type: %s", eventType)
		}

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
		} else {
			log.Fatalf("Unknown command %s", eventType)
		}

	}

}

//report|0.7|1576146008.006099|smtp-in|link-connect|7641df9771b4ed00|mail.openbsd.org|pass|199.185.178.25:33174|45.77.67.80:25
