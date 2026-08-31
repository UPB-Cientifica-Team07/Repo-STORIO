package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authenticateMethod = "/sync.SyncService/Authenticate"

// =====================================
// VALIDAR IDENTIDAD
// =====================================

func authenticateContext(
	ctx context.Context,
	authClient *Client,
) (context.Context, error) {

	if authClient == nil {
		return nil,
			status.Error(
				codes.Internal,
				"Auth Client no inicializado",
			)
	}

	md, ok :=
		metadata.FromIncomingContext(
			ctx,
		)

	if !ok {
		return nil,
			status.Error(
				codes.Unauthenticated,
				"metadata de autenticación requerida",
			)
	}

	values :=
		md.Get(
			"authorization",
		)

	if len(values) == 0 {
		return nil,
			status.Error(
				codes.Unauthenticated,
				"token de autenticación requerido",
			)
	}

	authorization :=
		strings.TrimSpace(
			values[0],
		)

	const prefix = "Bearer "

	if !strings.HasPrefix(
		authorization,
		prefix,
	) {
		return nil,
			status.Error(
				codes.Unauthenticated,
				"formato Authorization inválido",
			)
	}

	token :=
		strings.TrimSpace(
			strings.TrimPrefix(
				authorization,
				prefix,
			),
		)

	if token == "" {
		return nil,
			status.Error(
				codes.Unauthenticated,
				"token de autenticación requerido",
			)
	}

	result, err :=
		authClient.ValidateToken(
			token,
		)

	if err != nil {
		return nil,
			status.Errorf(
				codes.Unavailable,
				"no fue posible validar token: %v",
				err,
			)
	}

	if !result.Valid ||
		result.UserID == "" {

		return nil,
			status.Error(
				codes.Unauthenticated,
				"token inválido",
			)
	}

	ctx =
		WithIdentity(
			ctx,
			Identity{
				UserID: result.UserID,
				Role: strings.ToUpper(
					strings.TrimSpace(
						result.Role,
					),
				),
			},
		)

	return ctx, nil
}

// =====================================
// UNARY INTERCEPTOR
// =====================================

func UnaryServerInterceptor(
	authClient *Client,
) grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		// Authenticate mantiene compatibilidad:
		// recibe token dentro del request.
		if info.FullMethod ==
			authenticateMethod {

			return handler(
				ctx,
				req,
			)
		}

		authenticatedCtx, err :=
			authenticateContext(
				ctx,
				authClient,
			)

		if err != nil {
			return nil, err
		}

		return handler(
			authenticatedCtx,
			req,
		)
	}
}

// =====================================
// STREAM CONTEXT WRAPPER
// =====================================

type authenticatedServerStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (s *authenticatedServerStream) Context() context.Context {
	return s.ctx
}

// =====================================
// STREAM INTERCEPTOR
// =====================================

func StreamServerInterceptor(
	authClient *Client,
) grpc.StreamServerInterceptor {

	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		authenticatedCtx, err :=
			authenticateContext(
				stream.Context(),
				authClient,
			)

		if err != nil {
			return err
		}

		wrapped :=
			&authenticatedServerStream{
				ServerStream: stream,
				ctx:          authenticatedCtx,
			}

		return handler(
			srv,
			wrapped,
		)
	}
}
