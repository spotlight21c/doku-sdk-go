package repayment

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var (
	productionURL string = "https://pay.doku.com"
	statingURL    string = "https://staging.doku.com"
)

type Client struct {
	mallID    string
	sharedKey string
	endpoint  string
}

type InquiryRequest struct {
	MallID         string
	ChainMerchant  string
	PaymentChannel string
	PaymentCode    string
}

type InquiryResponse struct {
	XMLName          xml.Name `xml:"INQUIRY_RESPONSE"`
	PaymentCode      string   `xml:"PAYMENTCODE"`
	Amount           string   `xml:"AMOUNT"`
	PurchaseAmount   string   `xml:"PURCHASEAMOUNT"`
	MinAmount        string   `xml:"MINAMOUNT"`
	MaxAmount        string   `xml:"MAXAMOUNT"`
	TransIDMerchant  string   `xml:"TRANSIDMERCHANT"`
	Words            string   `xml:"WORDS"`
	RequestDatetime  string   `xml:"REQUESTDATETIME"`
	Currency         string   `xml:"CURRENCY"`
	PurchaseCurrency string   `xml:"PURCHASECURRENCY"`
	SessionID        string   `xml:"SESSIONID"`
	Email            string   `xml:"EMAIL"`
	Basket           string   `xml:"BASKET"`
	AdditionalData   string   `xml:"ADDITIONALDATA"`
	Name             string   `xml:"NAME"`
	ResponseCode     string   `xml:"RESPONSECODE"`
}

type NotifyRequest struct {
	Amount             string
	TransIDMerchant    string
	StatusType         string
	ResponseCode       string
	ApprovalCode       string
	ResultMsg          string
	PaymentChannel     string
	PaymentCode        string
	SessionID          string
	Bank               string
	MCN                string
	PaymentDatetime    string
	VerifyID           string
	VerifyScore        string
	VerifyStatus       string
	Currency           string
	PurchaseCurrency   string
	Brand              string
	Chname             string
	ThreedSecureStatus string
	Liability          string
	EduStatus          string
	CustomerID         string
	TokenID            string
}

type CheckStatusResponse struct {
	XMLName            xml.Name `xml:"PAYMENT_STATUS"`
	Amount             string   `xml:"AMOUNT"`
	TransIDMerchant    string   `xml:"TRANSIDMERCHANT"`
	Words              string   `xml:"WORDS"`
	ResponseCode       string   `xml:"RESPONSECODE"`
	ApprovalCode       string   `xml:"APPROVALCODE"`
	ResultMsg          string   `xml:"RESULTMSG"`
	PaymentChannel     string   `xml:"PAYMENTCHANNEL"`
	PaymentCode        string   `xml:"PAYMENTCODE"`
	SessionID          string   `xml:"SESSIONID"`
	Bank               string   `xml:"BANK"`
	Mcn                string   `xml:"MCN"`
	PaymentDatetime    string   `xml:"PAYMENTDATETIME"`
	VerifyID           string   `xml:"VERIFYID"`
	VerifyScore        string   `xml:"VERIFYSCORE"`
	VerifyStatus       string   `xml:"VERIFYSTATUS"`
	Currency           string   `xml:"CURRENCY"`
	PurchaseCurrency   string   `xml:"PURCHASECURRENCY"`
	Brand              string   `xml:"BRAND"`
	Chname             string   `xml:"CHNAME"`
	ThreedSecureStatus string   `xml:"THREEDSECURESTATUS"`
	Liability          string   `xml:"LIABILITY"`
	EduStatus          string   `xml:"EDUSTATUS"`
}

func New(mallID string, sharedKey string, isProduction bool) *Client {
	url := statingURL

	if isProduction {
		url = productionURL
	}

	return &Client{
		mallID:    mallID,
		sharedKey: sharedKey,
		endpoint:  url,
	}
}

func (c *Client) ParseInquiryRequest(r *http.Request) (*InquiryRequest, error) {
	mallID := r.FormValue("MALLID")
	chainMerchant := r.FormValue("CHAINMERCHANT")
	paymentChannel := r.FormValue("PAYMENTCHANNEL")
	paymentCode := r.FormValue("PAYMENTCODE")
	words := r.FormValue("WORDS")

	if paymentCode == "" {
		return nil, errors.New("no PAYMENTCODE")
	}

	if words == "" {
		return nil, errors.New("no WORDS")
	}

	if words != c.MakeWordsForInquiry(paymentCode) {
		return nil, errors.New("invalid request")
	}

	inquiryRequest := &InquiryRequest{
		MallID:         mallID,
		ChainMerchant:  chainMerchant,
		PaymentChannel: paymentChannel,
		PaymentCode:    paymentCode,
	}

	return inquiryRequest, nil
}

func (c *Client) ParseNotifyRequest(r *http.Request) (*NotifyRequest, error) {
	amount := r.FormValue("AMOUNT")
	transIDMerchant := r.FormValue("TRANSIDMERCHANT")
	words := r.FormValue("WORDS")
	statusType := r.FormValue("STATUSTYPE")
	responseCode := r.FormValue("RESPONSECODE")
	approvalCode := r.FormValue("APPROVALCODE")
	resultMsg := r.FormValue("RESULTMSG")
	paymentChannel := r.FormValue("PAYMENTCHANNEL")
	paymentCode := r.FormValue("PAYMENTCODE")
	sessionID := r.FormValue("SESSIONID")
	bank := r.FormValue("BANK")
	mcn := r.FormValue("MCN")
	paymentDatetime := r.FormValue("PAYMENTDATETIME")
	verifyID := r.FormValue("VERIFYID")
	verifyScore := r.FormValue("VERIFYSCORE")
	verifyStatus := r.FormValue("VERIFYSTATUS")
	currency := r.FormValue("CURRENCY")
	purchaseCurrency := r.FormValue("PURCHASECURRENCY")
	brand := r.FormValue("BRAND")
	chname := r.FormValue("CHNAME")
	threedSecureStatus := r.FormValue("THREEDSECURESTATUS")
	liability := r.FormValue("LIABILITY")
	eduStatus := r.FormValue("EDUSTATUS")
	customerID := r.FormValue("CUSTOMERID")
	tokenID := r.FormValue("TOKENID")

	if amount == "" {
		return nil, errors.New("no AMOUNT")
	}

	if transIDMerchant == "" {
		return nil, errors.New("no TRANSIDMERCHANT")
	}

	if resultMsg == "" {
		return nil, errors.New("no RESULTMSG")
	}

	if verifyStatus == "" {
		return nil, errors.New("no VERIFYSTATUS")
	}

	if words == "" {
		return nil, errors.New("no WORDS")
	}

	if paymentCode == "" {
		return nil, errors.New("no PAYMENTCODE")
	}

	if words != c.MakeWordsForNotify(amount, transIDMerchant, resultMsg, verifyStatus) {
		return nil, errors.New("invalid request")
	}

	notifyRequest := &NotifyRequest{
		Amount:             amount,
		TransIDMerchant:    transIDMerchant,
		StatusType:         statusType,
		ResponseCode:       responseCode,
		ApprovalCode:       approvalCode,
		ResultMsg:          resultMsg,
		PaymentChannel:     paymentChannel,
		PaymentCode:        paymentCode,
		SessionID:          sessionID,
		Bank:               bank,
		MCN:                mcn,
		PaymentDatetime:    paymentDatetime,
		VerifyID:           verifyID,
		VerifyScore:        verifyScore,
		VerifyStatus:       verifyStatus,
		Currency:           currency,
		PurchaseCurrency:   purchaseCurrency,
		Brand:              brand,
		Chname:             chname,
		ThreedSecureStatus: threedSecureStatus,
		Liability:          liability,
		EduStatus:          eduStatus,
		CustomerID:         customerID,
		TokenID:            tokenID,
	}

	return notifyRequest, nil
}

func (c *Client) CheckStatus(transIDMerchant string, sessionID string) (*CheckStatusResponse, error) {
	data := url.Values{}
	data.Set("MALLID", c.mallID)
	data.Add("CHAINMERCHANT", "NA")
	data.Add("TRANSIDMERCHANT", transIDMerchant)
	data.Add("SESSIONID", sessionID)
	data.Add("WORDS", c.MakeWordsForCheckStatus(transIDMerchant))

	client := &http.Client{}
	req, _ := http.NewRequest("POST", c.endpoint+"/Suite/CheckStatus", strings.NewReader(data.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(body))

	if response.StatusCode != http.StatusOK {
		errorResponse := &CheckStatusResponse{}
		if err := json.Unmarshal(body, errorResponse); err != nil {
			return nil, err
		}

		return nil, errors.New(errorResponse.ResultMsg)
	}

	checkStatusResponse := &CheckStatusResponse{}
	if err := xml.Unmarshal(body, checkStatusResponse); err != nil {
		return nil, err
	}

	if checkStatusResponse.ResponseCode != "0000" {
		return nil, errors.New(checkStatusResponse.ResultMsg)
	}

	return checkStatusResponse, nil
}

func (c *Client) MakeWords(amount float64, transIDMerchant string) string {
	value := fmt.Sprintf("%.2f%s%s%s", amount, c.mallID, c.sharedKey, transIDMerchant)

	h := sha1.New()
	io.WriteString(h, value)
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) MakeWordsForInquiry(paymentCode string) string {
	value := fmt.Sprintf("%s%s%s", c.mallID, c.sharedKey, paymentCode)

	h := sha1.New()
	io.WriteString(h, value)
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) MakeWordsForNotify(amount, transIDMerchant, resultMsg, verifyStatus string) string {
	value := fmt.Sprintf("%s%s%s%s%s%s", amount, c.mallID, c.sharedKey, transIDMerchant, resultMsg, verifyStatus)

	h := sha1.New()
	io.WriteString(h, value)
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) MakeWordsForCheckStatus(transIDMerchant string) string {
	value := fmt.Sprintf("%s%s%s", c.mallID, c.sharedKey, transIDMerchant)

	h := sha1.New()
	io.WriteString(h, value)
	return hex.EncodeToString(h.Sum(nil))
}
