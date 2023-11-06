package nanny

type NoRecorder struct{}

func (n NoRecorder) Record(string, any) CanRecord { return n }
func (n NoRecorder) Group(name string) CanRecord {
	return n
}
