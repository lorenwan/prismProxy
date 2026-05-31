package grpc

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/diff"
	pb "prismproxy/proto/gen/go"
)

// DiffServiceImpl DiffService gRPC 实现
type DiffServiceImpl struct {
	pb.UnimplementedDiffServiceServer
	engine *diff.DiffEngine
}

// RegisterDiffServiceImpl 注册 DiffService
func RegisterDiffServiceImpl(s *grpc.Server, engine *diff.DiffEngine) {
	pb.RegisterDiffServiceServer(s, &DiffServiceImpl{engine: engine})
}

// CompareHeaders 对比 Headers
func (s *DiffServiceImpl) CompareHeaders(ctx context.Context, req *pb.CompareHeadersRequest) (*pb.DiffResultProto, error) {
	left := protoToHeaderMap(req.GetLeft())
	right := protoToHeaderMap(req.GetRight())

	result := s.engine.CompareHeaders(left, right)
	return diffResultToProto(result), nil
}

// CompareBody 对比 Body
func (s *DiffServiceImpl) CompareBody(ctx context.Context, req *pb.CompareBodyRequest) (*pb.DiffResultProto, error) {
	result := s.engine.CompareBody(req.GetLeft(), req.GetRight())
	return diffResultToProto(result), nil
}

// CompareJSON 对比 JSON
func (s *DiffServiceImpl) CompareJSON(ctx context.Context, req *pb.CompareJSONRequest) (*pb.JSONDiffResultProto, error) {
	result, err := s.engine.CompareJSONStrings(req.GetLeft(), req.GetRight())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "JSON 对比失败: %v", err)
	}

	return jsonDiffResultToProto(result), nil
}

// CompareQuery 对比 Query 参数
func (s *DiffServiceImpl) CompareQuery(ctx context.Context, req *pb.CompareQueryRequest) (*pb.DiffResultProto, error) {
	left := protoToQueryMap(req.GetLeft())
	right := protoToQueryMap(req.GetRight())

	result := s.engine.CompareQuery(left, right)
	return diffResultToProto(result), nil
}

// === proto ↔ Go 转换函数 ===

func diffResultToProto(result diff.DiffResult) *pb.DiffResultProto {
	entries := make([]*pb.DiffEntry, len(result.Entries))
	for i, e := range result.Entries {
		entries[i] = &pb.DiffEntry{
			Path:   e.Path,
			Left:   e.Left,
			Right:  e.Right,
			Status: diffStatusToProto(e.Status),
		}
	}

	return &pb.DiffResultProto{
		Type:    diffTypeToProto(result.Type),
		Entries: entries,
	}
}

func jsonDiffResultToProto(result diff.JSONDiffResult) *pb.JSONDiffResultProto {
	diffs := make([]*pb.JSONDiffEntry, len(result.Diffs))
	for i, d := range result.Diffs {
		leftStr := ""
		rightStr := ""
		if d.Left != nil {
			leftStr = marshalJSON(d.Left)
		}
		if d.Right != nil {
			rightStr = marshalJSON(d.Right)
		}

		diffs[i] = &pb.JSONDiffEntry{
			Path:   d.Path,
			Left:   leftStr,
			Right:  rightStr,
			Status: diffStatusToProto(d.Status),
			Type:   d.Type,
		}
	}

	return &pb.JSONDiffResultProto{
		Diffs: diffs,
		Summary: &pb.DiffSummary{
			TotalFields: int32(result.Summary.TotalFields),
			Added:       int32(result.Summary.Added),
			Removed:     int32(result.Summary.Removed),
			Modified:    int32(result.Summary.Modified),
			Unchanged:   int32(result.Summary.Unchanged),
		},
	}
}

func diffStatusToProto(status diff.DiffStatus) pb.DiffStatus {
	switch status {
	case diff.StatusAdded:
		return pb.DiffStatus_DIFF_STATUS_ADDED
	case diff.StatusRemoved:
		return pb.DiffStatus_DIFF_STATUS_REMOVED
	case diff.StatusModified:
		return pb.DiffStatus_DIFF_STATUS_MODIFIED
	case diff.StatusUnchanged:
		return pb.DiffStatus_DIFF_STATUS_UNCHANGED
	default:
		return pb.DiffStatus_DIFF_STATUS_UNSPECIFIED
	}
}

func diffTypeToProto(t diff.DiffType) pb.DiffType {
	switch t {
	case diff.DiffTypeHeaders:
		return pb.DiffType_DIFF_TYPE_HEADERS
	case diff.DiffTypeBody:
		return pb.DiffType_DIFF_TYPE_BODY
	case diff.DiffTypeJSON:
		return pb.DiffType_DIFF_TYPE_JSON
	case diff.DiffTypeQuery:
		return pb.DiffType_DIFF_TYPE_QUERY
	default:
		return pb.DiffType_DIFF_TYPE_UNSPECIFIED
	}
}

func protoToHeaderMap(headers map[string]*pb.StringList) http.Header {
	result := make(http.Header)
	for k, v := range headers {
		if v != nil {
			result[k] = v.GetValues()
		}
	}
	return result
}

func protoToQueryMap(params map[string]*pb.StringList) map[string][]string {
	result := make(map[string][]string)
	for k, v := range params {
		if v != nil {
			result[k] = v.GetValues()
		}
	}
	return result
}

// 确保 DiffServiceImpl 实现了 pb.DiffServiceServer
var _ pb.DiffServiceServer = (*DiffServiceImpl)(nil)
