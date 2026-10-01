package detector

type Call struct {
	Type  string
	Depth int
	To    string
	Input string
	Value string
	Error string
}

type Finding struct {
	Depth     int
	To        string
	Selector  string
	Label     string
	MethodID  int
	Supported bool
	Unknown   bool
}

type Classification struct {
	Calls       []Finding
	Unknowns    []Finding
	Unsupported []Finding
	Supported   bool
}
