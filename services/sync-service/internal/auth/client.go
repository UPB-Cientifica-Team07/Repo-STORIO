package auth

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type ValidationResult struct {
	Valid   bool
	UserID  string
	Role    string
	Message string
}

type LoginResult struct {
	Success bool
	UserID  string
	Role    string
	Token   string
	Message string
}

type Client struct {
	baseURL    string
	httpClient *http.Client
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

// =====================================
// LOGIN
// =====================================

func (c *Client) Login(
	username string,
	password string,
) (*LoginResult, error) {

	if username == "" {
		return nil, errors.New(
			"el username es obligatorio",
		)
	}

	if password == "" {
		return nil, errors.New(
			"el password es obligatorio",
		)
	}

	endpoint :=
		c.baseURL +
			"/internal/auth/login"

	form :=
		url.Values{}

	form.Set(
		"username",
		username,
	)

	form.Set(
		"password",
		password,
	)

	request, err :=
		http.NewRequest(
			http.MethodPost,
			endpoint,
			strings.NewReader(
				form.Encode(),
			),
		)

	if err != nil {
		return nil, fmt.Errorf(
			"error creando petición de login: %w",
			err,
		)
	}

	request.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	response, err :=
		c.httpClient.Do(
			request,
		)

	if err != nil {
		return nil, fmt.Errorf(
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
		return nil, fmt.Errorf(
			"error leyendo respuesta de Auth Service: %w",
			err,
		)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"Auth Service respondió HTTP %d: %s",
			response.StatusCode,
			string(body),
		)
	}

	parts :=
		strings.SplitN(
			string(body),
			"|",
			5,
		)

	if len(parts) != 5 {
		return nil, errors.New(
			"respuesta inválida del Auth Service",
		)
	}

	return &LoginResult{
		Success: parts[0] == "true",
		UserID:  parts[1],
		Role:    parts[2],
		Token:   parts[3],
		Message: parts[4],
	}, nil
}

// =====================================
// VALIDATE TOKEN
// =====================================

func (c *Client) ValidateToken(
	token string,
) (*ValidationResult, error) {

	if token == "" {
		return nil, errors.New(
			"el token es obligatorio",
		)
	}

	endpoint :=
		c.baseURL +
			"/internal/auth/validate"

	request, err :=
		http.NewRequest(
			http.MethodGet,
			endpoint,
			nil,
		)

	if err != nil {
		return nil, fmt.Errorf(
			"error creando petición de validación: %w",
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
		return nil, fmt.Errorf(
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
		return nil, fmt.Errorf(
			"error leyendo respuesta de Auth Service: %w",
			err,
		)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"Auth Service respondió HTTP %d: %s",
			response.StatusCode,
			string(body),
		)
	}

	parts :=
		strings.SplitN(
			string(body),
			"|",
			4,
		)

	if len(parts) != 4 {
		return nil, errors.New(
			"respuesta inválida del Auth Service",
		)
	}

	return &ValidationResult{
		Valid:   parts[0] == "true",
		UserID:  parts[1],
		Role:    parts[2],
		Message: parts[3],
	}, nil
}
