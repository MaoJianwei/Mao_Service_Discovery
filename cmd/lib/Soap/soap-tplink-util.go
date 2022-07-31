package Soap

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	SOAP_HEADER_KEY = "SOAPAction"


	/* ============================== ipc ============================== */
	SOAP_URL_WANIPConnection   = "http://192.168.1.1:1900/ipc"

	// ========== GetUptime ==========
	SOAP_HEADER_VALUE_GetUptime 		= "urn:schemas-upnp-org:service:WANIPConnection:1#GetStatusInfo"
	SOAP_MSG_GetUptime          		= "<s:Body><u:GetStatusInfo></u:GetStatusInfo></s:Body>"
	SOAP_KEYWORD_GetUptime				= "NewUptime"

	// ========== GetExternalIPAddress ==========
	SOAP_HEADER_VALUE_GetExternalIPAddress	= "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1#GetExternalIPAddress"
	SOAP_MSG_GetExternalIPAddress			= "<s:Body><u:GetExternalIPAddress></u:GetExternalIPAddress></s:Body>"
	SOAP_KEYWORD_GetExternalIPAddress		= "NewExternalIPAddress"


	/* ============================== ifc ============================== */
	SOAP_URL_WANCommonInterfaceConfig   = "http://192.168.1.1:1900/ifc"

	// ========== GetTotalBytesSent ==========
	SOAP_HEADER_VALUE_GetTotalBytesSent = "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1#GetTotalBytesSent"
	SOAP_MSG_GetTotalBytesSent          = "<s:Body><u:GetTotalBytesSent></u:GetTotalBytesSent></s:Body>"
	SOAP_KEYWORD_GetTotalBytesSent		= "NewTotalBytesSent"

	// ========== GetTotalBytesReceived ==========
	SOAP_HEADER_VALUE_GetTotalBytesReceived = "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1#GetTotalBytesReceived"
	SOAP_MSG_GetTotalBytesReceived          = "<s:Body><u:GetTotalBytesReceived></u:GetTotalBytesReceived></s:Body>"
	SOAP_KEYWORD_GetTotalBytesReceived		= "NewTotalBytesReceived"

	// ========== GetTotalPacketsSent ==========
	SOAP_HEADER_VALUE_GetTotalPacketsSent = "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1#GetTotalPacketsSent"
	SOAP_MSG_GetTotalPacketsSent          = "<s:Body><u:GetTotalPacketsSent></u:GetTotalPacketsSent></s:Body>"
	SOAP_KEYWORD_GetTotalPacketsSent		= "NewTotalPacketsSent"

	// ========== GetTotalPacketsReceived ==========
	SOAP_HEADER_VALUE_GetTotalPacketsReceived = "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1#GetTotalPacketsReceived"
	SOAP_MSG_GetTotalPacketsReceived          = "<s:Body><u:GetTotalPacketsReceived></u:GetTotalPacketsReceived></s:Body>"
	SOAP_KEYWORD_GetTotalPacketsReceived		= "NewTotalPacketsReceived"
)

func requestSoapData(soapUrl, soapHeader, soapBody string) (*[]byte, error) {

	postData := bytes.NewReader([]byte(soapBody))
	req, err := http.NewRequest("POST", soapUrl, postData)
	if err != nil {
		return nil, err
	}

	req.Header.Set(SOAP_HEADER_KEY, soapHeader)

	client := http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return &body, nil
}

func getDataString(data string, keyword string) (string, error) {
	_, after, found := strings.Cut(data, "<" + keyword + ">")
	if !found {
		return "", errors.New("Can't find the keyword at before: " + keyword)
	}
	before, _, found := strings.Cut(after, "</" + keyword + ">")
	if !found {
		return "", errors.New("Can't find the keyword at after: " + keyword)
	}
	return before, nil
}

func GetTotalBytesSent() (uint64, error) {
	body, err := requestSoapData(SOAP_URL_WANCommonInterfaceConfig, SOAP_HEADER_VALUE_GetTotalBytesSent, SOAP_MSG_GetTotalBytesSent)
	if err != nil {
		return 0, err
	}
	data, err := getDataString(string(*body), SOAP_KEYWORD_GetTotalBytesSent)
	if err != nil {
		return 0, err
	}
	uintData, err := strconv.ParseUint(data, 10, 64)
	if err != nil {
		return 0, err
	}
	return uintData, nil
}

func GetTotalBytesReceived() (uint64, error) {
	body, err := requestSoapData(SOAP_URL_WANCommonInterfaceConfig, SOAP_HEADER_VALUE_GetTotalBytesReceived, SOAP_MSG_GetTotalBytesReceived)
	if err != nil {
		return 0, err
	}
	data, err := getDataString(string(*body), SOAP_KEYWORD_GetTotalBytesReceived)
	if err != nil {
		return 0, err
	}
	uintData, err := strconv.ParseUint(data, 10, 64)
	if err != nil {
		return 0, err
	}
	return uintData, nil
}

func GetTotalPacketsSent() (uint64, error) {
	body, err := requestSoapData(SOAP_URL_WANCommonInterfaceConfig, SOAP_HEADER_VALUE_GetTotalPacketsSent, SOAP_MSG_GetTotalPacketsSent)
	if err != nil {
		return 0, err
	}
	data, err := getDataString(string(*body), SOAP_KEYWORD_GetTotalPacketsSent)
	if err != nil {
		return 0, err
	}
	uintData, err := strconv.ParseUint(data, 10, 64)
	if err != nil {
		return 0, err
	}
	return uintData, nil
}

func GetTotalPacketsReceived() (uint64, error) {
	body, err := requestSoapData(SOAP_URL_WANCommonInterfaceConfig, SOAP_HEADER_VALUE_GetTotalPacketsReceived, SOAP_MSG_GetTotalPacketsReceived)
	if err != nil {
		return 0, err
	}
	data, err := getDataString(string(*body), SOAP_KEYWORD_GetTotalPacketsReceived)
	if err != nil {
		return 0, err
	}
	uintData, err := strconv.ParseUint(data, 10, 64)
	if err != nil {
		return 0, err
	}
	return uintData, nil
}

func GetUptime() (uint64, error) {
	body, err := requestSoapData(SOAP_URL_WANIPConnection, SOAP_HEADER_VALUE_GetUptime, SOAP_MSG_GetUptime)
	if err != nil {
		return 0, err
	}
	data, err := getDataString(string(*body), SOAP_KEYWORD_GetUptime)
	if err != nil {
		return 0, err
	}
	uintData, err := strconv.ParseUint(data[:len(data)-1], 10, 64)
	if err != nil {
		return 0, err
	}
	return uintData, nil
}

func GetExternalIPAddress() (string, error) {
	body, err := requestSoapData(SOAP_URL_WANIPConnection, SOAP_HEADER_VALUE_GetExternalIPAddress, SOAP_MSG_GetExternalIPAddress)
	if err != nil {
		return "", err
	}
	data, err := getDataString(string(*body), SOAP_KEYWORD_GetExternalIPAddress)
	if err != nil {
		return "", err
	}
	return data, nil
}

