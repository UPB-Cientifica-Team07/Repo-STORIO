package monitoring

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

type RuntimeMetrics struct {
	totalRequests     int64
	activeConnections int64
}

func NewRuntimeMetrics() *RuntimeMetrics {
	return &RuntimeMetrics{}
}

func (m *RuntimeMetrics) TotalRequests() int64 {
	return atomic.LoadInt64(
		&m.totalRequests,
	)
}

func (m *RuntimeMetrics) ActiveConnections() int64 {
	return atomic.LoadInt64(
		&m.activeConnections,
	)
}

// =====================================
// UNARY RPC
// =====================================

func (m *RuntimeMetrics) UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	atomic.AddInt64(
		&m.totalRequests,
		1,
	)

	return handler(
		ctx,
		req,
	)
}

// =====================================
// STREAM RPC
// =====================================

func (m *RuntimeMetrics) StreamInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {

	atomic.AddInt64(
		&m.totalRequests,
		1,
	)

	return handler(
		srv,
		stream,
	)
}

// =====================================
// GRPC CONNECTION STATS
// =====================================

type ConnectionStatsHandler struct {
	metrics *RuntimeMetrics
}

func NewConnectionStatsHandler(
	metrics *RuntimeMetrics,
) *ConnectionStatsHandler {

	return &ConnectionStatsHandler{
		metrics: metrics,
	}
}

func (h *ConnectionStatsHandler) TagRPC(
	ctx context.Context,
	info *stats.RPCTagInfo,
) context.Context {

	return ctx
}

func (h *ConnectionStatsHandler) HandleRPC(
	ctx context.Context,
	stat stats.RPCStats,
) {
}

func (h *ConnectionStatsHandler) TagConn(
	ctx context.Context,
	info *stats.ConnTagInfo,
) context.Context {

	return ctx
}

func (h *ConnectionStatsHandler) HandleConn(
	ctx context.Context,
	stat stats.ConnStats,
) {

	switch stat.(type) {

	case *stats.ConnBegin:

		atomic.AddInt64(
			&h.metrics.activeConnections,
			1,
		)

	case *stats.ConnEnd:

		atomic.AddInt64(
			&h.metrics.activeConnections,
			-1,
		)
	}
}
