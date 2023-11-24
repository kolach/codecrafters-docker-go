package dockerhub

const (
	tokenURL    = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/%s:pull"
	manifestURL = "https://registry-1.docker.io/v2/library/%s/manifests/%s"
	layerURL    = "https://registry.hub.docker.com/v2/library/%s/blobs/%s"

	manifestListAcceptHeader = "application/vnd.docker.distribution.manifest.list.v2+json"
)

type tokenResponse struct {
	Token string `json:"token"`
}

type manifestMeta struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
	Platform  struct {
		Architecture string `json:"architecture"`
		Os           string `json:"os"`
	} `json:"platform"`
}

type manifestList struct {
	SchemaVersion int             `json:"schemaVersion"`
	MediaType     string          `json:"mediaType"`
	Manifests     []*manifestMeta `json:"manifests"`
}

type manifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"layers"`
}
