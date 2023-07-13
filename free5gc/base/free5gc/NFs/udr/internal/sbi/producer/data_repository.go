package producer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/internal/util"
	"github.com/free5gc/udr/pkg/factory"
	"github.com/free5gc/util/httpwrapper"
	"github.com/free5gc/util/mongoapi"
)

const (
	APPDATA_INFLUDATA_DB_COLLECTION_NAME       = "applicationData.influenceData"
	APPDATA_INFLUDATA_SUBSC_DB_COLLECTION_NAME = "applicationData.influenceData.subsToNotify"
	APPDATA_PFD_DB_COLLECTION_NAME             = "applicationData.pfds"
)

var CurrentResourceUri string

func getDataFromDB(collName string, filter bson.M) (map[string]interface{}, *models.ProblemDetails) {
	data, err := mongoapi.RestfulAPIGetOne(collName, filter)
	if err != nil {
		return nil, openapi.ProblemDetailsSystemFailure(err.Error())
	}
	if data == nil {
		return nil, util.ProblemDetailsNotFound("DATA_NOT_FOUND")
	}
	return data, nil
}

func deleteDataFromDB(collName string, filter bson.M) {
	if err := mongoapi.RestfulAPIDeleteOne(collName, filter); err != nil {
		logger.DataRepoLog.Errorf("deleteDataFromDB: %+v", err)
	}
}

func HandleCreateAccessAndMobilityData(request *httpwrapper.Request) *httpwrapper.Response {
	return httpwrapper.NewResponse(http.StatusOK, nil, map[string]interface{}{})
}

func HandleDeleteAccessAndMobilityData(request *httpwrapper.Request) *httpwrapper.Response {
	return httpwrapper.NewResponse(http.StatusOK, nil, map[string]interface{}{})
}

func HandleQueryAccessAndMobilityData(request *httpwrapper.Request) *httpwrapper.Response {
	return httpwrapper.NewResponse(http.StatusOK, nil, map[string]interface{}{})
}

func HandleQueryAmData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QueryAmData")

	collName := "subscriptionData.provisionedData.amData"
	ueId := request.Params["ueId"]
	servingPlmnId := request.Params["servingPlmnId"]
	response, problemDetails := QueryAmDataProcedure(collName, ueId, servingPlmnId)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func QueryAmDataProcedure(collName string, ueId string, servingPlmnId string) (*map[string]interface{},
	*models.ProblemDetails,
) {
	filter := bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QueryAmDataProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleAmfContext3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle AmfContext3gpp")
	collName := "subscriptionData.contextData.amf3gppAccess"
	patchItem := request.Body.([]models.PatchItem)
	ueId := request.Params["ueId"]

	problemDetails := AmfContext3gppProcedure(collName, ueId, patchItem)
	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func patchDataToDBAndNotify(collName string, ueId string, patchItem []models.PatchItem, filter bson.M) error {
	var err error
	origValue, err := mongoapi.RestfulAPIGetOne(collName, filter)
	if err != nil {
		return err
	}

	patchJSON, err := json.Marshal(patchItem)
	if err != nil {
		return err
	}

	if err = mongoapi.RestfulAPIJSONPatch(collName, filter, patchJSON); err != nil {
		return err
	}

	newValue, err := mongoapi.RestfulAPIGetOne(collName, filter)
	if err != nil {
		return err
	}
	PreHandleOnDataChangeNotify(ueId, CurrentResourceUri, patchItem, origValue, newValue)
	return nil
}

func AmfContext3gppProcedure(collName string, ueId string, patchItem []models.PatchItem) *models.ProblemDetails {
	filter := bson.M{"ueId": ueId}
	if err := patchDataToDBAndNotify(collName, ueId, patchItem, filter); err != nil {
		logger.DataRepoLog.Errorf("AmfContext3gppProcedure err: %+v", err)
		return util.ProblemDetailsModifyNotAllowed("")
	}
	return nil
}

func HandleCreateAmfContext3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle CreateAmfContext3gpp")

	Amf3GppAccessRegistration := request.Body.(models.Amf3GppAccessRegistration)
	ueId := request.Params["ueId"]
	collName := "subscriptionData.contextData.amf3gppAccess"

	CreateAmfContext3gppProcedure(collName, ueId, Amf3GppAccessRegistration)

	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func CreateAmfContext3gppProcedure(collName string, ueId string,
	Amf3GppAccessRegistration models.Amf3GppAccessRegistration,
) {
	filter := bson.M{"ueId": ueId}
	putData := util.ToBsonM(Amf3GppAccessRegistration)
	putData["ueId"] = ueId

	if _, err := mongoapi.RestfulAPIPutOne(collName, filter, putData); err != nil {
		logger.DataRepoLog.Errorf("CreateAmfContext3gppProcedure err: %+v", err)
	}
}

func HandleQueryAmfContext3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QueryAmfContext3gpp")

	ueId := request.Params["ueId"]
	collName := "subscriptionData.contextData.amf3gppAccess"

	response, problemDetails := QueryAmfContext3gppProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QueryAmfContext3gppProcedure(collName string, ueId string) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QueryAmfContext3gppProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleAmfContextNon3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle AmfContextNon3gpp")

	ueId := request.Params["ueId"]
	collName := "subscriptionData.contextData.amfNon3gppAccess"
	patchItem := request.Body.([]models.PatchItem)
	filter := bson.M{"ueId": ueId}

	problemDetails := AmfContextNon3gppProcedure(ueId, collName, patchItem, filter)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func AmfContextNon3gppProcedure(ueId string, collName string, patchItem []models.PatchItem,
	filter bson.M,
) *models.ProblemDetails {
	if err := patchDataToDBAndNotify(collName, ueId, patchItem, filter); err != nil {
		logger.DataRepoLog.Errorf("AmfContextNon3gppProcedure err: %+v", err)
		return util.ProblemDetailsModifyNotAllowed("")
	}
	return nil
}

func HandleCreateAmfContextNon3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle CreateAmfContextNon3gpp")

	AmfNon3GppAccessRegistration := request.Body.(models.AmfNon3GppAccessRegistration)
	collName := "subscriptionData.contextData.amfNon3gppAccess"
	ueId := request.Params["ueId"]

	CreateAmfContextNon3gppProcedure(AmfNon3GppAccessRegistration, collName, ueId)

	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func CreateAmfContextNon3gppProcedure(AmfNon3GppAccessRegistration models.AmfNon3GppAccessRegistration,
	collName string, ueId string,
) {
	putData := util.ToBsonM(AmfNon3GppAccessRegistration)
	putData["ueId"] = ueId
	filter := bson.M{"ueId": ueId}

	if _, err := mongoapi.RestfulAPIPutOne(collName, filter, putData); err != nil {
		logger.DataRepoLog.Errorf("CreateAmfContextNon3gppProcedure err: %+v", err)
	}
}

func HandleQueryAmfContextNon3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QueryAmfContextNon3gpp")

	collName := "subscriptionData.contextData.amfNon3gppAccess"
	ueId := request.Params["ueId"]

	response, problemDetails := QueryAmfContextNon3gppProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QueryAmfContextNon3gppProcedure(collName string, ueId string) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QueryAmfContextNon3gppProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleModifyAuthentication(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ModifyAuthentication")

	collName := "subscriptionData.authenticationData.authenticationSubscription"
	ueId := request.Params["ueId"]
	patchItem := request.Body.([]models.PatchItem)

	problemDetails := ModifyAuthenticationProcedure(collName, ueId, patchItem)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func ModifyAuthenticationProcedure(collName string, ueId string, patchItem []models.PatchItem) *models.ProblemDetails {
	filter := bson.M{"ueId": ueId}
	if err := patchDataToDBAndNotify(collName, ueId, patchItem, filter); err != nil {
		logger.DataRepoLog.Errorf("ModifyAuthenticationProcedure err: %+v", err)
		return util.ProblemDetailsModifyNotAllowed("")
	}
	return nil
}

func HandleQueryAuthSubsData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QueryAuthSubsData")

	collName := "subscriptionData.authenticationData.authenticationSubscription"
	ueId := request.Params["ueId"]

	response, problemDetails := QueryAuthSubsDataProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QueryAuthSubsDataProcedure(collName string, ueId string) (map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		if pd.Status == http.StatusNotFound {
			logger.DataRepoLog.Warnf("QueryAuthSubsDataProcedure err: %s", pd.Title)
		} else {
			logger.DataRepoLog.Errorf("QueryAuthSubsDataProcedure err: %s", pd.Detail)
		}
		return nil, pd
	}
	return data, nil
}

func HandleCreateAuthenticationSoR(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle CreateAuthenticationSoR")
	putData := util.ToBsonM(request.Body)
	ueId := request.Params["ueId"]
	collName := "subscriptionData.ueUpdateConfirmationData.sorData"

	CreateAuthenticationSoRProcedure(collName, ueId, putData)

	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func CreateAuthenticationSoRProcedure(collName string, ueId string, putData bson.M) {
	filter := bson.M{"ueId": ueId}
	putData["ueId"] = ueId

	if _, err := mongoapi.RestfulAPIPutOne(collName, filter, putData); err != nil {
		logger.DataRepoLog.Errorf("CreateAuthenticationSoRProcedure err: %+v", err)
	}
}

func HandleQueryAuthSoR(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QueryAuthSoR")

	ueId := request.Params["ueId"]
	collName := "subscriptionData.ueUpdateConfirmationData.sorData"

	response, problemDetails := QueryAuthSoRProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QueryAuthSoRProcedure(collName string, ueId string) (map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QueryAuthSoRProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return data, nil
}

func HandleCreateAuthenticationStatus(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle CreateAuthenticationStatus")

	putData := util.ToBsonM(request.Body)
	ueId := request.Params["ueId"]
	collName := "subscriptionData.authenticationData.authenticationStatus"

	CreateAuthenticationStatusProcedure(collName, ueId, putData)

	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func CreateAuthenticationStatusProcedure(collName string, ueId string, putData bson.M) {
	filter := bson.M{"ueId": ueId}
	putData["ueId"] = ueId

	if _, err := mongoapi.RestfulAPIPutOne(collName, filter, putData); err != nil {
		logger.DataRepoLog.Errorf("CreateAuthenticationStatusProcedure err: %+v", err)
	}
}

func HandleQueryAuthenticationStatus(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QueryAuthenticationStatus")

	ueId := request.Params["ueId"]
	collName := "subscriptionData.authenticationData.authenticationStatus"

	response, problemDetails := QueryAuthenticationStatusProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QueryAuthenticationStatusProcedure(collName string, ueId string) (*map[string]interface{},
	*models.ProblemDetails,
) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QueryAuthenticationStatusProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleApplicationDataInfluenceDataGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataInfluenceDataGet")
	collName := "applicationData.influenceData"
	var filter []bson.M
	influenceIdsParam := request.Query["influence-Ids"]
	dnnsParam := request.Query["dnns"]
	internalGroupIdsParam := request.Query["internal-Group-Ids"]
	supisParam := request.Query["supis"]
	snssaisParam := request.Query["snssais"]
	if len(influenceIdsParam) != 0 {
		influenceIds := strings.Split(influenceIdsParam[0], ",")
		filter = append(filter, bson.M{"influenceId": bson.M{"$in": influenceIds}})
	}
	if len(dnnsParam) != 0 {
		dnns := strings.Split(dnnsParam[0], ",")
		filter = append(filter, bson.M{"dnn": bson.M{"$in": dnns}})
	}
	if len(internalGroupIdsParam) != 0 {
		internalGroupIds := strings.Split(internalGroupIdsParam[0], ",")
		withAnyUeIndFilter := []bson.M{
			{
				"interGroupId": bson.M{"$in": internalGroupIds},
			},
			{
				"interGroupId": "AnyUE",
			},
		}
		filter = append(filter, bson.M{"$or": withAnyUeIndFilter})
	} else if len(supisParam) != 0 {
		supis := strings.Split(supisParam[0], ",")
		withAnyUeIndFilter := []bson.M{
			{
				"supi": bson.M{"$in": supis},
			},
			{
				"interGroupId": "AnyUE",
			},
		}
		filter = append(filter, bson.M{"$or": withAnyUeIndFilter})
	}
	if len(snssaisParam) != 0 {
		snssais := parseSnssaisFromQueryParam(snssaisParam[0])
		// NOTE: The following code would have bugs with several tries that return null value from Mongo DB, while most of
		//       tries would be correct. The errors seem to occur only when the receiving filters on Mongo DB have reverse
		//       orders of snssai fields, i.e. first sd then sst, even though bson.M{} is used
		// matchList := buildSnssaiMatchList(snssais)
		// filter = append(filter, bson.M{"snssai": bson.M{"$in": matchList}})
		matchList := buildSnssaiMatchList(snssais)
		filter = append(filter, bson.M{"$or": matchList})
	}
	response := ApplicationDataInfluenceDataGetProcedure(collName, filter)
	return httpwrapper.NewResponse(http.StatusOK, nil, response)
}

func parseSnssaisFromQueryParam(snssaiStr string) []models.Snssai {
	var snssais []models.Snssai
	err := json.Unmarshal([]byte("["+snssaiStr+"]"), &snssais)
	if err != nil {
		logger.DataRepoLog.Warnln("Unmarshal Error in snssaiStruct", err)
	}
	return snssais
}

// func buildSnssaiMatchList(snssais []models.Snssai) (matchList []bson.M) {
// 	for _, v := range snssais {
// 		bsonData := util.ToBsonM(v)
// 		matchList = append(matchList, bsonData)
// 	}
// 	return
// }

func buildSnssaiMatchList(snssais []models.Snssai) (matchList []bson.M) {
	for _, v := range snssais {
		matchList = append(matchList, bson.M{"snssai.sst": v.Sst, "snssai.sd": v.Sd})
	}
	return
}

func ApplicationDataInfluenceDataGetProcedure(collName string, filter []bson.M) (
	response *[]map[string]interface{},
) {
	influenceDataArray := make([]map[string]interface{}, 0)
	if len(filter) != 0 {
		var err error
		influenceDataArray, err = mongoapi.RestfulAPIGetMany(collName, bson.M{"$and": filter})
		if err != nil {
			logger.DataRepoLog.Errorf("ApplicationDataInfluenceDataGetProcedure err: %+v", err)
			return nil
		}
	}
	for _, influenceData := range influenceDataArray {
		groupUri := udr_context.GetSelf().GetIPv4GroupUri(udr_context.NUDR_DR)
		influenceData["resUri"] = fmt.Sprintf("%s/application-data/influenceData/%s",
			groupUri, influenceData["influenceId"].(string))
		delete(influenceData, "_id")
		delete(influenceData, "influenceId")
	}
	return &influenceDataArray
}

func HandleApplicationDataInfluenceDataInfluenceIdDelete(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataInfluenceDataInfluenceIdDelete")

	collName := "applicationData.influenceData"
	influenceId := request.Params["influenceId"]
	status := ApplicationDataInfluenceDataInfluenceIdDeleteProcedure(collName, influenceId)

	if status == http.StatusNoContent {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, nil)
	}

	problemDetails := &models.ProblemDetails{
		Status: http.StatusInternalServerError,
		Cause:  "UNSPECIFIED",
	}
	return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
}

func ApplicationDataInfluenceDataInfluenceIdDeleteProcedure(collName, influenceId string) int {
	filter := bson.M{"influenceId": influenceId}

	mapData, err := mongoapi.RestfulAPIGetOne(collName, filter)
	if err != nil {
		logger.DataRepoLog.Errorf("ApplicationDataInfluenceDataInfluenceIdDeleteProcedure err: %+v", err)
		return http.StatusInternalServerError
	}
	var original *models.TrafficInfluData
	if len(mapData) != 0 {
		original = new(models.TrafficInfluData)
		byteData, err := json.Marshal(mapData)
		if err != nil {
			logger.DataRepoLog.Errorf(err.Error())
			return http.StatusInternalServerError
		}
		err = json.Unmarshal(byteData, &original)
		if err != nil {
			logger.DataRepoLog.Errorf(err.Error())
			return http.StatusInternalServerError
		}
	}

	if err := mongoapi.RestfulAPIDeleteOne(collName, filter); err != nil {
		logger.DataRepoLog.Errorf("InfluIdDelProcedure: %+v", err)
		return http.StatusInternalServerError
	}

	// Notify the change of influence data
	PreHandleInfluenceDataUpdateNotification(influenceId, original, nil)

	return http.StatusNoContent
}

func HandleApplicationDataInfluenceDataInfluenceIdPatch(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataInfluenceDataInfluenceIdPatch")

	collName := "application.influenceData"
	influenceId := request.Params["influenceId"]
	trafficInfluDataPatch := request.Body.(models.TrafficInfluDataPatch)

	response, status := ApplicationDataInfluenceDataInfluenceIdPatchProcedure(
		collName, influenceId, &trafficInfluDataPatch)

	if status == http.StatusOK || status == http.StatusNoContent {
		return httpwrapper.NewResponse(status, nil, response)
	} else if status == http.StatusNotFound {
		problemDetails := models.ProblemDetails{
			Status: http.StatusNotFound,
			Detail: "Resource not found",
		}
		return httpwrapper.NewResponse(status, nil, problemDetails)
	}

	problemDetails := &models.ProblemDetails{
		Status: http.StatusInternalServerError,
		Cause:  "UNSPECIFIED",
	}
	return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
}

func ApplicationDataInfluenceDataInfluenceIdPatchProcedure(
	collName, influenceId string, request *models.TrafficInfluDataPatch) (
	*models.TrafficInfluDataPatch, int,
) {
	filter := bson.M{"influenceId": influenceId}

	mapData, err := mongoapi.RestfulAPIGetOne(collName, filter)
	if err != nil {
		logger.DataRepoLog.Errorf(err.Error())
		return nil, http.StatusInternalServerError
	}
	var original *models.TrafficInfluData
	if len(mapData) != 0 {
		original = new(models.TrafficInfluData)
		byteData, err := json.Marshal(mapData)
		if err != nil {
			logger.DataRepoLog.Errorf(err.Error())
			return nil, http.StatusInternalServerError
		}
		err = json.Unmarshal(byteData, &original)
		if err != nil {
			logger.DataRepoLog.Errorf(err.Error())
			return nil, http.StatusInternalServerError
		}
	} else {
		return nil, http.StatusNotFound
	}

	patchTrafficInfluData := models.TrafficInfluData{
		UpPathChgNotifCorreId: request.UpPathChgNotifCorreId,
		AppReloInd:            request.AppReloInd,
		AfAppId:               original.AfAppId,
		Dnn:                   request.Dnn,
		EthTrafficFilters:     request.EthTrafficFilters,
		Snssai:                request.Snssai,
		InterGroupId:          request.InternalGroupId,
		Supi:                  request.Supi,
		TrafficFilters:        request.TrafficFilters,
		TrafficRoutes:         request.TrafficRoutes,
		TraffCorreInd:         original.TraffCorreInd,
		ValidStartTime:        request.ValidStartTime,
		ValidEndTime:          request.ValidEndTime,
		TempValidities:        original.TempValidities,
		NwAreaInfo:            request.NwAreaInfo,
		UpPathChgNotifUri:     request.UpPathChgNotifUri,
		SubscribedEvents:      original.SubscribedEvents,
		DnaiChgType:           original.DnaiChgType,
		AfAckInd:              original.AfAckInd,
		AddrPreserInd:         original.AddrPreserInd,
		SupportedFeatures:     original.SupportedFeatures,
		ResUri:                original.ResUri,
	}

	if reflect.DeepEqual(*original, patchTrafficInfluData) {
		return nil, http.StatusNoContent
	} else {
		putData := util.ToBsonM(patchTrafficInfluData)
		putData["influenceId"] = influenceId
		if _, err := mongoapi.RestfulAPIPutOne(collName, filter, putData); err != nil {
			return nil, http.StatusInternalServerError
		}
		// Notify the change of influence data
		PreHandleInfluenceDataUpdateNotification(influenceId, original, &patchTrafficInfluData)
		return request, http.StatusOK
	}
}

func HandleApplicationDataInfluenceDataInfluenceIdPut(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataInfluenceDataInfluenceIdPut")

	collName := "applicationData.influenceData"
	influenceId := request.Params["influenceId"]
	trafficInfluData := request.Body.(models.TrafficInfluData)

	response, problemDetails, status := ApplicationDataInfluenceDataInfluenceIdPutProcedure(
		collName, influenceId, &trafficInfluData)
	if status == http.StatusCreated {
		// According to 3GPP TS 29.519 V16.5.0 clause 6.2.6.3.1
		// Contain the URI of the newly created resource with `Location` key in the header
		groupUri := udr_context.GetSelf().GetIPv4GroupUri(udr_context.NUDR_DR)
		resourceUri := fmt.Sprintf("%s/application-data/influenceData/%s", groupUri, influenceId)
		header := http.Header{
			"Location": {resourceUri},
		}
		return httpwrapper.NewResponse(http.StatusCreated, header, response)
	} else if status == http.StatusOK {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if status == http.StatusNoContent {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, nil)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	problemDetails = &models.ProblemDetails{
		Status: http.StatusInternalServerError,
		Cause:  "UNSPECIFIED",
	}
	return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
}

func ApplicationDataInfluenceDataInfluenceIdPutProcedure(
	collName, influenceId string, request *models.TrafficInfluData) (
	*models.TrafficInfluData, *models.ProblemDetails, int,
) {
	putData := util.ToBsonM(*request)
	putData["influenceId"] = influenceId
	filter := bson.M{"influenceId": influenceId}

	var original *models.TrafficInfluData

	if mapData, err := mongoapi.RestfulAPIGetOne(collName, filter); err != nil {
		logger.DataRepoLog.Errorf(err.Error())
		problemDetails := &models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
		}
		return nil, problemDetails, http.StatusInternalServerError
	} else {
		if len(mapData) != 0 {
			original = new(models.TrafficInfluData)
			byteData, err := json.Marshal(mapData)
			if err != nil {
				logger.DataRepoLog.Errorf(err.Error())
				problemDetails := &models.ProblemDetails{
					Status: http.StatusInternalServerError,
					Detail: err.Error(),
				}
				return nil, problemDetails, http.StatusInternalServerError
			}
			err = json.Unmarshal(byteData, &original)
			if err != nil {
				logger.DataRepoLog.Errorf(err.Error())
				problemDetails := &models.ProblemDetails{
					Status: http.StatusInternalServerError,
					Detail: err.Error(),
				}
				return nil, problemDetails, http.StatusInternalServerError
			}
		}
	}

	isExisted, err := mongoapi.RestfulAPIPutOne(collName, filter, putData)
	if err != nil {
		logger.DataRepoLog.Errorf("ApplicationDataInfluenceDataInfluenceIdPutProcedure err: %+v", err)
		problemDetails := &models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
		}
		return nil, problemDetails, http.StatusInternalServerError
	}
	if original == nil || !reflect.DeepEqual(*original, *request) {
		// Notify the change of influence data
		PreHandleInfluenceDataUpdateNotification(influenceId, original, request)
	}

	if isExisted {
		return request, nil, http.StatusOK
	} else {
		return request, nil, http.StatusCreated
	}
}

func HandleApplicationDataInfluenceDataSubsToNotifyGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataInfluenceDataSubsToNotifyGet")

	dnn := request.Query.Get("dnn")
	internalGroupId := request.Query.Get("internal-Group-Id")
	supi := request.Query.Get("supi")

	var snssai *models.Snssai
	if request.Query.Get("snssai") != "" {
		snssai = new(models.Snssai)
		err := openapi.Deserialize(snssai, []byte(request.Query.Get("snssai")), "application/json")
		if err != nil {
			problemDetails := models.ProblemDetails{
				Status: http.StatusBadRequest,
				Detail: err.Error(),
			}
			return httpwrapper.NewResponse(http.StatusBadRequest, nil, problemDetails)
		}
	}

	if dnn == "" && snssai == nil && internalGroupId == "" && supi == "" {
		problemDetails := models.ProblemDetails{
			Status: http.StatusBadRequest,
			Detail: "At least one of DNNs, S-NSSAIs, Internal Group IDs or SUPIs shall be provided",
		}
		return httpwrapper.NewResponse(http.StatusBadRequest, nil, problemDetails)
	}

	rspData, problemDetails := ApplicationDataInfluenceDataSubsToNotifyGetProcedure(dnn, snssai, internalGroupId, supi)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, rspData)
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func ApplicationDataInfluenceDataSubsToNotifyGetProcedure(
	dnn string, snssai *models.Snssai, internalGroupId, supi string) (
	[]models.TrafficInfluSub, *models.ProblemDetails,
) {
	var response []models.TrafficInfluSub

	udrSelf := udr_context.GetSelf()
	udrSelf.InfluenceDataSubscriptions.Range(func(key, value interface{}) bool {
		subs, ok := value.(*models.TrafficInfluSub)
		if !ok {
			logger.DataRepoLog.Errorf("Failed to load influence Data subscription ID [%+v]", key)
			return true
		} else if dnn != "" && !util.Contain(dnn, subs.Dnns) {
			return true
		} else if snssai != nil && !util.Contain(*snssai, subs.Snssais) {
			return true
		} else if internalGroupId != "" && !util.Contain(internalGroupId, subs.InternalGroupIds) {
			return true
		} else if supi != "" && !util.Contain(supi, subs.Supis) {
			return true
		} else {
			response = append(response, *subs)
		}
		return true
	})
	return response, nil
}

func HandleApplicationDataInfluenceDataSubsToNotifyPost(
	request *httpwrapper.Request,
) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataInfluenceDataSubsToNotifyPost")

	requestMsg := request.Body.(models.TrafficInfluSub)

	subscriptionId := udr_context.NewInfluenceDataSubscriptionId()

	rspData, problemDetails := ApplicationDataInfluenceDataSubsToNotifySubscriptionIdPutProcedure(
		subscriptionId, &requestMsg)

	if rspData != nil {
		header := http.Header{
			"Location": {
				fmt.Sprintf("%s/application-data/influenceData/subs-to-notify/%s",
					udr_context.GetSelf().GetIPv4GroupUri(udr_context.NUDR_DR), subscriptionId),
			},
		}
		return httpwrapper.NewResponse(http.StatusCreated, header, rspData)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	problemDetails = &models.ProblemDetails{
		Status: http.StatusForbidden,
		Cause:  "UNSPECIFIED",
	}
	return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
}

func HandleApplicationDataInfluenceDataSubsToNotifySubscriptionIdDelete(
	request *httpwrapper.Request,
) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataInfluenceDataSubsToNotifySubscriptionIdDelete")

	subscriptionId := request.Params["subscriptionId"]

	udrSelf := udr_context.GetSelf()

	udrSelf.InfluenceDataSubscriptions.Delete(subscriptionId)

	return httpwrapper.NewResponse(http.StatusNoContent, nil, nil)
}

func HandleApplicationDataInfluenceDataSubsToNotifySubscriptionIdGet(
	request *httpwrapper.Request,
) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataInfluenceDataSubsToNotifySubscriptionIdGet")

	subscriptionId := request.Params["subscriptionId"]

	udrSelf := udr_context.GetSelf()

	if subscription, ok := udrSelf.InfluenceDataSubscriptions.Load(subscriptionId); ok {
		return httpwrapper.NewResponse(http.StatusOK, nil, subscription)
	} else {
		problemDetails := models.ProblemDetails{
			Status: http.StatusNotFound,
			Detail: "Resource not found",
		}
		return httpwrapper.NewResponse(http.StatusNotFound, nil, problemDetails)
	}
}

func HandleApplicationDataInfluenceDataSubsToNotifySubscriptionIdPut(
	request *httpwrapper.Request,
) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataInfluenceDataSubsToNotifySubscriptiondIdPut")

	subscriptionId := request.Params["subscriptionId"]
	requestMsg := request.Body.(models.TrafficInfluSub)

	rspData, problemDetails := ApplicationDataInfluenceDataSubsToNotifySubscriptionIdPutProcedure(
		subscriptionId, &requestMsg)

	if rspData != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, rspData)
	} else if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, nil)
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func ApplicationDataInfluenceDataSubsToNotifySubscriptionIdPutProcedure(
	subscriptionId string, request *models.TrafficInfluSub) (
	*models.TrafficInfluSub, *models.ProblemDetails,
) {
	if len(request.Dnns) == 0 &&
		len(request.Snssais) == 0 &&
		len(request.InternalGroupIds) == 0 &&
		len(request.Supis) == 0 {
		return nil, &models.ProblemDetails{
			Status: http.StatusBadRequest,
			Detail: "At least one of DNNs, S-NSSAIs, Internal Group IDs or SUPIs shall be provided",
		}
	}

	if request.NotificationUri == "" {
		return nil, &models.ProblemDetails{
			Status: http.StatusBadRequest,
			Detail: "Notification URI shall be provided",
		}
	}

	udrSelf := udr_context.GetSelf()
	if subs, ok := udrSelf.InfluenceDataSubscriptions.Load(subscriptionId); ok && reflect.DeepEqual(*request, subs) {
		return nil, nil
	} else {
		udrSelf.InfluenceDataSubscriptions.Store(subscriptionId, request)
		return request, nil
	}
}

func HandleApplicationDataPfdsAppIdDelete(appID string) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataPfdsAppIdDelete: appID=%q", appID)

	deleteApplicationDataIndividualPfdFromDB(appID)

	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func deleteApplicationDataIndividualPfdFromDB(appID string) {
	filter := bson.M{"applicationId": appID}
	deleteDataFromDB(APPDATA_PFD_DB_COLLECTION_NAME, filter)
}

func HandleApplicationDataPfdsAppIdGet(appID string) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataPfdsAppIdGet: appID=%q", appID)

	response, problemDetails := getApplicationDataIndividualPfdFromDB(appID)

	if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
	return httpwrapper.NewResponse(http.StatusOK, nil, response)
}

func getApplicationDataIndividualPfdFromDB(appID string) (map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"applicationId": appID}
	data, pd := getDataFromDB(APPDATA_PFD_DB_COLLECTION_NAME, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("getApplicationDataIndividualPfdFromDB err: %s", pd.Detail)
		return nil, pd
	}
	return data, nil
}

func HandleApplicationDataPfdsAppIdPut(appID string, pfdDataForApp *models.PfdDataForApp) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataPfdsAppIdPut: appID=%q", appID)

	response, status := putApplicationDataIndividualPfdToDB(appID, pfdDataForApp)

	return httpwrapper.NewResponse(status, nil, response)
}

func putApplicationDataIndividualPfdToDB(appID string, pfdDataForApp *models.PfdDataForApp) (bson.M, int) {
	filter := bson.M{"applicationId": appID}
	data := util.ToBsonM(*pfdDataForApp)

	existed, err := mongoapi.RestfulAPIPutOne(APPDATA_PFD_DB_COLLECTION_NAME, filter, data)
	if err != nil {
		logger.DataRepoLog.Errorf("putApplicationDataIndividualPfdToDB err: %+v", err)
		return nil, http.StatusInternalServerError
	}

	if existed {
		return data, http.StatusOK
	}
	return data, http.StatusCreated
}

func HandleApplicationDataPfdsGet(pfdsAppIDs []string) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ApplicationDataPfdsGet: pfdsAppIDs=%#v", pfdsAppIDs)

	// TODO: Parse appID with separator ','
	// Ex: "app1,app2,..."
	response := getApplicationDataPfdsFromDB(pfdsAppIDs)

	return httpwrapper.NewResponse(http.StatusOK, nil, response)
}

func getApplicationDataPfdsFromDB(pfdsAppIDs []string) (response []map[string]interface{}) {
	filter := bson.M{}

	var matchedPfds []map[string]interface{}
	if len(pfdsAppIDs) == 0 {
		var err error
		matchedPfds, err = mongoapi.RestfulAPIGetMany(APPDATA_PFD_DB_COLLECTION_NAME, filter)
		if err != nil {
			logger.DataRepoLog.Errorf("getApplicationDataPfdsFromDB err: %+v", err)
			return nil
		}
	} else {
		for _, v := range pfdsAppIDs {
			filter := bson.M{"applicationId": v}
			data, pd := getDataFromDB(APPDATA_PFD_DB_COLLECTION_NAME, filter)
			if pd == nil {
				matchedPfds = append(matchedPfds, data)
			}
		}
	}
	return matchedPfds
}

func HandleExposureDataSubsToNotifyPost(request *httpwrapper.Request) *httpwrapper.Response {
	return httpwrapper.NewResponse(http.StatusOK, nil, map[string]interface{}{})
}

func HandleExposureDataSubsToNotifySubIdDelete(request *httpwrapper.Request) *httpwrapper.Response {
	return httpwrapper.NewResponse(http.StatusOK, nil, map[string]interface{}{})
}

func HandleExposureDataSubsToNotifySubIdPut(request *httpwrapper.Request) *httpwrapper.Response {
	return httpwrapper.NewResponse(http.StatusOK, nil, map[string]interface{}{})
}

func HandlePolicyDataBdtDataBdtReferenceIdDelete(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataBdtDataBdtReferenceIdDelete")

	collName := "policyData.bdtData"
	bdtReferenceId := request.Params["bdtReferenceId"]

	PolicyDataBdtDataBdtReferenceIdDeleteProcedure(collName, bdtReferenceId)
	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func PolicyDataBdtDataBdtReferenceIdDeleteProcedure(collName string, bdtReferenceId string) {
	filter := bson.M{"bdtReferenceId": bdtReferenceId}
	deleteDataFromDB(collName, filter)
}

func HandlePolicyDataBdtDataBdtReferenceIdGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataBdtDataBdtReferenceIdGet")

	collName := "policyData.bdtData"
	bdtReferenceId := request.Params["bdtReferenceId"]

	response, problemDetails := PolicyDataBdtDataBdtReferenceIdGetProcedure(collName, bdtReferenceId)
	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func PolicyDataBdtDataBdtReferenceIdGetProcedure(collName string, bdtReferenceId string) (*map[string]interface{},
	*models.ProblemDetails,
) {
	filter := bson.M{"bdtReferenceId": bdtReferenceId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("PolicyDataBdtDataBdtReferenceIdGetProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandlePolicyDataBdtDataBdtReferenceIdPut(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataBdtDataBdtReferenceIdPut")

	collName := "policyData.bdtData"
	bdtReferenceId := request.Params["bdtReferenceId"]
	bdtData := request.Body.(models.BdtData)

	response := PolicyDataBdtDataBdtReferenceIdPutProcedure(collName, bdtReferenceId, bdtData)
	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func PolicyDataBdtDataBdtReferenceIdPutProcedure(collName string, bdtReferenceId string,
	bdtData models.BdtData,
) bson.M {
	putData := util.ToBsonM(bdtData)
	putData["bdtReferenceId"] = bdtReferenceId
	filter := bson.M{"bdtReferenceId": bdtReferenceId}

	existed, err := mongoapi.RestfulAPIPutOne(collName, filter, putData)
	if err != nil {
		logger.DataRepoLog.Errorf("putApplicationDataIndividualPfdToDB err: %+v", err)
		return nil
	}

	if existed {
		PreHandlePolicyDataChangeNotification("", bdtReferenceId, bdtData)
	}
	return putData
}

func HandlePolicyDataBdtDataGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataBdtDataGet")

	collName := "policyData.bdtData"

	response := PolicyDataBdtDataGetProcedure(collName)
	return httpwrapper.NewResponse(http.StatusOK, nil, response)
}

func PolicyDataBdtDataGetProcedure(collName string) *[]map[string]interface{} {
	filter := bson.M{}
	bdtDataArray, err := mongoapi.RestfulAPIGetMany(collName, filter)
	if err != nil {
		logger.DataRepoLog.Errorf("PolicyDataBdtDataGetProcedure err: %+v", err)
		return nil
	}
	return &bdtDataArray
}

func HandlePolicyDataPlmnsPlmnIdUePolicySetGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataPlmnsPlmnIdUePolicySetGet")

	collName := "policyData.plmns.uePolicySet"
	plmnId := request.Params["plmnId"]

	response, problemDetails := PolicyDataPlmnsPlmnIdUePolicySetGetProcedure(collName, plmnId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func PolicyDataPlmnsPlmnIdUePolicySetGetProcedure(collName string,
	plmnId string,
) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"plmnId": plmnId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("PolicyDataPlmnsPlmnIdUePolicySetGetProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandlePolicyDataSponsorConnectivityDataSponsorIdGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataSponsorConnectivityDataSponsorIdGet")

	collName := "policyData.sponsorConnectivityData"
	sponsorId := request.Params["sponsorId"]

	response, status := PolicyDataSponsorConnectivityDataSponsorIdGetProcedure(collName, sponsorId)

	if status == http.StatusOK {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if status == http.StatusNoContent {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func PolicyDataSponsorConnectivityDataSponsorIdGetProcedure(collName string,
	sponsorId string,
) (*map[string]interface{}, int) {
	filter := bson.M{"sponsorId": sponsorId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("PolicyDataSponsorConnectivityDataSponsorIdGetProcedure err: %s", pd.Detail)
		return nil, http.StatusNoContent
	}
	return &data, http.StatusOK
}

func HandlePolicyDataSubsToNotifyPost(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataSubsToNotifyPost")

	PolicyDataSubscription := request.Body.(models.PolicyDataSubscription)

	locationHeader := PolicyDataSubsToNotifyPostProcedure(PolicyDataSubscription)

	headers := http.Header{}
	headers.Set("Location", locationHeader)
	return httpwrapper.NewResponse(http.StatusCreated, headers, PolicyDataSubscription)
}

func PolicyDataSubsToNotifyPostProcedure(PolicyDataSubscription models.PolicyDataSubscription) string {
	udrSelf := udr_context.GetSelf()

	newSubscriptionID := strconv.Itoa(udrSelf.PolicyDataSubscriptionIDGenerator)
	udrSelf.PolicyDataSubscriptions[newSubscriptionID] = &PolicyDataSubscription
	udrSelf.PolicyDataSubscriptionIDGenerator++

	/* Contains the URI of the newly created resource, according
	   to the structure: {apiRoot}/subscription-data/subs-to-notify/{subsId} */
	locationHeader := fmt.Sprintf("%s/policy-data/subs-to-notify/%s", udrSelf.GetIPv4GroupUri(udr_context.NUDR_DR),
		newSubscriptionID)

	return locationHeader
}

func HandlePolicyDataSubsToNotifySubsIdDelete(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataSubsToNotifySubsIdDelete")

	subsId := request.Params["subsId"]

	problemDetails := PolicyDataSubsToNotifySubsIdDeleteProcedure(subsId)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func PolicyDataSubsToNotifySubsIdDeleteProcedure(subsId string) (problemDetails *models.ProblemDetails) {
	udrSelf := udr_context.GetSelf()
	_, ok := udrSelf.PolicyDataSubscriptions[subsId]
	if !ok {
		return util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}
	delete(udrSelf.PolicyDataSubscriptions, subsId)

	return nil
}

func HandlePolicyDataSubsToNotifySubsIdPut(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataSubsToNotifySubsIdPut")

	subsId := request.Params["subsId"]
	policyDataSubscription := request.Body.(models.PolicyDataSubscription)

	response, problemDetails := PolicyDataSubsToNotifySubsIdPutProcedure(subsId, policyDataSubscription)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func PolicyDataSubsToNotifySubsIdPutProcedure(subsId string,
	policyDataSubscription models.PolicyDataSubscription,
) (*models.PolicyDataSubscription, *models.ProblemDetails) {
	udrSelf := udr_context.GetSelf()
	_, ok := udrSelf.PolicyDataSubscriptions[subsId]
	if !ok {
		return nil, util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}

	udrSelf.PolicyDataSubscriptions[subsId] = &policyDataSubscription

	return &policyDataSubscription, nil
}

func HandlePolicyDataUesUeIdAmDataGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdAmDataGet")

	collName := "policyData.ues.amData"
	ueId := request.Params["ueId"]

	response, problemDetails := PolicyDataUesUeIdAmDataGetProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func PolicyDataUesUeIdAmDataGetProcedure(collName string,
	ueId string,
) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdAmDataGetProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandlePolicyDataUesUeIdOperatorSpecificDataGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdOperatorSpecificDataGet")

	collName := "policyData.ues.operatorSpecificData"
	ueId := request.Params["ueId"]

	response, problemDetails := PolicyDataUesUeIdOperatorSpecificDataGetProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func PolicyDataUesUeIdOperatorSpecificDataGetProcedure(collName string,
	ueId string,
) (*interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdOperatorSpecificDataGetProcedure err: %s", pd.Detail)
		return nil, pd
	}
	operatorSpecificDataContainerMap := data["operatorSpecificDataContainerMap"]
	return &operatorSpecificDataContainerMap, nil
}

func HandlePolicyDataUesUeIdOperatorSpecificDataPatch(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdOperatorSpecificDataPatch")

	collName := "policyData.ues.operatorSpecificData"
	ueId := request.Params["ueId"]
	patchItem := request.Body.([]models.PatchItem)

	problemDetails := PolicyDataUesUeIdOperatorSpecificDataPatchProcedure(collName, ueId, patchItem)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func PolicyDataUesUeIdOperatorSpecificDataPatchProcedure(collName string, ueId string,
	patchItem []models.PatchItem,
) *models.ProblemDetails {
	filter := bson.M{"ueId": ueId}

	patchJSON, err := json.Marshal(patchItem)
	if err != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdOperatorSpecificDataPatchProcedure err: %+v", err)
		return util.ProblemDetailsModifyNotAllowed("")
	}

	if err := mongoapi.RestfulAPIJSONPatchExtend(collName, filter, patchJSON,
		"operatorSpecificDataContainerMap"); err != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdOperatorSpecificDataPatchProcedure err: %+v", err)
		return util.ProblemDetailsModifyNotAllowed("")
	}
	return nil
}

func HandlePolicyDataUesUeIdOperatorSpecificDataPut(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdOperatorSpecificDataPut")

	// json.NewDecoder(c.Request.Body).Decode(&operatorSpecificDataContainerMap)

	collName := "policyData.ues.operatorSpecificData"
	ueId := request.Params["ueId"]
	OperatorSpecificDataContainer := request.Body.(map[string]models.OperatorSpecificDataContainer)

	PolicyDataUesUeIdOperatorSpecificDataPutProcedure(collName, ueId, OperatorSpecificDataContainer)

	return httpwrapper.NewResponse(http.StatusOK, nil, map[string]interface{}{})
}

func PolicyDataUesUeIdOperatorSpecificDataPutProcedure(collName string, ueId string,
	OperatorSpecificDataContainer map[string]models.OperatorSpecificDataContainer,
) {
	filter := bson.M{"ueId": ueId}

	putData := map[string]interface{}{"operatorSpecificDataContainerMap": OperatorSpecificDataContainer}
	putData["ueId"] = ueId

	_, err := mongoapi.RestfulAPIPutOne(collName, filter, putData)
	if err != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdOperatorSpecificDataPutProcedure err: %+v", err)
	}
}

func HandlePolicyDataUesUeIdSmDataGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdSmDataGet")

	collName := "policyData.ues.smData"
	ueId := request.Params["ueId"]
	sNssai := models.Snssai{}
	sNssaiQuery := request.Query.Get("snssai")
	err := json.Unmarshal([]byte(sNssaiQuery), &sNssai)
	if err != nil {
		logger.DataRepoLog.Warnln(err)
	}
	dnn := request.Query.Get("dnn")

	response, problemDetails := PolicyDataUesUeIdSmDataGetProcedure(collName, ueId, sNssai, dnn)
	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func PolicyDataUesUeIdSmDataGetProcedure(collName string, ueId string, snssai models.Snssai,
	dnn string,
) (*models.SmPolicyData, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}

	if !reflect.DeepEqual(snssai, models.Snssai{}) {
		filter["smPolicySnssaiData."+util.SnssaiModelsToHex(snssai)] = bson.M{"$exists": true}
	}
	if !reflect.DeepEqual(snssai, models.Snssai{}) && dnn != "" {
		dnnKey := util.EscapeDnn(dnn)
		filter["smPolicySnssaiData."+util.SnssaiModelsToHex(snssai)+".smPolicyDnnData."+dnnKey] = bson.M{"$exists": true}
	}

	smPolicyData, pd := getDataFromDB(collName, filter)
	if pd != nil {
		return nil, pd
	}

	var smPolicyDataResp models.SmPolicyData
	err := json.Unmarshal(util.MapToByte(smPolicyData), &smPolicyDataResp)
	if err != nil {
		logger.DataRepoLog.Warnln(err)
	}
	tmpSmPolicySnssaiData := make(map[string]models.SmPolicySnssaiData)
	for snssai, snssaiData := range smPolicyDataResp.SmPolicySnssaiData {
		tmpSmPolicyDnnData := make(map[string]models.SmPolicyDnnData)
		for escapedDnn, dnnData := range snssaiData.SmPolicyDnnData {
			dnn := util.UnescapeDnn(escapedDnn)
			tmpSmPolicyDnnData[dnn] = dnnData
		}
		snssaiData.SmPolicyDnnData = tmpSmPolicyDnnData
		tmpSmPolicySnssaiData[snssai] = snssaiData
	}
	smPolicyDataResp.SmPolicySnssaiData = tmpSmPolicySnssaiData
	filter = bson.M{"ueId": ueId}
	usageMonDataMapArray, err := mongoapi.RestfulAPIGetMany("policyData.ues.smData.usageMonData", filter)
	if err != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdSmDataGetProcedure err: %+v", err)
	}

	if !reflect.DeepEqual(usageMonDataMapArray, []map[string]interface{}{}) {
		var usageMonDataArray []models.UsageMonData
		if err := json.Unmarshal(util.MapArrayToByte(usageMonDataMapArray), &usageMonDataArray); err != nil {
			logger.DataRepoLog.Warnln(err)
		}
		smPolicyDataResp.UmData = make(map[string]models.UsageMonData)
		for _, element := range usageMonDataArray {
			smPolicyDataResp.UmData[element.LimitId] = element
		}
	}
	return &smPolicyDataResp, nil
}

func HandlePolicyDataUesUeIdSmDataPatch(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdSmDataPatch")

	collName := "policyData.ues.smData.usageMonData"
	ueId := request.Params["ueId"]
	usageMonData := request.Body.(map[string]models.UsageMonData)

	problemDetails := PolicyDataUesUeIdSmDataPatchProcedure(collName, ueId, usageMonData)
	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func PolicyDataUesUeIdSmDataPatchProcedure(collName string, ueId string,
	UsageMonData map[string]models.UsageMonData,
) *models.ProblemDetails {
	filter := bson.M{"ueId": ueId}

	successAll := true
	for k, usageMonData := range UsageMonData {
		limitId := k
		filterTmp := bson.M{"ueId": ueId, "limitId": limitId}
		if err := mongoapi.RestfulAPIMergePatch(collName, filterTmp, util.ToBsonM(usageMonData)); err != nil {
			successAll = false
		} else {
			var usageMonData models.UsageMonData
			usageMonDataBsonM, pd := getDataFromDB(collName, filter)
			if pd != nil && pd.Status == http.StatusInternalServerError {
				logger.DataRepoLog.Errorf("PolicyDataUesUeIdSmDataPatchProcedure err: %s", pd.Detail)
				return pd
			}
			if err := json.Unmarshal(util.MapToByte(usageMonDataBsonM), &usageMonData); err != nil {
				logger.DataRepoLog.Warnln(err)
			}
			PreHandlePolicyDataChangeNotification(ueId, limitId, usageMonData)
		}
	}

	if successAll {
		smPolicyDataBsonM, pd := getDataFromDB(collName, filter)
		if pd != nil {
			logger.DataRepoLog.Errorf("PolicyDataUesUeIdSmDataPatchProcedure err: %s", pd.Detail)
			return pd
		}
		var smPolicyData models.SmPolicyData
		if err := json.Unmarshal(util.MapToByte(smPolicyDataBsonM), &smPolicyData); err != nil {
			logger.DataRepoLog.Warnln(err)
		}

		collName := "policyData.ues.smData.usageMonData"
		filter := bson.M{"ueId": ueId}
		usageMonDataMapArray, err := mongoapi.RestfulAPIGetMany(collName, filter)
		if err != nil {
			logger.DataRepoLog.Errorf("PolicyDataUesUeIdSmDataPatchProcedure err: %+v", err)
		}

		if !reflect.DeepEqual(usageMonDataMapArray, []map[string]interface{}{}) {
			var usageMonDataArray []models.UsageMonData
			if err := json.Unmarshal(util.MapArrayToByte(usageMonDataMapArray), &usageMonDataArray); err != nil {
				logger.DataRepoLog.Warnln(err)
			}
			smPolicyData.UmData = make(map[string]models.UsageMonData)
			for _, element := range usageMonDataArray {
				smPolicyData.UmData[element.LimitId] = element
			}
		}
		PreHandlePolicyDataChangeNotification(ueId, "", smPolicyData)
		return nil
	}
	return util.ProblemDetailsModifyNotAllowed("")
}

func HandlePolicyDataUesUeIdSmDataUsageMonIdDelete(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdSmDataUsageMonIdDelete")

	collName := "policyData.ues.smData.usageMonData"
	ueId := request.Params["ueId"]
	usageMonId := request.Params["usageMonId"]

	PolicyDataUesUeIdSmDataUsageMonIdDeleteProcedure(collName, ueId, usageMonId)
	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func PolicyDataUesUeIdSmDataUsageMonIdDeleteProcedure(collName string, ueId string, usageMonId string) {
	filter := bson.M{"ueId": ueId, "usageMonId": usageMonId}
	deleteDataFromDB(collName, filter)
}

func HandlePolicyDataUesUeIdSmDataUsageMonIdGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdSmDataUsageMonIdGet")

	collName := "policyData.ues.smData.usageMonData"
	ueId := request.Params["ueId"]
	usageMonId := request.Params["usageMonId"]

	response := PolicyDataUesUeIdSmDataUsageMonIdGetProcedure(collName, usageMonId, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	}
}

func PolicyDataUesUeIdSmDataUsageMonIdGetProcedure(collName string, usageMonId string,
	ueId string,
) *map[string]interface{} {
	filter := bson.M{"ueId": ueId, "usageMonId": usageMonId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdSmDataUsageMonIdGetProcedure err: %s", pd.Detail)
		return nil
	}
	return &data
}

func HandlePolicyDataUesUeIdSmDataUsageMonIdPut(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdSmDataUsageMonIdPut")

	ueId := request.Params["ueId"]
	usageMonId := request.Params["usageMonId"]
	usageMonData := request.Body.(models.UsageMonData)
	collName := "policyData.ues.smData.usageMonData"

	response := PolicyDataUesUeIdSmDataUsageMonIdPutProcedure(collName, ueId, usageMonId, usageMonData)

	return httpwrapper.NewResponse(http.StatusCreated, nil, response)
}

func PolicyDataUesUeIdSmDataUsageMonIdPutProcedure(collName string, ueId string, usageMonId string,
	usageMonData models.UsageMonData,
) *bson.M {
	putData := util.ToBsonM(usageMonData)
	putData["ueId"] = ueId
	putData["usageMonId"] = usageMonId
	filter := bson.M{"ueId": ueId, "usageMonId": usageMonId}

	_, err := mongoapi.RestfulAPIPutOne(collName, filter, putData)
	if err != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdSmDataUsageMonIdPutProcedure err: %+v", err)
	}
	return &putData
}

func HandlePolicyDataUesUeIdUePolicySetGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdUePolicySetGet")

	ueId := request.Params["ueId"]
	collName := "policyData.ues.uePolicySet"

	response, problemDetails := PolicyDataUesUeIdUePolicySetGetProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func PolicyDataUesUeIdUePolicySetGetProcedure(collName string, ueId string) (*map[string]interface{},
	*models.ProblemDetails,
) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdUePolicySetGetProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandlePolicyDataUesUeIdUePolicySetPatch(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdUePolicySetPatch")

	collName := "policyData.ues.uePolicySet"
	ueId := request.Params["ueId"]
	UePolicySet := request.Body.(models.UePolicySet)

	problemDetails := PolicyDataUesUeIdUePolicySetPatchProcedure(collName, ueId, UePolicySet)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func PolicyDataUesUeIdUePolicySetPatchProcedure(collName string, ueId string,
	UePolicySet models.UePolicySet,
) *models.ProblemDetails {
	patchData := util.ToBsonM(UePolicySet)
	patchData["ueId"] = ueId
	filter := bson.M{"ueId": ueId}

	if err := mongoapi.RestfulAPIMergePatch(collName, filter, patchData); err != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdUePolicySetPatchProcedure err: %+v", err)
		return util.ProblemDetailsModifyNotAllowed("")
	}

	var uePolicySet models.UePolicySet
	uePolicySetBsonM, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdUePolicySetPatchProcedure err: %s", pd.Detail)
		return pd
	}
	if err := json.Unmarshal(util.MapToByte(uePolicySetBsonM), &uePolicySet); err != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdUePolicySetPatchProcedure err: %+v", err)
		return openapi.ProblemDetailsSystemFailure(err.Error())
	}
	PreHandlePolicyDataChangeNotification(ueId, "", uePolicySet)
	return nil
}

func HandlePolicyDataUesUeIdUePolicySetPut(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PolicyDataUesUeIdUePolicySetPut")

	collName := "policyData.ues.uePolicySet"
	ueId := request.Params["ueId"]
	UePolicySet := request.Body.(models.UePolicySet)

	response, status := PolicyDataUesUeIdUePolicySetPutProcedure(collName, ueId, UePolicySet)

	if status == http.StatusNoContent {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else if status == http.StatusCreated {
		return httpwrapper.NewResponse(http.StatusCreated, nil, response)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func PolicyDataUesUeIdUePolicySetPutProcedure(collName string, ueId string,
	UePolicySet models.UePolicySet,
) (bson.M, int) {
	putData := util.ToBsonM(UePolicySet)
	putData["ueId"] = ueId
	filter := bson.M{"ueId": ueId}

	existed, err := mongoapi.RestfulAPIPutOne(collName, filter, putData)
	if err != nil {
		logger.DataRepoLog.Errorf("PolicyDataUesUeIdUePolicySetPutProcedure err: %+v", err)
		return nil, http.StatusInternalServerError
	}
	if existed {
		return nil, http.StatusNoContent
	}
	return putData, http.StatusCreated
}

func HandleCreateAMFSubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle CreateAMFSubscriptions")

	ueId := request.Params["ueId"]
	subsId := request.Params["subsId"]
	AmfSubscriptionInfo := request.Body.([]models.AmfSubscriptionInfo)

	problemDetails := CreateAMFSubscriptionsProcedure(subsId, ueId, AmfSubscriptionInfo)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func CreateAMFSubscriptionsProcedure(subsId string, ueId string,
	AmfSubscriptionInfo []models.AmfSubscriptionInfo,
) *models.ProblemDetails {
	udrSelf := udr_context.GetSelf()
	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		return util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}
	UESubsData := value.(*udr_context.UESubsData)

	_, ok = UESubsData.EeSubscriptionCollection[subsId]
	if !ok {
		return util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}

	UESubsData.EeSubscriptionCollection[subsId].AmfSubscriptionInfos = AmfSubscriptionInfo
	return nil
}

func HandleRemoveAmfSubscriptionsInfo(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle RemoveAmfSubscriptionsInfo")

	ueId := request.Params["ueId"]
	subsId := request.Params["subsId"]

	problemDetails := RemoveAmfSubscriptionsInfoProcedure(subsId, ueId)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func RemoveAmfSubscriptionsInfoProcedure(subsId string, ueId string) *models.ProblemDetails {
	udrSelf := udr_context.GetSelf()
	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		return util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}

	UESubsData := value.(*udr_context.UESubsData)
	_, ok = UESubsData.EeSubscriptionCollection[subsId]

	if !ok {
		return util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}

	if UESubsData.EeSubscriptionCollection[subsId].AmfSubscriptionInfos == nil {
		return util.ProblemDetailsNotFound("AMFSUBSCRIPTION_NOT_FOUND")
	}

	UESubsData.EeSubscriptionCollection[subsId].AmfSubscriptionInfos = nil

	return nil
}

func HandleModifyAmfSubscriptionInfo(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ModifyAmfSubscriptionInfo")

	patchItem := request.Body.([]models.PatchItem)
	ueId := request.Params["ueId"]
	subsId := request.Params["subsId"]

	problemDetails := ModifyAmfSubscriptionInfoProcedure(ueId, subsId, patchItem)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func ModifyAmfSubscriptionInfoProcedure(ueId string, subsId string,
	patchItem []models.PatchItem,
) *models.ProblemDetails {
	udrSelf := udr_context.GetSelf()
	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		return util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}
	UESubsData := value.(*udr_context.UESubsData)

	_, ok = UESubsData.EeSubscriptionCollection[subsId]

	if !ok {
		return util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}

	if UESubsData.EeSubscriptionCollection[subsId].AmfSubscriptionInfos == nil {
		return util.ProblemDetailsNotFound("AMFSUBSCRIPTION_NOT_FOUND")
	}
	var patchJSON []byte
	if patchJSONtemp, err := json.Marshal(patchItem); err != nil {
		logger.DataRepoLog.Errorln(err)
	} else {
		patchJSON = patchJSONtemp
	}
	var patch jsonpatch.Patch
	if patchtemp, err := jsonpatch.DecodePatch(patchJSON); err != nil {
		logger.DataRepoLog.Errorln(err)
		return util.ProblemDetailsModifyNotAllowed("PatchItem attributes are invalid")
	} else {
		patch = patchtemp
	}
	original, err := json.Marshal((UESubsData.EeSubscriptionCollection[subsId]).AmfSubscriptionInfos)
	if err != nil {
		logger.DataRepoLog.Warnln(err)
	}

	modified, err := patch.Apply(original)
	if err != nil {
		return util.ProblemDetailsModifyNotAllowed("Occur error when applying PatchItem")
	}
	var modifiedData []models.AmfSubscriptionInfo
	err = json.Unmarshal(modified, &modifiedData)
	if err != nil {
		logger.DataRepoLog.Error(err)
	}

	UESubsData.EeSubscriptionCollection[subsId].AmfSubscriptionInfos = modifiedData
	return nil
}

func HandleGetAmfSubscriptionInfo(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle GetAmfSubscriptionInfo")

	ueId := request.Params["ueId"]
	subsId := request.Params["subsId"]

	response, problemDetails := GetAmfSubscriptionInfoProcedure(subsId, ueId)
	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func GetAmfSubscriptionInfoProcedure(subsId string, ueId string) (*[]models.AmfSubscriptionInfo,
	*models.ProblemDetails,
) {
	udrSelf := udr_context.GetSelf()

	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		return nil, util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}

	UESubsData := value.(*udr_context.UESubsData)
	_, ok = UESubsData.EeSubscriptionCollection[subsId]

	if !ok {
		return nil, util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}

	if UESubsData.EeSubscriptionCollection[subsId].AmfSubscriptionInfos == nil {
		return nil, util.ProblemDetailsNotFound("AMFSUBSCRIPTION_NOT_FOUND")
	}
	return &UESubsData.EeSubscriptionCollection[subsId].AmfSubscriptionInfos, nil
}

func HandleQueryEEData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QueryEEData")

	ueId := request.Params["ueId"]
	collName := "subscriptionData.eeProfileData"

	response, problemDetails := QueryEEDataProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QueryEEDataProcedure(collName string, ueId string) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QueryEEDataProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleRemoveEeGroupSubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle RemoveEeGroupSubscriptions")

	ueGroupId := request.Params["ueGroupId"]
	subsId := request.Params["subsId"]

	problemDetails := RemoveEeGroupSubscriptionsProcedure(ueGroupId, subsId)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func RemoveEeGroupSubscriptionsProcedure(ueGroupId string, subsId string) *models.ProblemDetails {
	udrSelf := udr_context.GetSelf()
	value, ok := udrSelf.UEGroupCollection.Load(ueGroupId)
	if !ok {
		return util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}

	UEGroupSubsData := value.(*udr_context.UEGroupSubsData)
	_, ok = UEGroupSubsData.EeSubscriptions[subsId]

	if !ok {
		return util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}
	delete(UEGroupSubsData.EeSubscriptions, subsId)

	return nil
}

func HandleUpdateEeGroupSubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle UpdateEeGroupSubscriptions")

	ueGroupId := request.Params["ueGroupId"]
	subsId := request.Params["subsId"]
	EeSubscription := request.Body.(models.EeSubscription)

	problemDetails := UpdateEeGroupSubscriptionsProcedure(ueGroupId, subsId, EeSubscription)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func UpdateEeGroupSubscriptionsProcedure(ueGroupId string, subsId string,
	EeSubscription models.EeSubscription,
) *models.ProblemDetails {
	udrSelf := udr_context.GetSelf()
	value, ok := udrSelf.UEGroupCollection.Load(ueGroupId)
	if !ok {
		return util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}

	UEGroupSubsData := value.(*udr_context.UEGroupSubsData)
	_, ok = UEGroupSubsData.EeSubscriptions[subsId]

	if !ok {
		return util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}
	UEGroupSubsData.EeSubscriptions[subsId] = &EeSubscription

	return nil
}

func HandleCreateEeGroupSubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle CreateEeGroupSubscriptions")

	ueGroupId := request.Params["ueGroupId"]
	EeSubscription := request.Body.(models.EeSubscription)

	locationHeader := CreateEeGroupSubscriptionsProcedure(ueGroupId, EeSubscription)

	headers := http.Header{}
	headers.Set("Location", locationHeader)
	return httpwrapper.NewResponse(http.StatusCreated, headers, EeSubscription)
}

func CreateEeGroupSubscriptionsProcedure(ueGroupId string, EeSubscription models.EeSubscription) string {
	udrSelf := udr_context.GetSelf()

	value, ok := udrSelf.UEGroupCollection.Load(ueGroupId)
	if !ok {
		udrSelf.UEGroupCollection.Store(ueGroupId, new(udr_context.UEGroupSubsData))
		value, _ = udrSelf.UEGroupCollection.Load(ueGroupId)
	}
	UEGroupSubsData := value.(*udr_context.UEGroupSubsData)
	if UEGroupSubsData.EeSubscriptions == nil {
		UEGroupSubsData.EeSubscriptions = make(map[string]*models.EeSubscription)
	}

	newSubscriptionID := strconv.Itoa(udrSelf.EeSubscriptionIDGenerator)
	UEGroupSubsData.EeSubscriptions[newSubscriptionID] = &EeSubscription
	udrSelf.EeSubscriptionIDGenerator++

	/* Contains the URI of the newly created resource, according
	   to the structure: {apiRoot}/nudr-dr/v1/subscription-data/group-data/{ueGroupId}/ee-subscriptions */
	locationHeader := fmt.Sprintf("%s"+factory.UdrDrResUriPrefix+"/subscription-data/group-data/%s/ee-subscriptions/%s",
		udrSelf.GetIPv4GroupUri(udr_context.NUDR_DR), ueGroupId, newSubscriptionID)

	return locationHeader
}

func HandleQueryEeGroupSubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QueryEeGroupSubscriptions")

	ueGroupId := request.Params["ueGroupId"]

	response, problemDetails := QueryEeGroupSubscriptionsProcedure(ueGroupId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QueryEeGroupSubscriptionsProcedure(ueGroupId string) ([]models.EeSubscription, *models.ProblemDetails) {
	udrSelf := udr_context.GetSelf()

	value, ok := udrSelf.UEGroupCollection.Load(ueGroupId)
	if !ok {
		return nil, util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}

	UEGroupSubsData := value.(*udr_context.UEGroupSubsData)
	var eeSubscriptionSlice []models.EeSubscription

	for _, v := range UEGroupSubsData.EeSubscriptions {
		eeSubscriptionSlice = append(eeSubscriptionSlice, *v)
	}
	return eeSubscriptionSlice, nil
}

func HandleRemoveeeSubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle RemoveeeSubscriptions")

	ueId := request.Params["ueId"]
	subsId := request.Params["subsId"]

	problemDetails := RemoveeeSubscriptionsProcedure(ueId, subsId)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func RemoveeeSubscriptionsProcedure(ueId string, subsId string) *models.ProblemDetails {
	udrSelf := udr_context.GetSelf()
	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		return util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}

	UESubsData := value.(*udr_context.UESubsData)
	_, ok = UESubsData.EeSubscriptionCollection[subsId]

	if !ok {
		return util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}
	delete(UESubsData.EeSubscriptionCollection, subsId)
	return nil
}

func HandleUpdateEesubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle UpdateEesubscriptions")

	ueId := request.Params["ueId"]
	subsId := request.Params["subsId"]
	EeSubscription := request.Body.(models.EeSubscription)

	problemDetails := UpdateEesubscriptionsProcedure(ueId, subsId, EeSubscription)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func UpdateEesubscriptionsProcedure(ueId string, subsId string,
	EeSubscription models.EeSubscription,
) *models.ProblemDetails {
	udrSelf := udr_context.GetSelf()
	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		return util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}

	UESubsData := value.(*udr_context.UESubsData)
	_, ok = UESubsData.EeSubscriptionCollection[subsId]

	if !ok {
		return util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}
	UESubsData.EeSubscriptionCollection[subsId].EeSubscriptions = &EeSubscription

	return nil
}

func HandleCreateEeSubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle CreateEeSubscriptions")

	ueId := request.Params["ueId"]
	EeSubscription := request.Body.(models.EeSubscription)

	locationHeader := CreateEeSubscriptionsProcedure(ueId, EeSubscription)

	headers := http.Header{}
	headers.Set("Location", locationHeader)
	return httpwrapper.NewResponse(http.StatusCreated, headers, EeSubscription)
}

func CreateEeSubscriptionsProcedure(ueId string, EeSubscription models.EeSubscription) string {
	udrSelf := udr_context.GetSelf()

	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		udrSelf.UESubsCollection.Store(ueId, new(udr_context.UESubsData))
		value, _ = udrSelf.UESubsCollection.Load(ueId)
	}
	UESubsData := value.(*udr_context.UESubsData)
	if UESubsData.EeSubscriptionCollection == nil {
		UESubsData.EeSubscriptionCollection = make(map[string]*udr_context.EeSubscriptionCollection)
	}

	newSubscriptionID := strconv.Itoa(udrSelf.EeSubscriptionIDGenerator)
	UESubsData.EeSubscriptionCollection[newSubscriptionID] = new(udr_context.EeSubscriptionCollection)
	UESubsData.EeSubscriptionCollection[newSubscriptionID].EeSubscriptions = &EeSubscription
	udrSelf.EeSubscriptionIDGenerator++

	/* Contains the URI of the newly created resource, according
	   to the structure: {apiRoot}/subscription-data/{ueId}/context-data/ee-subscriptions/{subsId} */
	locationHeader := fmt.Sprintf("%s/subscription-data/%s/context-data/ee-subscriptions/%s",
		udrSelf.GetIPv4GroupUri(udr_context.NUDR_DR), ueId, newSubscriptionID)

	return locationHeader
}

func HandleQueryeesubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle Queryeesubscriptions")

	ueId := request.Params["ueId"]

	response, problemDetails := QueryeesubscriptionsProcedure(ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QueryeesubscriptionsProcedure(ueId string) ([]models.EeSubscription, *models.ProblemDetails) {
	udrSelf := udr_context.GetSelf()

	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		return nil, util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}

	UESubsData := value.(*udr_context.UESubsData)
	var eeSubscriptionSlice []models.EeSubscription

	for _, v := range UESubsData.EeSubscriptionCollection {
		eeSubscriptionSlice = append(eeSubscriptionSlice, *v.EeSubscriptions)
	}
	return eeSubscriptionSlice, nil
}

func HandlePatchOperSpecData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PatchOperSpecData")

	collName := "subscriptionData.operatorSpecificData"
	ueId := request.Params["ueId"]
	patchItem := request.Body.([]models.PatchItem)

	problemDetails := PatchOperSpecDataProcedure(collName, ueId, patchItem)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func PatchOperSpecDataProcedure(collName string, ueId string, patchItem []models.PatchItem) *models.ProblemDetails {
	filter := bson.M{"ueId": ueId}
	if err := patchDataToDBAndNotify(collName, ueId, patchItem, filter); err != nil {
		logger.DataRepoLog.Errorf("PatchOperSpecDataProcedure err: %+v", err)
		return util.ProblemDetailsModifyNotAllowed("")
	}
	return nil
}

func HandleQueryOperSpecData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QueryOperSpecData")

	ueId := request.Params["ueId"]
	collName := "subscriptionData.operatorSpecificData"

	response, problemDetails := QueryOperSpecDataProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QueryOperSpecDataProcedure(collName string, ueId string) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	// The key of the map is operator specific data element name and the value is the operator specific data of the UE.
	if pd != nil {
		logger.DataRepoLog.Errorf("QueryOperSpecDataProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleGetppData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle GetppData")

	collName := "subscriptionData.ppData"
	ueId := request.Params["ueId"]

	response, problemDetails := GetppDataProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func GetppDataProcedure(collName string, ueId string) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("GetppDataProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleCreateSessionManagementData(request *httpwrapper.Request) *httpwrapper.Response {
	return httpwrapper.NewResponse(http.StatusOK, nil, map[string]interface{}{})
}

func HandleDeleteSessionManagementData(request *httpwrapper.Request) *httpwrapper.Response {
	return httpwrapper.NewResponse(http.StatusOK, nil, map[string]interface{}{})
}

func HandleQuerySessionManagementData(request *httpwrapper.Request) *httpwrapper.Response {
	return httpwrapper.NewResponse(http.StatusOK, nil, map[string]interface{}{})
}

func HandleQueryProvisionedData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QueryProvisionedData")

	var provisionedDataSets models.ProvisionedDataSets
	ueId := request.Params["ueId"]
	servingPlmnId := request.Params["servingPlmnId"]

	response, problemDetails := QueryProvisionedDataProcedure(ueId, servingPlmnId, provisionedDataSets)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QueryProvisionedDataProcedure(ueId string, servingPlmnId string,
	provisionedDataSets models.ProvisionedDataSets,
) (*models.ProvisionedDataSets, *models.ProblemDetails) {
	var collName string
	var filter bson.M

	collName = "subscriptionData.provisionedData.amData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	accessAndMobilitySubscriptionData, pd := getDataFromDB(collName, filter)
	if pd != nil && pd.Status == http.StatusInternalServerError {
		logger.DataRepoLog.Errorf(
			"QueryProvisionedDataProcedure get accessAndMobilitySubscriptionData err: %s", pd.Detail)
		return nil, pd
	}
	if accessAndMobilitySubscriptionData != nil {
		var tmp models.AccessAndMobilitySubscriptionData
		if err := mapstructure.Decode(accessAndMobilitySubscriptionData, &tmp); err != nil {
			logger.DataRepoLog.Errorf(
				"QueryProvisionedDataProcedure accessAndMobilitySubscriptionData decode err: %+v", err)
			return nil, openapi.ProblemDetailsSystemFailure(err.Error())
		}
		provisionedDataSets.AmData = &tmp
	}

	collName = "subscriptionData.provisionedData.smfSelectionSubscriptionData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	smfSelectionSubscriptionData, pd := getDataFromDB(collName, filter)
	if pd != nil && pd.Status == http.StatusInternalServerError {
		logger.DataRepoLog.Errorf("QueryProvisionedDataProcedure get smfSelectionSubscriptionData err: %s", pd.Detail)
		return nil, pd
	}
	if smfSelectionSubscriptionData != nil {
		var tmp models.SmfSelectionSubscriptionData
		if err := mapstructure.Decode(smfSelectionSubscriptionData, &tmp); err != nil {
			logger.DataRepoLog.Errorf(
				"QueryProvisionedDataProcedure smfSelectionSubscriptionData decode err: %+v", err)
			return nil, openapi.ProblemDetailsSystemFailure(err.Error())
		}
		provisionedDataSets.SmfSelData = &tmp
	}

	collName = "subscriptionData.provisionedData.smsData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	smsSubscriptionData, pd := getDataFromDB(collName, filter)
	if pd != nil && pd.Status == http.StatusInternalServerError {
		logger.DataRepoLog.Errorf("QueryProvisionedDataProcedure get smsSubscriptionData err: %s", pd.Detail)
		return nil, pd
	}
	if smsSubscriptionData != nil {
		var tmp models.SmsSubscriptionData
		if err := mapstructure.Decode(smsSubscriptionData, &tmp); err != nil {
			logger.DataRepoLog.Errorf(
				"QueryProvisionedDataProcedure smsSubscriptionData decode err: %+v", err)
			return nil, openapi.ProblemDetailsSystemFailure(err.Error())
		}
		provisionedDataSets.SmsSubsData = &tmp
	}

	collName = "subscriptionData.provisionedData.smData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	sessionManagementSubscriptionDatas, err := mongoapi.RestfulAPIGetMany(collName, filter)
	if err != nil {
		logger.DataRepoLog.Errorf("QueryProvisionedDataProcedure get sessionManagementSubscriptionDatas err: %+v", err)
		return nil, openapi.ProblemDetailsSystemFailure(err.Error())
	}
	if sessionManagementSubscriptionDatas != nil {
		var tmp []models.SessionManagementSubscriptionData
		if err := mapstructure.Decode(sessionManagementSubscriptionDatas, &tmp); err != nil {
			logger.DataRepoLog.Errorf(
				"QueryProvisionedDataProcedure sessionManagementSubscriptionDatas decode err: %+v", err)
			return nil, openapi.ProblemDetailsSystemFailure(err.Error())
		}
		for _, smData := range tmp {
			dnnConfigurations := smData.DnnConfigurations
			tmpDnnConfigurations := make(map[string]models.DnnConfiguration)
			for escapedDnn, dnnConf := range dnnConfigurations {
				dnn := util.UnescapeDnn(escapedDnn)
				tmpDnnConfigurations[dnn] = dnnConf
			}
			smData.DnnConfigurations = tmpDnnConfigurations
		}
		provisionedDataSets.SmData = tmp
	}

	collName = "subscriptionData.provisionedData.traceData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	traceData, pd := getDataFromDB(collName, filter)
	if pd != nil && pd.Status == http.StatusInternalServerError {
		logger.DataRepoLog.Errorf("QueryProvisionedDataProcedure get traceData err: %s", pd.Detail)
		return nil, pd
	}
	if traceData != nil {
		var tmp models.TraceData
		if err := mapstructure.Decode(traceData, &tmp); err != nil {
			logger.DataRepoLog.Errorf("QueryProvisionedDataProcedure traceData decode err: %+v", err)
			return nil, openapi.ProblemDetailsSystemFailure(err.Error())
		}
		provisionedDataSets.TraceData = &tmp
	}

	collName = "subscriptionData.provisionedData.smsMngData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	smsManagementSubscriptionData, pd := getDataFromDB(collName, filter)
	if pd != nil && pd.Status == http.StatusInternalServerError {
		logger.DataRepoLog.Errorf(
			"QueryProvisionedDataProcedure get smsManagementSubscriptionData err: %s", pd.Detail)
		return nil, pd
	}
	if smsManagementSubscriptionData != nil {
		var tmp models.SmsManagementSubscriptionData
		if err := mapstructure.Decode(smsManagementSubscriptionData, &tmp); err != nil {
			logger.DataRepoLog.Errorf(
				"QueryProvisionedDataProcedure smsManagementSubscriptionData decode err: %+v", err)
			return nil, openapi.ProblemDetailsSystemFailure(err.Error())
		}
		provisionedDataSets.SmsMngData = &tmp
	}

	if reflect.DeepEqual(provisionedDataSets, models.ProvisionedDataSets{}) {
		return nil, util.ProblemDetailsNotFound("DATA_NOT_FOUND")
	}
	return &provisionedDataSets, nil
}

func HandleModifyPpData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle ModifyPpData")

	collName := "subscriptionData.ppData"
	patchItem := request.Body.([]models.PatchItem)
	ueId := request.Params["ueId"]

	problemDetails := ModifyPpDataProcedure(collName, ueId, patchItem)
	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func ModifyPpDataProcedure(collName string, ueId string, patchItem []models.PatchItem) *models.ProblemDetails {
	filter := bson.M{"ueId": ueId}
	if err := patchDataToDBAndNotify(collName, ueId, patchItem, filter); err != nil {
		logger.DataRepoLog.Errorf("ModifyPpDataProcedure err: %+v", err)
		return util.ProblemDetailsModifyNotAllowed("")
	}
	return nil
}

func HandleGetIdentityData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle GetIdentityData")

	ueId := request.Params["ueId"]
	collName := "subscriptionData.identityData"

	response, problemDetails := GetIdentityDataProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func GetIdentityDataProcedure(collName string, ueId string) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("GetIdentityDataProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleGetOdbData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle GetOdbData")

	ueId := request.Params["ueId"]
	collName := "subscriptionData.operatorDeterminedBarringData"

	response, problemDetails := GetOdbDataProcedure(collName, ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func GetOdbDataProcedure(collName string, ueId string) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("GetOdbDataProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleGetSharedData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle GetSharedData")

	var sharedDataIds []string
	if len(request.Query["shared-data-ids"]) != 0 {
		sharedDataIds = request.Query["shared-data-ids"]
		if strings.Contains(sharedDataIds[0], ",") {
			sharedDataIds = strings.Split(sharedDataIds[0], ",")
		}
	}
	collName := "subscriptionData.sharedData"

	response, problemDetails := GetSharedDataProcedure(collName, sharedDataIds)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func GetSharedDataProcedure(collName string, sharedDataIds []string) (*[]map[string]interface{},
	*models.ProblemDetails,
) {
	var sharedDataArray []map[string]interface{}
	for _, sharedDataId := range sharedDataIds {
		filter := bson.M{"sharedDataId": sharedDataId}
		sharedData, pd := getDataFromDB(collName, filter)
		if pd != nil && pd.Status == http.StatusInternalServerError {
			logger.DataRepoLog.Errorf("GetSharedDataProcedure err: %s", pd.Detail)
			return nil, pd
		}
		if sharedData != nil {
			sharedDataArray = append(sharedDataArray, sharedData)
		}
	}

	if sharedDataArray == nil {
		return nil, util.ProblemDetailsNotFound("DATA_NOT_FOUND")
	}
	return &sharedDataArray, nil
}

func HandleRemovesdmSubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle RemovesdmSubscriptions")

	ueId := request.Params["ueId"]
	subsId := request.Params["subsId"]

	problemDetails := RemovesdmSubscriptionsProcedure(ueId, subsId)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func RemovesdmSubscriptionsProcedure(ueId string, subsId string) *models.ProblemDetails {
	udrSelf := udr_context.GetSelf()
	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		return util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}

	UESubsData := value.(*udr_context.UESubsData)
	_, ok = UESubsData.SdmSubscriptions[subsId]

	if !ok {
		return util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}
	delete(UESubsData.SdmSubscriptions, subsId)

	return nil
}

func HandleUpdatesdmsubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle Updatesdmsubscriptions")

	ueId := request.Params["ueId"]
	subsId := request.Params["subsId"]
	SdmSubscription := request.Body.(models.SdmSubscription)

	problemDetails := UpdatesdmsubscriptionsProcedure(ueId, subsId, SdmSubscription)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func UpdatesdmsubscriptionsProcedure(ueId string, subsId string,
	SdmSubscription models.SdmSubscription,
) *models.ProblemDetails {
	udrSelf := udr_context.GetSelf()
	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		return util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}

	UESubsData := value.(*udr_context.UESubsData)
	_, ok = UESubsData.SdmSubscriptions[subsId]

	if !ok {
		return util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}
	SdmSubscription.SubscriptionId = subsId
	UESubsData.SdmSubscriptions[subsId] = &SdmSubscription

	return nil
}

func HandleCreateSdmSubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle CreateSdmSubscriptions")

	SdmSubscription := request.Body.(models.SdmSubscription)
	collName := "subscriptionData.contextData.amfNon3gppAccess"
	ueId := request.Params["ueId"]

	locationHeader, SdmSubscription := CreateSdmSubscriptionsProcedure(SdmSubscription, collName, ueId)

	headers := http.Header{}
	headers.Set("Location", locationHeader)
	return httpwrapper.NewResponse(http.StatusCreated, headers, SdmSubscription)
}

func CreateSdmSubscriptionsProcedure(SdmSubscription models.SdmSubscription,
	collName string, ueId string,
) (string, models.SdmSubscription) {
	udrSelf := udr_context.GetSelf()

	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		udrSelf.UESubsCollection.Store(ueId, new(udr_context.UESubsData))
		value, _ = udrSelf.UESubsCollection.Load(ueId)
	}
	UESubsData := value.(*udr_context.UESubsData)
	if UESubsData.SdmSubscriptions == nil {
		UESubsData.SdmSubscriptions = make(map[string]*models.SdmSubscription)
	}

	newSubscriptionID := strconv.Itoa(udrSelf.SdmSubscriptionIDGenerator)
	SdmSubscription.SubscriptionId = newSubscriptionID
	UESubsData.SdmSubscriptions[newSubscriptionID] = &SdmSubscription
	udrSelf.SdmSubscriptionIDGenerator++

	/* Contains the URI of the newly created resource, according
	   to the structure: {apiRoot}/subscription-data/{ueId}/context-data/sdm-subscriptions/{subsId}' */
	locationHeader := fmt.Sprintf("%s/subscription-data/%s/context-data/sdm-subscriptions/%s",
		udrSelf.GetIPv4GroupUri(udr_context.NUDR_DR), ueId, newSubscriptionID)

	return locationHeader, SdmSubscription
}

func HandleQuerysdmsubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle Querysdmsubscriptions")

	ueId := request.Params["ueId"]

	response, problemDetails := QuerysdmsubscriptionsProcedure(ueId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QuerysdmsubscriptionsProcedure(ueId string) (*[]models.SdmSubscription, *models.ProblemDetails) {
	udrSelf := udr_context.GetSelf()

	value, ok := udrSelf.UESubsCollection.Load(ueId)
	if !ok {
		return nil, util.ProblemDetailsNotFound("USER_NOT_FOUND")
	}

	UESubsData := value.(*udr_context.UESubsData)
	var sdmSubscriptionSlice []models.SdmSubscription

	for _, v := range UESubsData.SdmSubscriptions {
		sdmSubscriptionSlice = append(sdmSubscriptionSlice, *v)
	}
	return &sdmSubscriptionSlice, nil
}

func HandleQuerySmData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QuerySmData")

	collName := "subscriptionData.provisionedData.smData"
	ueId := request.Params["ueId"]
	servingPlmnId := request.Params["servingPlmnId"]
	singleNssai := models.Snssai{}
	singleNssaiQuery := request.Query.Get("single-nssai")
	err := json.Unmarshal([]byte(singleNssaiQuery), &singleNssai)
	if err != nil {
		logger.DataRepoLog.Warnln(err)
	}

	dnn := request.Query.Get("dnn")
	response := QuerySmDataProcedure(collName, ueId, servingPlmnId, singleNssai, dnn)

	return httpwrapper.NewResponse(http.StatusOK, nil, response)
}

func QuerySmDataProcedure(collName string, ueId string, servingPlmnId string,
	singleNssai models.Snssai, dnn string,
) *[]map[string]interface{} {
	filter := bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}

	if !reflect.DeepEqual(singleNssai, models.Snssai{}) {
		if singleNssai.Sd == "" {
			filter["singleNssai.sst"] = singleNssai.Sst
		} else {
			filter["singleNssai.sst"] = singleNssai.Sst
			filter["singleNssai.sd"] = singleNssai.Sd
		}
	}

	if dnn != "" {
		dnnKey := util.EscapeDnn(dnn)
		filter["dnnConfigurations."+dnnKey] = bson.M{"$exists": true}
	}

	sessionManagementSubscriptionDatas, err := mongoapi.RestfulAPIGetMany(collName, filter)
	if err != nil {
		logger.DataRepoLog.Errorf("QuerySmDataProcedure err: %+v", err)
		return nil
	}
	for _, smData := range sessionManagementSubscriptionDatas {
		var tmpSmData models.SessionManagementSubscriptionData
		err := json.Unmarshal(util.MapToByte(smData), &tmpSmData)
		if err != nil {
			logger.DataRepoLog.Debug("SmData Unmarshal error")
			continue
		}
		dnnConfigurations := tmpSmData.DnnConfigurations
		tmpDnnConfigurations := make(map[string]models.DnnConfiguration)
		for escapedDnn, dnnConf := range dnnConfigurations {
			dnn := util.UnescapeDnn(escapedDnn)
			tmpDnnConfigurations[dnn] = dnnConf
		}
		smData["DnnConfigurations"] = tmpDnnConfigurations
	}
	return &sessionManagementSubscriptionDatas
}

func HandleCreateSmfContextNon3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle CreateSmfContextNon3gpp")

	SmfRegistration := request.Body.(models.SmfRegistration)
	collName := "subscriptionData.contextData.smfRegistrations"
	ueId := request.Params["ueId"]
	pduSessionId, err := strconv.ParseInt(request.Params["pduSessionId"], 10, 64)
	if err != nil {
		logger.DataRepoLog.Warnln(err)
	}

	response, status := CreateSmfContextNon3gppProcedure(SmfRegistration, collName, ueId, pduSessionId)

	if status == http.StatusCreated {
		return httpwrapper.NewResponse(http.StatusCreated, nil, response)
	} else if status == http.StatusOK {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func CreateSmfContextNon3gppProcedure(SmfRegistration models.SmfRegistration,
	collName string, ueId string, pduSessionIdInt int64,
) (bson.M, int) {
	putData := util.ToBsonM(SmfRegistration)
	putData["ueId"] = ueId
	putData["pduSessionId"] = int32(pduSessionIdInt)

	filter := bson.M{"ueId": ueId, "pduSessionId": pduSessionIdInt}
	existed, err := mongoapi.RestfulAPIPutOne(collName, filter, putData)
	if err != nil {
		logger.DataRepoLog.Errorf("CreateSmfContextNon3gppProcedure err: %+v", err)
		return nil, http.StatusInternalServerError
	}

	if existed {
		return putData, http.StatusOK
	}
	return putData, http.StatusCreated
}

func HandleDeleteSmfContext(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle DeleteSmfContext")

	collName := "subscriptionData.contextData.smfRegistrations"
	ueId := request.Params["ueId"]
	pduSessionId := request.Params["pduSessionId"]

	DeleteSmfContextProcedure(collName, ueId, pduSessionId)
	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func DeleteSmfContextProcedure(collName string, ueId string, pduSessionId string) {
	pduSessionIdInt, err := strconv.ParseInt(pduSessionId, 10, 32)
	if err != nil {
		logger.DataRepoLog.Error(err)
	}
	filter := bson.M{"ueId": ueId, "pduSessionId": pduSessionIdInt}
	deleteDataFromDB(collName, filter)
}

func HandleQuerySmfRegistration(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QuerySmfRegistration")

	ueId := request.Params["ueId"]
	pduSessionId := request.Params["pduSessionId"]
	collName := "subscriptionData.contextData.smfRegistrations"

	response, problemDetails := QuerySmfRegistrationProcedure(collName, ueId, pduSessionId)
	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QuerySmfRegistrationProcedure(collName string, ueId string,
	pduSessionId string,
) (*map[string]interface{}, *models.ProblemDetails) {
	pduSessionIdInt, err := strconv.ParseInt(pduSessionId, 10, 32)
	if err != nil {
		logger.DataRepoLog.Error(err)
	}

	filter := bson.M{"ueId": ueId, "pduSessionId": pduSessionIdInt}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QuerySmfRegistrationProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleQuerySmfRegList(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QuerySmfRegList")

	collName := "subscriptionData.contextData.smfRegistrations"
	ueId := request.Params["ueId"]
	response := QuerySmfRegListProcedure(collName, ueId)

	if response == nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, []map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	}
}

func QuerySmfRegListProcedure(collName string, ueId string) *[]map[string]interface{} {
	filter := bson.M{"ueId": ueId}
	smfRegList, err := mongoapi.RestfulAPIGetMany(collName, filter)
	if err != nil {
		logger.DataRepoLog.Errorf("QuerySmfRegListProcedure err: %+v", err)
		return nil
	}

	if smfRegList != nil {
		return &smfRegList
	}
	return nil
}

func HandleQuerySmfSelectData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QuerySmfSelectData")

	collName := "subscriptionData.provisionedData.smfSelectionSubscriptionData"
	ueId := request.Params["ueId"]
	servingPlmnId := request.Params["servingPlmnId"]
	response, problemDetails := QuerySmfSelectDataProcedure(collName, ueId, servingPlmnId)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func QuerySmfSelectDataProcedure(collName string, ueId string,
	servingPlmnId string,
) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QuerySmfSelectDataProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleCreateSmsfContext3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle CreateSmsfContext3gpp")

	SmsfRegistration := request.Body.(models.SmsfRegistration)
	collName := "subscriptionData.contextData.smsf3gppAccess"
	ueId := request.Params["ueId"]

	CreateSmsfContext3gppProcedure(collName, ueId, SmsfRegistration)

	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func CreateSmsfContext3gppProcedure(collName string, ueId string, SmsfRegistration models.SmsfRegistration) {
	putData := util.ToBsonM(SmsfRegistration)
	putData["ueId"] = ueId
	filter := bson.M{"ueId": ueId}

	_, err := mongoapi.RestfulAPIPutOne(collName, filter, putData)
	if err != nil {
		logger.DataRepoLog.Errorf("CreateSmsfContext3gppProcedure err: %+v", err)
	}
}

func HandleDeleteSmsfContext3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle DeleteSmsfContext3gpp")

	collName := "subscriptionData.contextData.smsf3gppAccess"
	ueId := request.Params["ueId"]

	DeleteSmsfContext3gppProcedure(collName, ueId)
	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func DeleteSmsfContext3gppProcedure(collName string, ueId string) {
	filter := bson.M{"ueId": ueId}
	deleteDataFromDB(collName, filter)
}

func HandleQuerySmsfContext3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QuerySmsfContext3gpp")

	collName := "subscriptionData.contextData.smsf3gppAccess"
	ueId := request.Params["ueId"]

	response, problemDetails := QuerySmsfContext3gppProcedure(collName, ueId)
	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QuerySmsfContext3gppProcedure(collName string, ueId string) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QuerySmsfContext3gppProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleCreateSmsfContextNon3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle CreateSmsfContextNon3gpp")

	SmsfRegistration := request.Body.(models.SmsfRegistration)
	collName := "subscriptionData.contextData.smsfNon3gppAccess"
	ueId := request.Params["ueId"]

	CreateSmsfContextNon3gppProcedure(SmsfRegistration, collName, ueId)

	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func CreateSmsfContextNon3gppProcedure(SmsfRegistration models.SmsfRegistration, collName string, ueId string) {
	putData := util.ToBsonM(SmsfRegistration)
	putData["ueId"] = ueId
	filter := bson.M{"ueId": ueId}

	_, err := mongoapi.RestfulAPIPutOne(collName, filter, putData)
	if err != nil {
		logger.DataRepoLog.Errorf("CreateSmsfContextNon3gppProcedure err: %+v", err)
	}
}

func HandleDeleteSmsfContextNon3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle DeleteSmsfContextNon3gpp")

	collName := "subscriptionData.contextData.smsfNon3gppAccess"
	ueId := request.Params["ueId"]

	DeleteSmsfContextNon3gppProcedure(collName, ueId)
	return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
}

func DeleteSmsfContextNon3gppProcedure(collName string, ueId string) {
	filter := bson.M{"ueId": ueId}
	deleteDataFromDB(collName, filter)
}

func HandleQuerySmsfContextNon3gpp(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QuerySmsfContextNon3gpp")

	ueId := request.Params["ueId"]
	collName := "subscriptionData.contextData.smsfNon3gppAccess"

	response, problemDetails := QuerySmsfContextNon3gppProcedure(collName, ueId)
	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QuerySmsfContextNon3gppProcedure(collName string, ueId string) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QuerySmsfContextNon3gppProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleQuerySmsMngData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QuerySmsMngData")

	collName := "subscriptionData.provisionedData.smsMngData"
	ueId := request.Params["ueId"]
	servingPlmnId := request.Params["servingPlmnId"]
	response, problemDetails := QuerySmsMngDataProcedure(collName, ueId, servingPlmnId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QuerySmsMngDataProcedure(collName string, ueId string,
	servingPlmnId string,
) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QuerySmsMngDataProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandleQuerySmsData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QuerySmsData")

	ueId := request.Params["ueId"]
	servingPlmnId := request.Params["servingPlmnId"]
	collName := "subscriptionData.provisionedData.smsData"

	response, problemDetails := QuerySmsDataProcedure(collName, ueId, servingPlmnId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QuerySmsDataProcedure(collName string, ueId string,
	servingPlmnId string,
) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QuerySmsDataProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}

func HandlePostSubscriptionDataSubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle PostSubscriptionDataSubscriptions")

	SubscriptionDataSubscriptions := request.Body.(models.SubscriptionDataSubscriptions)

	locationHeader := PostSubscriptionDataSubscriptionsProcedure(SubscriptionDataSubscriptions)

	headers := http.Header{}
	headers.Set("Location", locationHeader)
	return httpwrapper.NewResponse(http.StatusCreated, headers, SubscriptionDataSubscriptions)
}

func PostSubscriptionDataSubscriptionsProcedure(
	SubscriptionDataSubscriptions models.SubscriptionDataSubscriptions,
) string {
	udrSelf := udr_context.GetSelf()

	newSubscriptionID := strconv.Itoa(udrSelf.SubscriptionDataSubscriptionIDGenerator)
	udrSelf.SubscriptionDataSubscriptions[newSubscriptionID] = &SubscriptionDataSubscriptions
	udrSelf.SubscriptionDataSubscriptionIDGenerator++

	/* Contains the URI of the newly created resource, according
	   to the structure: {apiRoot}/subscription-data/subs-to-notify/{subsId} */
	locationHeader := fmt.Sprintf("%s/subscription-data/subs-to-notify/%s",
		udrSelf.GetIPv4GroupUri(udr_context.NUDR_DR), newSubscriptionID)

	return locationHeader
}

func HandleRemovesubscriptionDataSubscriptions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle RemovesubscriptionDataSubscriptions")

	subsId := request.Params["subsId"]

	problemDetails := RemovesubscriptionDataSubscriptionsProcedure(subsId)

	if problemDetails == nil {
		return httpwrapper.NewResponse(http.StatusNoContent, nil, map[string]interface{}{})
	} else {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func RemovesubscriptionDataSubscriptionsProcedure(subsId string) *models.ProblemDetails {
	udrSelf := udr_context.GetSelf()
	_, ok := udrSelf.SubscriptionDataSubscriptions[subsId]
	if !ok {
		return util.ProblemDetailsNotFound("SUBSCRIPTION_NOT_FOUND")
	}
	delete(udrSelf.SubscriptionDataSubscriptions, subsId)
	return nil
}

func HandleQueryTraceData(request *httpwrapper.Request) *httpwrapper.Response {
	logger.DataRepoLog.Infof("Handle QueryTraceData")

	collName := "subscriptionData.provisionedData.traceData"
	ueId := request.Params["ueId"]
	servingPlmnId := request.Params["servingPlmnId"]

	response, problemDetails := QueryTraceDataProcedure(collName, ueId, servingPlmnId)

	if response != nil {
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	pd := util.ProblemDetailsUpspecified("")
	return httpwrapper.NewResponse(int(pd.Status), nil, pd)
}

func QueryTraceDataProcedure(collName string, ueId string,
	servingPlmnId string,
) (*map[string]interface{}, *models.ProblemDetails) {
	filter := bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	data, pd := getDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QueryTraceDataProcedure err: %s", pd.Detail)
		return nil, pd
	}
	return &data, nil
}
