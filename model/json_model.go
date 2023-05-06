package model

type Methods struct {
	rpcFunction string `json:"rpc_function"`
	httpMethod  string `json:"http_method"`
	httpPath    string `json:"http_path"`
}

type Api struct {
	name    string     `json:"name"`
	methods []*Methods `json:"methods"`
}

type Conf struct {
	serviceName   string `json:"service_name"`
	listenPort    int    `json:"listen_port"`
	listenIp      string `json:"listen_ip"`
	consulAddRess string `json:"consul_address"`
	api           []*Api `json:"api"`
}

func (c *Conf) GetServiceName() string {
	return c.serviceName
}

func (c *Conf) GetListenPort() int {
	return c.listenPort
}

func (c *Conf) GetConsulAddRess() string {
	return c.consulAddRess
}

func (c *Conf) GetApi() []*Api {
	return c.api
}

func (a *Api) GetName() string {
	return a.name
}

func (a *Api) GetMethods() []*Methods {
	return a.methods
}

func (m *Methods) GetRpcFunction() string {
	return m.rpcFunction
}

func (m *Methods) GetHttpMethod() string {
	return m.httpMethod
}

func (m *Methods) GetHttpPath() string {
	return m.httpPath
}
