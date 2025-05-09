package helpers

import (
	"strconv"

	"github.com/Bbanks/internal/structs"
	"github.com/gin-gonic/gin"
)

// GetQueryParams returns required parameters for the API
func GetQueryParams(ctx *gin.Context) QueryParamsStruct {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	sort := ctx.DefaultQuery("sort", "createdAt")
	search := ctx.DefaultQuery("search", "")

	return QueryParamsStruct{Page: page, PageSize: pageSize, Sort: sort, Search: search}
}
