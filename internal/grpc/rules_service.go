package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/rules"
	pb "prismproxy/proto/gen/go"
)

// RulesServiceImpl RulesService gRPC 实现
type RulesServiceImpl struct {
	pb.UnimplementedRulesServiceServer
	engine *rules.Engine
}

// RegisterRulesServiceImpl 注册 RulesService
func RegisterRulesServiceImpl(s *grpc.Server, engine *rules.Engine) {
	pb.RegisterRulesServiceServer(s, &RulesServiceImpl{engine: engine})
}

// List 获取规则列表
func (s *RulesServiceImpl) List(ctx context.Context, req *pb.Empty) (*pb.RuleListResponse, error) {
	rulesList, err := s.engine.ListRules()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询规则失败: %v", err)
	}

	items := make([]*pb.Rule, len(rulesList))
	for i, r := range rulesList {
		items[i] = ruleToProto(r)
	}

	return &pb.RuleListResponse{Rules: items}, nil
}

// Get 获取单条规则
func (s *RulesServiceImpl) Get(ctx context.Context, req *pb.RuleGetRequest) (*pb.Rule, error) {
	r, err := s.engine.GetRule(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询规则失败: %v", err)
	}
	if r == nil {
		return nil, status.Errorf(codes.NotFound, "规则不存在: %s", req.GetId())
	}

	return ruleToProto(r), nil
}

// Create 创建规则
func (s *RulesServiceImpl) Create(ctx context.Context, req *pb.Rule) (*pb.Rule, error) {
	rule := protoToRule(req)
	if err := s.engine.CreateRule(rule); err != nil {
		return nil, status.Errorf(codes.Internal, "创建规则失败: %v", err)
	}

	return ruleToProto(rule), nil
}

// Update 更新规则
func (s *RulesServiceImpl) Update(ctx context.Context, req *pb.Rule) (*pb.Rule, error) {
	rule := protoToRule(req)
	if err := s.engine.UpdateRule(rule); err != nil {
		return nil, status.Errorf(codes.Internal, "更新规则失败: %v", err)
	}

	return ruleToProto(rule), nil
}

// Delete 删除规则
func (s *RulesServiceImpl) Delete(ctx context.Context, req *pb.RuleDeleteRequest) (*pb.Empty, error) {
	if err := s.engine.DeleteRule(req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "删除规则失败: %v", err)
	}

	return &pb.Empty{}, nil
}

// Toggle 切换规则启用状态
func (s *RulesServiceImpl) Toggle(ctx context.Context, req *pb.RuleToggleRequest) (*pb.Rule, error) {
	r, err := s.engine.GetRule(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询规则失败: %v", err)
	}
	if r == nil {
		return nil, status.Errorf(codes.NotFound, "规则不存在: %s", req.GetId())
	}

	enabled, err := s.engine.ToggleRule(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "切换规则状态失败: %v", err)
	}

	r.Enabled = enabled
	return ruleToProto(r), nil
}

// Stats 获取规则统计
func (s *RulesServiceImpl) Stats(ctx context.Context, req *pb.Empty) (*pb.RuleStats, error) {
	stats, err := s.engine.GetStats()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取统计失败: %v", err)
	}

	return &pb.RuleStats{
		TotalRules:    int32(stats.TotalRules),
		EnabledRules:  int32(stats.EnabledRules),
		DisabledRules: int32(stats.DisabledRules),
	}, nil
}

// === proto ↔ Go 转换函数 ===

// ruleToProto 将 rules.Rule 转换为 pb.Rule
func ruleToProto(r *rules.Rule) *pb.Rule {
	if r == nil {
		return nil
	}

	return &pb.Rule{
		Id:       r.ID,
		Name:     r.Name,
		Enabled:  r.Enabled,
		Priority: int32(r.Priority),
		Match:    ruleMatchToProto(r.Match),
		Action:   ruleActionToProto(r.Action),
	}
}

// protoToRule 将 pb.Rule 转换为 rules.Rule
func protoToRule(r *pb.Rule) *rules.Rule {
	if r == nil {
		return nil
	}

	return &rules.Rule{
		ID:       r.GetId(),
		Name:     r.GetName(),
		Enabled:  r.GetEnabled(),
		Priority: int(r.GetPriority()),
		Match:    protoToRuleMatch(r.GetMatch()),
		Action:   protoToRuleAction(r.GetAction()),
	}
}

// ruleMatchToProto 转换匹配条件
func ruleMatchToProto(m rules.RuleMatch) *pb.RuleMatch {
	return &pb.RuleMatch{
		UrlPattern:  m.URLPattern,
		UrlWildcard: m.URLWildcard,
		HostPattern: m.HostPattern,
		Methods:     m.Methods,
		HeaderMatch: headerMatchToProto(m.HeaderMatch),
		ContentType: m.ContentType,
	}
}

// protoToRuleMatch 转换匹配条件
func protoToRuleMatch(m *pb.RuleMatch) rules.RuleMatch {
	if m == nil {
		return rules.RuleMatch{}
	}

	rm := rules.RuleMatch{
		URLPattern:  m.GetUrlPattern(),
		URLWildcard: m.GetUrlWildcard(),
		HostPattern: m.GetHostPattern(),
		Methods:     m.GetMethods(),
		ContentType: m.GetContentType(),
	}

	if m.GetHeaderMatch() != nil {
		rm.HeaderMatch = &rules.HeaderMatch{
			Name:      m.GetHeaderMatch().GetName(),
			Value:     m.GetHeaderMatch().GetValue(),
			MatchType: m.GetHeaderMatch().GetMatchType(),
		}
	}

	return rm
}

// headerMatchToProto 转换 Header 匹配
func headerMatchToProto(hm *rules.HeaderMatch) *pb.HeaderMatch {
	if hm == nil {
		return nil
	}
	return &pb.HeaderMatch{
		Name:      hm.Name,
		Value:     hm.Value,
		MatchType: hm.MatchType,
	}
}

// ruleActionToProto 转换规则动作
func ruleActionToProto(a rules.RuleAction) *pb.RuleAction {
	pa := &pb.RuleAction{
		Type:      string(a.Type),
		LocalPath: a.LocalPath,
		RemoteUrl: a.RemoteURL,
		DelayMs:   int32(a.DelayMs),
	}

	if a.Modify != nil {
		pa.Modify = &pb.ModifySpec{
			AddHeaders:    a.Modify.AddHeaders,
			RemoveHeaders: a.Modify.RemoveHeaders,
			SetHeaders:    a.Modify.SetHeaders,
			AddQuery:      a.Modify.AddQuery,
			RemoveQuery:   a.Modify.RemoveQuery,
			SetQuery:      a.Modify.SetQuery,
			BodyReplace:   a.Modify.BodyReplace,
		}
	}

	if a.BlockResponse != nil {
		pa.BlockResponse = &pb.BlockSpec{
			StatusCode: int32(a.BlockResponse.StatusCode),
			Headers:    a.BlockResponse.Headers,
			Body:       a.BlockResponse.Body,
		}
	}

	return pa
}

// protoToRuleAction 转换规则动作
func protoToRuleAction(a *pb.RuleAction) rules.RuleAction {
	if a == nil {
		return rules.RuleAction{}
	}

	ra := rules.RuleAction{
		Type:      rules.ActionType(a.GetType()),
		LocalPath: a.GetLocalPath(),
		RemoteURL: a.GetRemoteUrl(),
		DelayMs:   int(a.GetDelayMs()),
	}

	if a.GetModify() != nil {
		m := a.GetModify()
		ra.Modify = &rules.ModifySpec{
			AddHeaders:    m.GetAddHeaders(),
			RemoveHeaders: m.GetRemoveHeaders(),
			SetHeaders:    m.GetSetHeaders(),
			AddQuery:      m.GetAddQuery(),
			RemoveQuery:   m.GetRemoveQuery(),
			SetQuery:      m.GetSetQuery(),
			BodyReplace:   m.GetBodyReplace(),
		}
	}

	if a.GetBlockResponse() != nil {
		b := a.GetBlockResponse()
		ra.BlockResponse = &rules.BlockSpec{
			StatusCode: int(b.GetStatusCode()),
			Headers:    b.GetHeaders(),
			Body:       b.GetBody(),
		}
	}

	return ra
}

// 确保 RulesServiceImpl 实现了 pb.RulesServiceServer
var _ pb.RulesServiceServer = (*RulesServiceImpl)(nil)

// 确保未使用的导入被使用
var _ = fmt.Sprintf
