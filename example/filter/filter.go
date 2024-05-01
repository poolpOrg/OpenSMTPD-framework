package main

import (
	"fmt"
	"os"
	"time"

	"github.com/poolpOrg/OpenSMTPD-framework/filter"
)

func linkConnectCb(timestamp time.Time, sessionId string, rdns string, fcrdns string, src string, dest string) {
	fmt.Fprintf(os.Stderr, "%s: %s: link-connect: %s|%s|%s|%s\n", timestamp, sessionId, rdns, fcrdns, src, dest)
}

func linkDisconnectCb(timestamp time.Time, sessionId string) {
	fmt.Fprintf(os.Stderr, "%s: %s: link-disconnect\n", timestamp, sessionId)
}

func linkGreetingCb(timestamp time.Time, sessionId string, hostname string) {
	fmt.Fprintf(os.Stderr, "%s: %s: link-greeting: %s\n", timestamp, sessionId, hostname)
}

func linkIdentifyCb(timestamp time.Time, sessionId string, method string, hostname string) {
	fmt.Fprintf(os.Stderr, "%s: %s: link-identify: %s|%s\n", timestamp, sessionId, method, hostname)
}

func linkAuthCb(timestamp time.Time, sessionId string, result string, username string) {
	fmt.Fprintf(os.Stderr, "%s: %s: link-auth: %s|%s\n", timestamp, sessionId, result, username)
}

func linkTLSCb(timestamp time.Time, sessionId string, tlsString string) {
	fmt.Fprintf(os.Stderr, "%s: %s: link-tls: %s\n", timestamp, sessionId, tlsString)
}

func txResetCb(timestamp time.Time, sessionId string, messageId string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-reset: %s\n", timestamp, sessionId, messageId)
}

func txBeginCb(timestamp time.Time, sessionId string, messageId string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-begin: %s\n", timestamp, sessionId, messageId)
}

func txMailCb(timestamp time.Time, sessionId string, messageId string, result string, from string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-mail: %s|%s|%s\n", timestamp, sessionId, messageId, result, from)
}

func txRcptCb(timestamp time.Time, sessionId string, messageId string, result string, to string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-rcpt: %s|%s|%s\n", timestamp, sessionId, messageId, result, to)
}

func txEnvelopeCb(timestamp time.Time, sessionId string, messageId string, envelopeId string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-envelope: %s|%s\n", timestamp, sessionId, messageId, envelopeId)
}

func txDataCb(timestamp time.Time, sessionId string, messageId string, result string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-data: %s|%s\n", timestamp, sessionId, messageId, result)
}

func txCommmitCb(timestamp time.Time, sessionId string, messageId string, messageSize int) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-commit: %s|%d\n", timestamp, sessionId, messageId, messageSize)
}

func txRollbackCb(timestamp time.Time, sessionId string, messageId string) {
	fmt.Fprintf(os.Stderr, "%s: %s: tx-rollback: %s\n", timestamp, sessionId, messageId)
}

func protocolClientCb(timestamp time.Time, sessionId string, command string) {
	fmt.Fprintf(os.Stderr, "%s: %s: protocol-client: %s\n", timestamp, sessionId, command)
}

func protocolServerCb(timestamp time.Time, sessionId string, response string) {
	fmt.Fprintf(os.Stderr, "%s: %s: protocol-server: %s\n", timestamp, sessionId, response)
}

func filterReportCb(timestamp time.Time, sessionId string, filterKind string, name string, message string) {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-report: %s|%s|%s\n", timestamp, sessionId, filterKind, name, message)
}

func filterResponseCb(timestamp time.Time, sessionId string, phase string, response string, param ...string) {
	fmt.Fprintf(os.Stderr, "%s: %s: filter-response: %s|%s|%v\n", timestamp, sessionId, phase, response, param)
}

func timeoutCb(timestamp time.Time, sessionId string) {
	fmt.Fprintf(os.Stderr, "%s: %s: timeout\n", timestamp, sessionId)
}

func main() {
	filter.Init()

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

	filter.Dispatch()
}