package checkout

import (
	"Kaizen/adyen"
	"Kaizen/common"
	"Kaizen/logs"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	//"os"
	"strconv"
	"time"
)

/*
	An implementation of the adyen encryption library for Footlocker in Kaizen.
	Note that this code won't compile without the other internal libraries.
*/

type CheckoutResp struct {
	Order struct {
		AppliedCoupons           []interface{} `json:"appliedCoupons"`
		AppliedOrderPromotions   []interface{} `json:"appliedOrderPromotions"`
		AppliedProductPromotions []interface{} `json:"appliedProductPromotions"`
		AppliedVouchers          []interface{} `json:"appliedVouchers"`
		Calculated               bool          `json:"calculated"`
		Code                     string        `json:"code"`
	} `json:"order"`
}

func SubmitPayment(client *http.Client, csrfToken string, xFLRequestId string, cartGUIDCookie *http.Cookie, xFlAPISessionID string, profile common.ProfileDetails, taskid string) (int, string) {

	adyenKey := "A237060180D24CDEF3E4E27D828BDB6A13E12C6959820770D7F2C1671DD0AEF4729670C20C6C5967C664D18955058B69549FBE8BF3609EF64832D7C033008A818700A9B0458641C5824F5FCBB9FF83D5A83EBDF079E73B81ACA9CA52FDBCAD7CD9D6A337A4511759FA21E34CD166B9BABD512DB7B2293C0FE48B97CAB3DE8F6F1A8E49C08D23A98E986B8A995A8F382220F06338622631435736FA064AEAC5BD223BAF42AF2B66F1FEA34EF3C297F09C10B364B994EA287A5602ACF153D0B4B09A604B987397684D19DBC5E6FE7E4FFE72390D28D6E21CA3391FA3CAADAD80A729FEF4823F6BE9711D4D51BF4DFCB6A3607686B34ACCE18329D415350FD0654D"

	encryption := adyen.NewAdyen(adyenKey)

	encryptedCCNumber,encryptedExpMonth, encryptedExpYear, encryptedCvc, err := encryption.EncryptCreditcardDetails(profile.CardNumber, profile.ExpiryMonth, profile.ExpiryYear, profile.CVV)

	if err != nil {
		logs.RedPrintln("Error encrypting payment", taskid, "")
		return 400, ""
	}

	submitPaymentURL := profile.StoreURL + "/api/v2/users/orders?timestamp="  + strconv.FormatInt(time.Now().UnixNano() / int64(time.Millisecond), 10)

	browserInfo := map[string]string{
		"colorDepth":"24",
		"javaEnabled": "false",
		"language": "en-US",
		"screenHeight": "1440",
		"screenWidth": "3440",
		"timeZoneOffset": "480",
	}
	submitPayment := map[string]interface{} {
		"browserInfo": browserInfo,
		"cartId": cartGUIDCookie.Value,
		"encryptedCardNumber": encryptedCCNumber,
		"encryptedExpiryMonth": encryptedExpMonth,
		"encryptedExpiryYear": encryptedExpYear,
		"encryptedSecurityCode": encryptedCvc,
		"paymentMethod": "CREDITCARD",
		"preferredLanguage": "en",
		"returnUrl": profile.StoreURL + "/adyen/checkout",
		"termsAndCondition": "false",
	}

	submitPaymentJsonValue, _ := json.Marshal(submitPayment)

	submitPaymentRequest, err := http.NewRequest("POST", submitPaymentURL, bytes.NewBuffer(submitPaymentJsonValue))

	if err != nil {
		logs.RedPrintln("Error initializing payment", taskid, "")
		return 400, ""
	}

	submitPaymentRequest.Header.Set("accept",  common.AcceptJson)
	submitPaymentRequest.Header.Set("accept-language", common.AcceptLanguage)
	submitPaymentRequest.Header.Set("content-type",  common.AcceptJson)
	submitPaymentRequest.Header.Set("origin", profile.StoreURL)
	submitPaymentRequest.Header.Set("referer", profile.StoreURL + "/checkout")
	submitPaymentRequest.Header.Set("user-agent", common.UserAgent)
	submitPaymentRequest.Header.Set("x-csrf-token", csrfToken)
	submitPaymentRequest.Header.Set("x-fl-request-id", xFLRequestId)
	submitPaymentRequest.Header.Set("x-flapi-session-id", xFlAPISessionID)

	submitPaymentRequest.AddCookie(cartGUIDCookie)
	submitPaymentResponse, err := client.Do(submitPaymentRequest)

	if err != nil {
		logs.RedPrintln("Error initializing submit email", taskid, "")
		return 400, ""
	}

	defer submitPaymentResponse.Body.Close()

	switch submitPaymentResponse.StatusCode {
	case 200:
		logs.GreenPrintln("(200) Checkout success", taskid, "")
		pagejson, _ := ioutil.ReadAll(submitPaymentResponse.Body)
		unmarshaledCheckoutResponse := CheckoutResp{}
		json.Unmarshal(pagejson, &unmarshaledCheckoutResponse)
		orderNum := unmarshaledCheckoutResponse.Order.Code
		if orderNum == "" {
			orderNum = "null"
		}
		return 200, orderNum
	case 201:
		logs.GreenPrintln("(201) Checkout success", taskid, "")
		pagejson, _ := ioutil.ReadAll(submitPaymentResponse.Body)
		unmarshaledCheckoutResponse := CheckoutResp{}
		json.Unmarshal(pagejson, &unmarshaledCheckoutResponse)
		orderNum := unmarshaledCheckoutResponse.Order.Code
		if orderNum == "" {
			orderNum = "null"
		}
		return 201, orderNum
	case 400:
		logs.RedPrintln("(400) Checkout failed, payment declined", taskid, "")
	case 403:
		logs.YellowPrintln(fmt.Sprintf("(%v) Error submitting payment", submitPaymentResponse.StatusCode), taskid, "")
	}

	return submitPaymentResponse.StatusCode, ""
}
