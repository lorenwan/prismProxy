package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/rewrite"
	"prismproxy/internal/rules"
	pb "prismproxy/proto/gen/go"
)

// RewritesServiceImpl RewritesService gRPC 实现
type RewritesServiceImpl struct {
	pb.UnimplementedRewritesServiceServer
	engine *rewrite.Engine
}

// RegisterRewritesServiceImpl 注册 RewritesService
func RegisterRewritesServiceImpl(s *grpc.Server, engine *rewrite.Engine) {
	pb.RegisterRewritesServiceServer(s, &RewritesServiceImpl{engine: engine})
}

// List 获取重写规则列表
func (s *RewritesServiceImpl) List(ctx context.Context, req *pb.Empty) (*pb.RewriteListResponse, error) {
	rulesList, err := s.engine.ListRules()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询重写规则失败: %v", err)
	}

	items := make([]*pb.RewriteRule, len(rulesList))
	for i, r := range rulesList {
		items[i] = rewriteRuleToProto(r)
	}

	return &pb.RewriteListResponse{Rules: items}, nil
}

// Get 获取单条重写规则
func (s *RewritesServiceImpl) Get(ctx context.Context, req *pb.RewriteGetRequest) (*pb.RewriteRule, error) {
	r, err := s.engine.GetRule(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询重写规则失败: %v", err)
	}
	if r == nil {
		return nil, status.Errorf(codes.NotFound, "重写规则不存在: %s", req.GetId())
	}

	return rewriteRuleToProto(r), nil
}

// Create 创建重写规则
func (s *RewritesServiceImpl) Create(ctx context.Context, req *pb.RewriteRule) (*pb.RewriteRule, error) {
	rule := protoToRewriteRule(req)
	if err := s.engine.CreateRule(rule); err != nil {
		return nil, status.Errorf(codes.Internal, "创建重写规则失败: %v", err)
	}

	return rewriteRuleToProto(rule), nil
}

// Update 更新重写规则
func (s *RewritesServiceImpl) Update(ctx context.Context, req *pb.RewriteRule) (*pb.RewriteRule, error) {
	rule := protoToRewriteRule(req)
	if err := s.engine.UpdateRule(rule); err != nil {
		return nil, status.Errorf(codes.Internal, "更新重写规则失败: %v", err)
	}

	return rewriteRuleToProto(rule), nil
}

// Delete 删除重写规则
func (s *RewritesServiceImpl) Delete(ctx context.Context, req *pb.RewriteDeleteRequest) (*pb.Empty, error) {
	if err := s.engine.DeleteRule(req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "删除重写规则失败: %v", err)
	}

	return &pb.Empty{}, nil
}

// Toggle 切换重写规则启用状态
func (s *RewritesServiceImpl) Toggle(ctx context.Context, req *pb.RewriteToggleRequest) (*pb.RewriteRule, error) {
	r, err := s.engine.GetRule(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询重写规则失败: %v", err)
	}
	if r == nil {
		return nil, status.Errorf(codes.NotFound, "重写规则不存在: %s", req.GetId())
	}

	enabled, err := s.engine.ToggleRule(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "切换重写规则状态失败: %v", err)
	}

	r.Enabled = enabled
	return rewriteRuleToProto(r), nil
}

// === proto ↔ Go 转换函数 ===

// rewriteRuleToProto 将 rewrite.RewriteRule 转换为 pb.RewriteRule
func rewriteRuleToProto(r *rewrite.RewriteRule) *pb.RewriteRule {
	if r == nil {
		return nil
	}

	actions := make([]*pb.RewriteAction, len(r.Actions))
	for i, a := range r.Actions {
		actions[i] = rewriteActionToProto(a)
	}

	return &pb.RewriteRule{
		Id:       r.ID,
		Name:     r.Name,
		Enabled:  r.Enabled,
		Priority: int32(r.Priority),
		Match:    ruleMatchToProto(r.Match),
		Actions:  actions,
	}
}

// protoToRewriteRule 将 pb.RewriteRule 转换为 rewrite.RewriteRule
func protoToRewriteRule(r *pb.RewriteRule) *rewrite.RewriteRule {
	if r == nil {
		return nil
	}

	actions := make([]rewrite.RewriteAction, len(r.GetActions()))
	for i, a := range r.GetActions() {
		actions[i] = protoToRewriteAction(a)
	}

	return &rewrite.RewriteRule{
		ID:       r.GetId(),
		Name:     r.GetName(),
		Enabled:  r.GetEnabled(),
		Priority: int(r.GetPriority()),
		Match:    protoToRuleMatch(r.GetMatch()),
		Actions:  actions,
	}
}

// rewriteActionToProto 转换重写动作
func rewriteActionToProto(a rewrite.RewriteAction) *pb.RewriteAction {
	return &pb.RewriteAction{
		Type:   rewriteTypeToProto(a.Type),
		Where:  rewriteWhereToProto(a.Where),
		Key:    a.Key,
		Value:  a.Value,
		Target: a.Target,
	}
}

// protoToRewriteAction 转换重写动作
func protoToRewriteAction(a *pb.RewriteAction) rewrite.RewriteAction {
	if a == nil {
		return rewrite.RewriteAction{}
	}
	return rewrite.RewriteAction{
		Type:   protoToRewriteType(a.GetType()),
		Where:  protoToRewriteWhere(a.GetWhere()),
		Key:    a.GetKey(),
		Value:  a.GetValue(),
		Target: a.GetTarget(),
	}
}

// rewriteTypeToProto 重写类型映射
func rewriteTypeToProto(t rewrite.RewriteType) pb.RewriteType {
	switch t {
	case rewrite.RewriteAddHeader:
		return pb.RewriteType_REWRITE_TYPE_ADD
	case rewrite.RewriteRemoveHeader:
		return pb.RewriteType_REWRITE_TYPE_REMOVE
	case rewrite.RewriteReplaceHeader, rewrite.RewriteReplaceBody, rewrite.RewriteReplaceURL:
		return pb.RewriteType_REWRITE_TYPE_REPLACE
	case rewrite.RewriteMapLocal, rewrite.RewriteMapRemote:
		return pb.RewriteType_REWRITE_TYPE_SET
	default:
		return pb.RewriteType_REWRITE_TYPE_UNSPECIFIED
	}
}

// protoToRewriteType 重写类型映射
func protoToRewriteType(t pb.RewriteType) rewrite.RewriteType {
	switch t {
	case pb.RewriteType_REWRITE_TYPE_ADD:
		return rewrite.RewriteAddHeader
	case pb.RewriteType_REWRITE_TYPE_REMOVE:
		return rewrite.RewriteRemoveHeader
	case pb.RewriteType_REWRITE_TYPE_REPLACE:
		return rewrite.RewriteReplaceHeader
	case pb.RewriteType_REWRITE_TYPE_SET:
		return rewrite.RewriteMapRemote
	default:
		return ""
	}
}

// rewriteWhereToProto 重写位置映射
func rewriteWhereToProto(w rewrite.RewriteWhere) pb.RewriteWhere {
	switch w {
	case rewrite.RewriteWhereRequest:
		return pb.RewriteWhere_REWRITE_WHERE_REQUEST_HEADER
	case rewrite.RewriteWhereResponse:
		return pb.RewriteWhere_REWRITE_WHERE_RESPONSE_HEADER
	default:
		return pb.RewriteWhere_REWRITE_WHERE_UNSPECIFIED
	}
}

// protoToRewriteWhere 重写位置映射
func protoToRewriteWhere(w pb.RewriteWhere) rewrite.RewriteWhere {
	switch w {
	case pb.RewriteWhere_REWRITE_WHERE_REQUEST_HEADER, pb.RewriteWhere_REWRITE_WHERE_REQUEST_BODY, pb.RewriteWhere_REWRITE_WHERE_URL_QUERY, pb.RewriteWhere_REWRITE_WHERE_URL_PATH:
		return rewrite.RewriteWhereRequest
	case pb.RewriteWhere_REWRITE_WHERE_RESPONSE_HEADER, pb.RewriteWhere_REWRITE_WHERE_RESPONSE_BODY:
		return rewrite.RewriteWhereResponse
	default:
		return ""
	}
}

// 确保 RewritesServiceImpl 实现了 pb.RewritesServiceServer
var _ pb.RewritesServiceServer = (*RewritesServiceImpl)(nil)

// 确保未使用的导入被使用
var _ = fmt.Sprintf
var _ = rules.Rule{}
