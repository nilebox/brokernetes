package brokerserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/nilebox/brokernetes/pkg/broker/brokerapi"
	"github.com/nilebox/brokernetes/pkg/broker/controller"
	"github.com/nilebox/brokernetes/pkg/broker/util"
	b9s_util "github.com/nilebox/brokernetes/pkg/util"

	"github.com/deckarep/golang-set"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
)

type server struct {
	controller controller.Controller
}
type failureResponseBody struct {
	Err         string `json:"error,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateHandler creates Broker HTTP handler based on an implementation
// of a controller.Controller interface.
func createHandler(ctx context.Context, c controller.Controller) http.Handler {
	s := server{
		controller: c,
	}

	var router = mux.NewRouter()

	router.HandleFunc("/v2/catalog", s.catalog).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}/last_operation", s.getServiceInstanceStatus).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}", s.createServiceInstance).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{instance_id}", s.updateServiceInstance).Methods("PATCH")
	router.HandleFunc("/v2/service_instances/{instance_id}", s.removeServiceInstance).Methods("DELETE")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", s.bind).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", s.unBind).Methods("DELETE")
	router.HandleFunc("/healthcheck", s.healthcheck).Methods("GET")

	n := negroni.New()

	// This is negroni's recovery function, pulled out into some handler code
	// so we can replace the logger with something that isn't hardcoded to [negroni]
	n.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		defer func() {
			if err := recover(); err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, false)]

				f := "PANIC: %s\n%s"
				log.Printf(f, err, stack)
				fmt.Fprintf(rw, f, err, stack)
			}
		}()

		next(rw, r)
	})

	n.UseHandler(router)
	return n
}

// Run creates the HTTP handler based on an implementation of a
// controller.Controller interface, and begins to listen on the specified port.
func Run(ctx context.Context, addr string, c controller.Controller) error {
	log.Printf("Starting server on %s", addr)
	srv := &http.Server{
		Addr:    addr,
		Handler: createHandler(ctx, c),
	}
	return b9s_util.StartStopServer(ctx, srv, 3*time.Second)
}

func handleError(err error, w http.ResponseWriter) {
	if v := controller.GetControllerError(err); v != nil {
		util.WriteResponse(w, v.Code, failureResponseBody{
			Err:         v.Err,
			Description: v.Description,
		})
		return
	}

	log.Printf("Internal Server Error: %+v", err)
	util.WriteErrorResponse(w, http.StatusInternalServerError, err)
}

func (s *server) catalog(w http.ResponseWriter, r *http.Request) {
	log.Print("Get Service Broker Catalog...")

	if result, err := s.controller.Catalog(r.Context()); err == nil {
		util.WriteResponse(w, http.StatusOK, result)
	} else {
		handleError(err, w)
	}
}

func (s *server) getServiceInstanceStatus(w http.ResponseWriter, r *http.Request) {
	instanceID := mux.Vars(r)["instance_id"]
	q := r.URL.Query()
	serviceID := first(q["service_id"])
	planID := first(q["plan_id"])
	operation := first(q["operation"])
	log.Printf("GetServiceInstanceStatus ... %s", instanceID)

	if result, err := s.controller.GetServiceInstanceStatus(r.Context(), instanceID, serviceID, planID, operation); err == nil {
		util.WriteResponse(w, http.StatusOK, result)
	} else {
		handleError(err, w)
	}
}

func (s *server) createServiceInstance(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["instance_id"]
	log.Printf("CreateServiceInstance %s...", id)

	var req brokerapi.CreateServiceInstanceRequest
	if err := util.BodyToObject(r, &req); err != nil {
		log.Printf("[ERROR] error unmarshalling: %v", err)
		util.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	missing := appendMissingParam(req.OrgID, "OrgID", nil)
	missing = appendMissingParam(req.PlanID, "PlanID", missing)
	missing = appendMissingParam(req.ServiceID, "ServiceID", missing)
	missing = appendMissingParam(req.SpaceID, "SpaceID", missing)
	if len(missing) != 0 {
		util.WriteErrorResponse(w, http.StatusBadRequest, errors.Errorf("missing parameters: %s", missing))
		return
	}

	queryParams := r.URL.Query()
	acceptsIncomplete := queryParams.Get("accepts_incomplete") == "true"

	result, err := s.controller.CreateServiceInstance(r.Context(), id, acceptsIncomplete, &req)
	if err != nil {
		handleError(err, w)
		return
	}

	if result.Async {
		util.WriteResponse(w, http.StatusAccepted, result)
	} else {
		util.WriteResponse(w, http.StatusCreated, result)
	}
}

func (s *server) updateServiceInstance(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["instance_id"]
	log.Printf("UpdateServiceInstance %s...", id)

	var req brokerapi.UpdateServiceInstanceRequest
	if err := util.BodyToObject(r, &req); err != nil {
		log.Printf("[ERROR] error unmarshalling: %v", err)
		util.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	missing := appendMissingParam(req.ServiceID, "ServiceID", nil)
	if len(missing) != 0 {
		util.WriteErrorResponse(w, http.StatusBadRequest, errors.Errorf("missing parameters: %s", missing))
		return
	}

	queryParams := r.URL.Query()
	acceptsIncomplete := queryParams.Get("accepts_incomplete") == "true"

	result, err := s.controller.UpdateServiceInstance(r.Context(), id, acceptsIncomplete, &req)
	if err != nil {
		handleError(err, w)
		return
	}

	if result.Async {
		util.WriteResponse(w, http.StatusAccepted, result)
	} else {
		util.WriteResponse(w, http.StatusOK, result)
	}
}

func (s *server) removeServiceInstance(w http.ResponseWriter, r *http.Request) {
	instanceID := mux.Vars(r)["instance_id"]
	q := r.URL.Query()
	serviceID := first(q["service_id"])
	planID := first(q["plan_id"])
	acceptsIncomplete := first(q["accepts_incomplete"]) == "true"
	log.Printf("RemoveServiceInstance %s...", instanceID)

	missing := appendMissingParam(serviceID, "service_id", nil)
	missing = appendMissingParam(planID, "plan_id", missing)
	if len(missing) != 0 {
		util.WriteErrorResponse(w, http.StatusBadRequest, errors.Errorf("missing query params: %s", missing))
		return
	}

	result, err := s.controller.RemoveServiceInstance(r.Context(), instanceID, serviceID, planID, acceptsIncomplete)
	if err != nil {
		handleError(err, w)
		return
	}

	if result.Async {
		util.WriteResponse(w, http.StatusAccepted, result)
	} else {
		util.WriteResponse(w, http.StatusOK, result)
	}
}

func (s *server) bind(w http.ResponseWriter, r *http.Request) {
	bindingID := mux.Vars(r)["binding_id"]
	instanceID := mux.Vars(r)["instance_id"]

	log.Printf("Bind binding_id=%s, instance_id=%s", bindingID, instanceID)

	var req brokerapi.BindingRequest

	if err := util.BodyToObject(r, &req); err != nil {
		log.Printf("[ERROR] Failed to unmarshall request: %v", err)
		util.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	missing := appendMissingParam(req.ServiceID, "ServiceID", nil)
	missing = appendMissingParam(req.PlanID, "PlanID", missing)
	if len(missing) != 0 {
		util.WriteErrorResponse(w, http.StatusBadRequest, errors.Errorf("missing parameters: %s", missing))
		return
	}

	if result, err := s.controller.Bind(r.Context(), instanceID, bindingID, &req); err == nil {
		util.WriteResponse(w, http.StatusOK, result)
	} else {
		handleError(err, w)
	}
}

func (s *server) unBind(w http.ResponseWriter, r *http.Request) {
	instanceID := mux.Vars(r)["instance_id"]
	bindingID := mux.Vars(r)["binding_id"]
	q := r.URL.Query()
	serviceID := first(q["service_id"])
	planID := first(q["plan_id"])
	log.Printf("UnBind: Service instance guid: %s:%s", bindingID, instanceID)

	missing := appendMissingParam(serviceID, "service_id", nil)
	if len(missing) != 0 {
		util.WriteErrorResponse(w, http.StatusBadRequest, errors.Errorf("missing query params: %s", missing))
		return
	}

	if err := s.controller.UnBind(r.Context(), instanceID, bindingID, serviceID, planID); err == nil {
		util.WriteEmptyResponse(w, http.StatusOK)
	} else {
		handleError(err, w)
	}
}

func (s *server) healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func first(values []string) string {
	// TODO do proper error handling and validation
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func appendMissingParam(s string, name string, missing []string) []string {
	if s == "" {
		return append(missing, name)
	}
	return missing
}
