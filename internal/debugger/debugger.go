package debugger

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"prismproxy/internal/rules"
	"prismproxy/internal/traffic"
	"prismproxy/internal/websocket"
)

// BreakpointEvent 断点事件
type BreakpointEvent struct {
	Type      string             `json:"type"` // hit / resolved
	Session   *BreakpointSession `json:"session,omitempty"`
	Timestamp string             `json:"timestamp"`
}

// Debugger 断点调试器
type Debugger struct {
	store   *Store
	matcher *rules.Matcher
	hub     *websocket.Hub
	mu      sync.RWMutex
	// 活跃会话通道，用于阻塞等待用户操作
	sessions map[string]chan *SessionResult
	// 事件订阅者
	subscribers map[chan<- BreakpointEvent]struct{}
}

// SessionResult 会话操作结果
type SessionResult struct {
	Action     string               `json:"action"` // release, modify, drop
	Modified   *traffic.Transaction `json:"modified,omitempty"`
}

// NewDebugger 创建新的调试器
func NewDebugger(db *sql.DB, hub *websocket.Hub) *Debugger {
	return &Debugger{
		store:       NewStore(db),
		matcher:     rules.NewMatcher(),
		hub:         hub,
		sessions:    make(map[string]chan *SessionResult),
		subscribers: make(map[chan<- BreakpointEvent]struct{}),
	}
}

// Subscribe 订阅断点事件
func (d *Debugger) Subscribe(ch chan<- BreakpointEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.subscribers[ch] = struct{}{}
}

// Unsubscribe 取消订阅断点事件
func (d *Debugger) Unsubscribe(ch chan<- BreakpointEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.subscribers, ch)
}

// fireEvent 向所有订阅者发送事件
func (d *Debugger) fireEvent(event BreakpointEvent) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for ch := range d.subscribers {
		select {
		case ch <- event:
		default:
			// 队列满时跳过
		}
	}
}

// Init 初始化调试器
func (d *Debugger) Init() error {
	return d.store.Init()
}

// CreateBreakpoint 创建断点
func (d *Debugger) CreateBreakpoint(bp *Breakpoint) error {
	return d.store.Create(bp)
}

// GetBreakpoint 获取断点
func (d *Debugger) GetBreakpoint(id string) (*Breakpoint, error) {
	return d.store.GetByID(id)
}

// ListBreakpoints 获取断点列表
func (d *Debugger) ListBreakpoints() ([]*Breakpoint, error) {
	return d.store.List()
}

// UpdateBreakpoint 更新断点
func (d *Debugger) UpdateBreakpoint(bp *Breakpoint) error {
	return d.store.Update(bp)
}

// DeleteBreakpoint 删除断点
func (d *Debugger) DeleteBreakpoint(id string) error {
	return d.store.Delete(id)
}

// ToggleBreakpoint 启用/禁用断点
func (d *Debugger) ToggleBreakpoint(id string) (bool, error) {
	return d.store.Toggle(id)
}

// GetActiveSessions 获取活跃会话列表
func (d *Debugger) GetActiveSessions() ([]*BreakpointSession, error) {
	return d.store.GetActiveSessions()
}

// CheckRequest 检查请求是否匹配断点
func (d *Debugger) CheckRequest(method, url, host string, headers http.Header, tx *traffic.Transaction) *BreakpointSession {
	breakpoints, err := d.store.GetEnabled()
	if err != nil {
		log.Printf("[ERROR] 获取断点失败: %v", err)
		return nil
	}

	for _, bp := range breakpoints {
		if bp.Phase != PhaseRequest {
			continue
		}

		if d.matcher.MatchRequest(&rules.Rule{
			Enabled: true,
			Match:   bp.Match,
		}, method, url, host, headers) {
			// 命中断点，创建会话
			session := &BreakpointSession{
				BreakpointID:  bp.ID,
				TransactionID: tx.ID,
				Phase:         PhaseRequest,
				Original:      tx,
			}

			if err := d.store.CreateSession(session); err != nil {
				log.Printf("[ERROR] 创建断点会话失败: %v", err)
				continue
			}

			// 增加命中次数
			d.store.IncrementHitCount(bp.ID)

			// 通知前端
			d.notifyBreakpointHit(session)

			// 如果是暂停动作，等待用户操作
			if bp.Action.Type == BreakActionPause {
				return session
			}

			// 如果是自动修改动作
			if bp.Action.Type == BreakActionAutoModify && bp.Action.Modifications != nil {
				session.Status = SessionStatusModified
				now := time.Now()
				session.ResolvedAt = &now
				d.store.UpdateSession(session)
				return session
			}

			// 如果是丢弃动作
			if bp.Action.Type == BreakActionDrop {
				session.Status = SessionStatusDropped
				now := time.Now()
				session.ResolvedAt = &now
				d.store.UpdateSession(session)
				return session
			}
		}
	}

	return nil
}

// CheckResponse 检查响应是否匹配断点
func (d *Debugger) CheckResponse(method, url, host string, headers http.Header, tx *traffic.Transaction) *BreakpointSession {
	breakpoints, err := d.store.GetEnabled()
	if err != nil {
		log.Printf("[ERROR] 获取断点失败: %v", err)
		return nil
	}

	for _, bp := range breakpoints {
		if bp.Phase != PhaseResponse {
			continue
		}

		if d.matcher.MatchRequest(&rules.Rule{
			Enabled: true,
			Match:   bp.Match,
		}, method, url, host, headers) {
			session := &BreakpointSession{
				BreakpointID:  bp.ID,
				TransactionID: tx.ID,
				Phase:         PhaseResponse,
				Original:      tx,
			}

			if err := d.store.CreateSession(session); err != nil {
				log.Printf("[ERROR] 创建断点会话失败: %v", err)
				continue
			}

			d.store.IncrementHitCount(bp.ID)
			d.notifyBreakpointHit(session)

			if bp.Action.Type == BreakActionPause {
				return session
			}

			if bp.Action.Type == BreakActionAutoModify && bp.Action.Modifications != nil {
				session.Status = SessionStatusModified
				now := time.Now()
				session.ResolvedAt = &now
				d.store.UpdateSession(session)
				return session
			}

			if bp.Action.Type == BreakActionDrop {
				session.Status = SessionStatusDropped
				now := time.Now()
				session.ResolvedAt = &now
				d.store.UpdateSession(session)
				return session
			}
		}
	}

	return nil
}

// WaitForUserAction 等待用户操作（阻塞）
func (d *Debugger) WaitForUserAction(sessionID string) (*SessionResult, error) {
	d.mu.Lock()
	ch := make(chan *SessionResult, 1)
	d.sessions[sessionID] = ch
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		delete(d.sessions, sessionID)
		d.mu.Unlock()
	}()

	// 等待用户操作或超时
	select {
	case result := <-ch:
		return result, nil
	case <-time.After(5 * time.Minute):
		// 超时自动释放
		return &SessionResult{Action: "release"}, nil
	}
}

// ReleaseSession 释放会话
func (d *Debugger) ReleaseSession(sessionID string) error {
	session, err := d.store.GetSessionByID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("会话不存在")
	}
	if session.Status != SessionStatusPaused {
		return fmt.Errorf("会话已处理")
	}

	session.Status = SessionStatusReleased
	now := time.Now()
	session.ResolvedAt = &now

	if err := d.store.UpdateSession(session); err != nil {
		return err
	}

	// 通知等待的 goroutine
	d.mu.RLock()
	ch, ok := d.sessions[sessionID]
	d.mu.RUnlock()

	if ok {
		ch <- &SessionResult{Action: "release"}
	}

	d.notifySessionResolved(session)
	return nil
}

// ModifySession 修改后释放会话
func (d *Debugger) ModifySession(sessionID string, modified *traffic.Transaction) error {
	session, err := d.store.GetSessionByID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("会话不存在")
	}
	if session.Status != SessionStatusPaused {
		return fmt.Errorf("会话已处理")
	}

	session.Status = SessionStatusModified
	session.Modified = modified
	now := time.Now()
	session.ResolvedAt = &now

	if err := d.store.UpdateSession(session); err != nil {
		return err
	}

	// 通知等待的 goroutine
	d.mu.RLock()
	ch, ok := d.sessions[sessionID]
	d.mu.RUnlock()

	if ok {
		ch <- &SessionResult{Action: "modify", Modified: modified}
	}

	d.notifySessionResolved(session)
	return nil
}

// DropSession 丢弃会话
func (d *Debugger) DropSession(sessionID string) error {
	session, err := d.store.GetSessionByID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("会话不存在")
	}
	if session.Status != SessionStatusPaused {
		return fmt.Errorf("会话已处理")
	}

	session.Status = SessionStatusDropped
	now := time.Now()
	session.ResolvedAt = &now

	if err := d.store.UpdateSession(session); err != nil {
		return err
	}

	// 通知等待的 goroutine
	d.mu.RLock()
	ch, ok := d.sessions[sessionID]
	d.mu.RUnlock()

	if ok {
		ch <- &SessionResult{Action: "drop"}
	}

	d.notifySessionResolved(session)
	return nil
}

// notifyBreakpointHit 通知前端断点命中
func (d *Debugger) notifyBreakpointHit(session *BreakpointSession) {
	if d.hub != nil {
		d.hub.Broadcast(&websocket.Message{
			Type:    "breakpoint_hit",
			Payload: session,
		})
	}

	// 向 gRPC 订阅者发送事件
	go d.fireEvent(BreakpointEvent{
		Type:      "hit",
		Session:   session,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// notifySessionResolved 通知前端会话已处理
func (d *Debugger) notifySessionResolved(session *BreakpointSession) {
	if d.hub != nil {
		d.hub.Broadcast(&websocket.Message{
			Type:    "breakpoint_resolved",
			Payload: session,
		})
	}

	// 向 gRPC 订阅者发送事件
	go d.fireEvent(BreakpointEvent{
		Type:      "resolved",
		Session:   session,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
