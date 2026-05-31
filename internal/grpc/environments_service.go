package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/environment"
	pb "prismproxy/proto/gen/go"
)

// EnvironmentsServiceImpl EnvironmentsService gRPC 实现
type EnvironmentsServiceImpl struct {
	pb.UnimplementedEnvironmentsServiceServer
	manager *environment.Manager
}

// RegisterEnvironmentsServiceImpl 注册 EnvironmentsService
func RegisterEnvironmentsServiceImpl(s *grpc.Server, mgr *environment.Manager) {
	pb.RegisterEnvironmentsServiceServer(s, &EnvironmentsServiceImpl{manager: mgr})
}

// List 获取环境列表
func (s *EnvironmentsServiceImpl) List(ctx context.Context, req *pb.Empty) (*pb.EnvironmentListResponse, error) {
	envs, err := s.manager.List()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询环境失败: %v", err)
	}

	items := make([]*pb.Environment, len(envs))
	for i, env := range envs {
		items[i] = environmentToProto(env)
	}

	return &pb.EnvironmentListResponse{Environments: items}, nil
}

// Get 获取单个环境
func (s *EnvironmentsServiceImpl) Get(ctx context.Context, req *pb.EnvironmentGetRequest) (*pb.Environment, error) {
	env, err := s.manager.Get(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询环境失败: %v", err)
	}
	if env == nil {
		return nil, status.Errorf(codes.NotFound, "环境不存在: %s", req.GetId())
	}

	return environmentToProto(env), nil
}

// Create 创建环境
func (s *EnvironmentsServiceImpl) Create(ctx context.Context, req *pb.Environment) (*pb.Environment, error) {
	env := protoToEnvironment(req)
	if err := s.manager.Create(env); err != nil {
		return nil, status.Errorf(codes.Internal, "创建环境失败: %v", err)
	}

	return environmentToProto(env), nil
}

// Update 更新环境
func (s *EnvironmentsServiceImpl) Update(ctx context.Context, req *pb.Environment) (*pb.Environment, error) {
	env := protoToEnvironment(req)
	if err := s.manager.Update(env); err != nil {
		return nil, status.Errorf(codes.Internal, "更新环境失败: %v", err)
	}

	return environmentToProto(env), nil
}

// Delete 删除环境
func (s *EnvironmentsServiceImpl) Delete(ctx context.Context, req *pb.EnvironmentDeleteRequest) (*pb.Empty, error) {
	if err := s.manager.Delete(req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "删除环境失败: %v", err)
	}

	return &pb.Empty{}, nil
}

// Activate 激活环境
func (s *EnvironmentsServiceImpl) Activate(ctx context.Context, req *pb.EnvironmentActivateRequest) (*pb.Environment, error) {
	if err := s.manager.SetActive(req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "激活环境失败: %v", err)
	}

	env, err := s.manager.Get(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询环境失败: %v", err)
	}
	if env == nil {
		return nil, status.Errorf(codes.NotFound, "环境不存在: %s", req.GetId())
	}

	return environmentToProto(env), nil
}

// Export 导出环境
func (s *EnvironmentsServiceImpl) Export(ctx context.Context, req *pb.EnvironmentGetRequest) (*pb.EnvironmentExport, error) {
	export, err := s.manager.Export(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "导出环境失败: %v", err)
	}

	return environmentExportToProto(export), nil
}

// Import 导入环境
func (s *EnvironmentsServiceImpl) Import(ctx context.Context, req *pb.EnvironmentExport) (*pb.Environment, error) {
	export := protoToEnvironmentExport(req)
	env, err := s.manager.Import(export)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "导入环境失败: %v", err)
	}

	return environmentToProto(env), nil
}

// === proto ↔ Go 转换函数 ===

// environmentToProto 转换环境
func environmentToProto(env *environment.Environment) *pb.Environment {
	if env == nil {
		return nil
	}

	variables := make([]*pb.Variable, len(env.Variables))
	for i, v := range env.Variables {
		variables[i] = &pb.Variable{
			Id:          v.ID,
			Key:         v.Key,
			Value:       v.Value,
			Description: v.Description,
			Enabled:     v.Enabled,
			IsSecret:    v.IsSecret,
		}
	}

	return &pb.Environment{
		Id:        env.ID,
		Name:      env.Name,
		Variables: variables,
		IsActive:  env.IsActive,
		IsDefault: env.IsDefault,
		BaseUrl:   env.BaseURL,
	}
}

// protoToEnvironment 转换环境
func protoToEnvironment(env *pb.Environment) *environment.Environment {
	if env == nil {
		return nil
	}

	variables := make([]environment.Variable, len(env.GetVariables()))
	for i, v := range env.GetVariables() {
		variables[i] = environment.Variable{
			ID:          v.GetId(),
			Key:         v.GetKey(),
			Value:       v.GetValue(),
			Description: v.GetDescription(),
			Enabled:     v.GetEnabled(),
			IsSecret:    v.GetIsSecret(),
		}
	}

	return &environment.Environment{
		ID:        env.GetId(),
		Name:      env.GetName(),
		Variables: variables,
		IsActive:  env.GetIsActive(),
		IsDefault: env.GetIsDefault(),
		BaseURL:   env.GetBaseUrl(),
	}
}

// environmentExportToProto 转换环境导出
func environmentExportToProto(export *environment.EnvironmentExport) *pb.EnvironmentExport {
	if export == nil {
		return nil
	}

	variables := make([]*pb.VariableExport, len(export.Variables))
	for i, v := range export.Variables {
		variables[i] = &pb.VariableExport{
			Key:         v.Key,
			Value:       v.Value,
			Enabled:     v.Enabled,
			Description: v.Description,
		}
	}

	return &pb.EnvironmentExport{
		Version:   export.Version,
		Name:      export.Name,
		Variables: variables,
	}
}

// protoToEnvironmentExport 转换环境导出
func protoToEnvironmentExport(export *pb.EnvironmentExport) *environment.EnvironmentExport {
	if export == nil {
		return nil
	}

	variables := make([]environment.VariableExport, len(export.GetVariables()))
	for i, v := range export.GetVariables() {
		variables[i] = environment.VariableExport{
			Key:         v.GetKey(),
			Value:       v.GetValue(),
			Enabled:     v.GetEnabled(),
			Description: v.GetDescription(),
		}
	}

	return &environment.EnvironmentExport{
		Version:   export.GetVersion(),
		Name:      export.GetName(),
		Variables: variables,
	}
}

// 确保 EnvironmentsServiceImpl 实现了 pb.EnvironmentsServiceServer
var _ pb.EnvironmentsServiceServer = (*EnvironmentsServiceImpl)(nil)

// 确保未使用的导入被使用
var _ = fmt.Sprintf
