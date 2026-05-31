package environment

import (
	"fmt"
)

// Manager 环境管理器
type Manager struct {
	store *Store
}

// NewManager 创建新的环境管理器
func NewManager(store *Store) *Manager {
	return &Manager{store: store}
}

// Init 初始化环境管理器
func (m *Manager) Init() error {
	return m.store.Init()
}

// Create 创建环境
func (m *Manager) Create(env *Environment) error {
	if env.Name == "" {
		return fmt.Errorf("环境名称不能为空")
	}

	// 设置默认变量
	if env.Variables == nil {
		env.Variables = []Variable{}
	}

	return m.store.Create(env)
}

// Get 获取环境详情
func (m *Manager) Get(id string) (*Environment, error) {
	return m.store.GetByID(id)
}

// List 获取环境列表
func (m *Manager) List() ([]*Environment, error) {
	return m.store.List()
}

// Update 更新环境
func (m *Manager) Update(env *Environment) error {
	if env.ID == "" {
		return fmt.Errorf("环境 ID 不能为空")
	}
	if env.Name == "" {
		return fmt.Errorf("环境名称不能为空")
	}
	return m.store.Update(env)
}

// Delete 删除环境
func (m *Manager) Delete(id string) error {
	return m.store.Delete(id)
}

// SetActive 设置活跃环境
func (m *Manager) SetActive(id string) error {
	return m.store.SetActive(id)
}

// GetActive 获取活跃环境
func (m *Manager) GetActive() (*Environment, error) {
	return m.store.GetActive()
}

// GetVariables 获取环境变量
func (m *Manager) GetVariables(id string) (map[string]string, error) {
	return m.store.GetVariables(id)
}

// AddVariable 添加变量
func (m *Manager) AddVariable(envID string, variable Variable) error {
	env, err := m.store.GetByID(envID)
	if err != nil {
		return err
	}
	if env == nil {
		return fmt.Errorf("环境不存在")
	}

	// 检查变量名是否已存在
	for _, v := range env.Variables {
		if v.Key == variable.Key {
			return fmt.Errorf("变量 %s 已存在", variable.Key)
		}
	}

	if variable.ID == "" {
		variable.ID = fmt.Sprintf("%s_%s", envID, variable.Key)
	}
	if !variable.Enabled {
		variable.Enabled = true
	}

	env.Variables = append(env.Variables, variable)
	return m.store.Update(env)
}

// UpdateVariable 更新变量
func (m *Manager) UpdateVariable(envID string, variable Variable) error {
	env, err := m.store.GetByID(envID)
	if err != nil {
		return err
	}
	if env == nil {
		return fmt.Errorf("环境不存在")
	}

	found := false
	for i, v := range env.Variables {
		if v.ID == variable.ID || v.Key == variable.Key {
			env.Variables[i] = variable
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("变量不存在")
	}

	return m.store.Update(env)
}

// DeleteVariable 删除变量
func (m *Manager) DeleteVariable(envID, variableID string) error {
	env, err := m.store.GetByID(envID)
	if err != nil {
		return err
	}
	if env == nil {
		return fmt.Errorf("环境不存在")
	}

	found := false
	for i, v := range env.Variables {
		if v.ID == variableID {
			env.Variables = append(env.Variables[:i], env.Variables[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("变量不存在")
	}

	return m.store.Update(env)
}

// Export 导出环境
func (m *Manager) Export(id string) (*EnvironmentExport, error) {
	env, err := m.store.GetByID(id)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, fmt.Errorf("环境不存在")
	}

	export := &EnvironmentExport{
		Version: "1.0",
		Name:    env.Name,
	}

	for _, v := range env.Variables {
		export.Variables = append(export.Variables, VariableExport{
			Key:         v.Key,
			Value:       v.Value,
			Enabled:     v.Enabled,
			Description: v.Description,
		})
	}

	return export, nil
}

// Import 导入环境
func (m *Manager) Import(export *EnvironmentExport) (*Environment, error) {
	if export == nil {
		return nil, fmt.Errorf("导入数据为空")
	}
	if export.Name == "" {
		return nil, fmt.Errorf("环境名称为空")
	}

	env := &Environment{
		Name: export.Name,
	}

	for _, v := range export.Variables {
		env.Variables = append(env.Variables, Variable{
			Key:         v.Key,
			Value:       v.Value,
			Enabled:     v.Enabled,
			Description: v.Description,
		})
	}

	if err := m.store.Create(env); err != nil {
		return nil, err
	}

	return env, nil
}

// Duplicate 复制环境
func (m *Manager) Duplicate(id, newName string) (*Environment, error) {
	env, err := m.store.GetByID(id)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, fmt.Errorf("环境不存在")
	}

	newEnv := &Environment{
		Name:      newName,
		BaseURL:   env.BaseURL,
		Variables: make([]Variable, len(env.Variables)),
	}

	// 复制变量
	copy(newEnv.Variables, env.Variables)
	for i := range newEnv.Variables {
		newEnv.Variables[i].ID = ""
	}

	if err := m.store.Create(newEnv); err != nil {
		return nil, err
	}

	return newEnv, nil
}
