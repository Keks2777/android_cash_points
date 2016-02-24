package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/tarantool/go-tarantool"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type TestRequest struct {
	RequestType string
	EndpointUrl string
	HandlerUrl  string
	Data        string
}

type TestResponse struct {
	Code int
	Data []byte
}

func readResponse(w *httptest.ResponseRecorder) (TestResponse, error) {
	response := TestResponse{}

	response.Code = w.Code

	data, err := ioutil.ReadAll(w.Body)
	if err != nil {
		err = errors.New("Cannot read response body: " + err.Error())
	} else {
		response.Data = data
	}
	return response, err
}

func diff(expected, received []byte) (string, error) {
	differ := gojsondiff.New()
	d, err := differ.Compare(expected, received)
	if err != nil {
		return "", errors.New("Failed to compare json pair: " + err.Error())
	}

	if !d.Modified() {
		return "", nil
	}

	var expectedJson map[string]interface{}
	json.Unmarshal(expected, &expectedJson)
	formatter := formatter.NewAsciiFormatter(expectedJson)
	formatter.ShowArrayIndex = true
	diffString, err := formatter.Format(d)
	if err != nil {
		// No error can occur
	}

	return diffString, nil
}

func tarantoolConnect() (*tarantool.Connection, error) {
	tntUrl := "localhost:3301"
	tntOpts := tarantool.Opts{
		User: "admin",
		Pass: "admin",
	}

	return tarantool.Connect(tntUrl, tntOpts)
}

func testRequest(request TestRequest, handler EndpointCallback) *httptest.ResponseRecorder {
	var req *http.Request = nil

	if request.Data != "" {
		req, _ = http.NewRequest(request.RequestType, request.EndpointUrl, bytes.NewBufferString(request.Data))
	} else {
		req, _ = http.NewRequest(request.RequestType, request.EndpointUrl, nil)
	}

	req.Header.Add("Id", "1")

	w := httptest.NewRecorder()
	m := mux.NewRouter()
	if request.HandlerUrl == "" {
		request.HandlerUrl = request.EndpointUrl
	}
	m.HandleFunc(request.HandlerUrl, handler).Methods(request.RequestType)
	m.ServeHTTP(w, req)

	return w
}

func checkHttpCode(t *testing.T, got, expected int) {
	if got != expected {
		t.Errorf("Expected %d %s but got %d", expected, http.StatusText(expected), got)
	}
}

func checkJsonResponse(t *testing.T, got, expected []byte) {
	diffStr, err := diff(expected, got)
	if err != nil {
		t.Errorf("Failed to compare json pair: %v", err)
	}
	if diffStr != "" {
		t.Errorf("\n%s", diffStr)
	}
}

// ======================================================================

func TestPing(t *testing.T) {
	tntUrl := "localhost:3301"
	tntOpts := tarantool.Opts{
		User: "admin",
		Pass: "admin",
	}

	tnt, err := tarantool.Connect(tntUrl, tntOpts)
	if err != nil {
		t.Errorf("Connection to tarantool failed: %v", err)
	}
	defer tnt.Close()

	url, handler := handlerPing(tnt)
	request := TestRequest{RequestType: "GET", EndpointUrl: url}
	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusOK)

	expected := Message{Text: "pong"}
	expectedJson, _ := json.Marshal(expected)

	checkJsonResponse(t, response.Data, expectedJson)
}

// ======================================================================

type Town struct {
	Id             uint32  `json:"id"`
	Name           string  `json:"name"`
	NameTr         string  `json:"name_tr"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	RegionId       uint32  `json:"region_id"`
	RegionalCenter bool    `json:"regional_center"`
	Big            bool    `json:"big"`
	Zoom           uint32  `json:"zoom"`
}

func TestTown(t *testing.T) {
	tnt, err := tarantoolConnect()
	if err != nil {
		t.Errorf("Connection to tarantool failed: %v", err)
	}
	defer tnt.Close()

	url, handler := handlerTown(tnt)
	request := TestRequest{RequestType: "GET", EndpointUrl: "/town/4", HandlerUrl: url}
	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}
	checkHttpCode(t, response.Code, http.StatusOK)

	expected := Town{
		Id:             4,
		Name:           "Москва",
		NameTr:         "Moskva",
		Longitude:      37.61775970459,
		Latitude:       55.755771636963,
		RegionId:       3,
		RegionalCenter: true,
		Big:            true,
		Zoom:           10,
	}
	expectedJson, _ := json.Marshal(expected)

	checkJsonResponse(t, response.Data, expectedJson)
}

// ======================================================================

type CashpointShort struct {
	Id             uint32  `json:"id,omitempty"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	Type           string  `json:"type"`
	BankId         uint32  `json:"bank_id"`
	TownId         uint32  `json:"town_id"`
	Address        string  `json:"address"`
	AddressComment string  `json:"address_comment"`
	MetroName      string  `json:"metro_name"`
	FreeAccess     bool    `json:"free_access"`
	MainOffice     bool    `json:"main_office"`
	WithoutWeekend bool    `json:"without_weekend"`
	RoundTheClock  bool    `json:"round_the_clock"`
	WorksAsShop    bool    `json:"works_as_shop"`
	Schedule       string  `json:"schedule"`
	Tel            string  `json:"tel"`
	Additional     string  `json:"additional"`
	Rub            bool    `json:"rub"`
	Usd            bool    `json:"usd"`
	Eur            bool    `json:"eur"`
	CashIn         bool    `json:"cash_in"`
}

type CashpointFull struct {
	CashpointShort
	Version        uint32  `json:"version"`
	//	Timestamp      uint64  `json:"timestamp"` // TODO: timestamp on server
	Approved bool `json:"approved"`
}

func TestCashpoint(t *testing.T) {
	tnt, err := tarantoolConnect()
	if err != nil {
		t.Errorf("Connection to tarantool failed: %v", err)
	}
	defer tnt.Close()

	url, handler := handlerCashpoint(tnt)
	var id uint32 = 7138832
	request := TestRequest{
		RequestType: "GET",
		EndpointUrl: "/cashpoint/" + strconv.FormatUint(uint64(id), 10),
		HandlerUrl:  url,
	}
	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusOK)

	cpShort := CashpointShort {
		Id:             id,
		Longitude:      37.562019348145,
		Latitude:       55.6633644104,
		Type:           "atm",
		BankId:         2764,
		TownId:         4,
		Address:        "г. Москва, ул. Новочеремушкинская, д. 69",
		AddressComment: "ОАО «Вниизарубежгеология»",
		MetroName:      "",
		FreeAccess:     true,
		MainOffice:     false,
		WithoutWeekend: false,
		RoundTheClock:  false,
		WorksAsShop:    true,
		Schedule:       "",
		Tel:            "",
		Additional:     "",
		Rub:            true,
		Usd:            false,
		Eur:            false,
		CashIn:         false,
	}

	cp := CashpointFull{
		CashpointShort: cpShort,
		Version:        0,
		//Timestamp: 0,
		Approved: true,
	}
	expectedJson, _ := json.Marshal(cp)

	diffStr, err := diff(expectedJson, response.Data)
	if err != nil {
		t.Errorf("Failed to compare json pair: %v", err)
	}
	if diffStr != "" {
		t.Errorf("\n%s", diffStr)
	}
}

// ======================================================================

type QuadKeyRequest struct {
	Longitude float64 `json:"longitude,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Zoom      uint32  `json:"zoom,omitempty"`
}

type QuadKeyResponse struct {
	QuadKey string `json:"quadkey"`
}

func TestQuadKeyFromCoord(t *testing.T) {
	tnt, err := tarantoolConnect()
	if err != nil {
		t.Errorf("Connection to tarantool failed: %v", err)
	}
	defer tnt.Close()

	url, handler := handlerCoordToQuadKey(tnt)

	// empty request
	quadKeyReq := QuadKeyRequest{}
	reqJson, _ := json.Marshal(quadKeyReq)

	request := TestRequest{
		RequestType: "POST",
		EndpointUrl: url,
		Data: string(reqJson),
	}

	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusBadRequest)

	// request with missing Latitude
	quadKeyReq.Longitude = 56.6
	reqJson, _ = json.Marshal(quadKeyReq)

	request.Data = string(reqJson)

	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusBadRequest)

	// normal request
	quadKeyReq.Latitude = 34.84
	reqJson, _ = json.Marshal(quadKeyReq)

	request.Data = string(reqJson)

	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := QuadKeyResponse{QuadKey: "3032100220113311"}
	expectedJson, _ := json.Marshal(expected)

	checkHttpCode(t, response.Code, http.StatusOK)
	checkJsonResponse(t, response.Data, expectedJson)

	// request with zoom
	quadKeyReq.Zoom = 16
	reqJson, _ = json.Marshal(quadKeyReq)

	request.Data = string(reqJson)

	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusOK)
	checkJsonResponse(t, response.Data, expectedJson)

	// request with lower zoom
	quadKeyReq.Zoom = 12
	reqJson, _ = json.Marshal(quadKeyReq)

	request.Data = string(reqJson)

	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	expected.QuadKey = "303210022011"
	expectedJson, _ = json.Marshal(expected)

	checkHttpCode(t, response.Code, http.StatusOK)
	checkJsonResponse(t, response.Data, expectedJson)
}

// ======================================================================

func TestQuadTreeBranch(t *testing.T) {
	tnt, err := tarantoolConnect()
	if err != nil {
		t.Errorf("Connection to tarantool failed: %v", err)
	}
	defer tnt.Close()

	url, handler := handlerQuadTreeBranch(tnt)
	request := TestRequest{
		RequestType: "GET",
		EndpointUrl: "/quadtree/branch/3201323213002023",
		HandlerUrl: url,
	}

	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	if response.Code != http.StatusOK {
		t.Errorf("Expected 200 OK but got %d", response.Code)
	}

	// test short quadkey
	request.EndpointUrl = "/quadtree/branch/3201323213002"
	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	if response.Code != http.StatusOK {
		t.Errorf("Expected 200 OK but got %d", response.Code)
	}

	// test empty quadkey
	request.EndpointUrl = "/quadtree/branch/"
	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected 200 OK but got %d", response.Code)
	}

	// test too long quadkey
	request.EndpointUrl = "/quadtree/branch/320132321300211100"
	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 OK but got %d", response.Code)
	}

	// test wrong quadkey
	request.EndpointUrl = "/quadtree/branch/3201323253002023"
	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected 200 OK but got %d", response.Code)
	}
}

// ======================================================================

type CashpointCreateRequest struct {
	UserId uint32         `json:"user_id"`
	Data   CashpointShort `json:"data"`
}

func TestCashpointCreate(t *testing.T) {
	tnt, err := tarantoolConnect()
	if err != nil {
		t.Errorf("Connection to tarantool failed: %v", err)
	}
	defer tnt.Close()

	cp := CashpointShort{
		Longitude: 37.62644,
		Latitude: 55.75302,
		Type: "atm",
		BankId: 322, // Sberbank
		TownId: 4, // Moscow
		Address: "",
		AddressComment: "",
//		MetroName: "",
		FreeAccess: true,
		MainOffice: false,
		WithoutWeekend: true,
		RoundTheClock: false,
		WorksAsShop: false,
		Schedule: "",
		Tel: "",
		Additional: "",
		Rub: true,
		Usd: false,
		Eur: false,
		CashIn: true,
	}

	reqData := CashpointCreateRequest{
		UserId: 0, // TODO: check against real user
		Data: cp,
	}
	reqJson, _ := json.Marshal(reqData)

	url, handlerCreate := handlerCashpointCreate(tnt)
	request := TestRequest{
		RequestType: "POST",
		EndpointUrl: "/cashpoint",
		HandlerUrl: url,
		Data: string(reqJson),
	}

	response, err := readResponse(testRequest(request, handlerCreate))
	if err != nil {
		t.Errorf("%v", err)
	}
	if response.Code != http.StatusOK {
		t.Errorf("Expected 200 OK but got %d", response.Code)
	}

	var cashpointId uint64 = 0
	err = json.Unmarshal(response.Data, &cashpointId)
	if err != nil {
		t.Errorf("Cannot unpack cashpoint id response: %v => %s", err, string(response.Data))
	}

	// now delete created cashpoint

	url, handlerDelete := handlerCashpointDelete(tnt)
	request = TestRequest{
		RequestType: "DELETE",
		EndpointUrl: "/cashpoint/" + strconv.FormatUint(cashpointId, 10),
		HandlerUrl: url,
	}

	response, err = readResponse(testRequest(request, handlerDelete))
	if err != nil {
		t.Errorf("%v", err)
	}
	if response.Code != http.StatusOK {
		t.Errorf("Expected 200 OK but got %d", response.Code)
	}

	// try to get deleted cashpoint

	url, handlerGet := handlerCashpoint(tnt)
	request = TestRequest{
		RequestType: "GET",
		EndpointUrl: "/cashpoint/" + strconv.FormatUint(cashpointId, 10),
		HandlerUrl: url,
	}

	response, err = readResponse(testRequest(request, handlerGet))
	if err != nil {
		t.Errorf("%v", err)
	}
	if response.Code != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found but got %d", response.Code)
	}
}

// TODO: approved hack test