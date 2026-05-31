package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/search"
	pb "prismproxy/proto/gen/go"
)

// SearchServiceImpl SearchService gRPC 实现
type SearchServiceImpl struct {
	pb.UnimplementedSearchServiceServer
	engine      *search.SearchEngine
	filterStore *search.FilterStore
}

// RegisterSearchServiceImpl 注册 SearchService
func RegisterSearchServiceImpl(s *grpc.Server, engine *search.SearchEngine, filterStore *search.FilterStore) {
	pb.RegisterSearchServiceServer(s, &SearchServiceImpl{engine: engine, filterStore: filterStore})
}

// Search 全文搜索
func (s *SearchServiceImpl) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResultProto, error) {
	page, pageSize := paginationToPage(req.GetPagination())

	query := &search.SearchQuery{
		Query:    req.GetQuery(),
		Filters:  protoToFilters(req.GetFilters()),
		Sort:     req.GetSort(),
		Page:     page,
		PageSize: pageSize,
	}

	result, err := s.engine.Search(query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "搜索失败: %v", err)
	}

	return searchResultToProto(result), nil
}

// SearchByMethod 按方法搜索
func (s *SearchServiceImpl) SearchByMethod(ctx context.Context, req *pb.SearchByMethodRequest) (*pb.SearchResultProto, error) {
	page, pageSize := paginationToPage(req.GetPagination())

	result, err := s.engine.SearchByMethod(req.GetMethod(), page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "搜索失败: %v", err)
	}

	return searchResultToProto(result), nil
}

// SearchByHost 按主机搜索
func (s *SearchServiceImpl) SearchByHost(ctx context.Context, req *pb.SearchByHostRequest) (*pb.SearchResultProto, error) {
	page, pageSize := paginationToPage(req.GetPagination())

	result, err := s.engine.SearchByHost(req.GetHost(), page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "搜索失败: %v", err)
	}

	return searchResultToProto(result), nil
}

// SearchByStatusCode 按状态码搜索
func (s *SearchServiceImpl) SearchByStatusCode(ctx context.Context, req *pb.SearchByStatusCodeRequest) (*pb.SearchResultProto, error) {
	page, pageSize := paginationToPage(req.GetPagination())

	result, err := s.engine.SearchByStatusCode(int(req.GetStatusCode()), page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "搜索失败: %v", err)
	}

	return searchResultToProto(result), nil
}

// SearchSlowRequests 搜索慢请求
func (s *SearchServiceImpl) SearchSlowRequests(ctx context.Context, req *pb.SearchSlowRequestsRequest) (*pb.SearchResultProto, error) {
	page, pageSize := paginationToPage(req.GetPagination())

	result, err := s.engine.SearchSlowRequests(req.GetThresholdMs(), page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "搜索失败: %v", err)
	}

	return searchResultToProto(result), nil
}

// GetSearchStats 获取搜索统计
func (s *SearchServiceImpl) GetSearchStats(ctx context.Context, req *pb.SearchRequest) (*pb.SearchStatsResponse, error) {
	query := &search.SearchQuery{
		Query:   req.GetQuery(),
		Filters: protoToFilters(req.GetFilters()),
	}

	stats, err := s.engine.BuildSearchStats(query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取搜索统计失败: %v", err)
	}

	return searchStatsToProto(stats), nil
}

// SaveFilter 保存过滤器
func (s *SearchServiceImpl) SaveFilter(ctx context.Context, req *pb.SavedFilter) (*pb.SavedFilter, error) {
	f := search.SavedFilter{
		ID:      req.GetId(),
		Name:    req.GetName(),
		Query:   req.GetQuery(),
		Filters: protoToFilters(req.GetFilters()),
	}

	saved, err := s.filterStore.SaveFilter(f)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "保存过滤器失败: %v", err)
	}

	return savedFilterToProto(*saved), nil
}

// ListFilters 获取保存的过滤器列表
func (s *SearchServiceImpl) ListFilters(ctx context.Context, req *pb.Empty) (*pb.SavedFilterListResponse, error) {
	filters, err := s.filterStore.ListSavedFilters()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询过滤器失败: %v", err)
	}

	items := make([]*pb.SavedFilter, len(filters))
	for i, f := range filters {
		items[i] = savedFilterToProto(f)
	}

	return &pb.SavedFilterListResponse{Filters: items}, nil
}

// DeleteFilter 删除过滤器
func (s *SearchServiceImpl) DeleteFilter(ctx context.Context, req *pb.DeleteFilterRequest) (*pb.Empty, error) {
	if err := s.filterStore.DeleteFilter(req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "删除过滤器失败: %v", err)
	}

	return &pb.Empty{}, nil
}

// === proto ↔ Go 转换函数 ===

func paginationToPage(p *pb.Pagination) (int, int) {
	page := 1
	pageSize := 20
	if p != nil {
		if p.GetPage() > 0 {
			page = int(p.GetPage())
		}
		if p.GetPageSize() > 0 {
			pageSize = int(p.GetPageSize())
		}
	}
	return page, pageSize
}

func protoToFilters(filters []*pb.SearchFilter) []search.Filter {
	result := make([]search.Filter, len(filters))
	for i, f := range filters {
		op := search.FilterOpEq
		switch f.GetOperator() {
		case pb.FilterOperator_FILTER_OPERATOR_EQ:
			op = search.FilterOpEq
		case pb.FilterOperator_FILTER_OPERATOR_NE:
			op = search.FilterOpNe
		case pb.FilterOperator_FILTER_OPERATOR_GT:
			op = search.FilterOpGt
		case pb.FilterOperator_FILTER_OPERATOR_LT:
			op = search.FilterOpLt
		case pb.FilterOperator_FILTER_OPERATOR_CONTAINS:
			op = search.FilterOpContains
		case pb.FilterOperator_FILTER_OPERATOR_REGEX:
			op = search.FilterOpRegex
		}

		result[i] = search.Filter{
			Field:    f.GetField(),
			Operator: op,
			Value:    f.GetValue(),
		}
	}
	return result
}

func searchResultToProto(result *search.SearchResult) *pb.SearchResultProto {
	if result == nil {
		return &pb.SearchResultProto{}
	}

	items := make([]*pb.TrafficEntry, len(result.Items))
	for i, tx := range result.Items {
		items[i] = trafficToProto(tx)
	}

	return &pb.SearchResultProto{
		Items:      items,
		Total:      int32(result.Total),
		Page:       int32(result.Page),
		PageSize:   int32(result.PageSize),
		TotalPages: int32(result.TotalPages),
	}
}

func savedFilterToProto(f search.SavedFilter) *pb.SavedFilter {
	filters := make([]*pb.SearchFilter, len(f.Filters))
	for i, fl := range f.Filters {
		op := pb.FilterOperator_FILTER_OPERATOR_UNSPECIFIED
		switch fl.Operator {
		case search.FilterOpEq:
			op = pb.FilterOperator_FILTER_OPERATOR_EQ
		case search.FilterOpNe:
			op = pb.FilterOperator_FILTER_OPERATOR_NE
		case search.FilterOpGt:
			op = pb.FilterOperator_FILTER_OPERATOR_GT
		case search.FilterOpLt:
			op = pb.FilterOperator_FILTER_OPERATOR_LT
		case search.FilterOpContains:
			op = pb.FilterOperator_FILTER_OPERATOR_CONTAINS
		case search.FilterOpRegex:
			op = pb.FilterOperator_FILTER_OPERATOR_REGEX
		}

		filters[i] = &pb.SearchFilter{
			Field:    fl.Field,
			Operator: op,
			Value:    marshalJSON(fl.Value),
		}
	}

	return &pb.SavedFilter{
		Id:        f.ID,
		Name:      f.Name,
		Query:     f.Query,
		Filters:   filters,
		CreatedAt: f.CreatedAt.Format(time.RFC3339),
	}
}

func searchStatsToProto(stats map[string]interface{}) *pb.SearchStatsResponse {
	resp := &pb.SearchStatsResponse{
		Methods:     make(map[string]int64),
		StatusCodes: make(map[string]int64),
	}

	if total, ok := stats["total"].(int); ok {
		resp.Total = int32(total)
	}

	if methods, ok := stats["methods"].(map[string]int); ok {
		for k, v := range methods {
			resp.Methods[k] = int64(v)
		}
	}

	if statuses, ok := stats["status_codes"].(map[int]int); ok {
		for k, v := range statuses {
			resp.StatusCodes[intToString(k)] = int64(v)
		}
	}

	return resp
}

func intToString(n int) string {
	return fmt.Sprintf("%d", n)
}

// 确保 SearchServiceImpl 实现了 pb.SearchServiceServer
var _ pb.SearchServiceServer = (*SearchServiceImpl)(nil)
