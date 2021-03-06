package shortener

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"strings"
	"shortener/models"
	"net/http"
	"time"
)

const BaseUrl  = "http://localhost:8080"


func render(w http.ResponseWriter, context interface{}, template_path string) {
	tmpl := template.New("PAGE")

	tmpl = template.Must(template.ParseFiles(template_path))
	template_paths := strings.Split(template_path, "/")
	template_name := template_paths[len(template_paths)-1]

	if err := tmpl.ExecuteTemplate(w, template_name, context); err != nil {
		fmt.Fprint(w, err)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	render(w, nil, "shortener/templates/index.html")
}

func CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	originUrl := r.FormValue("originUrl")
	log.Print(originUrl)
	ctx := appengine.NewContext(r)

	entity := models.UrlHistory{OriginalUrl:originUrl, Created:time.Now()}
	key := entity.CreateKey(ctx)
	if _, err := datastore.Put(ctx, key, &entity); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	shorterUrl := BaseUrl + "/" + entity.ShortUrl
	item := &memcache.Item{
		Key:   entity.ShortUrl,
		Value: []byte(originUrl),
		Expiration: time.Second*60*15,
	}
	if err := memcache.Add(ctx, item); err != nil {
		fmt.Println(err)
	}
	context := map[string]interface{}{
		"resultUrl": shorterUrl,
	}
	render(w, context, "shortener/templates/result.html")
}

func OriginalRedirect(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	ctx := appengine.NewContext(r)
	log.Println(vars["urlHash"])
	urlItem, err := memcache.Get(ctx, vars["urlHash"])
	var url string
	if err != nil{
		entity := new(models.UrlHistory)
		k := datastore.NewKey(ctx, "UrlHistory", vars["urlHash"], 0, nil)
		if err := datastore.Get(ctx, k, *entity); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		url = entity.OriginalUrl
	} else {
		url = string(urlItem.Value)
	}
	if !strings.HasPrefix(url, "http"){
		url = "http://" + url
	}
	http.Redirect(w, r, url , http.StatusSeeOther)
}
