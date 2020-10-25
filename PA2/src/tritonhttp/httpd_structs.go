package tritonhttp

type HttpServer struct {
	ServerPort string
	DocRoot    string
	MIMEPath   string
	MIMEMap    map[string]string
}

type HttpResponseHeader struct {
	Date          string
	LastModified  string
	ContentType   string
	ContentLength string
	Connection    string
	InitialLine   string
	Server        string
	FilePath      string
}

type HttpRequestHeader struct {
	Host        string
	Connection  string
	InitialLine string
}
