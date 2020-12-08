package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/api/option"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
)

type AuthInfo struct {
	Type         string `json:"type"`
	ProjectID    string `json:"project_id"`
	PrivateKeyID string `json:"private_key_id"`

	PrivateKey             string `json:"private_key"`
	ClientEmail            string `json:"client_email"`
	ClientID               string `json:"client_id"`
	AuthURI                string `json:"auth_uri"`
	TokenURI               string `json:"token_uri"`
	AuthProviderX509CerURI string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL      string `json:"client_x509_cert_url"`
}

// DialogflowProcessor has all the information for connecting with Dialogflow
type DialogflowProcessor struct {
	auth             AuthInfo
	authJSONFilePath string
	lang             string
	timeZone         string
	sessionClient    *dialogflow.SessionsClient
	ctx              context.Context
}

// NLPResponse is the struct for the response
type NLPResponse struct {
	Intent     string            `json:"intent"`
	Confidence float32           `json:"confidence"`
	Entities   map[string]string `json:"entities"`
}

var dp DialogflowProcessor

func main() {

	data, err := ioutil.ReadFile("./animeshiki-84b60-1b87ee34e207.json")
	if err != nil {
		log.Fatalf("Err while getting auth file: %v", err)
	}

	var cert = &AuthInfo{}
	err = json.Unmarshal(data, cert)
	if err != nil {
		log.Fatalf("Err while unmarshalling auth file: %v", err)
	}

	dp.init(*cert, "animeshiki-84b60-1b87ee34e207.json", "ru", "America/Montevideo")
	talkWithMe()
	fmt.Println("Started listening...")
	http.ListenAndServe(":5000", nil)
}

func talkWithMe() {
	var stop bool
	fmt.Println("Привет поболтаем?")
	for !stop {
		inputReader := bufio.NewReader(os.Stdin)
		message, _ := inputReader.ReadString('\n')

		//fmt.Scanf("%s", &message)
		// Use NLP
		response := dp.processNLP(message, "testUser")
		handleResult(response)
		// fmt.Printf("%#v", response)
		// data, _ := json.Marshal(response)
		// fmt.Printf("%s", string(data))
		//json.NewEncoder(w).Encode(response)
		//json.NewEncoder(w).Encode(response)
	}
}

func handleResult(response NLPResponse) {
	/*
		Узнать описание
	*/
	fmt.Printf("\nentity %s with %f' \n", response.Intent, response.Confidence)
	for k, v := range response.Entities {
		fmt.Printf("entity %s %s\n", k, v)
	}
}

func (dp *DialogflowProcessor) init(
	auth AuthInfo, path, lang, timeZone string,
) (err error) {
	dp.auth = auth
	dp.authJSONFilePath = path
	dp.lang = lang
	dp.timeZone = timeZone

	// Auth process: https://dialogflow.com/docs/reference/v2-auth-setup

	dp.ctx = context.Background()
	sessionClient, err := dialogflow.NewSessionsClient(dp.ctx, option.WithCredentialsFile(dp.authJSONFilePath))
	if err != nil {
		log.Fatal("Error in auth with Dialogflow")
	}
	dp.sessionClient = sessionClient

	return
}

func (dp *DialogflowProcessor) processNLP(rawMessage string, username string) (r NLPResponse) {
	sessionID := username
	log.Println("rawMessage", rawMessage)
	request := dialogflowpb.DetectIntentRequest{
		Session: fmt.Sprintf("projects/%s/agent/sessions/%s", dp.auth.ProjectID, sessionID),
		QueryInput: &dialogflowpb.QueryInput{
			Input: &dialogflowpb.QueryInput_Text{
				Text: &dialogflowpb.TextInput{
					Text:         rawMessage,
					LanguageCode: dp.lang,
				},
			},
		},
		QueryParams: &dialogflowpb.QueryParameters{
			TimeZone: dp.timeZone,
		},
	}
	response, err := dp.sessionClient.DetectIntent(dp.ctx, &request)
	if err != nil {
		log.Fatalf("Error in communication with Dialogflow %s", err.Error())
		return
	}
	fmt.Println("response is", response)
	queryResult := response.GetQueryResult()
	if queryResult.Intent != nil {
		r.Intent = queryResult.Intent.DisplayName
		r.Confidence = float32(queryResult.IntentDetectionConfidence)
	}
	r.Entities = make(map[string]string)
	params := queryResult.Parameters.GetFields()
	if len(params) > 0 {
		for paramName, p := range params {
			//fmt.Printf("Param %s: %s (%s)", paramName, p.GetStringValue(), p.String())
			extractedValue := extractDialogflowEntities(p)
			r.Entities[paramName] = extractedValue
			// v := p.GetStructValue()
			// v.AsMap()
			// for key, value := range v.AsMap() {
			// 	fmt.Println("look", key, value)
			// 	str := value.(string)
			// 	if str != "" {
			// 		r.Entities[paramName] = str
			// 		break
			// 	}
			// }
		}
	}
	return
}

func extractDialogflowEntities(p *structpb.Value) (extractedEntity string) {
	kind := p.GetKind()
	switch kind.(type) {
	case *structpb.Value_StringValue:
		return p.GetStringValue()
	case *structpb.Value_NumberValue:
		return strconv.FormatFloat(p.GetNumberValue(), 'f', 6, 64)
	case *structpb.Value_BoolValue:
		return strconv.FormatBool(p.GetBoolValue())
	case *structpb.Value_StructValue:
		s := p.GetStructValue()
		fields := s.GetFields()
		extractedEntity = ""
		for key, value := range fields {
			if key == "amount" {
				extractedEntity = fmt.Sprintf("%s%s", extractedEntity, strconv.FormatFloat(value.GetNumberValue(), 'f', 6, 64))
			}
			if key == "unit" {
				extractedEntity = fmt.Sprintf("%s%s", extractedEntity, value.GetStringValue())
			}
			if key == "date_time" {
				extractedEntity = fmt.Sprintf("%s%s", extractedEntity, value.GetStringValue())
			}
			// @TODO: Other entity types can be added here
		}
		return extractedEntity
	case *structpb.Value_ListValue:
		list := p.GetListValue()
		if len(list.GetValues()) > 1 {
			// @TODO: Extract more values
		}
		extractedEntity = extractDialogflowEntities(list.GetValues()[0])
		return extractedEntity
	default:
		return ""
	}
}
