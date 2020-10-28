package tritonhttp

type HttpServer struct {
	ServerPort string
	DocRoot    string
	MIMEPath   string
	MIMEMap    map[string]string
}

type HttpResponseHeader struct {
	InitialLine   string
	Date          string
	LastModified  string
	ContentType   string
	ContentLength string
	Connection    string
	Server        string
	FilePath      string
}

type HttpRequestHeader struct {
	InitialLine      string
	Host             string
	Connection       string
	IsBadRequest     bool
	IsPartialRequest bool
}
