package common

type GoDataType int

const (
	Int GoDataType = iota
	Float64
	String
	Bool
)

func Convert(goDataType GoDataType) string {
	switch goDataType {
	case Int:
		return "int"
	case Float64:
		return "float64"
	case String:
		return "string"
	case Bool:
		return "bool"
	default:
		return ""
	}
}
