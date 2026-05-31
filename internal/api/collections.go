package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prismproxy/internal/collection"
)

// listCollections 获取集合列表
func (a *API) listCollections(c *gin.Context) {
	collections, err := a.collectionManager.ListCollections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": collections})
}

// createCollection 创建集合
func (a *API) createCollection(c *gin.Context) {
	var col collection.Collection
	if err := c.ShouldBindJSON(&col); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	if err := a.collectionManager.CreateCollection(&col); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, col)
}

// getCollection 获取集合详情
func (a *API) getCollection(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "集合 ID 不能为空"})
		return
	}

	col, err := a.collectionManager.GetCollection(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if col == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "集合不存在"})
		return
	}

	c.JSON(http.StatusOK, col)
}

// updateCollection 更新集合
func (a *API) updateCollection(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "集合 ID 不能为空"})
		return
	}

	var col collection.Collection
	if err := c.ShouldBindJSON(&col); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}
	col.ID = id

	if err := a.collectionManager.UpdateCollection(&col); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, col)
}

// deleteCollection 删除集合
func (a *API) deleteCollection(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "集合 ID 不能为空"})
		return
	}

	if err := a.collectionManager.DeleteCollection(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// listCollectionItems 获取集合项目列表
func (a *API) listCollectionItems(c *gin.Context) {
	collectionID := c.Param("id")
	parentID := c.Query("parent_id")

	items, err := a.collectionManager.GetItems(collectionID, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// createCollectionItem 创建集合项目
func (a *API) createCollectionItem(c *gin.Context) {
	collectionID := c.Param("id")
	parentID := c.Query("parent_id")

	var item collection.CollectionItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	if err := a.collectionManager.CreateItem(collectionID, parentID, &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// getCollectionItem 获取集合项目详情
func (a *API) getCollectionItem(c *gin.Context) {
	id := c.Param("itemId")

	item, err := a.collectionManager.GetItem(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// updateCollectionItem 更新集合项目
func (a *API) updateCollectionItem(c *gin.Context) {
	id := c.Param("itemId")

	var item collection.CollectionItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}
	item.ID = id

	if err := a.collectionManager.UpdateItem(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// deleteCollectionItem 删除集合项目
func (a *API) deleteCollectionItem(c *gin.Context) {
	id := c.Param("itemId")

	if err := a.collectionManager.DeleteItem(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// createFolder 创建文件夹
func (a *API) createFolder(c *gin.Context) {
	collectionID := c.Param("id")
	parentID := c.Query("parent_id")

	var input struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	folder, err := a.collectionManager.CreateFolder(collectionID, parentID, input.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

// createRequest 创建请求
func (a *API) createRequest(c *gin.Context) {
	collectionID := c.Param("id")
	parentID := c.Query("parent_id")

	var req collection.APIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	item, err := a.collectionManager.CreateRequest(collectionID, parentID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// executeRequest 执行请求
func (a *API) executeRequest(c *gin.Context) {
	id := c.Param("itemId")

	item, err := a.collectionManager.GetItem(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if item == nil || item.Request == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "请求不存在"})
		return
	}

	// 获取环境变量
	envVars := make(map[string]string)
	envID := c.Query("environment_id")
	if envID != "" {
		vars, err := a.environmentManager.GetVariables(envID)
		if err == nil {
			envVars = vars
		}
	}

	// 执行请求
	runner := collection.NewRunner()
	runner.SetEnvironment(envVars)

	result, err := runner.Execute(item.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// generateCode 生成代码
func (a *API) generateCode(c *gin.Context) {
	id := c.Param("itemId")
	language := c.Query("language")
	if language == "" {
		language = "curl"
	}

	item, err := a.collectionManager.GetItem(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if item == nil || item.Request == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "请求不存在"})
		return
	}

	// 构建 cURL 命令
	curlCmd := collection.BuildCurlCommand(item.Request)

	c.JSON(http.StatusOK, gin.H{
		"language": language,
		"code":     curlCmd,
	})
}
