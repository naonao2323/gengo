package common

import "github.com/naonao2323/testgen/pkg/extractor"

type GoDataType int

const (
	Int GoDataType = iota
	Float64
	String
	Bool
)

func Convert(goDataType extractor.GoDataType) GoDataType {
	switch goDataType {
	case extractor.Int:
		return Int
	case extractor.Float64:
		return Float64
	case extractor.String:
		return String
	case extractor.Bool:
		return Bool
	default:
		return -1
	}
}
