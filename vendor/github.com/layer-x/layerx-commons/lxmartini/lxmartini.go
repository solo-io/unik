package lxmartini

import (
	"github.com/go-martini/martini"
	"github.com/Sirupsen/logrus"
	"net/http"
	"time"
	"github.com/layer-x/layerx-commons/lxerrors"
	"encoding/json"
)

func QuietMartini() *martini.ClassicMartini {
	r := martini.NewRouter()
	customMartini := martini.New()
	customMartini.Use(customLogger())
	customMartini.Use(martini.Recovery())
	customMartini.Use(martini.Static("public"))
	customMartini.MapTo(r, (*martini.Routes)(nil))
	customMartini.Action(r.Handle)
	return &martini.ClassicMartini{customMartini, r}
}

func customLogger() martini.Handler {
	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
		start := time.Now()
		addr := req.Header.Get("X-Real-IP")
		if addr == "" {
			addr = req.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = req.RemoteAddr
			}
		}

		logrus.Debugf("Started %s %s for %s", req.Method, req.URL.Path, addr)

		rw := res.(martini.ResponseWriter)
		c.Next()

		logrus.Debugf("Completed %v %s in %v\n", rw.Status(), http.StatusText(rw.Status()), time.Since(start))
	}
}

func Respond(res http.ResponseWriter, message interface{}) error {
	switch message.(type){
	case string:
		messageString := message.(string)
		data := []byte(messageString)
		_, err := res.Write(data)
		if err != nil {
			return lxerrors.New("writing data", err)
		}
		return nil
	case error:
		responseError := message.(error)
		http.Error(res, responseError.Error(), http.StatusInternalServerError)
		return nil
	}
	data, err := json.Marshal(message)
	if err != nil {
		return lxerrors.New("marshalling message to json", err)
	}
	_, err = res.Write(data)
	if err != nil {
		return lxerrors.New("writing data", err)
	}
	return nil
}
