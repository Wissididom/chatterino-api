package oembed

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/dyatlov/go-oembed/oembed"
)

var (
	templateOEmbed = template.Must(template.New("oEmbedTemplateTooltip").Parse(templateStringOEmbed))

	oEmbedCache = cache.New("oEmbed", load, 1*time.Hour)

	oEmbed = oembed.NewOembed()
)

func New(cfg config.APIConfig) (resolvers []resolver.CustomURLManager) {

	data, err := ioutil.ReadFile(cfg.OembedProvidersPath)

	if err != nil {
		log.Println("[oEmbed] No providers.json file found, won't do oEmbed parsing")
		return
	}

	if cfg.OembedFacebookAppID != "" && cfg.OembedFacebookAppSecret != "" {
		if err := initFacebookAppAccessToken(cfg.OembedFacebookAppID, cfg.OembedFacebookAppSecret); err != nil {
			log.Println("[oEmbed] error loading facebook app access token", err)
		} else {
			log.Println("[oEmbed] Extra rich info loading enabled for Instagram and Facebook")
		}
	}

	oEmbed.ParseProviders(bytes.NewReader(data))

	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			return oEmbed.FindItem(url.String()) != nil
		},
		Run: func(url *url.URL, r *http.Request) ([]byte, error) {
			apiResponse := oEmbedCache.Get(url.String(), r)
			return json.Marshal(apiResponse)
		},
	})

	return
}
