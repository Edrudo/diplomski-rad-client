package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"path"
	"runtime"

	"http3-client-poc/cmd/exitcodes"
	"http3-client-poc/internal/utils"
)

var certPath string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		utils.DefaultLogger.Fatalf(errors.New("failed to get current frame"), exitcodes.ExitFailedInit)
	}

	certPath = path.Dir(filename)
}

// GetCertificatePaths returns the paths to certificate and key
func GetCertificatePaths() (string, string) {
	return path.Join(certPath, "cert.pem"), path.Join(certPath, "priv.key")
}

// GetTLSConfig returns a tls config for quic.clemente.io
func GetTLSConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair(GetCertificatePaths())
	if err != nil {
		utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedInit)
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
	}
}

// AddRootCA adds the root CA certificate to a cert pool
func AddRootCA(certPool *x509.CertPool) {
	caCertPath := path.Join(certPath, "ca.pem")
	caCertRaw, err := os.ReadFile(caCertPath)
	if err != nil {
		utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedInit)
	}
	if ok := certPool.AppendCertsFromPEM(caCertRaw); !ok {
		utils.DefaultLogger.Fatalf(errors.New("could not add root ceritificate to pool"), exitcodes.ExitFailedInit)
	}
}

// GetRootCA returns an x509.CertPool containing (only) the CA certificate
func GetRootCA() *x509.CertPool {
	pool := x509.NewCertPool()
	AddRootCA(pool)
	return pool
}
