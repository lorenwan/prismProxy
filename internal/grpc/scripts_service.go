package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/script"
	pb "prismproxy/proto/gen/go"
)

// ScriptsServiceImpl ScriptsService gRPC 实现
type ScriptsServiceImpl struct {
	pb.UnimplementedScriptsServiceServer
	store  *script.ScriptStore
	engine *script.ScriptEngine
}

// RegisterScriptsServiceImpl 注册 ScriptsService
func RegisterScriptsServiceImpl(s *grpc.Server, store *script.ScriptStore, engine *script.ScriptEngine) {
	pb.RegisterScriptsServiceServer(s, &ScriptsServiceImpl{store: store, engine: engine})
}

// List 获取脚本列表
func (s *ScriptsServiceImpl) List(ctx context.Context, req *pb.Empty) (*pb.ScriptListResponse, error) {
	scripts, err := s.store.List()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询脚本列表失败: %v", err)
	}

	items := make([]*pb.Script, len(scripts))
	for i, sc := range scripts {
		items[i] = scriptToProto(sc)
	}

	return &pb.ScriptListResponse{Scripts: items}, nil
}

// Get 获取单个脚本
func (s *ScriptsServiceImpl) Get(ctx context.Context, req *pb.ScriptGetRequest) (*pb.Script, error) {
	sc, err := s.store.Get(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询脚本失败: %v", err)
	}
	if sc == nil {
		return nil, status.Errorf(codes.NotFound, "脚本不存在: %s", req.GetId())
	}

	return scriptToProto(sc), nil
}

// Create 创建脚本
func (s *ScriptsServiceImpl) Create(ctx context.Context, req *pb.Script) (*pb.Script, error) {
	sc := protoToScript(req)
	if err := s.store.Create(sc); err != nil {
		return nil, status.Errorf(codes.Internal, "创建脚本失败: %v", err)
	}

	return scriptToProto(sc), nil
}

// Update 更新脚本
func (s *ScriptsServiceImpl) Update(ctx context.Context, req *pb.Script) (*pb.Script, error) {
	sc := protoToScript(req)
	if err := s.store.Update(sc); err != nil {
		return nil, status.Errorf(codes.Internal, "更新脚本失败: %v", err)
	}

	return scriptToProto(sc), nil
}

// Delete 删除脚本
func (s *ScriptsServiceImpl) Delete(ctx context.Context, req *pb.ScriptDeleteRequest) (*pb.Empty, error) {
	if err := s.store.Delete(req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "删除脚本失败: %v", err)
	}

	return &pb.Empty{}, nil
}

// Toggle 启用/禁用脚本
func (s *ScriptsServiceImpl) Toggle(ctx context.Context, req *pb.ScriptToggleRequest) (*pb.ToggleResponse, error) {
	enabled, err := s.store.Toggle(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "切换脚本状态失败: %v", err)
	}

	return &pb.ToggleResponse{Enabled: enabled}, nil
}

// Execute 执行脚本
func (s *ScriptsServiceImpl) Execute(ctx context.Context, req *pb.ExecuteScriptRequest) (*pb.ScriptExecution, error) {
	sc, err := s.store.Get(req.GetScriptId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询脚本失败: %v", err)
	}
	if sc == nil {
		return nil, status.Errorf(codes.NotFound, "脚本不存在: %s", req.GetScriptId())
	}

	// 构建执行数据
	data := make(map[string]interface{})
	for k, v := range req.GetData() {
		data[k] = v
	}

	phase := script.ScriptPhase(sc.Phase)
	exec, err := s.engine.Execute(ctx, sc, phase, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "执行脚本失败: %v", err)
	}

	return scriptExecutionToProto(exec), nil
}

// === proto ↔ Go 转换函数 ===

func scriptToProto(sc *script.Script) *pb.Script {
	if sc == nil {
		return nil
	}

	phase := pb.ScriptPhase_SCRIPT_PHASE_UNSPECIFIED
	switch sc.Phase {
	case script.PhaseRequest:
		phase = pb.ScriptPhase_SCRIPT_PHASE_REQUEST
	case script.PhaseResponse:
		phase = pb.ScriptPhase_SCRIPT_PHASE_RESPONSE
	}

	return &pb.Script{
		Id:        sc.ID,
		Name:      sc.Name,
		Content:   sc.Content,
		Phase:     phase,
		Enabled:   sc.Enabled,
		Priority:  int32(sc.Priority),
		Language:  string(sc.Language),
		CreatedAt: sc.CreatedAt.Format(time.RFC3339),
		UpdatedAt: sc.UpdatedAt.Format(time.RFC3339),
	}
}

func protoToScript(p *pb.Script) *script.Script {
	if p == nil {
		return nil
	}

	phase := script.ScriptPhase("")
	switch p.GetPhase() {
	case pb.ScriptPhase_SCRIPT_PHASE_REQUEST:
		phase = script.PhaseRequest
	case pb.ScriptPhase_SCRIPT_PHASE_RESPONSE:
		phase = script.PhaseResponse
	}

	sc := &script.Script{
		ID:       p.GetId(),
		Name:     p.GetName(),
		Content:  p.GetContent(),
		Phase:    phase,
		Enabled:  p.GetEnabled(),
		Priority: int(p.GetPriority()),
		Language: script.ScriptType(p.GetLanguage()),
	}

	if p.GetCreatedAt() != "" {
		sc.CreatedAt, _ = time.Parse(time.RFC3339, p.GetCreatedAt())
	}
	if p.GetUpdatedAt() != "" {
		sc.UpdatedAt, _ = time.Parse(time.RFC3339, p.GetUpdatedAt())
	}

	return sc
}

func scriptExecutionToProto(exec *script.ScriptExecution) *pb.ScriptExecution {
	if exec == nil {
		return nil
	}

	return &pb.ScriptExecution{
		ScriptId:      exec.ScriptID,
		TransactionId: exec.TransactionID,
		Success:       exec.Success,
		Output:        exec.Output,
		DurationMs:    exec.Duration,
		Error:         exec.Error,
		ExecutedAt:    exec.ExecutedAt.Format(time.RFC3339),
	}
}

// 确保 ScriptsServiceImpl 实现了 pb.ScriptsServiceServer
var _ pb.ScriptsServiceServer = (*ScriptsServiceImpl)(nil)
