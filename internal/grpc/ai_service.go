package grpc

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/ai"
	"prismproxy/internal/traffic"
	pb "prismproxy/proto/gen/go"
)

// AIServiceImpl AIService gRPC 实现
type AIServiceImpl struct {
	pb.UnimplementedAIServiceServer
	service *ai.Service
	manager *traffic.Manager
}

// RegisterAIServiceImpl 注册 AIService
func RegisterAIServiceImpl(s *grpc.Server, svc *ai.Service, mgr *traffic.Manager) {
	pb.RegisterAIServiceServer(s, &AIServiceImpl{service: svc, manager: mgr})
}

// Chat 非流式聊天
func (s *AIServiceImpl) Chat(ctx context.Context, req *pb.ChatRequest) (*pb.ChatResponse, error) {
	chatReq := &ai.ChatRequest{
		Messages: protoToChatMessages(req.GetMessages()),
		Model:    req.GetModel(),
	}

	resp, err := s.service.Chat(ctx, chatReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "AI 聊天失败: %v", err)
	}

	return &pb.ChatResponse{
		Content:  resp.Content,
		Provider: resp.Provider,
		Model:    resp.Model,
		Usage:    usageToProto(resp.Usage),
	}, nil
}

// StreamChat 流式聊天
func (s *AIServiceImpl) StreamChat(req *pb.ChatRequest, stream pb.AIService_StreamChatServer) error {
	chatReq := &ai.ChatRequest{
		Messages: protoToChatMessages(req.GetMessages()),
		Model:    req.GetModel(),
		Stream:   true,
	}

	chunks, err := s.service.StreamChat(stream.Context(), chatReq)
	if err != nil {
		return status.Errorf(codes.Internal, "启动流式聊天失败: %v", err)
	}

	for chunk := range chunks {
		if err := stream.Send(&pb.ChatChunk{
			Content:  chunk.Content,
			Done:     chunk.Done,
			Provider: chunk.Provider,
		}); err != nil {
			return status.Errorf(codes.Internal, "发送流数据失败: %v", err)
		}
	}

	return nil
}

// AnalyzeTraffic 分析流量数据
func (s *AIServiceImpl) AnalyzeTraffic(ctx context.Context, req *pb.AnalyzeTrafficRequest) (*pb.AnalysisResult, error) {
	tx, err := s.manager.GetTransaction(req.GetTrafficId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取流量数据失败: %v", err)
	}
	if tx == nil {
		return nil, status.Errorf(codes.NotFound, "流量记录不存在: %d", req.GetTrafficId())
	}

	result, err := s.service.AnalyzeTraffic(ctx, []*traffic.Transaction{tx})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "流量分析失败: %v", err)
	}

	return analysisResultToProto(result), nil
}

// SecurityScan 安全扫描
func (s *AIServiceImpl) SecurityScan(ctx context.Context, req *pb.SecurityScanRequest) (*pb.SecurityReport, error) {
	tx, err := s.manager.GetTransaction(req.GetTrafficId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取流量数据失败: %v", err)
	}
	if tx == nil {
		return nil, status.Errorf(codes.NotFound, "流量记录不存在: %d", req.GetTrafficId())
	}

	report, err := s.service.SecurityCheck(ctx, tx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "安全扫描失败: %v", err)
	}

	return securityReportToProto(report), nil
}

// GenerateTests 生成测试用例
func (s *AIServiceImpl) GenerateTests(ctx context.Context, req *pb.GenerateTestsRequest) (*pb.GenerateTestsResponse, error) {
	tx, err := s.manager.GetTransaction(req.GetTrafficId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取流量数据失败: %v", err)
	}
	if tx == nil {
		return nil, status.Errorf(codes.NotFound, "流量记录不存在: %d", req.GetTrafficId())
	}

	tests, err := s.service.GenerateTests(ctx, []*traffic.Transaction{tx})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "生成测试用例失败: %v", err)
	}

	return &pb.GenerateTestsResponse{
		Tests: testCasesToProto(tests),
	}, nil
}

// CheckAvailability 检查 AI 服务可用性
func (s *AIServiceImpl) CheckAvailability(ctx context.Context, req *pb.Empty) (*pb.AvailabilityResponse, error) {
	providers := s.service.GetAvailableProviders()

	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = string(p)
	}

	return &pb.AvailabilityResponse{
		Available: len(providers) > 0,
		Providers: names,
	}, nil
}

// === proto ↔ Go 转换函数 ===

// protoToChatMessages 转换聊天消息
func protoToChatMessages(msgs []*pb.ChatMessage) []ai.ChatMessage {
	result := make([]ai.ChatMessage, len(msgs))
	for i, m := range msgs {
		result[i] = ai.ChatMessage{
			Role:    m.GetRole(),
			Content: m.GetContent(),
		}
	}
	return result
}

// analysisResultToProto 转换分析结果
func analysisResultToProto(r *ai.AnalysisResult) *pb.AnalysisResult {
	if r == nil {
		return nil
	}

	pbResult := &pb.AnalysisResult{
		Summary: r.Summary,
	}

	if len(r.Issues) > 0 {
		pbResult.Issues = make([]*pb.Issue, len(r.Issues))
		for i, issue := range r.Issues {
			pbResult.Issues[i] = &pb.Issue{
				Severity: issue.Severity,
				Type:     issue.Type,
				Title:    issue.Title,
				Detail:   issue.Detail,
			}
		}
	}

	if len(r.Suggestions) > 0 {
		pbResult.Suggestions = make([]*pb.Suggestion, len(r.Suggestions))
		for i, sug := range r.Suggestions {
			pbResult.Suggestions[i] = &pb.Suggestion{
				Category: sug.Category,
				Title:    sug.Title,
				Detail:   sug.Detail,
			}
		}
	}

	return pbResult
}

// securityReportToProto 转换安全报告
func securityReportToProto(r *ai.SecurityReport) *pb.SecurityReport {
	if r == nil {
		return nil
	}

	pbReport := &pb.SecurityReport{
		RiskLevel: r.RiskLevel,
		Summary:   r.Summary,
	}

	if len(r.Findings) > 0 {
		pbReport.Findings = make([]*pb.SecurityFinding, len(r.Findings))
		for i, f := range r.Findings {
			pbReport.Findings[i] = &pb.SecurityFinding{
				Severity:    f.Severity,
				Category:    f.Category,
				Title:       f.Title,
				Description: f.Description,
				Location:    f.Location,
				Remediation: f.Remediation,
			}
		}
	}

	return pbReport
}

// testCasesToProto 转换测试用例
func testCasesToProto(tests []*ai.TestCase) []*pb.TestCase {
	result := make([]*pb.TestCase, len(tests))
	for i, t := range tests {
		pbTest := &pb.TestCase{
			Name:        t.Name,
			Description: t.Description,
			Method:      t.Method,
			Url:         t.URL,
			Body:        t.Body,
		}

		if len(t.Headers) > 0 {
			pbTest.Headers = t.Headers
		}

		if len(t.Assertions) > 0 {
			pbTest.Assertions = make([]*pb.Assertion, len(t.Assertions))
			for j, a := range t.Assertions {
				pbTest.Assertions[j] = &pb.Assertion{
					Type:     a.Type,
					Target:   a.Target,
					Operator: a.Operator,
					Value:    a.Value,
				}
			}
		}

		result[i] = pbTest
	}
	return result
}

// usageToProto 将领域 Usage 转换为 Proto Usage
func usageToProto(u *ai.Usage) *pb.Usage {
	if u == nil {
		return nil
	}
	return &pb.Usage{
		PromptTokens:     int32(u.PromptTokens),
		CompletionTokens: int32(u.CompletionTokens),
		TotalTokens:      int32(u.TotalTokens),
	}
}

// 确保 AIServiceImpl 实现了 pb.AIServiceServer
var _ pb.AIServiceServer = (*AIServiceImpl)(nil)

// 确保未使用的导入被使用
var _ = fmt.Sprintf
var _ = log.Printf
