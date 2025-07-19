package audit

type Event struct {
	Kind                     string            `json:"kind"`
	APIVersion               string            `json:"apiVersion"`
	Level                    string            `json:"level"`
	AuditID                  string            `json:"auditID"`
	Stage                    string            `json:"stage"`
	RequestURI               string            `json:"requestURI"`
	Verb                     string            `json:"verb"`
	User                     User              `json:"user"`
	SourceIPs                []string          `json:"sourceIPs"`
	UserAgent                string            `json:"userAgent"`
	ObjectRef                ObjectRef         `json:"objectRef"`
	ResponseStatus           ResponseStatus    `json:"responseStatus"`
	RequestReceivedTimestamp string            `json:"requestReceivedTimestamp"`
	StageTimestamp           string            `json:"stageTimestamp"`
	Annotations              map[string]string `json:"annotations"`
}

type User struct {
	Username string              `json:"username"`
	Groups   []string            `json:"groups"`
	Extra    map[string][]string `json:"extra"`
}

type ObjectRef struct {
	Resource        string `json:"resource"`
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	UID             string `json:"uid"`
	APIGroup        string `json:"apiGroup"`
	APIVersion      string `json:"apiVersion"`
	ResourceVersion string `json:"resourceVersion"`
}

type ResponseStatus struct {
	Metadata map[string]interface{} `json:"metadata"`
	Code     int                    `json:"code"`
}
