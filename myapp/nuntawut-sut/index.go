package index

import (
    "appengine"
    "appengine/user"
    "appengine/channel"
    "appengine/datastore"
    "appengine/memcache"
    "http"
    "template"
    "os"
    "fmt"
)

var (
        mainTemplate    *template.Template
        mainTemplateErr os.Error
)

type Client struct {
	ClientID string
}

func init() {
    http.HandleFunc("/", main)
	http.HandleFunc("/msg", messageReceived )

    mainTemplate = template.New(nil)
    mainTemplate.SetDelims("{{", "}}")
    if err := mainTemplate.ParseFile("main.html"); err != nil {
    	mainTemplateErr = fmt.Errorf("tmpl.ParseFile failed: %v", err)
        return
    }
}


func main(w http.ResponseWriter, r *http.Request) {
   c := appengine.NewContext(r)
   u := user.Current(c)
	
   if u == nil {
        url, err := user.LoginURL(c, r.URL.String())
        if err != nil {
            http.Error(w, err.String(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Location", url)
        w.WriteHeader(http.StatusFound)
        return
    }
    
    tok, err := AddClient(c, u.Id)
    if err != nil {
        http.Error(w, err.String(), 500)
        return
    }
    
    err = mainTemplate.Execute(w, map[string]string{
        "token":    	tok,
        "client_id":	u.Id,
		"client_email":	u.Email,
    })
    
    if err != nil {
        c.Errorf("mainTemplate: %v", err)
    }
    
}

func AddClient(c appengine.Context, id string) (string, os.Error){

	q := datastore.NewQuery("Client")
	var gg []*Client
	var check = 0 
	
    if _, err := q.GetAll(c, &gg);err != nil {
    	return "",err
    }
    
	for _, client := range gg {
			if client.ClientID == id {
				check = check + 1
			}
    }
    
	if check == 0 {
		key := datastore.NewIncompleteKey("Client")
		client := Client{ClientID: id}
			
	    _, err := datastore.Put(c, key, &client)
	    if err != nil {
	        return "",err
	    }
	}
	
	memcache.Delete(c, "sut")
	
	return channel.Create(c, id)

}

func messageReceived(w http.ResponseWriter, r *http.Request) {
	
	var clients []Client
	
	c := appengine.NewContext(r)
	
	message := r.FormValue("name")+": "+r.FormValue("message")	

	_, err := memcache.JSON.Get(c, "sut", &clients)
	if err != nil && err != memcache.ErrCacheMiss {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	if err == memcache.ErrCacheMiss {
		q := datastore.NewQuery("Client")
		_, err = q.GetAll(c, &clients)
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		err = memcache.JSON.Set(c, &memcache.Item{
			Key: "sut", Object: clients,
		})
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
	}
	
	for _, client := range clients {
		channel.SendJSON(c, client.ClientID, map[string]string{
				"reply_message":message,
			})
	}
    
}
