package tlscreds

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/grpc/credentials"
)

// Set via linker parameters
var caCertPEM []byte

// This loads the tls identity credentials from a part of files with the prefix identityPath
func LoadCredentials(identityPath string) (credentials.TransportCredentials, error) {
	certFilePath := identityPath + "-cert.pem"
	keyFilePath := identityPath + "-key.pem"

	slog.Debug("LoadCredentials",
		"certFilePath", certFilePath,
		"keyFilePath", keyFilePath,
	)

	cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
	if err != nil {
		return nil, fmt.Errorf("loadX509KeyPair: %w", err)
	}

	// TODO: This is only used for clients, should probably add a flag to LoadCredentials
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("x509.ParseCertificate-Leaf: %w", err)
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(caCertPEM) {
		return nil, fmt.Errorf("appendCertsFromPEM: %w", err)
	}

	// Note: some of these fields are only used by either the client or the server
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    capool,
		RootCAs:      capool,
		// TODO: Allowing override of ServerName may be convienient
		// Bundled test cert has the DNS name: 127.0.0.1, which means
		// that you need to use that as your name in the client Dial
		// options, but this is probably fine for now... the connection
		// string is hardcoded anyway
	}

	if keyLogPath := os.Getenv("INSECURE_DEBUG_KEYLOGGING_PATH"); keyLogPath != "" {
		slog.Warn("INSECURE_DEBUG_KEYLOGGING_PATH is set. Security is compromised")
		keyLogFile, err := os.Create(keyLogPath)
		if err != nil {
			return nil, fmt.Errorf("keylog create file: %w", err)
		}
		tlsConfig.KeyLogWriter = keyLogFile

	}

	return credentials.NewTLS(tlsConfig), nil
}
