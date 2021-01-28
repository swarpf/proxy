package swproxy

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"path"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
)

func setCA(rootCa tls.Certificate) error {
	var err error
	if rootCa.Leaf, err = x509.ParseCertificate(rootCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = rootCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&rootCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&rootCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&rootCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&rootCa)}
	return nil
}

func getRootCA(certDir string) tls.Certificate {
	appfs := afero.NewOsFs()

	dirExists, err := afero.DirExists(appfs, certDir)
	if err != nil {
		log.Fatal().Err(err).
			Str("cert_dir", certDir).
			Msg("Failed to check if the certificate directory exists")
	}

	if !dirExists {
		log.Warn().
			Str("cert_dir", certDir).
			Msg("certificate directory does not exist - trying to create it")

		if err := appfs.MkdirAll(certDir, 0755); err != nil {
			log.Fatal().Err(err).
				Str("cert_dir", certDir).
				Msg("Failed to create the certificate directory")
		}
	}

	caCertPath := path.Join(certDir, "ca.crt")
	caCertExists, err := afero.Exists(appfs, path.Join(certDir, caCertPath))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to check if the CA certificate exists")
	}

	caKeyPath := path.Join(certDir, "ca.key")
	caKeyExists, err := afero.Exists(appfs, caKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to check if the CA private key exists")
	}

	log.Debug().
		Str("cert_dir", certDir).
		Str("ca_cert_path", caCertPath).
		Str("ca_key_path", caKeyPath).
		Bool("ca_cert_exists", caCertExists).
		Bool("ca_key_exists", caKeyExists).
		Send()

	// create a new CA and write its certificate and private key to disk
	if !caCertExists || !caKeyExists {
		_ = appfs.Remove(caCertPath)
		_ = appfs.Remove(caKeyPath)

		log.Info().Msg("Generating new CA cert and key")

		caCert, caPrivKey := generateCA()
		if err := afero.WriteFile(appfs, caCertPath, caCert, 0644); err != nil {
			log.Fatal().Err(err).Msg("Failed to write CA certificate to disk")
		}
		if err := afero.WriteFile(appfs, caKeyPath, caPrivKey, 0600); err != nil {
			log.Fatal().Err(err).Msg("Failed to write CA private key to disk")
		}

		log.Trace().
			Bytes("ca_cert", caCert).
			Bytes("ca_key", caPrivKey).
			Msg("Generated CA cert/key pair")
	}

	// read (back) the certificate and private key from disk
	caCert, err := afero.ReadFile(appfs, caCertPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read CA certificate from disk")
	}

	caPrivKey, err := afero.ReadFile(appfs, caKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read CA private key from disk")
	}

	rootCa, err := tls.X509KeyPair(caCert, caPrivKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create X509 TLS key pair")
	}

	return rootCa
}

func generateCA() (caCert, privateKey []byte) {
	notBefore := time.Now().Add(-10 * time.Second)
	notAfter := notBefore.AddDate(1, 0, 0)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate serial number")
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"swarpf v2"},
			Locality:     []string{"Local Network"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed go generate private key")
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, privKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create certificate")
	}

	var certBuffer bytes.Buffer
	if err := pem.Encode(&certBuffer, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatal().Err(err).Msg("Failed to write data to cert.pem")
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to marshal private key")
	}
	var privateKeyBuffer bytes.Buffer
	if err := pem.Encode(&privateKeyBuffer, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatal().Err(err).Msg("Failed to write data to key.pem")
	}

	return certBuffer.Bytes(), privateKeyBuffer.Bytes()
}
