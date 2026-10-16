package domain

type Result interface {
	isResult()
}

type StringResult string

type ListResult []string

type MapResult map[string]string

func (ListResult) isResult() {}

func (StringResult) isResult() {}

func (MapResult) isResult() {}
