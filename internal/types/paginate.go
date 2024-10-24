package types

// Paginated Data
type PaginatedData[T any] struct {
	List       []T
	TotalCount int64
}

func NewPaginatedData[Tin any, Tout any](in []Tin, convertFunc func(in Tin) Tout, getTotalCountFunc func(in Tin) int64) PaginatedData[Tout] {
	totalCount := int64(0)

	out := make([]Tout, len(in))

	for idx, inputVal := range in {
		if idx == 0 && getTotalCountFunc != nil {
			totalCount = getTotalCountFunc(inputVal)
		}

		out[idx] = convertFunc(inputVal)
	}

	return PaginatedData[Tout]{
		List:       out,
		TotalCount: totalCount,
	}
}
