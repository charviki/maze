package httputil

import (
	"sync"

	gorillaws "github.com/gorilla/websocket"
)

// RelayWebSocket 在两个连接之间双向转发消息。
// 任一方向读取或写入失败时都会终止转发，避免一端断开后另一端继续悬挂。
func RelayWebSocket(a, b *gorillaws.Conn) error {
	var (
		wg    sync.WaitGroup
		errCh = make(chan error, 2)
	)

	forward := func(src, dst *gorillaws.Conn) {
		defer wg.Done()
		for {
			msgType, msg, err := src.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			if err := dst.WriteMessage(msgType, msg); err != nil {
				errCh <- err
				return
			}
		}
	}

	wg.Add(2)
	go forward(a, b)
	go forward(b, a)

	err := <-errCh
	_ = a.Close()
	_ = b.Close()
	wg.Wait()
	return err
}
