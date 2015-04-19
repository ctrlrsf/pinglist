package main

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
)

// Start HTTP server
func startHTTPServer(listenIPPort string, hostRegistry *HostRegistry, influxContext *InfluxContext) {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		&rest.Route{"GET", "/hosts", func(w rest.ResponseWriter, r *rest.Request) {
			hostRegistry.mutex.RLock()
			w.WriteJson(&hostRegistry.hosts)
			hostRegistry.mutex.RUnlock()
		}},
		&rest.Route{"PUT", "/hosts/#address", func(w rest.ResponseWriter, r *rest.Request) {
			hostJson := HostJson{}
			err := r.DecodeJsonPayload(&hostJson)
			if err != nil {
				rest.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if hostJson.Address == "" {
				rest.Error(w, "Address required", 400)
				return
			}

			if len(hostJson.Description) > 200 {
				rest.Error(w, "Description too long", 400)
				return
			}

			if !ValidIPOrHost(hostJson.Address) {
				rest.Error(w, "Invalid address or format", http.StatusInternalServerError)
				return
			}

			h := &Host{Address: hostJson.Address, Description: hostJson.Description}
			hostRegistry.RegisterHost(h)
		}},
		&rest.Route{"DELETE", "/hosts/#address", func(w rest.ResponseWriter, r *rest.Request) {
			address := r.PathParam("address")

			if !hostRegistry.Contains(address) {
				rest.Error(w, "Host doesn't exist", http.StatusInternalServerError)
				return
			}

			hostRegistry.RemoveHost(address)
		}},
		&rest.Route{"GET", "/history/#address/#metric", func(w rest.ResponseWriter, r *rest.Request) {
			address := r.PathParam("address")
			metric := r.PathParam("metric")

			// Make sure we only accept metrics we can handle
			if metric != "status" && metric != "latency" {
				rest.Error(w, "Invalid metric", http.StatusInternalServerError)
				return
			}

			result, err := influxContext.Query(address)
			if err != nil {
				rest.Error(w, "Error getting history", http.StatusInternalServerError)
				return
			}

			resultsToLogEntryList(result)

			w.WriteJson(result)
		}},
	)

	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))

	http.Handle("/", http.FileServer(http.Dir("static/")))

	log.Fatal(http.ListenAndServe(listenIPPort, nil))
}
