package domain

type Result interface {
	isResult()
}

type StringResult string

func (StringResult) isResult() {}

type ListResult []string

func (ListResult) isResult() {}
