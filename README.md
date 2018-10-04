# doku-sdk-go

unofficial doku payment solution(disbursement, repayment using VA direct) sdk by golang

# How to use

## Disbursement

```golang
dokuDisbursement "github.com/spotlight21c/doku-sdk-go/disbursement"

isProduction := false

dokuClient := dokuDisbursement.New("DOKU_AGENT_KEY", "DOKU_ENC_KEY", isProduction)

dokuBank := &dokuDisbursement.Bank{ID: "014", Name: "BANK CENTRAL ASIA (BCA)", Code: "CENAIDJA", CountryCode: "ID"},

dokuAccount := &dokuDisbursement.Account{
	Number:  "123456789",
	Name:    "kevin",
	Address: "Jakarta",
	City:    "Jakarta",
	Bank:    dokuBank,
}

amount := 10000.0

if inquiryResponse, err := dokuClient.Inquiry("REQUEST_ID1", amount, dokuAccount); err != nil {
	return nil, err
} else {
	sender := &dokuDisbursement.Person{
		Country: &dokuDisbursement.Country{
			Code: "ID",
		},
		FirstName:      "kovin",
		LastName:       "Lee",
		PhoneNumber:    "1",
		BirthDate:      time.Now().Format("2006-01-02"),
		PersonalIDType: "CITIZENID",
		PersonalID:     "1",
		PersonalIDCountry: &dokuDisbursement.Country{
			Code: "ID",
		},
	}

	beneficiary := &dokuDisbursement.Person{
		Country: &dokuDisbursement.Country{
			Code: "ID",
		},
		FirstName:   "kevin",
		LastName:    "Kim",
		PhoneNumber: "1",
	}

	if remitResponse, err := dokuClient.Remit("REQUEST_ID2", inquiryResponse.Inquiry.IDToken, amount, dokuAccount, sender, beneficiary, "note"); err != nil {
		return nil, err
	} else {
		fmt.Println(remitResponse.Remit.TransactionId)
	}
}
```
