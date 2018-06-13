package service

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/libgolang/one/clients"
	"github.com/libgolang/one/utils"

	"net/http"

	"github.com/libgolang/log"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
)

var (
	htmlTemplate        = utils.NewTemplate(htmlTemplateString)
	vhostTemplate       = utils.NewTemplate(vhostTemplateString)
	dockerProxyTemplate = utils.NewTemplate(dockerProxyTemplateString)
)

// Proxy interface for proxy
type Proxy interface {
	AddDockerProxySsl(name, hostIP string, httpRandPort int)
	RemoveDockerProxy(name string)
}

type proxy struct {
	publicIP     string
	privateIP    string
	baseDomain   string
	masterClient clients.MasterClient
	ticker       *time.Ticker
	proxyMap     map[string]*url.URL
}

// NewProxy constructor
func NewProxy(masterAddr, publicIP, privateIP, baseDomain string) Proxy {
	p := &proxy{}
	p.publicIP = publicIP
	p.privateIP = privateIP
	p.baseDomain = baseDomain
	p.masterClient = clients.NewMasterClient(masterAddr)
	p.ticker = time.NewTicker(time.Second * 30)
	p.proxyMap = make(map[string]*url.URL)
	go func() {
		for range p.ticker.C {
			p.checkWithMaster()
		}
	}()

	return p
}

// Start starts the ssl proxy and the http to https redirector
func (p *proxy) Start() {
	//
	//
	go func() {
		p.startHttpsRedirector()
	}()

	p.startSslProxy()
}

func (p *proxy) startSslProxy() {
	fwd, _ := forward.New()
	lb, _ := roundrobin.New(fwd)
	url1 = testutils.ParseURI("http://localhost:63450")
	lb.UpsertServer(url1)
	lb.UpsertServer(url2)
	buffer, _ := buffer.New(lb, buffer.Retry(`IsNetworkError() && Attempts() < 2`))

	proxyHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		hostname := req.URL.Hostname()
		bufferProxy := p.proxyMap[hostname]
		bufferProxy.ServeHTTP(w, req)
	})

	//
	//
	cfg := &tls.Config{}
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Panic("%s", err)
	}
	cfg.Certificates = append(cfg.Certificates, cert)
	cfg.BuildNameToCertificate()

	//
	//
	server := &http.Server{Addr: ":443", Handler: proxyHandler}
	server.TLSConfig = cfg
	server.ListenAndServeTLS("", "")
}

func (p *proxy) startHttpsRedirector() {
	//
	// Forwards incoming requests to whatever location URL points to, adds proper forwarding headers
	//
	redirect := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL.Scheme = "https"
		w.Header().Set("Location", req.URL.String())
		w.WriteHeader(301)
	})
	unsecureServer := &http.Server{Addr: ":80", Handler: redirect}
	unsecureServer.ListenAndServe()
}

func (p *proxy) checkWithMaster() {
	proxyDefMap := p.masterClient.GetProxyData()
	for name, proxyDef := range proxyDefMap {
		p.AddDockerProxySsl(name)
	}
}

func (p *proxy) RemoveDockerProxy(name string) {
	domain := fmt.Sprintf("%s.%s", name, p.baseDomain)
	availConfig := fmt.Sprintf("/etc/apache2/sites-enabled/%s.conf", domain)
	config := fmt.Sprintf("/etc/apache2/sites-available/%s.conf", domain)

	// a2dissite
	if utils.FileExists(availConfig) {
		_ = utils.ExecSilent("a2dissite", domain)
		p.restartApache()
	}

	// remove config and restart
	if utils.FileExists(config) {
		utils.Remove(config)
	}
}

func (p *proxy) AddDockerProxySsl(name, host string, port int) {
	domain := fmt.Sprintf("%s.%s", name, p.baseDomain)
	log.Info("Adding SSL Proxy for %s -> %s:%d", domain, host, port)

	p.requestSslCertificate(domain)

	log.Info("Creating proxy...")
	tpl := dockerProxyTemplate.Context().
		Set("listenIp", p.privateIP).
		Set("domain", domain).
		Set("host", host).
		Set("port", port).
		Parse()

	file := fmt.Sprintf("/etc/apache2/sites-available/%s.conf", domain)
	log.Debug("file: %s; tpl:%s\n", file, tpl)
	if err := ioutil.WriteFile(file, tpl, 0755); err != nil {
		panic(err)
	}
	_ = utils.ExecSilent("a2ensite", domain)
	p.restartApache()
}
func (p *proxy) requestSslCertificate(domain string) {
	log.Info("requesting ssl certificate for %s...", domain)

	//
	keyFile := fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", domain)
	/*
		certFile := fmt.Sprintf("/etc/letsencrypt/live/%s/cert.pem", domain)
		chainFile := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain)
	*/

	// if file already exists, then bail out
	if _, err := os.Stat(keyFile); !os.IsNotExist(err) {
		log.Info("...certificate already exists")
		return
	}

	//
	p.addVirtualHost(domain)

	//
	log.Info("...certbot...")
	err := utils.Exec("certbot", "certonly", "--webroot", "-w", fmt.Sprintf("/var/www/virtual/%s/htdocs", domain), "-d", domain)
	if err != nil {
		panic(fmt.Errorf("Unable to request ssl certificate: %s", err))
	}
	log.Info("...cert bot done")
	p.removeVirtualHost(domain)
}

func (p *proxy) addVirtualHost(domain string) {
	dir := fmt.Sprintf("/var/www/virtual/%s/htdocs", domain)

	//
	fmt.Printf("\tcreating vhost dir %s\n", dir)

	//
	if err := utils.Exec("mkdir", "-p", dir); err != nil {
		panic(err)
	}

	//
	byteArr := htmlTemplate.Context().Set("Domain", domain).Parse()
	if err := ioutil.WriteFile(path.Join(dir, "index.html"), byteArr, 0755); err != nil {
		panic(err)
	}

	//
	vhostConfig := fmt.Sprintf("/etc/apache2/sites-available/%s.conf", domain)
	tpl := vhostTemplate.Context().Set("domain", domain).Set("listenIp", p.publicIP).Parse()
	if err := ioutil.WriteFile(vhostConfig, tpl, 0755); err != nil {
		panic(err)
	}
	fmt.Printf("\tenabling site %s\n", domain)
	if err := utils.ExecSilent("a2ensite", domain); err != nil {
		panic(err)
	}
	p.restartApache()
}

func (p *proxy) removeVirtualHost(domain string) {
	fmt.Printf("\tdisabling site %s\n", domain)
	if err := utils.ExecSilent("a2dissite", domain); err != nil {
		panic(err)
	}
	/*
		if removeHostingDir {
			dir := fmt.Sprintf("/var/www/virtual/%s", domain);
			fmt.Printf("\tremoving dir %s\n", dir);
			utils.Exec("rm", "-fr", dir)
		}
	*/
	file := fmt.Sprintf("/etc/apache2/sites-available/%s.conf", domain)
	fmt.Printf("\tremoving file %s\n", file)
	if err := os.Remove(file); err != nil {
		panic(err)
	}
	p.restartApache()
}

func (p *proxy) restartApache() {
	log.Info("reloading apache")
	if err := utils.Exec("systemctl", "reload", "apache2.service"); err != nil {
		panic(err)
	}
}

const (
	htmlTemplateString = `<html>
<head>
<title>{{.Domain}}</title>
</head>
<body>
<h3>{{.Domain}}</h3>
</body>
</html>
`

	vhostTemplateString = `<VirtualHost {{.listenIp}}:80>
	ServerAdmin ricardo@rhamerica.com
	ServerName {{.domain}}
	ServerAlias {{.domain}}
	DocumentRoot /var/www/virtual/{{.domain}}/htdocs
	<Directory /var/www/virtual/{{.domain}}>
		#Options Indexes FollowSymLinks MultiViews
		Options FollowSymLinks MultiViews
		AllowOverride All
		Order allow,deny
		allow from all
	</Directory>
</VirtualHost>
`

	dockerProxyTemplateString = `<VirtualHost {{.listenIp}}:80>
	ServerAdmin ricardo@rhamerica.com
	ServerName {{.domain}}
	RewriteEngine On 
	RewriteCond %{HTTPS}  !=on 
	RewriteRule ^/?(.*) https://%{SERVER_NAME}/$1 [R,L] 
</VirtualHost>
<VirtualHost {{.listenIp}}:443>
	ServerAdmin ricardo@rhamerica.com
	ServerName {{.domain}}
	AllowEncodedSlashes NoDecode
	ProxyPreserveHost On
	ProxyPass "/"  "http://{{.host}}:{{.port}}/"
	ProxyPassReverse "/"  "http://{{.host}}:{{.port}}/"
	SSLEngine on
	SSLCertificateFile    /etc/letsencrypt/live/{{.domain}}/cert.pem
	SSLCertificateKeyFile /etc/letsencrypt/live/{{.domain}}/privkey.pem
	SSLCACertificateFile  /etc/letsencrypt/live/{{.domain}}/fullchain.pem
</VirtualHost>
`
)
