package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(
	authClient *Client,
) grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md,
			ok :=
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

		result,
			err :=
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

					Role: result.Role,
				},
			)

		return handler(
			ctx,
			req,
		)
	}
}
