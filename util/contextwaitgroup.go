package util

import "context"

type ContextWaitGroup struct {
	count int

	ctx context.Context
	nCh chan int
	wCh chan chan struct{}

	notify []chan struct{}
}

func NewContextWaitGroup(ctx context.Context) *ContextWaitGroup {
	wg := &ContextWaitGroup{
		ctx: ctx,
		nCh: make(chan int),
		wCh: make(chan chan struct{}),
	}
	go wg.loop()
	return wg
}
func (c *ContextWaitGroup) loop() {
	done := func() {
		for _, ch := range c.notify {
			close(ch)
		}
		c.notify = nil
	}
	defer done()
mainLoop:
	for {
		select {
		case ch := <-c.wCh:
			if c.count == 0 {
				close(ch)
				continue
			}
			c.notify = append(c.notify, ch)
		case <-c.ctx.Done():
			break mainLoop
		case n := <-c.nCh:
			c.count += n
			if c.count == 0 {
				done()
				continue
			}
			if c.count < 0 {
				panic("Done() called too many times")
			}
		}
	}

cleanup:
	for {
		select {
		case <-c.nCh:
		case ch := <-c.wCh:
			close(ch)
		default:
			break cleanup
		}
	}
}
func (c *ContextWaitGroup) Add(n int) {
	if c.ctx.Err() != nil {
		return
	}
	c.nCh <- n
}
func (c *ContextWaitGroup) WaitCh() <-chan struct{} {
	ch := make(chan struct{})
	c.wCh <- ch
	return ch
}
func (c *ContextWaitGroup) Done() {
	if c.ctx.Err() != nil {
		return
	}
	c.nCh <- -1
}
func (c *ContextWaitGroup) Wait() {
	<-c.WaitCh()
}
