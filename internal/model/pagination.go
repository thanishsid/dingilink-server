package model

type Connection[T any] struct {
	Edges    []Edge[T]
	PageInfo PageInfo
}

type PageInfo struct {
	EndCursor       *string
	HasNextPage     bool
	HasPreviousPage bool
}

type Edge[T any] struct {
	Node   T
	Cursor string
}
