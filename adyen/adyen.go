package adyen

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"strings"
	"time"
)

type Adyen struct {
	rsa     *adrsa
	aesKey     []byte
	aesNonce   []byte
}

type Data struct {
	Number              string `json:"number"`
	Cvc                 string `json:"cvc"`
	ExpiryMonth         string `json:"expiryMonth"`
	ExpiryYear          string `json:"expiryYear"`
	Generationtime      string `json:"generationtime"`
}


type ExpiryYear struct {
	ExpiryYear          string `json:"expiryYear"`
	Generationtime      string `json:"generationtime"`
}

type ExpiryMonth struct {
	ExpiryMonth         string `json:"expiryMonth"`
	Generationtime      string `json:"generationtime"`
}
type CVC struct {
	CVC         string `json:"cvc"`
	Generationtime      string `json:"generationtime"`
}
type CreditCardNumber struct {
	Number              string `json:"number"`
	Generationtime      string `json:"generationtime"`
}

/*
Creates an instance of adyen encryption.
Inputs:
- publicKey: the public adyen RSA key for the website
*/
func NewAdyen(publicKey string) *Adyen {
	y := &Adyen{}
	y.rsa = NewRsa()
	y.aesKey = make([]byte, 32)

	err := y.rsa.Init(publicKey, 65537)
	if err != nil {
		panic(err)
	}
	return y
}



func (y *Adyen) random(len int) []byte {
	ak := make([]byte, len)
	rand.Read(ak)
	return ak
}
func (y *Adyen) EncryptCreditcardDetails(CCNumber string,ExpMonth string, ExpYear string, Cvc string) (EncryptedCCNumber string, EncryptedExpMonth string, EncryptedExpYear string, EncryptedCvc string, err error){
	EncryptedCCNumber,err = y.EncryptCC(CCNumber,"","","")
	if err != nil {
		return "","","","",err
	}
	EncryptedExpMonth,err = y.EncryptCC("",ExpMonth,"","")
	if err != nil {
		return "","","","",err
	}
	EncryptedExpYear,err = y.EncryptCC("","",ExpYear,"")
	if err != nil {
		return "","","","",err
	}
	EncryptedCvc,err = y.EncryptCC("","","",Cvc)
	if err != nil {
		return "","","","",err
	}
	return
}
func (y *Adyen) EncryptCC(CCnumber string, ExpMonth string, ExpYear string, Cvc string) (string, error) {
	y.aesKey = y.random(32)
	y.aesNonce = y.random(12)
	gt := time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")
	bytes,_ := json.Marshal(Data{})

	if CCnumber != ""{
		info := CreditCardNumber{
			Number:          CCnumber,
			Generationtime: gt,
		}
		bytes,_ = json.Marshal(info)
	}
	if ExpMonth != "" {
		info := ExpiryMonth{
			ExpiryMonth:          ExpMonth,
			Generationtime: gt,
		}
		bytes,_ = json.Marshal(info)
	}
	if ExpYear != ""{
		info := ExpiryYear{
			ExpiryYear:          ExpYear,
			Generationtime: gt,
		}
		bytes,_ = json.Marshal(info)
	}
	if Cvc != "" {
		info := CVC{
			CVC:          Cvc,
			Generationtime: gt,
		}
		bytes,_ = json.Marshal(info)
	}

	y.aesKey = y.random(32)
	y.aesNonce = y.random(12)
	block, err := aes.NewCipher(y.aesKey)
	if err != nil {
		return "", err
	}
	cmer, err := NewCCM(block, 8, len(y.aesNonce))
	if err != nil {
		return "", err
	}

	cipherBytes := cmer.Seal(nil, y.aesNonce, bytes, nil)
	cipherBytes = append(y.aesNonce, cipherBytes...)
	cipherText := base64.StdEncoding.EncodeToString(cipherBytes)

	encryptedPublicKey, err := y.rsa.encryptWithAesKey(y.aesKey)
	if err != nil {
		return "", err
	}
	arr := []string{v18, encryptedPublicKey, "$", cipherText} // TODO: add support for different adyen versions
	return strings.Join(arr, ""), nil
}