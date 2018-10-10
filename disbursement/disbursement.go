package disbursement

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/spotlight21c/aesencryptor"
)

var (
	productionURL string = "https://kirimdoku.com/v2/api"
	statingURL    string = "https://staging.doku.com/apikirimdoku"
)

type Client struct {
	agentKey string
	encKey   string
	endpoint string
}

type MessageResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type Country struct {
	Code string `json:"code"`
}

type Currency struct {
	Code string `json:"code"`
}

type Channel struct {
	Code string `json:"code"`
}

type Bank struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	CountryCode string `json:"countryCode"`
}

type Account struct {
	Bank    *Bank  `json:"bank"`
	Number  string `json:"number"`
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
}

type Inquiry struct {
	IDToken string `json:"idToken"`
	Fund    struct {
		Fees struct {
			Total float64 `json:"total"`
		} `json:"fees"`
	} `json:"fund,omitempty"`
}

type InquiryRequest struct {
	SenderCountry       *Country  `json:"senderCountry"`
	SenderCurrency      *Currency `json:"senderCurrency"`
	BeneficiaryCountry  *Country  `json:"beneficiaryCountry"`
	BeneficiaryCurrency *Currency `json:"beneficiaryCurrency"`
	Channel             *Channel  `json:"channel"`
	SenderAmount        float64   `json:"senderAmount"`
	BeneficiaryAccount  *Account  `json:"beneficiaryAccount"`
}

type InquiryResponse struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Inquiry *Inquiry `json:"inquiry"`
}

type Person struct {
	IDToken           string   `json:"idToken"`
	Country           *Country `json:"country"`
	FirstName         string   `json:"firstName"`
	LastName          string   `json:"lastName"`
	PhoneNumber       string   `json:"phoneNumber"`
	BirthDate         string   `json:"birthDate,omitempty"`
	PersonalIDType    string   `json:"personalIdType,omitempty"`
	PersonalID        string   `json:"personalId,omitempty"`
	PersonalIDCountry *Country `json:"personalIdCountry,omitempty"`
}

type RemitRequest struct {
	SenderCountry       *Country  `json:"senderCountry"`
	SenderCurrency      *Currency `json:"senderCurrency"`
	BeneficiaryCountry  *Country  `json:"beneficiaryCountry"`
	BeneficiaryCurrency *Currency `json:"beneficiaryCurrency"`
	Channel             *Channel  `json:"channel"`
	SenderAmount        float64   `json:"senderAmount"`
	Inquiry             *Inquiry  `json:"inquiry"`
	SenderNote          string    `json:"senderNote"`
	Sender              *Person   `json:"sender"`
	Beneficiary         *Person   `json:"beneficiary"`
	BeneficiaryAccount  *Account  `json:"beneficiaryAccount"`
}

type RemitResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Remit   struct {
		TransactionId string `json:"transactionId"`
	} `json:"remit,omitempty"`
}

func New(agentKey string, encKey string, isProduction bool) *Client {
	url := statingURL

	if isProduction {
		url = productionURL
	}

	return &Client{
		agentKey: agentKey,
		encKey:   encKey,
		endpoint: url,
	}
}

// generateSignature it use AES/ECB/PKCS5Padding algorithm
func (c *Client) generateSignature(requestID string) string {
	if encValue, err := aesencryptor.Encrypt(c.agentKey+requestID, c.encKey); err == nil {
		return base64.StdEncoding.EncodeToString(encValue)
	}

	return ""
}

func (c *Client) addCredentialHeader(req *http.Request, requestID string) {
	// fmt.Println(c.endpoint)
	// fmt.Println(c.agentKey)
	// fmt.Println(requestID)
	// fmt.Println(c.generateSignature(requestID))

	// It looks like Doku server case sensitive
	req.Header["agentKey"] = []string{c.agentKey}
	req.Header["requestId"] = []string{requestID}
	req.Header["signature"] = []string{c.generateSignature(requestID)}

	// Do not send header like this
	// req.Header.Add("agentKey", "aaa")
	// req.Header.Add("requestId", "bbb")
	// req.Header.Add("signature", "ccc")
}

func (c *Client) Ping(requestID string) (*MessageResponse, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", c.endpoint+"/ping", nil)

	c.addCredentialHeader(req, requestID)

	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(body))

	if response.StatusCode != http.StatusOK {
		errorResponse := &MessageResponse{}
		if err := json.Unmarshal(body, errorResponse); err != nil {
			return nil, err
		}

		return nil, errors.New(errorResponse.Message)
	}

	messageResponse := &MessageResponse{}

	if err := json.Unmarshal(body, messageResponse); err != nil {
		return nil, err
	}

	return messageResponse, nil
}

func (c *Client) Inquiry(requestID string, amount float64, account *Account) (*InquiryResponse, error) {
	senderCountry := &Country{
		Code: "ID",
	}

	beneficiaryCountry := &Country{
		Code: "ID",
	}

	senderCurrency := &Currency{
		Code: "IDR",
	}

	beneficiaryCurrency := &Currency{
		Code: "IDR",
	}

	channel := &Channel{
		Code: "07",
	}

	payload := &InquiryRequest{
		SenderCountry:       senderCountry,
		SenderCurrency:      senderCurrency,
		BeneficiaryCountry:  beneficiaryCountry,
		BeneficiaryCurrency: beneficiaryCurrency,
		Channel:             channel,
		SenderAmount:        amount,
		BeneficiaryAccount:  account,
	}

	reqBytes, _ := json.Marshal(payload)

	reqBody := bytes.NewBufferString(string(reqBytes))

	client := &http.Client{}
	req, _ := http.NewRequest("POST", c.endpoint+"/cashin/inquiry", reqBody)

	c.addCredentialHeader(req, requestID)

	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(body))

	if response.StatusCode != http.StatusOK {
		errorResponse := &MessageResponse{}
		if err := json.Unmarshal(body, errorResponse); err != nil {
			return nil, err
		}

		return nil, errors.New(errorResponse.Message)
	}

	inquiryResponse := &InquiryResponse{}
	if err := json.Unmarshal(body, inquiryResponse); err != nil {
		return nil, err
	}

	return inquiryResponse, nil
}

func (c *Client) Remit(requestID, token string, amount float64, account *Account, sender *Person, beneficiary *Person, note string) (*RemitResponse, error) {
	senderCountry := &Country{
		Code: "ID",
	}

	beneficiaryCountry := &Country{
		Code: "ID",
	}

	senderCurrency := &Currency{
		Code: "IDR",
	}

	beneficiaryCurrency := &Currency{
		Code: "IDR",
	}

	channel := &Channel{
		Code: "07",
	}

	payload := &RemitRequest{
		SenderCountry:       senderCountry,
		SenderCurrency:      senderCurrency,
		BeneficiaryCountry:  beneficiaryCountry,
		BeneficiaryCurrency: beneficiaryCurrency,
		Channel:             channel,
		SenderAmount:        amount,
		Inquiry: &Inquiry{
			IDToken: token,
		},
		SenderNote:         note,
		Sender:             sender,
		Beneficiary:        beneficiary,
		BeneficiaryAccount: account,
	}

	reqBytes, _ := json.Marshal(payload)

	reqBody := bytes.NewBufferString(string(reqBytes))

	client := &http.Client{}
	req, _ := http.NewRequest("POST", c.endpoint+"/cashin/remit", reqBody)

	c.addCredentialHeader(req, requestID)

	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(body))

	if response.StatusCode != http.StatusOK {
		errorResponse := &MessageResponse{}
		if err := json.Unmarshal(body, errorResponse); err != nil {
			return nil, err
		}

		return nil, errors.New(errorResponse.Message)
	}

	remitResponse := &RemitResponse{}
	if err := json.Unmarshal(body, remitResponse); err != nil {
		return nil, err
	}

	return remitResponse, nil
}
