package auth

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type ValidationResult struct {
	Valid   bool
	UserID  string
	Role    string
	Message string
}

func NewClient(
	baseURL string,
) *Client {

	caFile :=
		os.Getenv(
			"AUTH_TLS_CA_FILE",
		)

	if caFile == "" {
		caFile =
			"security/pki/upb_dev_ca.crt"
	}

	caPEM, err :=
		os.ReadFile(
			caFile,
		)

	if err != nil {
		panic(
			"no se pudo leer CA de Auth Service: " +
				err.Error(),
		)
	}

	certPool :=
		x509.NewCertPool()

	if !certPool.AppendCertsFromPEM(
		caPEM,
	) {
		panic(
			"CA de Auth Service inválida",
		)
	}

	transport :=
		&http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,

				MinVersion: tls.VersionTLS12,
			},
		}

	return &Client{
		baseURL: strings.TrimRight(
			baseURL,
			"/",
		),

		httpClient: &http.Client{
			Timeout: 5 * time.Second,

			Transport: transport,
		},
	}
}

func (c *Client) ValidateToken(
	token string,
) (*ValidationResult, error) {

	token =
		strings.TrimSpace(
			token,
		)

	if token == "" {
		return nil,
			fmt.Errorf(
				"el token es obligatorio",
			)
	}

	request, err :=
		http.NewRequest(
			http.MethodGet,
			c.baseURL+
				"/internal/auth/validate",
			nil,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"error creando solicitud Auth: %w",
				err,
			)
	}

	request.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	response, err :=
		c.httpClient.Do(
			request,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"error conectando con Auth Service: %w",
				err,
			)
	}

	defer response.Body.Close()

	body, err :=
		io.ReadAll(
			response.Body,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"error leyendo respuesta Auth: %w",
				err,
			)
	}

	if response.StatusCode !=
		http.StatusOK {

		return nil,
			fmt.Errorf(
				"Auth Service respondió HTTP %d",
				response.StatusCode,
			)
	}

	parts :=
		strings.SplitN(
			strings.TrimSpace(
				string(body),
			),
			"|",
			4,
		)

	if len(parts) != 4 {
		return nil,
			fmt.Errorf(
				"respuesta inválida de Auth Service",
			)
	}

	return &ValidationResult{
		Valid: strings.EqualFold(
			parts[0],
			"true",
		),

		UserID: strings.TrimSpace(
			parts[1],
		),

		Role: strings.ToUpper(
			strings.TrimSpace(
				parts[2],
			),
		),

		Message: strings.TrimSpace(
			parts[3],
		),
	}, nil
}
