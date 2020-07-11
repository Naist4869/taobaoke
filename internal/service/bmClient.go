package service

import (
	"crypto/tls"
	"net/http"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"
)

func NewBmClient() (client *bm.Client, cf func(), err error) {
	var (
		cfg struct {
			Server *bm.ServerConfig
			Client *bm.ClientConfig
		}
	)

	var (
		ct paladin.TOML
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("bm").UnmarshalTOML(&cfg); err != nil {
		return
	}
	client = bm.NewClient(cfg.Client)
	t1 := &http.Transport{
		//DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
		//	dialer := &net.Dialer{Timeout: time.Second * 5}
		//	var proxyDialer proxy.Dialer
		//	proxyAddr := "127.0.0.1:8888" // fiddler
		//	if proxyDialer, err = proxy.SOCKS5("tcp", proxyAddr, nil, dialer); err != nil {
		//		log.Printf("proxy.SOCKS5(%s) error(%v)", proxyAddr, err)
		//		return nil, err
		//	}
		//	var conn net.Conn
		//	if conn, err = proxyDialer.Dial(network, addr); err != nil {
		//		log.Printf("proxyDialer.Dial(%s,%s) error(%v)", network, addr, err)
		//		if conn, err = dialer.Dial(network, addr); err != nil {
		//			log.Printf("dialer.Dial(%s,%s) error(%v)", network, addr, err)
		//			return nil, err
		//		}
		//
		//	}
		//	tlsConn := tls.Client(conn, cfg)
		//	if err = tlsConn.Handshake(); err != nil {
		//		log.Printf("tlsConn.Handshake() error(%v)", err)
		//		return nil, err
		//	}
		//	if !cfg.InsecureSkipVerify {
		//		if err = tlsConn.VerifyHostname(cfg.ServerName); err != nil {
		//			log.Printf("tlsConn.VerifyHostname(%s) error(%v)", cfg.ServerName, err)
		//			return nil, err
		//		}
		//	}
		//
		//	state := tlsConn.ConnectionState()
		//
		//	//if state.NegotiatedProtocol != http2.NextProtoTLS {
		//	//	err = fmt.Errorf("http2: unexpected ALPN protocol(%s) expect(%s)", state.NegotiatedProtocol, http2.NextProtoTLS)
		//	//	return nil, err
		//	//}
		//	if !state.NegotiatedProtocolIsMutual {
		//		err = errors.New("http2: could not negotiate protocol mutually")
		//		return nil, err
		//	}
		//	return tlsConn, nil
		//},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	//t1 := &http.Transport{
	//	TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	//	MaxIdleConns:        100,
	//	MaxIdleConnsPerHost: 100,
	//	IdleConnTimeout:     30 * time.Second,
	//}
	//if err = http2.ConfigureTransport(t1); err != nil {
	//	return
	//}
	client.SetTransport(t1)
	cf = func() {}
	return
}
