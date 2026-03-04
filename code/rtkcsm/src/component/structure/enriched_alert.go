package structure

type EnrichedAlert[T Stage] struct {
	Alert
	MetaStage T
}
