package collection

import (
	"fmt"
)

// Manager 集合管理器
type Manager struct {
	store *Store
}

// NewManager 创建新的集合管理器
func NewManager(store *Store) *Manager {
	return &Manager{store: store}
}

// Init 初始化集合管理器
func (m *Manager) Init() error {
	return m.store.Init()
}

// CreateCollection 创建集合
func (m *Manager) CreateCollection(col *Collection) error {
	if col.Name == "" {
		return fmt.Errorf("集合名称不能为空")
	}
	return m.store.Create(col)
}

// GetCollection 获取集合详情
func (m *Manager) GetCollection(id string) (*Collection, error) {
	return m.store.GetByID(id)
}

// ListCollections 获取集合列表
func (m *Manager) ListCollections() ([]*Collection, error) {
	return m.store.List()
}

// UpdateCollection 更新集合
func (m *Manager) UpdateCollection(col *Collection) error {
	if col.ID == "" {
		return fmt.Errorf("集合 ID 不能为空")
	}
	if col.Name == "" {
		return fmt.Errorf("集合名称不能为空")
	}
	return m.store.Update(col)
}

// DeleteCollection 删除集合
func (m *Manager) DeleteCollection(id string) error {
	return m.store.Delete(id)
}

// CreateItem 创建项目
func (m *Manager) CreateItem(collectionID, parentID string, item *CollectionItem) error {
	if collectionID == "" {
		return fmt.Errorf("集合 ID 不能为空")
	}
	if item.Name == "" {
		return fmt.Errorf("项目名称不能为空")
	}
	if item.Type == "" {
		item.Type = "request"
	}

	// 如果是请求类型，验证请求配置
	if item.Type == "request" && item.Request == nil {
		return fmt.Errorf("请求类型必须包含请求配置")
	}

	return m.store.CreateItem(collectionID, parentID, item)
}

// GetItem 获取项目详情
func (m *Manager) GetItem(id string) (*CollectionItem, error) {
	return m.store.GetItemByID(id)
}

// UpdateItem 更新项目
func (m *Manager) UpdateItem(item *CollectionItem) error {
	if item.ID == "" {
		return fmt.Errorf("项目 ID 不能为空")
	}
	if item.Name == "" {
		return fmt.Errorf("项目名称不能为空")
	}
	return m.store.UpdateItem(item)
}

// DeleteItem 删除项目
func (m *Manager) DeleteItem(id string) error {
	return m.store.DeleteItem(id)
}

// GetItems 获取项目列表
func (m *Manager) GetItems(collectionID, parentID string) ([]CollectionItem, error) {
	return m.store.GetItems(collectionID, parentID)
}

// CreateFolder 创建文件夹
func (m *Manager) CreateFolder(collectionID, parentID, name string) (*CollectionItem, error) {
	if name == "" {
		return nil, fmt.Errorf("文件夹名称不能为空")
	}

	item := &CollectionItem{
		Type: "folder",
		Name: name,
	}

	if err := m.store.CreateItem(collectionID, parentID, item); err != nil {
		return nil, err
	}

	return item, nil
}

// CreateRequest 创建请求
func (m *Manager) CreateRequest(collectionID, parentID string, req *APIRequest) (*CollectionItem, error) {
	if req == nil {
		return nil, fmt.Errorf("请求配置不能为空")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("请求名称不能为空")
	}
	if req.Method == "" {
		req.Method = "GET"
	}

	item := &CollectionItem{
		Type:    "request",
		Name:    req.Name,
		Request: req,
	}

	if err := m.store.CreateItem(collectionID, parentID, item); err != nil {
		return nil, err
	}

	return item, nil
}
