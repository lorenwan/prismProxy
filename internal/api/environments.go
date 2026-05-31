package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prismproxy/internal/environment"
)

// listEnvironments 获取环境列表
func (a *API) listEnvironments(c *gin.Context) {
	envs, err := a.environmentManager.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": envs})
}

// createEnvironment 创建环境
func (a *API) createEnvironment(c *gin.Context) {
	var env environment.Environment
	if err := c.ShouldBindJSON(&env); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	if err := a.environmentManager.Create(&env); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, env)
}

// getEnvironment 获取环境详情
func (a *API) getEnvironment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "环境 ID 不能为空"})
		return
	}

	env, err := a.environmentManager.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if env == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "环境不存在"})
		return
	}

	c.JSON(http.StatusOK, env)
}

// updateEnvironment 更新环境
func (a *API) updateEnvironment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "环境 ID 不能为空"})
		return
	}

	var env environment.Environment
	if err := c.ShouldBindJSON(&env); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}
	env.ID = id

	if err := a.environmentManager.Update(&env); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, env)
}

// deleteEnvironment 删除环境
func (a *API) deleteEnvironment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "环境 ID 不能为空"})
		return
	}

	if err := a.environmentManager.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// setActiveEnvironment 设置活跃环境
func (a *API) setActiveEnvironment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "环境 ID 不能为空"})
		return
	}

	if err := a.environmentManager.SetActive(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设置成功"})
}

// getActiveEnvironment 获取活跃环境
func (a *API) getActiveEnvironment(c *gin.Context) {
	env, err := a.environmentManager.GetActive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if env == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "没有活跃环境"})
		return
	}

	c.JSON(http.StatusOK, env)
}

// addEnvironmentVariable 添加环境变量
func (a *API) addEnvironmentVariable(c *gin.Context) {
	envID := c.Param("id")
	if envID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "环境 ID 不能为空"})
		return
	}

	var variable environment.Variable
	if err := c.ShouldBindJSON(&variable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	if err := a.environmentManager.AddVariable(envID, variable); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, variable)
}

// updateEnvironmentVariable 更新环境变量
func (a *API) updateEnvironmentVariable(c *gin.Context) {
	envID := c.Param("id")
	varID := c.Param("varId")

	var variable environment.Variable
	if err := c.ShouldBindJSON(&variable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}
	variable.ID = varID

	if err := a.environmentManager.UpdateVariable(envID, variable); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variable)
}

// deleteEnvironmentVariable 删除环境变量
func (a *API) deleteEnvironmentVariable(c *gin.Context) {
	envID := c.Param("id")
	varID := c.Param("varId")

	if err := a.environmentManager.DeleteVariable(envID, varID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// exportEnvironment 导出环境
func (a *API) exportEnvironment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "环境 ID 不能为空"})
		return
	}

	export, err := a.environmentManager.Export(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, export)
}

// importEnvironment 导入环境
func (a *API) importEnvironment(c *gin.Context) {
	var export environment.EnvironmentExport
	if err := c.ShouldBindJSON(&export); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	env, err := a.environmentManager.Import(&export)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, env)
}

// duplicateEnvironment 复制环境
func (a *API) duplicateEnvironment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "环境 ID 不能为空"})
		return
	}

	var input struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	env, err := a.environmentManager.Duplicate(id, input.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, env)
}
