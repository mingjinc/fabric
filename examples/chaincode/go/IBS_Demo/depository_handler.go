/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
        "strings"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// consts associated with chaincode table
const (
	tableColumn       = "AssetsOwnership"
	columnAccountID   = "Account"
	columnContactInfo = "ContactInfo"
	columnAmount      = "Amount"
        columnCurrency    = "Currency"
)

//DepositoryHandler provides APIs used to perform operations on CC's KV store
type depositoryHandler struct {
}

// NewDepositoryHandler create a new reference to CertHandler
func NewDepositoryHandler() *depositoryHandler {
	return &depositoryHandler{}
}

// createTable initiates a new asset depository table in the chaincode state
// stub: chaincodestub
func (t *depositoryHandler) createTable(stub *shim.ChaincodeStub) error {

	// Create asset depository table
	return stub.CreateTable(tableColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: columnAccountID, Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: columnContactInfo, Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: columnAmount, Type: shim.ColumnDefinition_UINT64, Key: false},
		&shim.ColumnDefinition{Name: columnCurrency, Type: shim.ColumnDefinition_STRING, Key: false},
	})

}

// assign allocates assets to account IDs in the chaincode state for each of the
// account ID passed in.
// accountID: account ID to be allocated with requested amount
// contactInfo: contact information of the owner of the account ID passed in
// amount: amount to be allocated to this account ID
func (t *depositoryHandler) assign(stub *shim.ChaincodeStub,
	accountID string,
	contactInfo string,
	amount uint64,
	currency string) error {

	myLogger.Debugf("insert accountID= %v", accountID)

	//insert a new row for this account ID that includes contact information and balance
	ok, err := stub.InsertRow(tableColumn, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: accountID}},
			&shim.Column{Value: &shim.Column_String_{String_: contactInfo}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: amount}},
			&shim.Column{Value: &shim.Column_String_{String_: currency}}},
	})

	// you can only assign balances to new account IDs
	if !ok && err == nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("Asset was already assigned.")
	}

	return nil
}

// updateAccountBalance updates the balance amount of an account ID
// stub: chaincodestub
// accountID: account will be updated with the new balance
// contactInfo: contact information associated with the account owner (chaincode table does not allow me to perform updates on specific columns)
// amount: new amount to be udpated with
func (t *depositoryHandler) updateAccountBalance(stub *shim.ChaincodeStub,
	accountID string,
	contactInfo string,
	amount uint64,
        currency string) error {

	myLogger.Debugf("insert accountID= %v", accountID)

	//replace the old record row associated with the account ID with the new record row
	ok, err := stub.ReplaceRow(tableColumn, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: accountID}},
			&shim.Column{Value: &shim.Column_String_{String_: contactInfo}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: amount}},
			&shim.Column{Value: &shim.Column_String_{String_: currency}}},
	})

	if !ok && err == nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("failed to replace row with account Id." + accountID)
	}
	return nil
}

// deleteAccountRecord deletes the record row associated with an account ID on the chaincode state table
// stub: chaincodestub
// accountID: account ID (record matching this account ID will be deleted after calling this method)
func (t *depositoryHandler) deleteAccountRecord(stub *shim.ChaincodeStub, accountID string) error {

	myLogger.Debugf("insert accountID= %v", accountID)

	//delete record matching account ID passed in
	err := stub.DeleteRow(
		"AssetsOwnership",
		[]shim.Column{shim.Column{Value: &shim.Column_String_{String_: accountID}}},
	)

	if err != nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("error in deleting account record")
	}
	return nil
}

// transfer transfers X amount of assets from "from account IDs" to a new account ID
// stub: chaincodestub
// fromAccounts: from account IDs with assets to be transferred
// toAccount: a new account ID on the table that will get assets transfered to
// toContact: contact information of the owner of "to account ID"
func (t *depositoryHandler) transfer(stub *shim.ChaincodeStub, fromAccount string, toAccount string, toContact string, amount uint64, currency string) error {

        //myLogger.Debugf("insert params= %v , %v , %v , %v , %v", fromAccount, toAccount, toContact, amount, currency)
        
        fromContactInfo, fromAcctBalance, fromCurrency, _ := t.queryAccount(stub, fromAccount)
        fromRole, fromErr := t.getRoleFromAccountID(fromAccount)
        if fromErr != nil {
        	return errors.New("fromErr: from role is not expected")
        }
        
        toContactInfo, toAcctBalance, toCurrency, _ := t.queryAccount(stub, toAccount)
        toRole, toErr:= t.getRoleFromAccountID(toAccount)
        if toErr != nil {
        	return errors.New("toErr: to role is not expected")
        }
              
        if strings.EqualFold(fromRole, "client") {
            if strings.EqualFold(toRole, "getway_level_2") {
                rate, _ := t.getExchangeRate(stub, fromCurrency, currency)
                err := t.updateAccountBalance(stub, fromAccount, fromContactInfo, fromAcctBalance-uint64(float64(amount)*rate), fromCurrency)
                if err != nil {
                    return errors.New("update account " + fromAccount + " (-" + strconv.FormatInt(int64(float64(amount)*rate),10) + fromCurrency + ") fail")
                }
                err = t.updateAccountBalance(stub, toAccount, toContactInfo, toAcctBalance-uint64(float64(amount)*rate), toCurrency)
                if err != nil {
                    return errors.New("update account " + toAccount + " (-" + strconv.FormatInt(int64(float64(amount)*rate),10) + toCurrency + ") fail")
                }
            }
        }
        
        if strings.EqualFold(fromRole, "getway_level_2") {
            if strings.EqualFold(toRole, "getway_level_1") {
                rate, _ := t.getExchangeRate(stub, fromCurrency, currency)
                err := t.updateAccountBalance(stub, toAccount, toContactInfo, toAcctBalance+uint64(float64(amount)*rate), toCurrency)
                if err != nil {
                    return errors.New("update account " + toAccount + " (+" + strconv.FormatInt(int64(float64(amount)*rate),10) + toCurrency + ") fail")
                }
            }
        }
        
        if strings.EqualFold(fromRole, "getway_level_1") {
            if strings.EqualFold(toRole, "getway_level_1") {
               err := t.updateAccountBalance(stub, fromAccount, fromContactInfo, fromAcctBalance-amount, fromCurrency)
               if err != nil {
                   return errors.New("update account " + fromAccount + " (-" + strconv.FormatInt(int64(amount),10) + fromCurrency + ") fail")
               }
               err = t.updateAccountBalance(stub, toAccount, toContactInfo, toAcctBalance+amount, toCurrency)
               if err != nil {
                   return errors.New("update account " + toAccount + " (+" + strconv.FormatInt(int64(amount),10) + toCurrency + ") fail")
               }
            }
        }
        
        if strings.EqualFold(fromRole, "getway_level_1") {
            if strings.EqualFold(toRole, "getway_level_2") {
               err := t.updateAccountBalance(stub, fromAccount, fromContactInfo, fromAcctBalance-amount, fromCurrency)
               if err != nil {
                   return errors.New("update account " + fromAccount + " (-" + strconv.FormatInt(int64(amount),10) + fromCurrency + ") fail")
               }
               err = t.updateAccountBalance(stub, toAccount, toContactInfo, toAcctBalance+amount, toCurrency)
               if err != nil {
                   return errors.New("update account " + toAccount + " (+" + strconv.FormatInt(int64(amount),10) + toCurrency + ") fail")
               }
            }
        }
        
        if strings.EqualFold(fromRole, "getway_level_2") {
            if strings.EqualFold(toRole, "client") {
               err := t.updateAccountBalance(stub, toAccount, toContactInfo, toAcctBalance+amount, toCurrency)
               if err != nil {
                   return errors.New("update account " + toAccount + " (+" + strconv.FormatInt(int64(amount),10) + toCurrency + ") fail")
               }
            }
        }
        
        return nil

}

//validate fromAccount , fromRole , toAccount , toRole , toCurrency , fromAcctBalance
func (t *depositoryHandler) validateRoleAndAccount(stub *shim.ChaincodeStub, fromAccount string, toAccount string, amount uint64, currency string) (error) {
      
      //myLogger.Debugf("insert params= %v , %v , %v , %v , %v , %v", fromAccount, toAccount, amount, currency, time, fee)
      
      //1 check fromAccount, role and currency
      _, fromAcctBalance, fromCurrency, err := t.queryAccount(stub, fromAccount)
      if err != nil {
          return errors.New("The fromAccount is not existing")
      }
        
      //2 check toAccount, role and currency
      _, _, toCurrency, err := t.queryAccount(stub, toAccount)
      if err != nil {
          return errors.New("The toAccount is not existing")
      }
      if !strings.EqualFold(toCurrency, currency) {
          return errors.New("The toCurrency is not " + currency)
      }
      
      //3 check fromAcctBalance
      rate, err := t.getExchangeRate(stub, fromCurrency, toCurrency)
      if err != nil {
      	return err
      }
      var amountTemp float64  = float64(amount)
     // for i := 0 ; i < time ; i++{
     //   amountTemp /= (1-fee)
     // } 
      remaining := uint64(amountTemp * rate)
      if fromAcctBalance < remaining {
          return errors.New("The fromAcctBalance is not enough for transfer")
      }
      
      return nil
}

//get exchange rate according to the fromCurrency and toCurrency
func (t *depositoryHandler) getExchangeRate(stub *shim.ChaincodeStub, fromCurrency string, toCurrency string) (float64, error) {
	if strings.EqualFold(fromCurrency, "KZT") && strings.EqualFold(toCurrency, "RMB"){
    	return float64(30) , nil
    }
    return float64(1) , errors.New("The exchange rate is not support");
}

func (t *depositoryHandler) getRoleFromAccountID(accountID string) (string, error) {

     if strings.Contains(accountID, "00001") || strings.Contains(accountID, "00002") {
         return "getway_level_1", nil
     }
     if strings.Contains(accountID, "00003") {
         return "getway_level_2", nil
     }
     if strings.Contains(accountID, "00004") {
         return "client", nil
     }
     
     return "", errors.New("the role is not expected")
}
// queryContactInfo queries the contact information matching a correponding account ID on the chaincode state table
// stub: chaincodestub
// accountID: account ID
func (t *depositoryHandler) queryContactInfo(stub *shim.ChaincodeStub, accountID string) (string, error) {
	row, err := t.queryTable(stub, accountID)
	if err != nil {
		return "", err
	}

	return row.Columns[1].GetString_(), nil
}

// queryBalance queries the balance information matching a correponding account ID on the chaincode state table
// stub: chaincodestub
// accountID: account ID
func (t *depositoryHandler) queryBalance(stub *shim.ChaincodeStub, accountID string) (uint64, string, error) {

	myLogger.Debugf("insert accountID= %v", accountID)

	row, err := t.queryTable(stub, accountID)
	if err != nil {
		return 0, "", err
	}
	if len(row.Columns) == 0 || row.Columns[2] == nil || row.Columns[3] == nil{
		return 0, "", errors.New("row or column value not found")
	}

	return row.Columns[2].GetUint64(), row.Columns[3].GetString_(), nil
}

// queryAccount queries the balance and contact information matching a correponding account ID on the chaincode state table
// stub: chaincodestub
// accountID: account ID
func (t *depositoryHandler) queryAccount(stub *shim.ChaincodeStub, accountID string) (string, uint64, string, error) {
	row, err := t.queryTable(stub, accountID)
	if err != nil {
		return "", 0, "", err
	}
	if len(row.Columns) == 0 || row.Columns[2] == nil || row.Columns[3] == nil{
		return "", 0, "",  errors.New("row or column value not found")
	}

	return row.Columns[1].GetString_(), row.Columns[2].GetUint64(), row.Columns[3].GetString_(), nil
}

// queryTable returns the record row matching a correponding account ID on the chaincode state table
// stub: chaincodestub
// accountID: account ID
func (t *depositoryHandler) queryTable(stub *shim.ChaincodeStub, accountID string) (shim.Row, error) {

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: accountID}}
	columns = append(columns, col1)

	return stub.GetRow(tableColumn, columns)
}
