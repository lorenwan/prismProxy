package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/perf"
	pb "prismproxy/proto/gen/go"
)

// PerfServiceImpl PerfService gRPC 实现
type PerfServiceImpl struct {
	pb.UnimplementedPerfServiceServer
	analyzer *perf.PerfAnalyzer
}

// RegisterPerfServiceImpl 注册 PerfService
func RegisterPerfServiceImpl(s *grpc.Server, analyzer *perf.PerfAnalyzer) {
	pb.RegisterPerfServiceServer(s, &PerfServiceImpl{analyzer: analyzer})
}

// GetStats 获取性能统计
func (s *PerfServiceImpl) GetStats(ctx context.Context, req *pb.PerfStatsRequest) (*pb.PerfStats, error) {
	since := time.Time{}
	if req.GetSince() != "" {
		var err error
		since, err = time.Parse(time.RFC3339, req.GetSince())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "时间格式错误: %v", err)
		}
	} else {
		since = time.Now().Add(-24 * time.Hour)
	}

	stats, err := s.analyzer.GetStats(since)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取性能统计失败: %v", err)
	}

	return perfStatsToProto(stats), nil
}

// GetSlowRequests 获取慢请求
func (s *PerfServiceImpl) GetSlowRequests(ctx context.Context, req *pb.GetSlowRequestsRequest) (*pb.SlowRequestListResponse, error) {
	threshold := time.Duration(req.GetThresholdMs()) * time.Millisecond
	if threshold == 0 {
		threshold = 1000 * time.Millisecond
	}
	limit := int(req.GetLimit())
	if limit == 0 {
		limit = 50
	}

	requests, err := s.analyzer.GetSlowRequests(threshold, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取慢请求失败: %v", err)
	}

	items := make([]*pb.SlowRequest, len(requests))
	for i, r := range requests {
		items[i] = slowRequestToProto(r)
	}

	return &pb.SlowRequestListResponse{Requests: items}, nil
}

// GetDomainStats 获取域名统计
func (s *PerfServiceImpl) GetDomainStats(ctx context.Context, req *pb.Empty) (*pb.DomainStatsListResponse, error) {
	stats, err := s.analyzer.GetDomainStats()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取域名统计失败: %v", err)
	}

	items := make([]*pb.DomainStats, len(stats))
	for i, ds := range stats {
		items[i] = domainStatsToProto(ds)
	}

	return &pb.DomainStatsListResponse{Stats: items}, nil
}

// GetTimeline 获取时间线数据
func (s *PerfServiceImpl) GetTimeline(ctx context.Context, req *pb.GetTimelineRequest) (*pb.TimelineResponse, error) {
	since := time.Now().Add(-24 * time.Hour)
	if req.GetSince() != "" {
		var err error
		since, err = time.Parse(time.RFC3339, req.GetSince())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "时间格式错误: %v", err)
		}
	}

	interval := time.Duration(req.GetIntervalSeconds()) * time.Second
	if interval == 0 {
		interval = time.Hour
	}

	points, err := s.analyzer.GetTimelineStats(since, interval)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取时间线失败: %v", err)
	}

	items := make([]*pb.TimelinePoint, len(points))
	for i, p := range points {
		items[i] = timelinePointToProto(p)
	}

	return &pb.TimelineResponse{Points: items}, nil
}

// GetStatusCodeStats 获取状态码统计
func (s *PerfServiceImpl) GetStatusCodeStats(ctx context.Context, req *pb.PerfStatsRequest) (*pb.StatusCodeStatsResponse, error) {
	since := time.Now().Add(-24 * time.Hour)
	if req.GetSince() != "" {
		var err error
		since, err = time.Parse(time.RFC3339, req.GetSince())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "时间格式错误: %v", err)
		}
	}

	stats, err := s.analyzer.GetStatusCodeStats(since)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取状态码统计失败: %v", err)
	}

	items := make([]*pb.StatusCodeStats, len(stats))
	for i, ss := range stats {
		items[i] = &pb.StatusCodeStats{
			StatusCode: int32(ss.StatusCode),
			Count:      ss.Count,
		}
	}

	return &pb.StatusCodeStatsResponse{Stats: items}, nil
}

// GetMethodStats 获取请求方法统计
func (s *PerfServiceImpl) GetMethodStats(ctx context.Context, req *pb.PerfStatsRequest) (*pb.MethodStatsResponse, error) {
	since := time.Now().Add(-24 * time.Hour)
	if req.GetSince() != "" {
		var err error
		since, err = time.Parse(time.RFC3339, req.GetSince())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "时间格式错误: %v", err)
		}
	}

	stats, err := s.analyzer.GetMethodStats(since)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取方法统计失败: %v", err)
	}

	items := make([]*pb.MethodStats, len(stats))
	for i, ms := range stats {
		items[i] = &pb.MethodStats{
			Method: ms.Method,
			Count:  ms.Count,
		}
	}

	return &pb.MethodStatsResponse{Stats: items}, nil
}

// GetRecentStats 获取最近 N 分钟统计
func (s *PerfServiceImpl) GetRecentStats(ctx context.Context, req *pb.GetRecentStatsRequest) (*pb.PerfStats, error) {
	minutes := int(req.GetMinutes())
	if minutes <= 0 {
		minutes = 60
	}

	stats, err := s.analyzer.GetRecentStats(minutes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取最近统计失败: %v", err)
	}

	return perfStatsToProto(stats), nil
}

// === proto ↔ Go 转换函数 ===

func perfStatsToProto(stats *perf.PerfStats) *pb.PerfStats {
	if stats == nil {
		return nil
	}

	return &pb.PerfStats{
		TotalRequests:   stats.TotalRequests,
		AvgDurationMs:   stats.AvgDuration,
		P50Ms:           stats.P50,
		P90Ms:           stats.P90,
		P99Ms:           stats.P99,
		SlowRequests:    stats.SlowRequests,
		MinDurationMs:   stats.MinDuration,
		MaxDurationMs:   stats.MaxDuration,
		TotalDurationMs: stats.TotalDuration,
	}
}

func slowRequestToProto(r perf.SlowRequest) *pb.SlowRequest {
	return &pb.SlowRequest{
		TransactionId: r.TransactionID,
		Url:           r.URL,
		Method:        r.Method,
		Host:          r.Host,
		StatusCode:    int32(r.StatusCode),
		DurationMs:    r.Duration,
		Timestamp:     r.Timestamp.Format(time.RFC3339),
	}
}

func domainStatsToProto(ds perf.DomainStats) *pb.DomainStats {
	return &pb.DomainStats{
		Domain:          ds.Domain,
		RequestCount:    ds.RequestCount,
		AvgDurationMs:   ds.AvgDuration,
		ErrorCount:      ds.ErrorCount,
		ErrorRate:       ds.ErrorRate,
		TotalDurationMs: ds.TotalDuration,
	}
}

func timelinePointToProto(p perf.TimelinePoint) *pb.TimelinePoint {
	return &pb.TimelinePoint{
		Timestamp:     p.Timestamp.Format(time.RFC3339),
		RequestCount:  p.RequestCount,
		AvgDurationMs: p.AvgDuration,
		ErrorCount:    p.ErrorCount,
	}
}

// 确保 PerfServiceImpl 实现了 pb.PerfServiceServer
var _ pb.PerfServiceServer = (*PerfServiceImpl)(nil)
