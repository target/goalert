package ctxlock

//	Unlock will release a lock for the given ID.
//
//	If there are any pending locks, the next one will be granted.
//
// Note: Unlock will panic if called more times than Lock.
func (l *IDLocker[K]) Unlock(id K) {
	l.mx.Lock()
	defer l.mx.Unlock()

	for {
		if len(l.queue[id]) == 0 {
			// since there is nobody waiting, we can just decrement the count
			l.count[id]--
			if l.count[id] < 0 {
				panic("count < 0")
			}
			if l.count[id] == 0 {
				delete(l.count, id) // cleanup so the map doesn't grow forever
			}
			return
		}

		ch := l.queue[id][0]
		l.queue[id] = l.queue[id][1:]
		if len(l.queue[id]) == 0 {
			delete(l.queue, id)
		}

		_, ok := <-ch
		// Use the negative flow so that code coverage can detect both the
		// continue and return case.
		if !ok {
			// The channel was closed and we need to try again.
			//
			// This is rare but can happen if Unlock is called
			// after a Lock context has been canceled but beats
			// the Lock's cleanup code to the Mutex.
			//
			// Commenting out the continue will cause the test
			// TestIDLocker_Unlock_Abandoned to fail.
			continue
		}

		// If the channel is still open, someone got the lock.
		return
	}
}
