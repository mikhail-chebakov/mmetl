package slack

type memoryStorage struct {
	threads map[string]*IntermediatePost
}

func (s *memoryStorage) LookupThread(threadTS string) *IntermediatePost {
	rootPost, ok := s.threads[threadTS]
	if !ok {
		return nil
	}
	return rootPost
}

func (s *memoryStorage) HasThread(threadTS string) bool {
	return s.threads[threadTS] != nil
}

func (s *memoryStorage) StoreThread(threadTS string, rootPost *IntermediatePost) {
	s.threads[threadTS] = rootPost
}

func (s *memoryStorage) GetChangedThreads() []*IntermediatePost {
	result := make([]*IntermediatePost, 0, len(s.threads))
	for _, post := range s.threads {
		result = append(result, post)
	}
	return result
}

func newMemoryStorage() ThreadsStorage {
	return &memoryStorage{
		threads: make(map[string]*IntermediatePost),
	}
}
