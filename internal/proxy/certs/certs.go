package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"proxy/internal/network"
	"time"
)

type TLSConfig struct {
	Cfg  *tls.Config
	Conn *tls.Conn
}

func LoadCertificate(certPath, keyPath string) (*tls.Certificate, error) {
	ca := new(tls.Certificate)

	cert, err := loadCertificate(certPath)
	if err != nil {
		log.Println("error loading CA certificate:", err)

		return nil, err
	}

	privateKey, err := loadPrivateKey(keyPath)
	if err != nil {
		log.Println("error loading private key:", err)

		return nil, err
	}

	ca = &tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  privateKey,
		Leaf:        cert,
	}

	return ca, nil
}

func GetTLSConfig(req *network.HTTPRequest, cert *tls.Certificate) (*TLSConfig, error) {
	provisionalCert, err := getTLSCert(req.Host, cert)
	if err != nil {
		return nil, err
	}

	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	tlsCfg.Certificates = []tls.Certificate{*provisionalCert}

	var sconn *tls.Conn

	tlsCfg.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		cConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		cConfig.ServerName = hello.ServerName
		sconn, err = tls.Dial("tcp", fmt.Sprintf("%s:%s", req.Host, req.Port), cConfig)
		if err != nil {
			return nil, err
		}
		return getTLSCert(hello.ServerName, cert)
	}

	return &TLSConfig{
		Cfg:  tlsCfg,
		Conn: sconn,
	}, nil
}

func getTLSCert(host string, ca *tls.Certificate) (*tls.Certificate, error) {
	csr, priv, err := createRequestCertificate(host)
	if err != nil {
		return nil, err
	}

	certBytes, err := signCertificate(ca.Leaf, ca.PrivateKey, csr)
	if err != nil {
		log.Println("sign", err)
		return nil, err
	}

	cert := &tls.Certificate{
		Certificate: [][]byte{certBytes},
		PrivateKey:  priv,
	}

	return cert, nil
}

func createRequestCertificate(host string) (*x509.CertificateRequest, *ecdsa.PrivateKey, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: host,
		},
		DNSNames: []string{host},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, template, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate request: %v", err)
	}

	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate request: %v", err)
	}

	return csr, priv, nil
}

func signCertificate(caCert *x509.Certificate, caKey interface{}, csr *x509.CertificateRequest) ([]byte, error) {
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      csr.Subject,
		DNSNames:     csr.DNSNames,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, csr.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	return certBytes, nil
}

func loadCertificate(certPath string) (*x509.Certificate, error) {
	certBuf, err := os.ReadFile(certPath)
	if err != nil {
		log.Println("error reading CA certificate from file", err)

		return nil, err
	}

	block, _ := pem.Decode(certBuf)
	if block == nil || block.Type != "CERTIFICATE" {
		log.Println("error decoding CA certificate", err)

		return nil, err
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Println("error parsing CA certificate", err)

		return nil, err
	}

	return cert, nil
}

func loadPrivateKey(keyPath string) (interface{}, error) {
	keyBuf, err := os.ReadFile(keyPath)
	if err != nil {
		log.Println("error reading private key from file", err)

		return nil, err
	}

	block, _ := pem.Decode(keyBuf)
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Println("error decoding private key", err)

		return nil, err
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Println("error parsing private key", err)

		return nil, err
	}

	return key, nil
}
