package netx

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/tlsutils"
)

type TLSInspectResult struct {
	Description     string
	Raw             []byte
	RelativeDomains []string
	RelativeEmail   []string
	RelativeAccount []string
	RelativeURIs    []string
}

func (t TLSInspectResult) String() string {
	return t.Description
}

func (t TLSInspectResult) Show() {
	fmt.Println(t.Description)
}

func TLSInspectTimeout(addr string, seconds float64) ([]*TLSInspectResult, error) {
	host, port, _ := utils.ParseStringToHostPort(addr)
	if port <= 0 {
		port = 443
	}
	if host == "" {
		host = addr
	}

	dialTimeout := 10 * time.Second
	if seconds > 0 {
		dialTimeout = time.Duration(seconds) * time.Second
	}

	conn, err := DialTCPTimeout(dialTimeout, utils.HostPort(host, port))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var results []*TLSInspectResult
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: host,
		VerifyConnection: func(state tls.ConnectionState) error {
			for _, cert := range state.PeerCertificates {
				if cert == nil {
					continue
				}
				var domains []string

				var urls []string
				for _, u := range cert.URIs {
					urls = append(urls, u.String())
					host, _, _ := utils.ParseStringToHostPort(u.Hostname())
					if host == "" {
						host = u.Hostname()
					}
					if host == "" {
						continue
					}
					domains = append(domains, host)
				}

				domains = append(domains, cert.ExcludedURIDomains...)
				domains = append(domains, cert.PermittedURIDomains...)

				var emails []string
				domains = append(domains, cert.DNSNames...)
				domains = append(domains, cert.PermittedDNSDomains...)
				domains = append(domains, cert.ExcludedDNSDomains...)
				emails = append(emails, cert.EmailAddresses...)
				emails = append(emails, cert.PermittedEmailAddresses...)
				emails = append(emails, cert.ExcludedEmailAddresses...)
				emails = utils.RemoveRepeatStringSlice(emails)
				var accounts []string
				for _, e := range emails {
					if strings.Contains(e, "@") {
						r := strings.Split(e, "@")
						domains = append(domains, r[1])
						accounts = append(accounts, r[0])
					} else {
						accounts = append(accounts, e)
					}
				}
				domains = utils.RemoveRepeatStringSlice(domains)
				text, err := tlsutils.CertificateText(cert)
				if err != nil {
					continue
				}

				result := TLSInspectResult{
					Description:     text,
					Raw:             cert.Raw,
					RelativeDomains: domains,
					RelativeEmail:   emails,
					RelativeAccount: utils.RemoveRepeatStringSlice(accounts),
					RelativeURIs:    utils.RemoveRepeatStringSlice(urls),
				}
				results = append(results, &result)
			}
			return nil
		},
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionSSL30, // nolint[:staticcheck]
		MaxVersion:         tls.VersionTLS13,
		KeyLogWriter:       nil,
	})
	err = tlsConn.HandshakeContext(utils.TimeoutContextSeconds(5))
	if err != nil {
		log.Errorf("TLSInspect: handshake error: %s", err)
	}
	return results, nil
}

// Inspect 检查目标地址的TLS证书，并返回其证书信息与错误
// Example:
// ```
// cert, err := tls.Inspect("yaklang.io:443")
// ```
func TLSInspect(addr string) ([]*TLSInspectResult, error) {
	return TLSInspectTimeout(addr, 10)
}
