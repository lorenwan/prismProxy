package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/traffic"
	pb "prismproxy/proto/gen/go"
)

// TrafficServiceImpl TrafficService gRPC 实现
type TrafficServiceImpl struct {
	pb.UnimplementedTrafficServiceServer
	manager *traffic.Manager
}

// RegisterTrafficServiceImpl 注册 TrafficService
func RegisterTrafficServiceImpl(s *grpc.Server, mgr *traffic.Manager) {
	pb.RegisterTrafficServiceServer(s, &TrafficServiceImpl{manager: mgr})
}

// List 获取流量列表
func (s *TrafficServiceImpl) List(ctx context.Context, req *pb.TrafficListRequest) (*pb.TrafficListResponse, error) {
	// 解析分页参数
	page := int32(1)
	pageSize := int32(20)
	if req.GetPagination() != nil {
		if req.GetPagination().GetPage() > 0 {
			page = req.GetPagination().GetPage()
		}
		if req.GetPagination().GetPageSize() > 0 {
			pageSize = req.GetPagination().GetPageSize()
		}
	}

	limit := int(pageSize)
	offset := int((page - 1) * pageSize)

	// 使用过滤器查询
	filter := protoToFilter(req.GetFilter())
	transactions, total, err := s.manager.ListWithFilter(filter, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询流量失败: %v", err)
	}

	// 转换为 proto 格式
	items := make([]*pb.TrafficEntry, len(transactions))
	for i, tx := range transactions {
		items[i] = trafficToProto(tx)
	}

	return &pb.TrafficListResponse{
		Items: items,
		Meta: &pb.PageMeta{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

// Get 获取单条流量详情
func (s *TrafficServiceImpl) Get(ctx context.Context, req *pb.TrafficGetRequest) (*pb.TrafficEntry, error) {
	tx, err := s.manager.GetTransaction(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询流量失败: %v", err)
	}
	if tx == nil {
		return nil, status.Errorf(codes.NotFound, "流量记录不存在: %d", req.GetId())
	}

	return trafficToProto(tx), nil
}

// Delete 删除流量记录
func (s *TrafficServiceImpl) Delete(ctx context.Context, req *pb.TrafficDeleteRequest) (*pb.Empty, error) {
	for _, id := range req.GetIds() {
		if err := s.manager.DeleteTransaction(id); err != nil {
			return nil, status.Errorf(codes.Internal, "删除流量 %d 失败: %v", id, err)
		}
	}

	return &pb.Empty{}, nil
}

// Clear 清空所有流量
func (s *TrafficServiceImpl) Clear(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	if err := s.manager.ClearTransactions(); err != nil {
		return nil, status.Errorf(codes.Internal, "清空流量失败: %v", err)
	}

	return &pb.Empty{}, nil
}

// Stats 获取流量统计
func (s *TrafficServiceImpl) Stats(ctx context.Context, req *pb.TrafficStatsRequest) (*pb.TrafficStats, error) {
	stats, err := s.manager.GetStats()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取统计失败: %v", err)
	}

	return statsToProto(stats), nil
}

// Subscribe 订阅流量事件 (服务端流)
func (s *TrafficServiceImpl) Subscribe(req *pb.Empty, stream pb.TrafficService_SubscribeServer) error {
	// TODO: 实现流量事件订阅，需要 Manager 支持事件通知机制
	// 暂时阻塞等待客户端断开
	<-stream.Context().Done()
	return stream.Context().Err()
}

// === proto ↔ Go 转换函数 ===

// trafficToProto 将 traffic.Transaction 转换为 pb.TrafficEntry
func trafficToProto(tx *traffic.Transaction) *pb.TrafficEntry {
	if tx == nil {
		return nil
	}

	entry := &pb.TrafficEntry{
		Id:         tx.ID,
		Timestamp:  tx.Timestamp.Format(time.RFC3339),
		DurationMs: tx.DurationMs,
		Method:     tx.Method,
		Url:        tx.URL,
		Host:       tx.Host,
		Path:       tx.Path,
		Scheme:     tx.Scheme,
		Port:       tx.Port,
		ClientAddr: tx.ClientAddr,
		ServerIp:   tx.ServerIP,
		Bookmarked: tx.Bookmarked,
		Color:      tx.Color,
		Notes:      tx.Notes,
		Tags:       tx.Tags,
	}

	// 请求数据
	if tx.Request != nil {
		entry.Request = &pb.RequestData{
			Headers:     headersToProto(tx.Request.Headers),
			Body:        tx.Request.Body,
			BodySize:    tx.Request.BodySize,
			ContentType: tx.Request.ContentType,
			Raw:         tx.Request.Raw,
		}
	}

	// 响应数据
	if tx.Response != nil {
		entry.Response = &pb.ResponseData{
			StatusCode:  int32(tx.Response.StatusCode),
			StatusText:  tx.Response.StatusText,
			Headers:     headersToProto(tx.Response.Headers),
			Body:        tx.Response.Body,
			BodySize:    tx.Response.BodySize,
			ContentType: tx.Response.ContentType,
			Raw:         tx.Response.Raw,
		}
	}

	return entry
}

// headersToProto 将 http.Header 转换为 proto 格式
func headersToProto(headers http.Header) map[string]*pb.StringList {
	if headers == nil {
		return nil
	}

	result := make(map[string]*pb.StringList, len(headers))
	for k, v := range headers {
		result[k] = &pb.StringList{Values: v}
	}
	return result
}

// protoToFilter 将 pb.TrafficFilter 转换为 traffic.Filter
func protoToFilter(f *pb.TrafficFilter) *traffic.Filter {
	if f == nil {
		return &traffic.Filter{}
	}

	filter := &traffic.Filter{
		Method:      f.GetMethod(),
		Host:        f.GetHost(),
		Path:        f.GetPath(),
		ContentType: f.GetContentType(),
		MinDuration: f.GetMinDuration(),
		MaxDuration: f.GetMaxDuration(),
		Search:      f.GetSearch(),
		Tags:        f.GetTags(),
	}

	// 状态码转换 int32 -> int
	if len(f.GetStatusCode()) > 0 {
		statusCodes := make([]int, len(f.GetStatusCode()))
		for i, sc := range f.GetStatusCode() {
			statusCodes[i] = int(sc)
		}
		filter.StatusCode = statusCodes
	}

	// 书签
	if f.Bookmarked != nil {
		b := f.GetBookmarked()
		filter.Bookmarked = &b
	}

	// 时间范围
	if f.GetTimeRange() != nil {
		tr := &traffic.TimeRange{}
		if f.GetTimeRange().GetStart() != "" {
			tr.Start, _ = time.Parse(time.RFC3339, f.GetTimeRange().GetStart())
		}
		if f.GetTimeRange().GetEnd() != "" {
			tr.End, _ = time.Parse(time.RFC3339, f.GetTimeRange().GetEnd())
		}
		filter.TimeRange = tr
	}

	return filter
}

// statsToProto 将 traffic.TrafficStats 转换为 pb.TrafficStats
func statsToProto(stats *traffic.TrafficStats) *pb.TrafficStats {
	if stats == nil {
		return nil
	}

	result := &pb.TrafficStats{
		TotalRequests:  stats.TotalRequests,
		TotalResponses: stats.TotalResponses,
		AvgDurationMs:  stats.AvgDuration,
		MaxDurationMs:  stats.MaxDuration,
		MinDurationMs:  stats.MinDuration,
		ErrorCount:     stats.ErrorCount,
		SuccessCount:   stats.SuccessCount,
	}

	// 主机统计
	if len(stats.HostStats) > 0 {
		result.HostStats = make([]*pb.HostStat, len(stats.HostStats))
		for i, hs := range stats.HostStats {
			result.HostStats[i] = &pb.HostStat{
				Host:      hs.Host,
				Count:     hs.Count,
				AvgTimeMs: hs.AvgTime,
			}
		}
	}

	// 方法统计
	if len(stats.MethodStats) > 0 {
		result.MethodStats = make([]*pb.MethodStat, len(stats.MethodStats))
		for i, ms := range stats.MethodStats {
			result.MethodStats[i] = &pb.MethodStat{
				Method: ms.Method,
				Count:  ms.Count,
			}
		}
	}

	// 状态码统计
	if len(stats.StatusStats) > 0 {
		result.StatusStats = make([]*pb.StatusStat, len(stats.StatusStats))
		for i, ss := range stats.StatusStats {
			result.StatusStats[i] = &pb.StatusStat{
				StatusCode: int32(ss.StatusCode),
				Count:      ss.Count,
			}
		}
	}

	return result
}

// TODO: 需要在 traffic.Manager 中添加事件订阅机制
// 以下是事件转换的辅助函数，供后续实现使用

// trafficEventToProto 将流量事件转换为 proto 格式
func trafficEventToProto(eventType string, tx *traffic.Transaction) *pb.TrafficEvent {
	return &pb.TrafficEvent{
		Type:      eventType,
		Entry:     trafficToProto(tx),
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// protoToTraffic 将 pb.TrafficEntry 转换为 traffic.Transaction (用于调试器修改后回写)
func protoToTraffic(entry *pb.TrafficEntry) *traffic.Transaction {
	if entry == nil {
		return nil
	}

	tx := &traffic.Transaction{
		ID:         entry.GetId(),
		DurationMs: entry.GetDurationMs(),
		Method:     entry.GetMethod(),
		URL:        entry.GetUrl(),
		Host:       entry.GetHost(),
		Path:       entry.GetPath(),
		Scheme:     entry.GetScheme(),
		Port:       entry.GetPort(),
		ClientAddr: entry.GetClientAddr(),
		ServerIP:   entry.GetServerIp(),
		Bookmarked: entry.GetBookmarked(),
		Color:      entry.GetColor(),
		Notes:      entry.GetNotes(),
		Tags:       entry.GetTags(),
	}

	// 解析时间
	if entry.GetTimestamp() != "" {
		tx.Timestamp, _ = time.Parse(time.RFC3339, entry.GetTimestamp())
	}

	// 请求数据
	if entry.GetRequest() != nil {
		tx.Request = &traffic.RequestData{
			Headers:     protoToHeaders(entry.GetRequest().GetHeaders()),
			Body:        entry.GetRequest().GetBody(),
			BodySize:    entry.GetRequest().GetBodySize(),
			ContentType: entry.GetRequest().GetContentType(),
			Raw:         entry.GetRequest().GetRaw(),
		}
	}

	// 响应数据
	if entry.GetResponse() != nil {
		tx.Response = &traffic.ResponseData{
			StatusCode:  int(entry.GetResponse().GetStatusCode()),
			StatusText:  entry.GetResponse().GetStatusText(),
			Headers:     protoToHeaders(entry.GetResponse().GetHeaders()),
			Body:        entry.GetResponse().GetBody(),
			BodySize:    entry.GetResponse().GetBodySize(),
			ContentType: entry.GetResponse().GetContentType(),
			Raw:         entry.GetResponse().GetRaw(),
		}
	}

	return tx
}

// protoToHeaders 将 proto 格式 headers 转换为 http.Header
func protoToHeaders(headers map[string]*pb.StringList) http.Header {
	if headers == nil {
		return nil
	}

	result := make(http.Header, len(headers))
	for k, v := range headers {
		if v != nil {
			result[k] = v.GetValues()
		}
	}
	return result
}

// marshalJSON 辅助函数，序列化为 JSON 字符串
func marshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("[WARN] JSON 序列化失败: %v", err)
		return "{}"
	}
	return string(data)
}

// unmarshalJSON 辅助函数，反序列化 JSON 字符串
func unmarshalJSON(s string, v interface{}) error {
	if s == "" {
		return nil
	}
	return json.Unmarshal([]byte(s), v)
}

// ensure TrafficServiceImpl implements pb.TrafficServiceServer
var _ pb.TrafficServiceServer = (*TrafficServiceImpl)(nil)

// ensure unused imports are used
var _ = fmt.Sprintf
var _ = marshalJSON
var _ = unmarshalJSON
