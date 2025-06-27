package structs

// QueryParams holds the common query parameters for pagination and searching.
type QueryParams struct {
	// Page is the page number, defaulting to 1.
	Page int
	// PageSize is the number of items per page, defaulting to 10.
	PageSize int
	// Sort is the field to sort by, defaulting to "id".
	Sort string
	// Search is the search query.
	Search string
}
