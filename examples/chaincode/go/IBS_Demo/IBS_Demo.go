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
	//"encoding/base64"
	//"encoding/binary"
	"errors"
	"strconv"
	//"strings"
        
	"github.com/hyperledger/fabric/core/crypto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/op/go-logging"
)

var myLogger = logging.MustGetLogger("asset_mgm")

var cHandler = NewCertHandler()
var dHandler = NewDepositoryHandler()

var (
	administrator crypto.Client
	alice         crypto.Client
	bob           crypto.Client
)

//AssetManagementChaincode APIs exposed to chaincode callers
type AssetManagementChaincode struct {
}

// assignOwnership assigns assets to a given account ID, only entities with the "issuer" are allowed to call this function
// Note: this issuer can only allocate balance to one account ID at a time
// args[0]: investor's TCert
// args[1]: attribute name inside the investor's TCert that contains investor's account ID
// args[2]: amount to be assigned to this investor's account ID
// args[3]: currency
func (t *AssetManagementChaincode) assignOwnership(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++assignOwnership+++++++++++++++++++++++++++++++++")

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}
	
	//check is invoker has the correct role, only invokers with the "issuer" role is allowed to
	//assign asset to owner
	isAuthorized, err := cHandler.isAuthorizedByRole(stub, "getway_level_2")
	if !isAuthorized {
		myLogger.Errorf("system error %v", err)
		return nil, errors.New("user is not aurthorized to assign assets")
	}

	amount, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		myLogger.Errorf("system error %v", err)
		return nil, errors.New("Unable to parse amount" + args[2])
	}

	currency := args[3]
	if currency != "RMB" &&  currency != "KZT" {
		return nil, errors.New("Not valid currency" + args[3])
	}

	//call DeposistoryHandler.assign function to put the "amount" and "contact info" under this account ID
    	accountID := ""
	if args[0] == "alice" || args[0] == "bob"{
    		if args[1] == "11111-00004" {
			isAuthorized, err = cHandler.isAuthorizedByName(stub, "bank_a_getway_level_2")
                        myLogger.Errorf("by Name")
			if !isAuthorized {
                        	myLogger.Errorf("system error %v", err)
                        	return nil, errors.New("user is not aurthorized to assign assets to "+ args[0])
                        } 
    	    	        accountID = args[1]
    			if args[3] != "KZT" {
    		   		return nil, errors.New("currency "+ args[3] + " is not supported on alice's account")
    			}
    		} else if args[1] == "22222-00004" {
                        isAuthorized, err = cHandler.isAuthorizedByName(stub, "bank_b_getway_level_2")
                        if !isAuthorized {
                                myLogger.Errorf("system error %v", err)
                                return nil, errors.New("user is not aurthorized to assign assets to "+ args[0])
                        }
                        accountID = args[1]
                        if args[3] != "RMB" {
                                return nil, errors.New("currency "+ args[3] + " is not supported on bob's account")
                        }  
                } else {
    	        	return nil, errors.New("AccountId is not " + args[0] + "'s " + args[1])
		}
    	} else {
    		return nil, errors.New("Not valid client " + args[0])
    	}
    
    	contactInfo := "alice@gmail.com" 
	dHandler.assign(stub, accountID, contactInfo, amount, currency)
    	ret := "Successfully assgin "+ args[2] + args[3] + " to " + args[0] + "'s "+ args[1]+ "account" 
	//call DeposistoryHandler.assign function to put the "amount" and "contact info" under this account ID
	return []byte(ret), nil
	//return nil, dHandler.assign(stub, accountID, contactInfo, amount, currency)
}

// transferOwnership moves x number of assets from account A to account B
// args[0]: Investor TCert that has account IDs which will their balances deducted
// args[1]: attribute names inside TCert (arg[0]) that countain the account IDs
// args[2]: Investor TCert that has account IDs which will have their balances increased
// args[3]: attribute names inside TCert (arg[2]) that countain the account IDs
// args[4]: amount to be assigned to this investor's account ID
// args[5]: currency
func (t *AssetManagementChaincode) transferOwnership(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++transferOwnership+++++++++++++++++++++++++++++++++")

	if len(args) != 6 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}
       
        //fromOwner := args[0]
        isAuthorized, err := cHandler.isAuthorizedByName(stub, "alice")
        if !isAuthorized {
        	myLogger.Errorf("system error %v", err)
                return nil, errors.New("user is not aurthorized to transfer assets, user must be alice in this demo")
        }
        fromAccountId := args[1]

	//toOwner := args[2]
        toAccountId := args[3]

	amount, err := strconv.ParseUint(args[4], 10, 64)
	if err != nil {
		myLogger.Errorf("system error %v", err)
		return nil, errors.New("Unable to parse amount" + args[4])
	}

	currency := args[5]
	if currency != "RMB" &&  currency != "KZT" {
		return nil, errors.New("Not valid currency" + args[5])
	}

        validateError := dHandler.validateRoleAndAccount(stub, fromAccountId, toAccountId, amount, currency)
	if validateError != nil {
		return nil, errors.New("Not valiad Role or AccountId")
	}	

	contactInfo := "bob@gmail.com"

        transferError := dHandler.transfer(stub, "11111-00004", "11111-00003", contactInfo, amount, currency)
	transferError = dHandler.transfer(stub, "11111-00003", "11111-00001", contactInfo, amount, currency)
	transferError = dHandler.transfer(stub, "11111-00002", "22222-00002", contactInfo, amount, currency)
	transferError = dHandler.transfer(stub, "22222-00002", "22222-00003", contactInfo, amount, currency)
	transferError = dHandler.transfer(stub, "22222-00003", "22222-00004", contactInfo, amount, currency)
        // call dHandler.transfer to transfer to transfer "amount" from "from account" IDs to "to account" IDs
	return nil, transferError
}

// getOwnerContactInformation retrieves the contact information of the investor that owns a particular account ID
// Note: user contact information shall be encrypted with issuer's pub key or KA key
// between investor and issuer, so that only issuer can decrypt such information
// args[0]: one of the many account IDs owned by "some" investor
func (t *AssetManagementChaincode) getOwnerContactInformation(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++getOwnerContactInformation+++++++++++++++++++++++++++++++++")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	accountID := args[0]

	email, err := dHandler.queryContactInfo(stub, accountID)
	if err != nil {
		return nil, err
	}

	return []byte(email), nil
}

// getBalance retrieves the account balance information of the investor that owns a particular account ID
// args[0]: one of the many account IDs owned by "some" investor
func (t *AssetManagementChaincode) getBalance(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++getBalance+++++++++++++++++++++++++++++++++")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	accountID := args[0]

	isAuthorized := true
        err := errors.New("")

        if accountID == "11111-00001" {
		isAuthorized, err = cHandler.isAuthorizedByName(stub, "bank_a_getway_level_1")
		if !isAuthorized {
			myLogger.Errorf("system error %v", err)
			return nil, errors.New("user is not aurthorized to query assets in " + args[0])
		}
	} else if accountID == "11111-00002" {
		isAuthorized, err = cHandler.isAuthorizedByName(stub, "bank_a_getway_level_1")
		if !isAuthorized {
			myLogger.Errorf("system error %v", err)
			return nil, errors.New("user is not aurthorized to query assets in " + args[0])
		}
	} else if accountID == "11111-00003" {
		isAuthorized, err = cHandler.isAuthorizedByName(stub, "bank_a_getway_level_2")
		if !isAuthorized {
			myLogger.Errorf("system error %v", err)
			return nil, errors.New("user is not aurthorized to query assets in " + args[0])
		}
	} else if accountID == "11111-00004" {
		isAuthorized, err = cHandler.isAuthorizedByName(stub, "bank_a_getway_level_2")
		if !isAuthorized {
			isAuthorized, err = cHandler.isAuthorizedByName(stub, "alice")
			if !isAuthorized {
				myLogger.Errorf("system error %v", err)
				return nil, errors.New("user is not aurthorized to query assets in " + args[0])
			}
		}
	} else if accountID == "22222-00001" {
		isAuthorized, err = cHandler.isAuthorizedByName(stub, "bank_b_getway_level_1")
		if !isAuthorized {
			myLogger.Errorf("system error %v", err)
			return nil, errors.New("user is not aurthorized to query assets in " + args[0])
		}
	} else if accountID == "22222-00002" {
		isAuthorized, err = cHandler.isAuthorizedByName(stub, "bank_b_getway_level_1")
		if !isAuthorized {
			myLogger.Errorf("system error %v", err)
			return nil, errors.New("user is not aurthorized to query assets in " + args[0])
		}
	} else if accountID == "22222-00003" {
		isAuthorized, err = cHandler.isAuthorizedByName(stub, "bank_b_getway_level_2")
		if !isAuthorized {
			myLogger.Errorf("system error %v", err)
			return nil, errors.New("user is not aurthorized to query assets in " + args[0])
		}
	} else if accountID == "22222-00004" {
		isAuthorized, err = cHandler.isAuthorizedByName(stub, "bank_b_getway_level_2")
		if !isAuthorized {
			isAuthorized, err = cHandler.isAuthorizedByName(stub, "bob")
			if !isAuthorized {
				myLogger.Errorf("system error %v", err)
				return nil, errors.New("user is not aurthorized to query assets in " + args[0])
			}
		}
	}

	balance, currency,  err := dHandler.queryBalance(stub, accountID)
	if err != nil {
		return nil, err
	}

        myLogger.Debugf("balance=%d, currency = %s", balance, currency)
	//convert balance (uint64) to []byte (Big Endian)
	//ret := make([]byte, 8)
	//binary.BigEndian.PutUint64(ret, balance)

	ret := []byte(strconv.Itoa(int(balance))+ currency)
	return ret, nil
}

// Init initialization, this method will create asset despository in the chaincode state
func (t *AssetManagementChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	myLogger.Debugf("********************************Init****************************************")

	myLogger.Info("[AssetManagementChaincode] Init")
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}
	dHandler.createTable(stub)
	
	dHandler.assign(stub, "11111-00001", "bank_a_getway_level_1@gmail.com", 1000000, "KZT")
	dHandler.assign(stub, "11111-00002", "bank_a_getway_level_1@gmail.com", 100000, "RMB")
	dHandler.assign(stub, "11111-00003", "bank_a_getway_level_2@gmail.com", 100000, "KZT")
	dHandler.assign(stub, "22222-00001", "bank_b_getway_level_1@gmail.com", 1000000, "KZT")
	dHandler.assign(stub, "22222-00002", "bank_b_getway_level_1@gmail.com", 100000, "RMB")
	dHandler.assign(stub, "22222-00003", "bank_b_getway_level_2@gmail.com", 10000, "RMB")
	dHandler.assign(stub, "22222-00004", "bob@gmail.com", 1000, "RMB")
		
	return nil, nil
}

// Invoke  method is the interceptor of all invocation transactions, its job is to direct
// invocation transactions to intended APIs
func (t *AssetManagementChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	myLogger.Debugf("********************************Invoke****************************************")

	//	 Handle different functions
	if function == "assignOwnership" {
		// Assign ownership
		return t.assignOwnership(stub, args)
	} else if function == "transferOwnership" {
		// Transfer ownership
		return t.transferOwnership(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

// Query method is the interceptor of all invocation transactions, its job is to direct
// query transactions to intended APIs, and return the result back to callers
func (t *AssetManagementChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	myLogger.Debugf("********************************Query****************************************")

	// Handle different functions
	if function == "getOwnerContactInformation" {
		return t.getOwnerContactInformation(stub, args)
        } else if function == "getBalance" {
		return t.getBalance(stub, args)
	}

	return nil, errors.New("Received unknown function query invocation with function " + function)
}

func main() {

	//	primitives.SetSecurityLevel("SHA3", 256)
	err := shim.Start(new(AssetManagementChaincode))
	if err != nil {
		myLogger.Debugf("Error starting AssetManagementChaincode: %s", err)
	}

}
