package auth

import (
	"fmt"
	"io"
	"net/http"
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

	return &Client{
		baseURL: strings.TrimRight(
			baseURL,
			"/",
		),

		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// =====================================
// VALIDATE TOKEN
// =====================================

func (c *Client) ValidateToken(
	token string,
) (*ValidationResult, error) {

	token =
		strings.TrimSpace(
			token,
		)

	if token == "" {
		return nil, fmt.Errorf(
			"el token es obligatorio",
		)
	}

	endpoint :=
		c.baseURL +
			"/internal/auth/validate"

	request,
		err :=
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

	response,
		err :=
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

	body,
		err :=
		io.ReadAll(
			response.Body,
		)

	if err != nil {
		return nil, fmt.Errorf(
			"error leyendo respuesta de Auth Service: %w",
			err,
		)
	}

	if response.StatusCode !=
		http.StatusOK {

		return nil, fmt.Errorf(
			"Auth Service respondió HTTP %d: %s",
			response.StatusCode,
			string(body),
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
		return nil, fmt.Errorf(
			"respuesta inválida del Auth Service",
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
