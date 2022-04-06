package swogrp

import "github.com/target/goalert/swo/swomsg"

type msgBuf struct {
	full  chan []swomsg.Message
	empty chan []swomsg.Message

	next chan swomsg.Message
}

func (buf *msgBuf) Append(msg swomsg.Message) {
	var msgs []swomsg.Message
	select {
	case msgs = <-buf.empty:
	case msgs = <-buf.full:
	}
	msgs = append(msgs, msg)
	buf.full <- msgs
}
func (buf *msgBuf) Next() <-chan swomsg.Message { return buf.next }

func newMsgBuf() *msgBuf {
	buf := &msgBuf{
		full:  make(chan []swomsg.Message, 1),
		empty: make(chan []swomsg.Message, 1),
		next:  make(chan swomsg.Message),
	}
	buf.empty <- nil
	go func() {
		for msgs := range buf.full {
			msg := msgs[0]
			if len(msgs) > 1 {
				buf.full <- msgs[1:]
			} else {
				buf.empty <- msgs[1:]
			}
			buf.next <- msg
		}
	}()
	return buf
}
