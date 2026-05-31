package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prismproxy/internal/rewrite"
)

// createRewrite 创建重写规则
func (a *API) createRewrite(c *gin.Context) {
	var rule rewrite.RewriteRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	if err := a.rewriteEngine.CreateRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// getRewrite 获取重写规则
func (a *API) getRewrite(c *gin.Context) {
	id := c.Param("id")

	rule, err := a.rewriteEngine.GetRule(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "重写规则不存在"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// listRewrites 获取重写规则列表
func (a *API) listRewrites(c *gin.Context) {
	rules, err := a.rewriteEngine.ListRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rules})
}

// updateRewrite 更新重写规则
func (a *API) updateRewrite(c *gin.Context) {
	id := c.Param("id")

	var rule rewrite.RewriteRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	rule.ID = id
	if err := a.rewriteEngine.UpdateRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// deleteRewrite 删除重写规则
func (a *API) deleteRewrite(c *gin.Context) {
	id := c.Param("id")

	if err := a.rewriteEngine.DeleteRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// toggleRewrite 启用/禁用重写规则
func (a *API) toggleRewrite(c *gin.Context) {
	id := c.Param("id")

	enabled, err := a.rewriteEngine.ToggleRule(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"enabled": enabled})
}

// reorderRewrites 重新排序重写规则
func (a *API) reorderRewrites(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	if err := a.rewriteEngine.ReorderRules(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "排序成功"})
}
