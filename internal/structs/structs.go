package data

// dataAffiliateStruct defines data type expected as `json:""`
type dataAffiliateStruct struct {
	_id            string   `json: "_id"`
	userId         string   `json: "userId"`
	affiliateSales []string `json: "affiliateSales"`
}

// QueryParamsStruct defines the query parameters for the API especially for pagination
type QueryParamsStruct struct {
	Page     int
	Pagesize int
	Sort     string
	Search   string
}
