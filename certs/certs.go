package certs

import (
	"crypto/tls"
	"fmt"
	"os"
	"sync"
	"time"
)

type Certificate struct {
	certificate *tls.Certificate
	sync.Mutex
}

func (c *Certificate) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return c.getCertificate()
}

func (c *Certificate) GetClientCertificate(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	return c.getCertificate()
}

func (c *Certificate) getCertificate() (*tls.Certificate, error) {
	c.Lock()
	defer c.Unlock()
	return c.certificate, nil
}

func (c *Certificate) setCertificate(cert *tls.Certificate) {
	c.Lock()
	defer c.Unlock()
	c.certificate = cert
}

func (c *Certificate) ResetCertificatePeriodically(certFilePath, keyFilePath string, interval time.Duration) {
	timer := time.NewTimer(0 * time.Second)

	fmt.Printf("reloading certificates every %s\n", interval)

	for {
		select {
		case <-timer.C:
			certificate, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading cert from '%s' and '%s': %s\n", certFilePath, keyFilePath, err.Error())
				os.Exit(2)
			}

			c.setCertificate(&certificate)
			fmt.Printf("reloaded certificate at %v\n", time.Now())

			timer.Reset(interval)
		}
	}

}
