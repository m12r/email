package smtp

import (
	"crypto/tls"
	"crypto/x509"
	"net"

	"github.com/m12r/email"
)

type Config struct {
	ServerAddr  string
	Username    string
	Password    string
	UseCRAMMD5  bool
	UseClear    bool
	UseStartTLS bool
}

func (c *Config) NewSender() (email.Sender, error) {
	server, _, err := net.SplitHostPort(c.ServerAddr)
	if err != nil {
		server = c.ServerAddr
	}

	auth := WithPlainAuth("", c.Username, c.Password, server)
	if c.UseCRAMMD5 {
		auth = WithCRAMMD5Auth(c.Username, c.Password)
	}

	opts := make([]SenderOpt, 0)
	if !c.UseClear {
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		tlsConfig := &tls.Config{
			RootCAs:    certPool,
			ServerName: server,
		}
		opts = append(opts, WithTLSConfig(tlsConfig))

		if c.UseStartTLS {
			opts = append(opts, WithStartTLS(true))
		}
	}

	sender := NewSender(c.ServerAddr, auth, opts...)
	return sender, nil
}
