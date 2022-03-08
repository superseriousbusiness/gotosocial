package mutexes

// func init() {
// 	log.SetFlags(log.Flags() | log.Lshortfile)
// }

// type debugMutex sync.Mutex

// func (mu *debugMutex) Lock() {
// 	log.Output(2, "Lock()")
// 	(*sync.Mutex)(mu).Lock()
// }

// func (mu *debugMutex) Unlock() {
// 	log.Output(2, "Unlock()")
// 	(*sync.Mutex)(mu).Unlock()
// }

// type debugRWMutex sync.RWMutex

// func (mu *debugRWMutex) Lock() {
// 	log.Output(2, "Lock()")
// 	(*sync.RWMutex)(mu).Lock()
// }

// func (mu *debugRWMutex) Unlock() {
// 	log.Output(2, "Unlock()")
// 	(*sync.RWMutex)(mu).Unlock()
// }

// func (mu *debugRWMutex) RLock() {
// 	log.Output(2, "RLock()")
// 	(*sync.RWMutex)(mu).RLock()
// }

// func (mu *debugRWMutex) RUnlock() {
// 	log.Output(2, "RUnlock()")
// 	(*sync.RWMutex)(mu).RUnlock()
// }
