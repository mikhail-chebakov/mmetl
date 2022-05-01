package slack

type ThreadsStorage interface {
	LookupThread(threadTS string) *IntermediatePost
	HasThread(threadTS string) bool
	StoreThread(threadTS string, rootPost *IntermediatePost)
	GetChangedThreads() []*IntermediatePost
}
