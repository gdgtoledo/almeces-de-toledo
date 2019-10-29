package main

import (
	"net/http"

	"github.com/gdgtoledo/linneo/plants"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.elastic.co/apm/module/apmgin"
	"go.elastic.co/apm/module/apmlogrus"
)

var routes = map[string]string{"plants": "/plants", "plant": "/plants/:id"}

func init() {
	// apmlogrus.Hook will send "error", "panic", and "fatal"
	// level log messages to Elastic APM.
	log.AddHook(&apmlogrus.Hook{})
}

func handleSearchItems(c *gin.Context) {
	// apmlogrus.TraceContext extracts the transaction and span (if any) from the given context,
	// and returns logrus.Fields containing the trace, transaction, and span IDs.
	traceContextFields := apmlogrus.TraceContext(c)
	log.WithFields(traceContextFields).Debug("handling request")

	searchQueryByIndexName := plants.SearchQueryByIndexName{
		IndexName: "plants",
		Query:     map[string]interface{}{},
		Context:   c,
	}

	res, err := plants.Search(searchQueryByIndexName)

	log.WithFields(log.Fields{
		"result": res,
	}).Info("Query Result")

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error querying database")
	}

	hits := res["hits"].(map[string]interface{})["hits"].([]interface{})

	if len(hits) == 0 {
		c.String(http.StatusNoContent, "There are no plants in the primary storage")
	} else {
		c.String(http.StatusOK, "YAY! There are %d plants in the primary storage", len(hits))
	}
}

func handleDeleteItem(c *gin.Context) {
	var plant plants.Model
	result, err := plants.Delete(plant.ID)

	log.WithFields(log.Fields{
		"result": result,
	}).Info("Delete Query Result")

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error deleting a plant")
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(apmgin.Middleware(r))

	r.GET(routes["plants"], handleSearchItems)
	r.DELETE(routes["plant"], handleDeleteItem)

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
