package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleError(err)
	
	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type Loadbalancer struct {
	port            string
	roundRobinCOunt int
	servers         []Server
}

func NewLoadBalancer(port string, servers []Server) *Loadbalancer {
	return &Loadbalancer{
		port:            port,
		roundRobinCOunt: 0,
		servers:         servers,
	}
}

func handleError(err error) {
	
	if err != nil {
		fmt.Printf("Error is : %v\n", err)
		os.Exit(1)
	}
}

func (s *simpleServer) Address() string {return s.addr}

func (s *simpleServer) IsAlive() bool{return true}

func (s *simpleServer) Serve(w http.ResponseWriter, r *http.Request){
	s.proxy.ServeHTTP(w,r) 
}

func (lb *Loadbalancer) getNextAvailableServer() Server{
	server:= lb.servers[lb.roundRobinCOunt%len(lb.servers)]
	for !server.IsAlive(){
		lb.roundRobinCOunt++
		server = lb.servers[lb.roundRobinCOunt%len(lb.servers)]
	}
	lb.roundRobinCOunt++
	return server
}

func (lb *Loadbalancer) serveProxy(w http.ResponseWriter, r *http.Request) {
	targetServer:= lb.getNextAvailableServer()
	fmt.Printf("Forwarding requests through address %q\n", targetServer.Address())
	targetServer.Serve(w, r)
}

func main() {
	servers := []Server{
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.google.com"),
		newSimpleServer("https://www.linkedin.com"),
		//You will get error on all three sites becasue you are 
		//opening the url from localhost, which is prohibited from 
		//the servers running these domains
	}

	lb := NewLoadBalancer("3000", servers)
	handleRedirect := func(w http.ResponseWriter, r *http.Request){
		lb.serveProxy(w,r)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("Serving requests at 'localhost : %s\n", lb.port)
	http.ListenAndServe(":" +lb.port, nil)
}
