package service

func (s *UserService) GetPool() *WorkerPool {
	return s.pool
}
