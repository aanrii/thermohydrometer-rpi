package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"golang.org/x/xerrors"
)

func TLSConfig(rootCAPath, certPath, privateKeyPath string) (*tls.Config, error) {
	rootCAFile, err := ioutil.ReadFile(rootCAPath)
	if err != nil {
		return nil, xerrors.Errorf("Failed to load a root CA file: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(rootCAFile) {
		return nil, xerrors.Errorf("Failed to append certs from %s", rootCAFile)
	}

	cert, err := tls.LoadX509KeyPair(certPath, privateKeyPath)
	if err != nil {
		return nil, xerrors.Errorf("Failed to load a X509 key pair: %w", err)
	}

	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, xerrors.Errorf("Failed to parse a certificate: %w", err)
	}

	return &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"x-amzn-mqtt-ca"},
	}, nil
}
