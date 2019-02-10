package p2p

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-peerstore"
	"sync"
	"time"
)

func (s *Services) bootstrap(ctx context.Context) {

	ready := false
	lock := sync.Mutex{}
	cond := sync.NewCond(&lock)

	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerInfo, _ := peerstore.InfoFromP2pAddr(peerAddr)
		ctx, _ = context.WithTimeout(ctx, 5*time.Second)

		go func(ctx context.Context) {
			err := s.host.Connect(ctx, *peerInfo)
			if err != nil {
				s.log(fmt.Sprintf("error connecting to bootstrap peer %s: %v", peerInfo.ID, err))
			} else {
				s.log(fmt.Sprintf("connected to bootstrap peer %s", peerInfo.ID))
				lock.Lock()
				ready = true
				cond.Broadcast()
				lock.Unlock()
			}
		}(ctx)
	}
	for {
		lock.Lock()
		if ready {
			lock.Unlock()
			break
		} else {
			cond.Wait()
			lock.Unlock()
		}

	}

}
