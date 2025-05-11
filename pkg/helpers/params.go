package helpers

import (
	"net/http"
	"strconv"

	"github.com/Bbanks/internal/structs"
	"github.com/gin-gonic/gin"
)

// GetQueryParams returns required parameters for the API
func GetQueryParams(ctx *gin.Context) QueryParamsStruct {
	pageStr := ctx.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)

	if err != nil || page < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return QueryParamsStruct{}
	}

	pageSizeStr := ctx.DefaultQuery("pageSize", "20")
	pageSize, err := strconv.Atoi(pageSizeStr)

	if err != nil || pageSize < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size"})
		return QueryParamsStruct{}
	}

	sort := ctx.DefaultQuery("sort", "created_at")
	search := ctx.DefaultQuery("search", "")

	return QueryParamsStruct{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Search:   search,
	}
}
