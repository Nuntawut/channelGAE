package index

import (
    "appengine"
    "appengine/user"
    "appengine/channel"
    "http"
    "template"
    "os"
    "fmt"
    //"strconv"
)


var (
        mainTemplate    *template.Template
        mainTemplateErr os.Error
)

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
    
	tok, err := channel.Create(c, u.Id)
	if err != nil {
		http.Error(w, "Couldn't create Channel", http.StatusInternalServerError)
        c.Errorf("channel.Create: %v", err)
        return
    }
    
    err = mainTemplate.Execute(w, map[string]string{
        "token":    tok,
        "me":       u.Id,
    })
    
    if err != nil {
        c.Errorf("mainTemplate: %v", err)
    }
}

func messageReceived(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
	u := user.Current(c)
	
	channel.SendJSON(c, u.Id, map[string]string{
        "reply_message":       "Hello, Nuntawut",
    })
    
}