package security

import (
	"context"
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	serviceTokenHeader = "x-monitoring-service-token"
)

type Identity struct {
	UserID string
	Role   string
}

type Interceptor struct {
	serviceToken string
	authURL      string
	httpClient   *http.Client
}

func NewInterceptor() (*Interceptor, error) {
	serviceToken :=
		strings.TrimSpace(
			os.Getenv(
				"MONITORING_SERVICE_TOKEN",
			),
		)

	if serviceToken == "" {
		return nil,
			fmt.Errorf(
				"MONITORING_SERVICE_TOKEN es obligatorio",
			)
	}

	authURL :=
		strings.TrimSpace(
			os.Getenv(
				"MONITORING_AUTH_SERVICE",
			),
		)

	if authURL == "" {
		authURL =
			"http://127.0.0.1:8081"
	}

	return &Interceptor{
		serviceToken: serviceToken,
		authURL:      strings.TrimRight(authURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (i *Interceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		switch {
		case strings.HasSuffix(
			info.FullMethod,
			"/ReportMetrics",
		),
			strings.HasSuffix(
				info.FullMethod,
				"/ReportStatus",
			):

			if err :=
				i.validateServiceToken(
					ctx,
				); err != nil {

				return nil, err
			}

		default:

			identity, err :=
				i.validateUserToken(
					ctx,
				)

			if err != nil {
				return nil, err
			}

			if isAdminMethod(
				info.FullMethod,
			) &&
				identity.Role != "ADMIN" {

				return nil,
					status.Error(
						codes.PermissionDenied,
						"rol ADMIN requerido",
					)
			}
		}

		return handler(
			ctx,
			req,
		)
	}
}

func (i *Interceptor) validateServiceToken(
	ctx context.Context,
) error {

	md, ok :=
		metadata.FromIncomingContext(
			ctx,
		)

	if !ok {
		return status.Error(
			codes.Unauthenticated,
			"service token requerido",
		)
	}

	values :=
		md.Get(
			serviceTokenHeader,
		)

	if len(values) != 1 {
		return status.Error(
			codes.Unauthenticated,
			"service token requerido",
		)
	}

	provided :=
		[]byte(
			values[0],
		)

	expected :=
		[]byte(
			i.serviceToken,
		)

	if len(provided) != len(expected) ||
		subtle.ConstantTimeCompare(
			provided,
			expected,
		) != 1 {

		return status.Error(
			codes.Unauthenticated,
			"service token inválido",
		)
	}

	return nil
}

func (i *Interceptor) validateUserToken(
	ctx context.Context,
) (*Identity, error) {

	md, ok :=
		metadata.FromIncomingContext(
			ctx,
		)

	if !ok {
		return nil,
			status.Error(
				codes.Unauthenticated,
				"Bearer token requerido",
			)
	}

	values :=
		md.Get(
			"authorization",
		)

	if len(values) != 1 {
		return nil,
			status.Error(
				codes.Unauthenticated,
				"Bearer token requerido",
			)
	}

	authorization :=
		strings.TrimSpace(
			values[0],
		)

	if !strings.HasPrefix(
		authorization,
		"Bearer ",
	) {
		return nil,
			status.Error(
				codes.Unauthenticated,
				"Bearer token inválido",
			)
	}

	token :=
		strings.TrimSpace(
			strings.TrimPrefix(
				authorization,
				"Bearer ",
			),
		)

	if token == "" {
		return nil,
			status.Error(
				codes.Unauthenticated,
				"Bearer token requerido",
			)
	}

	request, err :=
		http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			i.authURL+
				"/internal/auth/validate",
			nil,
		)

	if err != nil {
		return nil,
			status.Error(
				codes.Internal,
				"no fue posible crear solicitud de autenticación",
			)
	}

	request.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	response, err :=
		i.httpClient.Do(
			request,
		)

	if err != nil {
		return nil,
			status.Error(
				codes.Unavailable,
				"Auth Service no disponible",
			)
	}

	defer response.Body.Close()

	body, err :=
		io.ReadAll(
			response.Body,
		)

	if err != nil {
		return nil,
			status.Error(
				codes.Internal,
				"respuesta Auth inválida",
			)
	}

	parts :=
		strings.Split(
			strings.TrimSpace(
				string(body),
			),
			"|",
		)

	if response.StatusCode != http.StatusOK ||
		len(parts) < 4 ||
		parts[0] != "true" {

		return nil,
			status.Error(
				codes.Unauthenticated,
				"token inválido o expirado",
			)
	}

	return &Identity{
		UserID: parts[1],
		Role:   parts[2],
	}, nil
}

func isAdminMethod(
	method string,
) bool {

	return strings.HasSuffix(
		method,
		"/CreateAlertRule",
	) ||
		strings.HasSuffix(
			method,
			"/UpdateAlertRule",
		) ||
		strings.HasSuffix(
			method,
			"/DeleteAlertRule",
		)
}
