package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/collection"
	pb "prismproxy/proto/gen/go"
)

// CollectionsServiceImpl CollectionsService gRPC 实现
type CollectionsServiceImpl struct {
	pb.UnimplementedCollectionsServiceServer
	manager *collection.Manager
	runner  *collection.Runner
}

// RegisterCollectionsServiceImpl 注册 CollectionsService
func RegisterCollectionsServiceImpl(s *grpc.Server, mgr *collection.Manager, runner *collection.Runner) {
	pb.RegisterCollectionsServiceServer(s, &CollectionsServiceImpl{manager: mgr, runner: runner})
}

// List 获取集合列表
func (s *CollectionsServiceImpl) List(ctx context.Context, req *pb.Empty) (*pb.CollectionListResponse, error) {
	collections, err := s.manager.ListCollections()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询集合失败: %v", err)
	}

	items := make([]*pb.Collection, len(collections))
	for i, c := range collections {
		items[i] = collectionToProto(c)
	}

	return &pb.CollectionListResponse{Collections: items}, nil
}

// Get 获取单个集合
func (s *CollectionsServiceImpl) Get(ctx context.Context, req *pb.CollectionGetRequest) (*pb.Collection, error) {
	c, err := s.manager.GetCollection(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询集合失败: %v", err)
	}
	if c == nil {
		return nil, status.Errorf(codes.NotFound, "集合不存在: %s", req.GetId())
	}

	return collectionToProto(c), nil
}

// Create 创建集合
func (s *CollectionsServiceImpl) Create(ctx context.Context, req *pb.Collection) (*pb.Collection, error) {
	col := protoToCollection(req)
	if err := s.manager.CreateCollection(col); err != nil {
		return nil, status.Errorf(codes.Internal, "创建集合失败: %v", err)
	}

	return collectionToProto(col), nil
}

// Update 更新集合
func (s *CollectionsServiceImpl) Update(ctx context.Context, req *pb.Collection) (*pb.Collection, error) {
	col := protoToCollection(req)
	if err := s.manager.UpdateCollection(col); err != nil {
		return nil, status.Errorf(codes.Internal, "更新集合失败: %v", err)
	}

	return collectionToProto(col), nil
}

// Delete 删除集合
func (s *CollectionsServiceImpl) Delete(ctx context.Context, req *pb.CollectionDeleteRequest) (*pb.Empty, error) {
	if err := s.manager.DeleteCollection(req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "删除集合失败: %v", err)
	}

	return &pb.Empty{}, nil
}

// AddRequest 添加请求到集合
func (s *CollectionsServiceImpl) AddRequest(ctx context.Context, req *pb.AddRequestRequest) (*pb.CollectionItem, error) {
	apiReq := protoToAPIRequest(req.GetRequest())
	item, err := s.manager.CreateRequest(req.GetCollectionId(), req.GetParentItemId(), apiReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "添加请求失败: %v", err)
	}

	return collectionItemToProto(item), nil
}

// UpdateRequest 更新请求
func (s *CollectionsServiceImpl) UpdateRequest(ctx context.Context, req *pb.UpdateRequestRequest) (*pb.APIRequest, error) {
	item := &collection.CollectionItem{
		ID:   req.GetItemId(),
		Name: req.GetRequest().GetName(),
		Type: "request",
	}
	item.Request = protoToAPIRequest(req.GetRequest())
	item.Request.ID = req.GetItemId()

	if err := s.manager.UpdateItem(item); err != nil {
		return nil, status.Errorf(codes.Internal, "更新请求失败: %v", err)
	}

	return req.GetRequest(), nil
}

// DeleteRequest 删除请求
func (s *CollectionsServiceImpl) DeleteRequest(ctx context.Context, req *pb.DeleteRequestRequest) (*pb.Empty, error) {
	if err := s.manager.DeleteItem(req.GetItemId()); err != nil {
		return nil, status.Errorf(codes.Internal, "删除请求失败: %v", err)
	}

	return &pb.Empty{}, nil
}

// ExecuteRequest 执行请求
func (s *CollectionsServiceImpl) ExecuteRequest(ctx context.Context, req *pb.ExecuteRequestRequest) (*pb.ExecutionResult, error) {
	if s.runner == nil {
		return nil, status.Errorf(codes.Unavailable, "请求执行器未初始化")
	}

	// 获取请求定义
	item, err := s.manager.GetItem(req.GetRequest().GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询请求失败: %v", err)
	}
	if item == nil || item.Request == nil {
		return nil, status.Errorf(codes.NotFound, "请求不存在: %s", req.GetRequest().GetId())
	}

	result, err := s.runner.Execute(item.Request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "执行请求失败: %v", err)
	}

	return executionResultToProto(result), nil
}

// === proto ↔ Go 转换函数 ===

// collectionToProto 转换集合
func collectionToProto(c *collection.Collection) *pb.Collection {
	if c == nil {
		return nil
	}

	items := make([]*pb.CollectionItem, len(c.Items))
	for i, item := range c.Items {
		items[i] = collectionItemToProto(&item)
	}

	pbCol := &pb.Collection{
		Id:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		ParentId:    c.ParentID,
		Items:       items,
	}

	if !c.CreatedAt.IsZero() {
		pbCol.CreatedAt = c.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !c.UpdatedAt.IsZero() {
		pbCol.UpdatedAt = c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return pbCol
}

// protoToCollection 转换集合
func protoToCollection(c *pb.Collection) *collection.Collection {
	if c == nil {
		return nil
	}

	col := &collection.Collection{
		ID:          c.GetId(),
		Name:        c.GetName(),
		Description: c.GetDescription(),
		ParentID:    c.GetParentId(),
	}

	if t := c.GetCreatedAt(); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			col.CreatedAt = parsed
		}
	}
	if t := c.GetUpdatedAt(); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			col.UpdatedAt = parsed
		}
	}

	return col
}

// collectionItemToProto 转换集合项目
func collectionItemToProto(item *collection.CollectionItem) *pb.CollectionItem {
	if item == nil {
		return nil
	}

	pbItem := &pb.CollectionItem{
		Id:   item.ID,
		Type: item.Type,
		Name: item.Name,
	}

	if item.Request != nil {
		pbItem.Request = apiRequestToProto(item.Request)
	}

	// 映射子条目（文件夹嵌套）
	if len(item.Items) > 0 {
		pbItem.Items = make([]*pb.CollectionItem, len(item.Items))
		for i, child := range item.Items {
			pbItem.Items[i] = collectionItemToProto(&child)
		}
	}

	if !item.CreatedAt.IsZero() {
		pbItem.CreatedAt = item.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !item.UpdatedAt.IsZero() {
		pbItem.UpdatedAt = item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return pbItem
}

// apiRequestToProto 转换 API 请求
func apiRequestToProto(r *collection.APIRequest) *pb.APIRequest {
	if r == nil {
		return nil
	}

	pbReq := &pb.APIRequest{
		Id:          r.ID,
		Name:        r.Name,
		Method:      r.Method,
		Url:         r.URL,
		Description: r.Description,
		Metadata:    r.Metadata,
	}

	// 转换请求头
	if len(r.Headers) > 0 {
		pbReq.Headers = make([]*pb.KeyValue, len(r.Headers))
		for i, h := range r.Headers {
			pbReq.Headers[i] = &pb.KeyValue{
				Key:         h.Key,
				Value:       h.Value,
				Description: h.Description,
				Enabled:     h.Enabled,
			}
		}
	}

	// 转换查询参数
	if len(r.QueryParams) > 0 {
		pbReq.QueryParams = make([]*pb.KeyValue, len(r.QueryParams))
		for i, p := range r.QueryParams {
			pbReq.QueryParams[i] = &pb.KeyValue{
				Key:         p.Key,
				Value:       p.Value,
				Description: p.Description,
				Enabled:     p.Enabled,
			}
		}
	}

	// 转换请求体
	if r.Body != nil {
		pbReq.Body = &pb.RequestBody{
			Type:    r.Body.Type,
			Content: r.Body.Content,
			Binary:  r.Body.Binary,
		}
		if r.Body.GraphQL != nil {
			pbReq.Body.Graphql = &pb.GraphQLBody{
				Query:     r.Body.GraphQL.Query,
				Variables: r.Body.GraphQL.Variables,
			}
		}
	}

	// 转换认证配置
	if r.Auth != nil {
		pbReq.Auth = &pb.AuthConfig{
			Type:   r.Auth.Type,
			Config: r.Auth.Config,
		}
	}

	// 转换测试用例
	if len(r.Tests) > 0 {
		pbReq.Tests = make([]*pb.Test, len(r.Tests))
		for i, t := range r.Tests {
			pbReq.Tests[i] = &pb.Test{
				Name:     t.Name,
				Type:     t.Type,
				Target:   t.Target,
				Operator: t.Operator,
				Value:    t.Value,
				Enabled:  t.Enabled,
			}
		}
	}

	// 转换变量
	if len(r.Variables) > 0 {
		pbReq.Variables = make([]*pb.KeyValue, len(r.Variables))
		for i, v := range r.Variables {
			pbReq.Variables[i] = &pb.KeyValue{
				Key:         v.Key,
				Value:       v.Value,
				Description: v.Description,
				Enabled:     v.Enabled,
			}
		}
	}

	// 时间戳
	if !r.CreatedAt.IsZero() {
		pbReq.CreatedAt = r.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !r.UpdatedAt.IsZero() {
		pbReq.UpdatedAt = r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return pbReq
}

// protoToAPIRequest 转换 API 请求
func protoToAPIRequest(r *pb.APIRequest) *collection.APIRequest {
	if r == nil {
		return nil
	}

	req := &collection.APIRequest{
		ID:          r.GetId(),
		Name:        r.GetName(),
		Method:      r.GetMethod(),
		URL:         r.GetUrl(),
		Description: r.GetDescription(),
		Metadata:    r.GetMetadata(),
	}

	// 转换请求头
	if len(r.GetHeaders()) > 0 {
		req.Headers = make([]collection.KeyValue, len(r.GetHeaders()))
		for i, h := range r.GetHeaders() {
			req.Headers[i] = collection.KeyValue{
				Key:         h.GetKey(),
				Value:       h.GetValue(),
				Description: h.GetDescription(),
				Enabled:     h.GetEnabled(),
			}
		}
	}

	// 转换查询参数
	if len(r.GetQueryParams()) > 0 {
		req.QueryParams = make([]collection.KeyValue, len(r.GetQueryParams()))
		for i, p := range r.GetQueryParams() {
			req.QueryParams[i] = collection.KeyValue{
				Key:         p.GetKey(),
				Value:       p.GetValue(),
				Description: p.GetDescription(),
				Enabled:     p.GetEnabled(),
			}
		}
	}

	// 转换请求体
	if r.GetBody() != nil {
		req.Body = &collection.RequestBody{
			Type:    r.GetBody().GetType(),
			Content: r.GetBody().GetContent(),
			Binary:  r.GetBody().GetBinary(),
		}
		if r.GetBody().GetGraphql() != nil {
			req.Body.GraphQL = &collection.GraphQLBody{
				Query:     r.GetBody().GetGraphql().GetQuery(),
				Variables: r.GetBody().GetGraphql().GetVariables(),
			}
		}
	}

	// 转换认证配置
	if r.GetAuth() != nil {
		req.Auth = &collection.AuthConfig{
			Type:   r.GetAuth().GetType(),
			Config: r.GetAuth().GetConfig(),
		}
	}

	// 转换测试用例
	if len(r.GetTests()) > 0 {
		req.Tests = make([]collection.Test, len(r.GetTests()))
		for i, t := range r.GetTests() {
			req.Tests[i] = collection.Test{
				Name:     t.GetName(),
				Type:     t.GetType(),
				Target:   t.GetTarget(),
				Operator: t.GetOperator(),
				Value:    t.GetValue(),
				Enabled:  t.GetEnabled(),
			}
		}
	}

	// 转换变量
	if len(r.GetVariables()) > 0 {
		req.Variables = make([]collection.KeyValue, len(r.GetVariables()))
		for i, v := range r.GetVariables() {
			req.Variables[i] = collection.KeyValue{
				Key:         v.GetKey(),
				Value:       v.GetValue(),
				Description: v.GetDescription(),
				Enabled:     v.GetEnabled(),
			}
		}
	}

	// 时间戳
	if t := r.GetCreatedAt(); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			req.CreatedAt = parsed
		}
	}
	if t := r.GetUpdatedAt(); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			req.UpdatedAt = parsed
		}
	}

	return req
}

// executionResultToProto 转换执行结果
func executionResultToProto(r *collection.ExecutionResult) *pb.ExecutionResult {
	if r == nil {
		return nil
	}

	pbResult := &pb.ExecutionResult{
		RequestId:   r.RequestID,
		Status:      int32(r.Status),
		StatusText:  r.StatusText,
		Body:        r.Body,
		ContentType: r.ContentType,
		DurationMs:  r.Duration.Milliseconds(),
		Size:        r.Size,
		Error:       r.Error,
	}

	// 转换响应头
	if len(r.Headers) > 0 {
		pbResult.Headers = make([]*pb.KeyValue, len(r.Headers))
		for i, h := range r.Headers {
			pbResult.Headers[i] = &pb.KeyValue{
				Key:   h.Key,
				Value: h.Value,
			}
		}
	}

	return pbResult
}

// 确保 CollectionsServiceImpl 实现了 pb.CollectionsServiceServer
var _ pb.CollectionsServiceServer = (*CollectionsServiceImpl)(nil)

// 确保未使用的导入被使用
var _ = fmt.Sprintf
