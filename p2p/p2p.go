package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p-routing"
	"io"

	"log"
	"time"
)

const topic = "things"

type Services struct {
	host    host.Host
	dht     *dht.IpfsDHT
	ps      *pubsub.PubSub
	Handler func(*pubsub.Message)
	say     chan string
	out		io.Writer
}

func NewServices(writer io.Writer) (s *Services, err error) {
	s = &Services{
		say: make(chan string),
		out: writer,
	}
	return
}

func (s *Services) log(message string) {
	if s.Handler != nil {
		s.Handler(&pubsub.Message{
			Message: &pubsub_pb.Message{
				From: []byte("system"),
				Data: []byte(message),
			},
		})
	} else {
		log.Printf("%s", message)
	}
}

func (s *Services) initHost(ctx context.Context, prvKey crypto.PrivKey) (err error) {

	s.host, err = libp2p.New(
		ctx,
		libp2p.Identity(prvKey),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelay(),
		libp2p.NATPortMap(),
		libp2p.Routing(s.dhtRoutingFactory),
	)
	if err != nil {
		return err
	}
	return nil
}

//this is kind of awkward
func (s *Services) dhtRoutingFactory(host2 host.Host) (routing.PeerRouting, error) {
	s.host = host2
	s.initDht(context.Background())
	return s.dht, nil
}

func (s *Services) initDht(ctx context.Context) (err error) {

	if s.dht != nil {
		return
	}

	s.dht, err = dht.New(ctx, s.host)
	if err != nil {
		s.log(fmt.Sprintf("error creating dht: %v", err))
		return
	}
	err = s.dht.Bootstrap(ctx)
	if err != nil {
		s.log(fmt.Sprintf("error bootstrapping dht: %v", err))
	}
	return
}

func (s *Services) initGossipSub(ctx context.Context) (err error) {
	s.ps, err = pubsub.NewGossipSub(ctx, s.host)
	return
}

func (s *Services) ShowSelf() {

	addresses, err := s.host.Network().InterfaceListenAddresses()
	if err != nil {
		panic(err)
	}
	for _, address := range addresses {
		log.Printf("address: %s/p2p/%s", address, s.host.ID().Pretty())
	}
}

func (s *Services) ShowPeers() {
	for i, peer := range s.host.Network().Peers() {
		log.Printf("peer %d: %s", i, peer)
	}
}

func (s *Services) ShowTopics() {
	for _, topic := range s.ps.GetTopics() {
		log.Printf("topic: %s", topic)
		for _, topicPeer := range s.ps.ListPeers(topic) {
			topicPeerinfo := s.host.Peerstore().PeerInfo(topicPeer)
			log.Printf(" peer: %v", topicPeerinfo)

		}
	}
}

func (s *Services) ShowFoo() {

}

func (s *Services) subscribe(ctx context.Context) (err error) {

	mySubscription, err := s.ps.Subscribe(topic)
	if err != nil {
		s.log(fmt.Sprintf("error subscribing: %v", err))
		return err
	} else {
		fmt.Fprint(s.out, "subscribed")
	}

	go func(ctx context.Context) {
		for {
			m, err := mySubscription.Next(ctx)
			if err != nil {
				s.log(fmt.Sprintf("error subscribing: %v", err))
			}
			if s.Handler != nil {
				s.Handler(m)
			} else {
				fmt.Fprintf(s.out, "%s: %s\n", m.GetFrom().String(), m.Data)
			}

		}
	}(ctx)
	return
}

func (s *Services) dhtHook(ctx context.Context) (err error) {
	routingDiscovery := discovery.NewRoutingDiscovery(s.dht)
	discovery.Advertise(ctx, routingDiscovery, topic)
	peerChan, err := routingDiscovery.FindPeers(ctx, topic)
	if err != nil {
		fmt.Fprintf(s.out, "error during discovery: %v\n", err)
	}
	for peer := range peerChan {
		if peer.ID == s.host.ID() {
			continue
		}
		fmt.Fprintf(s.out, "%s joined chat\n", peer.ID)
		err = s.host.Connect(ctx, peer)
		if err != nil {
			fmt.Fprintf(s.out, "dht error connecting to %v\n", peer.ID)
		}
	}
	fmt.Fprint(s.out, "done discovering peers\n")
	return
}

func (s *Services) Init(ctx context.Context) (err error) {

	prvKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return
	}
	//s.log("identity created")

	err = s.initHost(ctx, prvKey)
	if err != nil {
		return err
	}
	//s.log("host initialized")

	err = s.initGossipSub(ctx)
	if err != nil {
		return err
	}
	//s.log("chat started")

	err = s.subscribe(ctx)
	if err != nil {
		return err
	}
	//s.log("subscribe started")

	err = s.initDht(ctx)
	if err != nil {
		return err
	}
	//s.log("internet discovery started")

	s.bootstrap(ctx)
	go func() {
		time.Sleep(10 * time.Second)
		s.dhtHook(ctx)
	}()

	return
}

func (s *Services) Run(ctx context.Context) (err error) {
	for {
		select {
		case message := <-s.say:
			s.handleSay(ctx, message)
		}
	}
}

func (s *Services) handleSay(ctx context.Context, message string) {
	err := s.ps.Publish(topic, []byte(message))
	if err != nil {
		s.log(fmt.Sprintf("error saying message %s, %v", message, err))
	}
}

func (s *Services) Say(message string) {
	s.say <- message
}
