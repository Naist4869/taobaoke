package common

const Taobaoke = iota + 1

type IDGenerator interface {
	Generate() (string, error)
}
