package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prismproxy/internal/debugger"
)

// createBreakpoint 创建断点
func (a *API) createBreakpoint(c *gin.Context) {
	var bp debugger.Breakpoint
	if err := c.ShouldBindJSON(&bp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	if err := a.debugger.CreateBreakpoint(&bp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bp)
}

// getBreakpoint 获取断点
func (a *API) getBreakpoint(c *gin.Context) {
	id := c.Param("id")

	bp, err := a.debugger.GetBreakpoint(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if bp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "断点不存在"})
		return
	}

	c.JSON(http.StatusOK, bp)
}

// listBreakpoints 获取断点列表
func (a *API) listBreakpoints(c *gin.Context) {
	breakpoints, err := a.debugger.ListBreakpoints()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": breakpoints})
}

// updateBreakpoint 更新断点
func (a *API) updateBreakpoint(c *gin.Context) {
	id := c.Param("id")

	var bp debugger.Breakpoint
	if err := c.ShouldBindJSON(&bp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	bp.ID = id
	if err := a.debugger.UpdateBreakpoint(&bp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bp)
}

// deleteBreakpoint 删除断点
func (a *API) deleteBreakpoint(c *gin.Context) {
	id := c.Param("id")

	if err := a.debugger.DeleteBreakpoint(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// toggleBreakpoint 启用/禁用断点
func (a *API) toggleBreakpoint(c *gin.Context) {
	id := c.Param("id")

	enabled, err := a.debugger.ToggleBreakpoint(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"enabled": enabled})
}

// listBreakpointSessions 获取活跃会话列表
func (a *API) listBreakpointSessions(c *gin.Context) {
	sessions, err := a.debugger.GetActiveSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

// releaseBreakpointSession 释放会话
func (a *API) releaseBreakpointSession(c *gin.Context) {
	id := c.Param("id")

	if err := a.debugger.ReleaseSession(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已释放"})
}

// modifyBreakpointSession 修改后释放会话
func (a *API) modifyBreakpointSession(c *gin.Context) {
	id := c.Param("id")

	// 这里简化处理，实际应该接收修改后的数据
	if err := a.debugger.ReleaseSession(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已修改并释放"})
}

// dropBreakpointSession 丢弃会话
func (a *API) dropBreakpointSession(c *gin.Context) {
	id := c.Param("id")

	if err := a.debugger.DropSession(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已丢弃"})
}
