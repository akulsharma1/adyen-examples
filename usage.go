package main

import (
	"adyen-examples/adyen"
	"fmt"
)

// an example implementation of our adyen library
func main() {
	// old footlocker adyen RSA key
	adyenKey := "A237060180D24CDEF3E4E27D828BDB6A13E12C6959820770D7F2C1671DD0AEF4729670C20C6C5967C664D18955058B69549FBE8BF3609EF64832D7C033008A818700A9B0458641C5824F5FCBB9FF83D5A83EBDF079E73B81ACA9CA52FDBCAD7CD9D6A337A4511759FA21E34CD166B9BABD512DB7B2293C0FE48B97CAB3DE8F6F1A8E49C08D23A98E986B8A995A8F382220F06338622631435736FA064AEAC5BD223BAF42AF2B66F1FEA34EF3C297F09C10B364B994EA287A5602ACF153D0B4B09A604B987397684D19DBC5E6FE7E4FFE72390D28D6E21CA3391FA3CAADAD80A729FEF4823F6BE9711D4D51BF4DFCB6A3607686B34ACCE18329D415350FD0654D"

	encryption := adyen.NewAdyen(adyenKey)

	encryptedCCNumber,encryptedExpMonth, encryptedExpYear, encryptedCvc, err := encryption.EncryptCreditcardDetails("4242424242424242", "05", "2028", "444")

	if err != nil {
		fmt.Printf("Error encrypting card details: %s\n", err)
		return
	}

	fmt.Printf("encrypted cc details:\ncc number: %s\nencrypted exp month: %s\nencrypted exp year: %s\nencrypted cvc: %s\n", encryptedCCNumber, encryptedExpMonth, encryptedExpYear, encryptedCvc)
}