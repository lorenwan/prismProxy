package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/debugger"
	"prismproxy/internal/rules"
	pb "prismproxy/proto/gen/go"
)

// BreakpointsServiceImpl BreakpointsService gRPC 实现
type BreakpointsServiceImpl struct {
	pb.UnimplementedBreakpointsServiceServer
	debugger *debugger.Debugger
}

// RegisterBreakpointsServiceImpl 注册 BreakpointsService
func RegisterBreakpointsServiceImpl(s *grpc.Server, dbg *debugger.Debugger) {
	pb.RegisterBreakpointsServiceServer(s, &BreakpointsServiceImpl{debugger: dbg})
}

// List 获取断点列表
func (s *BreakpointsServiceImpl) List(ctx context.Context, req *pb.Empty) (*pb.BreakpointListResponse, error) {
	breakpoints, err := s.debugger.ListBreakpoints()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询断点失败: %v", err)
	}

	items := make([]*pb.Breakpoint, len(breakpoints))
	for i, bp := range breakpoints {
		items[i] = breakpointToProto(bp)
	}

	return &pb.BreakpointListResponse{Breakpoints: items}, nil
}

// Get 获取单条断点
func (s *BreakpointsServiceImpl) Get(ctx context.Context, req *pb.BreakpointGetRequest) (*pb.Breakpoint, error) {
	bp, err := s.debugger.GetBreakpoint(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询断点失败: %v", err)
	}
	if bp == nil {
		return nil, status.Errorf(codes.NotFound, "断点不存在: %s", req.GetId())
	}

	return breakpointToProto(bp), nil
}

// Create 创建断点
func (s *BreakpointsServiceImpl) Create(ctx context.Context, req *pb.Breakpoint) (*pb.Breakpoint, error) {
	bp := protoToBreakpoint(req)
	if err := s.debugger.CreateBreakpoint(bp); err != nil {
		return nil, status.Errorf(codes.Internal, "创建断点失败: %v", err)
	}

	return breakpointToProto(bp), nil
}

// Update 更新断点
func (s *BreakpointsServiceImpl) Update(ctx context.Context, req *pb.Breakpoint) (*pb.Breakpoint, error) {
	bp := protoToBreakpoint(req)
	if err := s.debugger.UpdateBreakpoint(bp); err != nil {
		return nil, status.Errorf(codes.Internal, "更新断点失败: %v", err)
	}

	return breakpointToProto(bp), nil
}

// Delete 删除断点
func (s *BreakpointsServiceImpl) Delete(ctx context.Context, req *pb.BreakpointDeleteRequest) (*pb.Empty, error) {
	if err := s.debugger.DeleteBreakpoint(req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "删除断点失败: %v", err)
	}

	return &pb.Empty{}, nil
}

// Toggle 切换断点启用状态
func (s *BreakpointsServiceImpl) Toggle(ctx context.Context, req *pb.BreakpointToggleRequest) (*pb.Breakpoint, error) {
	bp, err := s.debugger.GetBreakpoint(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询断点失败: %v", err)
	}
	if bp == nil {
		return nil, status.Errorf(codes.NotFound, "断点不存在: %s", req.GetId())
	}

	enabled, err := s.debugger.ToggleBreakpoint(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "切换断点状态失败: %v", err)
	}

	bp.Enabled = enabled
	return breakpointToProto(bp), nil
}

// ListSessions 获取活跃会话列表
func (s *BreakpointsServiceImpl) ListSessions(ctx context.Context, req *pb.Empty) (*pb.BreakpointSessionListResponse, error) {
	sessions, err := s.debugger.GetActiveSessions()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询会话失败: %v", err)
	}

	items := make([]*pb.BreakpointSession, len(sessions))
	for i, sess := range sessions {
		items[i] = sessionToProto(sess)
	}

	return &pb.BreakpointSessionListResponse{Sessions: items}, nil
}

// ResolveSession 处理会话
func (s *BreakpointsServiceImpl) ResolveSession(ctx context.Context, req *pb.ResolveSessionRequest) (*pb.BreakpointSession, error) {
	sessionID := req.GetSessionId()

	switch req.GetAction() {
	case "release", "continue":
		if err := s.debugger.ReleaseSession(sessionID); err != nil {
			return nil, status.Errorf(codes.Internal, "释放会话失败: %v", err)
		}
	case "modify":
		if req.GetModifiedData() != nil {
			tx := protoToTraffic(req.GetModifiedData())
			if err := s.debugger.ModifySession(sessionID, tx); err != nil {
				return nil, status.Errorf(codes.Internal, "修改会话失败: %v", err)
			}
		}
	case "drop":
		if err := s.debugger.DropSession(sessionID); err != nil {
			return nil, status.Errorf(codes.Internal, "丢弃会话失败: %v", err)
		}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "未知的处理动作: %s", req.GetAction())
	}

	// 获取更新后的会话信息
	sessions, _ := s.debugger.GetActiveSessions()
	for _, sess := range sessions {
		if sess.ID == sessionID {
			return sessionToProto(sess), nil
		}
	}

	// 会话已处理，返回基本信息
	return &pb.BreakpointSession{
		Id:         sessionID,
		ResolvedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// Subscribe 订阅断点事件 (服务端流)
func (s *BreakpointsServiceImpl) Subscribe(req *pb.Empty, stream pb.BreakpointsService_SubscribeServer) error {
	// 创建事件通道
	ch := make(chan debugger.BreakpointEvent, 64)
	s.debugger.Subscribe(ch)
	defer s.debugger.Unsubscribe(ch)

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case event := <-ch:
			pbEvent := &pb.BreakpointEvent{
				Type:      event.Type,
				Timestamp: event.Timestamp,
			}
			if event.Session != nil {
				pbEvent.Session = sessionToProto(event.Session)
			}
			if err := stream.Send(pbEvent); err != nil {
				return err
			}
		}
	}
}

// === proto ↔ Go 转换函数 ===

// breakpointToProto 转换断点
func breakpointToProto(bp *debugger.Breakpoint) *pb.Breakpoint {
	if bp == nil {
		return nil
	}

	return &pb.Breakpoint{
		Id:       bp.ID,
		Name:     bp.Name,
		Enabled:  bp.Enabled,
		Phase:    phaseToProto(bp.Phase),
		Match:    ruleMatchToProto(bp.Match),
		Action:   breakActionToProto(bp.Action),
		HitCount: int32(bp.HitCount),
	}
}

// protoToBreakpoint 转换断点
func protoToBreakpoint(bp *pb.Breakpoint) *debugger.Breakpoint {
	if bp == nil {
		return nil
	}

	return &debugger.Breakpoint{
		ID:       bp.GetId(),
		Name:     bp.GetName(),
		Enabled:  bp.GetEnabled(),
		Phase:    protoToPhase(bp.GetPhase()),
		Match:    protoToRuleMatch(bp.GetMatch()),
		Action:   protoToBreakAction(bp.GetAction()),
		HitCount: int(bp.GetHitCount()),
	}
}

// phaseToProto 转换断点阶段
func phaseToProto(p debugger.Phase) pb.BreakPhase {
	switch p {
	case debugger.PhaseRequest:
		return pb.BreakPhase_BREAK_PHASE_REQUEST
	case debugger.PhaseResponse:
		return pb.BreakPhase_BREAK_PHASE_RESPONSE
	default:
		return pb.BreakPhase_BREAK_PHASE_UNSPECIFIED
	}
}

// protoToPhase 转换断点阶段
func protoToPhase(p pb.BreakPhase) debugger.Phase {
	switch p {
	case pb.BreakPhase_BREAK_PHASE_REQUEST:
		return debugger.PhaseRequest
	case pb.BreakPhase_BREAK_PHASE_RESPONSE:
		return debugger.PhaseResponse
	default:
		return ""
	}
}

// breakActionToProto 转换断点动作
func breakActionToProto(a debugger.BreakAction) *pb.BreakAction {
	pa := &pb.BreakAction{
		Type: breakActionTypeToProto(a.Type),
	}

	if a.Modifications != nil {
		pa.Modifications = &pb.ModifySpec{
			AddHeaders:    a.Modifications.AddHeaders,
			RemoveHeaders: a.Modifications.RemoveHeaders,
			SetHeaders:    a.Modifications.SetHeaders,
			BodyReplace:   a.Modifications.BodyReplace,
		}
	}

	return pa
}

// protoToBreakAction 转换断点动作
func protoToBreakAction(a *pb.BreakAction) debugger.BreakAction {
	if a == nil {
		return debugger.BreakAction{}
	}

	ba := debugger.BreakAction{
		Type: protoToBreakActionType(a.GetType()),
	}

	if a.GetModifications() != nil {
		m := a.GetModifications()
		ba.Modifications = &rules.ModifySpec{
			AddHeaders:    m.GetAddHeaders(),
			RemoveHeaders: m.GetRemoveHeaders(),
			SetHeaders:    m.GetSetHeaders(),
			BodyReplace:   m.GetBodyReplace(),
		}
	}

	return ba
}

// breakActionTypeToProto 转换断点动作类型
func breakActionTypeToProto(t debugger.BreakActionType) pb.BreakActionType {
	switch t {
	case debugger.BreakActionPause:
		return pb.BreakActionType_BREAK_ACTION_PAUSE
	case debugger.BreakActionAutoModify:
		return pb.BreakActionType_BREAK_ACTION_MODIFY
	case debugger.BreakActionDrop:
		return pb.BreakActionType_BREAK_ACTION_DROP
	default:
		return pb.BreakActionType_BREAK_ACTION_UNSPECIFIED
	}
}

// protoToBreakActionType 转换断点动作类型
func protoToBreakActionType(t pb.BreakActionType) debugger.BreakActionType {
	switch t {
	case pb.BreakActionType_BREAK_ACTION_PAUSE:
		return debugger.BreakActionPause
	case pb.BreakActionType_BREAK_ACTION_MODIFY:
		return debugger.BreakActionAutoModify
	case pb.BreakActionType_BREAK_ACTION_DROP:
		return debugger.BreakActionDrop
	default:
		return ""
	}
}

// sessionToProto 转换断点会话
func sessionToProto(sess *debugger.BreakpointSession) *pb.BreakpointSession {
	if sess == nil {
		return nil
	}

	pbSess := &pb.BreakpointSession{
		Id:            sess.ID,
		BreakpointId:  sess.BreakpointID,
		TransactionId: sess.TransactionID,
		Phase:         phaseToProto(sess.Phase),
		Status:        sessionStatusToProto(sess.Status),
		CreatedAt:     sess.CreatedAt.Format(time.RFC3339),
	}

	if sess.Original != nil {
		pbSess.Original = trafficToProto(sess.Original)
	}
	if sess.Modified != nil {
		pbSess.Modified = trafficToProto(sess.Modified)
	}
	if sess.ResolvedAt != nil {
		pbSess.ResolvedAt = sess.ResolvedAt.Format(time.RFC3339)
	}

	return pbSess
}

// sessionStatusToProto 转换会话状态
func sessionStatusToProto(s debugger.SessionStatus) pb.SessionStatus {
	switch s {
	case debugger.SessionStatusPaused:
		return pb.SessionStatus_SESSION_STATUS_WAITING
	case debugger.SessionStatusModified, debugger.SessionStatusReleased:
		return pb.SessionStatus_SESSION_STATUS_RESOLVED
	case debugger.SessionStatusDropped:
		return pb.SessionStatus_SESSION_STATUS_DROPPED
	default:
		return pb.SessionStatus_SESSION_STATUS_UNSPECIFIED
	}
}

// 确保 BreakpointsServiceImpl 实现了 pb.BreakpointsServiceServer
var _ pb.BreakpointsServiceServer = (*BreakpointsServiceImpl)(nil)
