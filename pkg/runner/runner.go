package runner

type RunnerItem interface {
	GetErr() error
}

func Runner(item RunnerItem) {

}
