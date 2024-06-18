package bootstrap

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"

	"http3-client-poc/cmd/exitcodes"
	"http3-client-poc/internal/application"
	"http3-client-poc/internal/tlsconfig"
	"http3-client-poc/internal/utils"
)

func NewClient() (*application.Client, *http3.RoundTripper) {
	utils.DefaultLogger.SetLogLevel(utils.LogLevelError)

	roundTripper := initilizeRoundTripper()
	httpClient := &http.Client{
		Transport: roundTripper,
	}

	return application.NewClient(sha256.New(), httpClient, roundTripper), roundTripper

}

func initilizeRoundTripper() *http3.RoundTripper {
	insecure := flag.Bool("insecure", false, "skip certificate verification")
	enableQlog := flag.Bool("qlog", false, "output a qlog (in the same directory)")
	flag.Parse()

	pool, err := x509.SystemCertPool()
	if err != nil {
		utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedInit)
	}
	tlsconfig.AddRootCA(pool)

	var qconf quic.Config
	if *enableQlog {
		qconf.Tracer = func(
			ctx context.Context,
			p logging.Perspective,
			connID quic.ConnectionID,
		) *logging.ConnectionTracer {
			filename := fmt.Sprintf("client_%s.qlog", connID)
			f, err := os.Create(filename)
			if err != nil {
				utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedInit)
			}
			log.Printf("Creating qlog file %s.\n", filename)
			return qlog.NewConnectionTracer(utils.NewBufferedWriteCloser(bufio.NewWriter(f), f), p, connID)
		}
	}

	return &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: *insecure,
		},
		QuicConfig: &qconf,
	}
}
