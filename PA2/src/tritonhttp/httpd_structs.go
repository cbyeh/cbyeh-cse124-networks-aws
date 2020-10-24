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
	ContentLength rune
	Connection    string
	InitialLine   string
	FilePath      string
}

type HttpRequestHeader struct {
	Host        string
	Connection  string
	InitialLine string
}
