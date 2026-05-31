package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/codegen"
	pb "prismproxy/proto/gen/go"
)

// CodeGenServiceImpl CodeGenService gRPC 实现
type CodeGenServiceImpl struct {
	pb.UnimplementedCodeGenServiceServer
	generator *codegen.Generator
}

// RegisterCodeGenServiceImpl 注册 CodeGenService
func RegisterCodeGenServiceImpl(s *grpc.Server, gen *codegen.Generator) {
	pb.RegisterCodeGenServiceServer(s, &CodeGenServiceImpl{generator: gen})
}

// Generate 生成代码
func (s *CodeGenServiceImpl) Generate(ctx context.Context, req *pb.CodeGenRequest) (*pb.CodeGenResult, error) {
	language := codegen.Language(req.GetLanguage())

	// 构建请求数据
	reqData := &codegen.RequestData{
		Method:   req.GetMethod(),
		URL:      req.GetUrl(),
		Headers:  req.GetHeaders(),
		Body:     req.GetBody(),
		BodyType: req.GetBodyType(),
	}

	// 转换查询参数
	reqData.QueryParams = req.GetQueryParams()

	// 转换认证配置
	if req.GetAuth() != nil {
		reqData.Auth = &codegen.AuthData{
			Type:     req.GetAuth().GetType(),
			Username: req.GetAuth().GetUsername(),
			Password: req.GetAuth().GetPassword(),
			Token:    req.GetAuth().GetToken(),
			APIKey:   req.GetAuth().GetApiKey(),
			APIValue: req.GetAuth().GetApiValue(),
			Location: req.GetAuth().GetLocation(),
		}
	}

	code, err := s.generator.Generate(language, reqData)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "生成代码失败: %v", err)
	}

	return &pb.CodeGenResult{
		Code:     code,
		Language: req.GetLanguage(),
	}, nil
}

// ListLanguages 获取支持的语言列表
func (s *CodeGenServiceImpl) ListLanguages(ctx context.Context, req *pb.Empty) (*pb.LanguageListResponse, error) {
	languages := s.generator.GetSupportedLanguages()

	items := make([]*pb.LanguageInfo, len(languages))
	for i, lang := range languages {
		items[i] = &pb.LanguageInfo{
			Id:          lang["id"],
			Name:        lang["name"],
			Description: lang["description"],
		}
	}

	return &pb.LanguageListResponse{Languages: items}, nil
}

// 确保 CodeGenServiceImpl 实现了 pb.CodeGenServiceServer
var _ pb.CodeGenServiceServer = (*CodeGenServiceImpl)(nil)

// 确保未使用的导入被使用
var _ = fmt.Sprintf
