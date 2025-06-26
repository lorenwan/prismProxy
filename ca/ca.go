package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

const (
	caCertFile = "ca.crt"
	caKeyFile  = "ca.key"
	rsaBits    = 2048
)

// LoadCA loads the certificate authority from disk.
// If the files don't exist, it generates a new CA.
func LoadCA() (tls.Certificate, error) {
	// Try to load existing certificate
	cert, err := tls.LoadX509KeyPair(caCertFile, caKeyFile)
	if err == nil {
		log.Println("Loaded existing CA certificate and key.")
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		return cert, err
	}

	// If loading fails (e.g., files not found), generate a new CA
	if os.IsNotExist(err) {
		log.Println("CA certificate not found, generating a new one...")
		return generateCA()
	}

	// For other errors, return the error
	return tls.Certificate{}, err
}

// generateCA creates a new certificate authority and saves it to disk.
func generateCA() (tls.Certificate, error) {
	// Create a template for the certificate
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"My Awesome Proxy"},
			CommonName:   "My Awesome Proxy CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // Valid for 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Generate a new private key
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Create the certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	// --- Save the CA certificate to ca.crt ---
	certOut, err := os.Create(caCertFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes})

	// --- Save the private key to ca.key ---
	keyOut, err := os.OpenFile(caKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return tls.Certificate{}, err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	log.Printf("Generated and saved new CA to %s and %s", caCertFile, caKeyFile)

	// Return the newly created certificate
	return tls.LoadX509KeyPair(caCertFile, caKeyFile)
}
