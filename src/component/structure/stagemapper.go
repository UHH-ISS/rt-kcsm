package structure

type StageMapper[T Stage] interface {
	DetermineStage(alert Alert) (T, error)
}
