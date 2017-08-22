package server

import (
	"github.com/artpar/api2go"
	"github.com/jmoiron/sqlx"
	"gopkg.in/gin-gonic/gin.v1"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"github.com/artpar/goms/server/resource"
	"github.com/artpar/goms/server/auth"
	"encoding/json"
	"net/http"
	"gopkg.in/Masterminds/squirrel.v1"
)

func CreateEventHandler(initConfig *resource.CmsConfig, fsmManager resource.FsmManager, cruds map[string]*resource.DbResource, db *sqlx.DB) func(context *gin.Context) {

	return func(gincontext *gin.Context) {

		currentUserReferenceId := gincontext.Request.Context().Value("user_id").(string)
		currentUsergroups := gincontext.Request.Context().Value("usergroup_id").([]auth.GroupPermission)

		pr := &http.Request{}
		pr.Method = "GET"
		req := api2go.Request{
			PlainRequest: gincontext.Request,
			QueryParams:  map[string][]string{},
		}

		objectStateMachineId := gincontext.Param("objectStateId")
		typename := gincontext.Param("typename")

		objectStateMachineResponse, err := cruds[typename+"_state"].FindOne(objectStateMachineId, req)
		if err != nil {
			log.Errorf("Failed to get object state machine: %v", err)
			gincontext.AbortWithError(400, err)
			return
		}

		objectStateMachine := objectStateMachineResponse.Result().(*api2go.Api2GoModel)

		stateObject := objectStateMachine.Data

		var subjectInstanceModel *api2go.Api2GoModel
		var stateMachineDescriptionInstance *api2go.Api2GoModel

		for _, included := range objectStateMachine.Includes {
			casted := included.(*api2go.Api2GoModel)
			if casted.GetTableName() == typename {
				subjectInstanceModel = casted
			} else if casted.GetTableName() == "smd" {
				stateMachineDescriptionInstance = casted
			}

		}

		stateMachineId := objectStateMachine.GetID()
		eventName := gincontext.Param("eventName")

		stateMachinePermission := cruds["smd"].GetRowPermission(stateMachineDescriptionInstance.GetAllAsAttributes())

		if !stateMachinePermission.CanExecute(currentUserReferenceId, currentUsergroups) {
			gincontext.AbortWithStatus(403)
			return
		}

		nextState, err := fsmManager.ApplyEvent(subjectInstanceModel.GetAllAsAttributes(), resource.NewStateMachineEvent(stateMachineId, eventName))
		if err != nil {
			gincontext.AbortWithError(400, err)
			return
		}

		stateObject["current_state"] = nextState

		s, v, err := squirrel.Update(typename + "_state").Set("current_state", nextState).Where(squirrel.Eq{"reference_id": stateMachineId}).ToSql()

		_, err = db.Exec(s, v...)
		if err != nil {
			gincontext.AbortWithError(500, err)
			return
		}

		gincontext.AbortWithStatus(200)

	}

}

func CreateEventStartHandler(fsmManager resource.FsmManager, cruds map[string]*resource.DbResource, db *sqlx.DB) func(context *gin.Context) {

	return func(gincontext *gin.Context) {

		uId := gincontext.Request.Context().Value("user_id")
		var currentUserReferenceId string
		currentUsergroups := make([]auth.GroupPermission, 0)

		if uId != nil {
			currentUserReferenceId = uId.(string)
		}
		ugId := gincontext.Request.Context().Value("usergroup_id")
		if ugId != nil {
			currentUsergroups = ugId.([]auth.GroupPermission)
		}

		jsBytes, err := ioutil.ReadAll(gincontext.Request.Body)
		if err != nil {
			log.Errorf("Failed to read post body: %v", err)
			gincontext.AbortWithError(400, err)
			return
		}

		m := make(map[string]interface{})
		json.Unmarshal(jsBytes, &m)

		typename := m["typeName"].(string)
		refId := m["referenceId"].(string)
		stateMachineId := gincontext.Param("stateMachineId")

		pr := &http.Request{}
		pr.Method = "GET"
		req := api2go.Request{
			PlainRequest: pr,
			QueryParams:  map[string][]string{},
		}

		pr = pr.WithContext(gincontext.Request.Context())
		response, err := cruds["smd"].FindOne(stateMachineId, req)
		if err != nil {
			gincontext.AbortWithError(400, err)
			return
		}

		stateMachineInstance := response.Result().(*api2go.Api2GoModel)
		stateMachineInstanceProperties := stateMachineInstance.GetAttributes()
		stateMachinePermission := cruds["smd"].GetRowPermission(stateMachineInstance.GetAllAsAttributes())

		if !stateMachinePermission.CanExecute(currentUserReferenceId, currentUsergroups) {
			gincontext.AbortWithStatus(403)
			return
		}

		subjectInstanceResponse, err := cruds[typename].FindOne(refId, req)
		if err != nil {
			gincontext.AbortWithError(400, err)
			return
		}
		subjectInstanceModel := subjectInstanceResponse.Result().(*api2go.Api2GoModel).GetAttributes()

		newStateMachine := make(map[string]interface{})

		newStateMachine["current_state"] = stateMachineInstanceProperties["initial_state"]
		newStateMachine[typename+"_smd"] = stateMachineInstanceProperties["reference_id"]
		newStateMachine["is_state_of_"+typename] = subjectInstanceModel["reference_id"]
		newStateMachine["permission"] = "750"

		req.PlainRequest.Method = "POST"

		resp, err := cruds[typename+"_state"].Create(api2go.NewApi2GoModelWithData(typename+"_state", nil, 0, nil, newStateMachine), req)

		//s, v, err := squirrel.Insert(typename + "_state").SetMap(newStateMachine).ToSql()
		//if err != nil {
		//  log.Errorf("Failed to create state insert query: %v", err)
		//  gincontext.AbortWithError(500, err)
		//}

		//_, err = db.Exec(s, v...)
		if err != nil {
			log.Errorf("Failed to execute state insert query: %v", err)
			gincontext.AbortWithError(500, err)
			return
		}

		gincontext.JSON(200, resp)

	}

}