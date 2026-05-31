package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"prismproxy/internal/collection"
	"prismproxy/internal/debugger"
	"prismproxy/internal/environment"
	"prismproxy/internal/rewrite"
	"prismproxy/internal/storage"
	"prismproxy/internal/websocket"
)

// API REST API 服务
type API struct {
	storage            *storage.Storage
	debugger           *debugger.Debugger
	rewriteEngine      *rewrite.Engine
	collectionManager  *collection.Manager
	environmentManager *environment.Manager
}

// NewAPI 创建新的 API 实例
func NewAPI(store *storage.Storage, hub *websocket.Hub) *API {
	db := store.DB
	return &API{
		storage:            store,
		debugger:           debugger.NewDebugger(db, hub),
		rewriteEngine:      rewrite.NewEngine(db),
		collectionManager:  collection.NewManager(collection.NewStore(db)),
		environmentManager: environment.NewManager(environment.NewStore(db)),
	}
}

// Init 初始化 API 服务
func (a *API) Init() error {
	if err := a.debugger.Init(); err != nil {
		return err
	}
	if err := a.rewriteEngine.Init(); err != nil {
		return err
	}
	if err := a.collectionManager.Init(); err != nil {
		return err
	}
	return a.environmentManager.Init()
}

// RegisterRoutes 注册路由
func (a *API) RegisterRoutes(r *gin.Engine) {
	// 抓包数据 API
	traffic := r.Group("/api/traffic")
	{
		traffic.GET("", a.getTrafficList)
		traffic.GET("/:id", a.getTrafficByID)
		traffic.DELETE("/:id", a.deleteTraffic)
		traffic.DELETE("", a.clearTraffic)
	}

	// 断点 API
	breakpoints := r.Group("/api/breakpoints")
	{
		breakpoints.GET("", a.listBreakpoints)
		breakpoints.POST("", a.createBreakpoint)
		breakpoints.GET("/:id", a.getBreakpoint)
		breakpoints.PUT("/:id", a.updateBreakpoint)
		breakpoints.DELETE("/:id", a.deleteBreakpoint)
		breakpoints.PATCH("/:id/toggle", a.toggleBreakpoint)
	}

	// 断点会话 API
	sessions := r.Group("/api/breakpoint-sessions")
	{
		sessions.GET("", a.listBreakpointSessions)
		sessions.POST("/:id/release", a.releaseBreakpointSession)
		sessions.POST("/:id/modify", a.modifyBreakpointSession)
		sessions.POST("/:id/drop", a.dropBreakpointSession)
	}

	// 重写规则 API
	rewrites := r.Group("/api/rewrites")
	{
		rewrites.GET("", a.listRewrites)
		rewrites.POST("", a.createRewrite)
		rewrites.GET("/:id", a.getRewrite)
		rewrites.PUT("/:id", a.updateRewrite)
		rewrites.DELETE("/:id", a.deleteRewrite)
		rewrites.PATCH("/:id/toggle", a.toggleRewrite)
		rewrites.POST("/reorder", a.reorderRewrites)
	}

	// 集合 API
	collections := r.Group("/api/collections")
	{
		collections.GET("", a.listCollections)
		collections.POST("", a.createCollection)
		collections.GET("/:id", a.getCollection)
		collections.PUT("/:id", a.updateCollection)
		collections.DELETE("/:id", a.deleteCollection)
		collections.GET("/:id/items", a.listCollectionItems)
		collections.POST("/:id/items", a.createCollectionItem)
		collections.POST("/:id/folders", a.createFolder)
		collections.POST("/:id/requests", a.createRequest)
		collections.GET("/:id/items/:itemId", a.getCollectionItem)
		collections.PUT("/:id/items/:itemId", a.updateCollectionItem)
		collections.DELETE("/:id/items/:itemId", a.deleteCollectionItem)
		collections.POST("/:id/items/:itemId/execute", a.executeRequest)
		collections.POST("/:id/items/:itemId/codegen", a.generateCode)
	}

	// 环境 API
	environments := r.Group("/api/environments")
	{
		environments.GET("", a.listEnvironments)
		environments.POST("", a.createEnvironment)
		environments.GET("/active", a.getActiveEnvironment)
		environments.GET("/:id", a.getEnvironment)
		environments.PUT("/:id", a.updateEnvironment)
		environments.DELETE("/:id", a.deleteEnvironment)
		environments.PATCH("/:id/activate", a.setActiveEnvironment)
		environments.POST("/:id/variables", a.addEnvironmentVariable)
		environments.PUT("/:id/variables/:varId", a.updateEnvironmentVariable)
		environments.DELETE("/:id/variables/:varId", a.deleteEnvironmentVariable)
		environments.POST("/:id/export", a.exportEnvironment)
		environments.POST("/import", a.importEnvironment)
		environments.POST("/:id/duplicate", a.duplicateEnvironment)
	}

	// 系统 API
	system := r.Group("/api/system")
	{
		system.GET("/status", a.getStatus)
	}
}

// getTrafficList 获取抓包列表
func (a *API) getTrafficList(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	list, err := a.storage.GetTrafficList(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	count, _ := a.storage.GetTrafficCount()

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": count,
		"limit": limit,
		"offset": offset,
	})
}

// getTrafficByID 根据 ID 获取抓包详情
func (a *API) getTrafficByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	data, err := a.storage.GetTrafficByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if data == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据不存在"})
		return
	}

	c.JSON(http.StatusOK, data)
}

// deleteTraffic 删除抓包数据
func (a *API) deleteTraffic(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	if err := a.storage.DeleteTrafficByID(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// clearTraffic 清空抓包数据
func (a *API) clearTraffic(c *gin.Context) {
	if err := a.storage.ClearTraffic(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "清空成功"})
}

// getStatus 获取系统状态
func (a *API) getStatus(c *gin.Context) {
	count, _ := a.storage.GetTrafficCount()

	c.JSON(http.StatusOK, gin.H{
		"status":  "running",
		"traffic": count,
	})
}
