/*
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"log"
	// State based endorsement
	//"github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Can be used to control the type of the contract. Not currently used.
const (
	// Created but not signed yet by other party.
	contractOpen = "O"
	// Created, signed and expired/closed. Collaboration between the two parties has ended.
	contractClosed = "C"
	// Created and signed by both parties and not expired or closed. 
	contractSigned = "S"
)

type SmartContract struct {
	contractapi.Contract
}

type Contract struct {
	// ID of the contract
	ID string `json:"contractID"`
	// Organisation that the client that calls the transaction belongs to
	OwnerOrg string `json:"ownerOrg"`
	// ID of the client that owns the contract e.g., Anders
	OwnerId string `json:"ownerId"`
	// The ID of the user that signs the contract e.g., Brian signs a contract with Anders
	SignerId string `json:"signerId"`  
	// Age of the contract creator
	OwnerAge string `json:"ownerAge"`
	// Price of the amount of electricity. We keep these as strings, because everything is converted to JSON anyway.
	Price string `json:"price"`
	// Describes the amount of electricity is being sold 
	Amount string `json:"amount"`
}

// The publicDescription string is not currently used, but it could be beneficial to allow users to write a small text to their contract. 
func (s *SmartContract) CreateContract(ctx contractapi.TransactionContextInterface, contractID string, price string, amount string, publicDescription string) error {
	// Client ID
	var clientId, fail = ctx.GetClientIdentity().GetID();
	// Client MSPID
	var id, err = ctx.GetClientIdentity().GetMSPID();
	// Peer MSPID
	var id2, err2 = shim.GetMSPID();
	// Made to show how we can retrieve attributes of a certificate and use them in our contracts.
	 var age, found, err3 = ctx.GetClientIdentity().GetAttributeValue("age");
	// Convert age value from string to int.
	var intVar, err4 = strconv.Atoi(age)
	var requiredAge = 18

	if fail != nil {
		return fmt.Errorf("Unable to retrieve id of client: %s", fail)
	}

	// Access control #1: If the client and peer org are not equal, then we return error.
	if id != id2 {
		return fmt.Errorf("client org and peer org are not equal and operation is therefore not allowed. ClientID: %s and PeerID: %s", id, id2)
	}
	if err != nil {
		return fmt.Errorf("client org and peer org are not equal and operation is therefore not allowed. %s", err)
	}
	if err2 != nil {
		return fmt.Errorf("client org and peer org are not equal and operation is therefore not allowed. %s", err2)
	}
	if err3 != nil {
		return fmt.Errorf("Unable to fetch 'age' attribute on cert: %v", err3)
	}
	if found == false {
		return fmt.Errorf("Unable to fetch 'age' attribute on cert: %v", found)
	}
	if err4 != nil {
		return fmt.Errorf("Unable to parse string to int");
	}
	// Access Control #2: Additional access control on specific attribute of the certificate.
	if intVar <= requiredAge {
		return fmt.Errorf("To create contracts you have to be atleast %v years old", requiredAge);
	} 

	// Add the contract and the owner name to the private data collection.
	// We convert the value to a byte array because that is what the method expects.
	var privateDataErr = ctx.GetStub().PutPrivateData(fmt.Sprintf("_implicit_org_%s", id), contractID, []byte("Jonas"))
	if privateDataErr != nil {
		return fmt.Errorf("failed to put Asset private details: %v", privateDataErr)
	}
	
	contract := Contract{
		ID: contractID,
		OwnerId: clientId,
		OwnerOrg: id,
		SignerId: "",
		OwnerAge: "10",
		Price: price,
		Amount: amount,
	}


	// Converts the contract object to json.
	contractBytes, err := json.Marshal(contract)
	if err != nil {
		return fmt.Errorf("failed to create asset JSON: %v", err)
	}

	// Adds the contract to the world state database. 
	// https://miro.medium.com/max/700/1*tMk4PCN3bewrTZW8hU17xg.png 
	err = ctx.GetStub().PutState(contract.ID, contractBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset in public data: %v", err)
	}
	return nil
}

func (s *SmartContract) ReadContract(ctx contractapi.TransactionContextInterface, contractID string) (*Contract, error) {
	// Anyone can read contracts, so we do not include any access control. This is from the public world state.
	
	// Client MSPID
	 // var id, err = ctx.GetClientIdentity().GetMSPID();
	// var value, found, err2 = ctx.GetClientIdentity().GetAttributeValue("email")
	
	contractJSON, err := ctx.GetStub().GetState(contractID)
	// If an error occurs.
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
/*
	if found == false {
		return nil, fmt.Errorf("Email does not exist on certificate");
	}
	if err2 != nil {
		return nil, fmt.Errorf("Failed to fetch certificate information %v", err2);
	}

	fmt.Printf(value); */
	// If we are attempting to read a contract that does not exist.
	if contractJSON == nil {
		return nil, fmt.Errorf("%s does not exist", contractID)
	}

	var contract *Contract
	err = json.Unmarshal(contractJSON, &contract)
	if err != nil {
		return nil, err
	}

	// Access control #3:  Restricts contract data to owners only. Not really useful in our use case but for testing purposes.
	/*
	if contract.OwnerOrg != id {
		return nil, fmt.Errorf("%s is not allowed to read contract %s owned by %s", id, contractID, contract.OwnerOrg)
	}
	*/
	return contract, nil
}

// Since data is immutable, this is actually just another transaction that indicates that selected contracted is to be considered deleted. 
func (s *SmartContract) DeleteContract(ctx contractapi.TransactionContextInterface, contractID string) error {
	// We start by retrieving the contract by the ID, such that we can check values before deleting it.
	var contractJSON, err1 = ctx.GetStub().GetState(contractID)
	
	// ID of the client calling the method.
	var id, err2 = ctx.GetClientIdentity().GetMSPID();

	if err1 != nil {
		return fmt.Errorf("No contract with id %s exists", contractID)
	}

	// Parse the contract JSON
	var contract *Contract
	var jsonErr = json.Unmarshal(contractJSON, &contract) 
	if jsonErr != nil {
		return jsonErr
	}
	if err2 != nil {
		return fmt.Errorf("Unable to retrieve MSPID of client")
	}

	// Only allow delete if the client is part of the org that owns the contract.
	if contract.OwnerOrg != id {
		return fmt.Errorf("%s is not allowed to delete a contract owned by %s", id, contract.OwnerOrg)
	}

	if contract.SignerId != "" {
		return fmt.Errorf("You cannot delete a contract that has been signed!")
	}

	// If nothing fails, we delete the contract.
	var delErr = ctx.GetStub().DelState(contractID)
	if delErr != nil {
		return fmt.Errorf("Unable to delete contract with id %s: %s", contractID, delErr)
	}

	return nil
}

// Allows for signing of a contract such that two parties can do business.
func (s *SmartContract) SignContract(ctx contractapi.TransactionContextInterface, contractID string) error {		
		// Client ID
		var clientId, fail = ctx.GetClientIdentity().GetID();
		// Client MSPID
		var id, err = ctx.GetClientIdentity().GetMSPID();
		// Peer MSPID
		var id2, err2 = shim.GetMSPID();
		// Made to show how we can retrieve attributes of a certificate and use them in our contracts.
		// var age, found, err3 = ctx.GetClientIdentity().GetAttributeValue("age");
		// Convert age value from string to int.
		// var intVar, err4 = strconv.Atoi(age)
		// You have to be 18 to sign a contract
	    // var requiredAge = 10
	
		if fail != nil {
			return fmt.Errorf("Unable to retrieve id of client: %s", fail)
		}
	
		// Access control #1: If the client and peer org are not equal, then we return error.
		if id != id2 {
			return fmt.Errorf("client org and peer org are not equal and operation is therefore not allowed.")
		}
		if err != nil {
			return fmt.Errorf("client org and peer org are not equal and operation is therefore not allowed.")
		}
		if err2 != nil {
			return fmt.Errorf("client org and peer org are not equal and operation is therefore not allowed.")
		}

		
		/* if err3 != nil {
			return fmt.Errorf("Unable to fetch 'age' attribute on cert: %v", err3)
		}
		if found == false {
			return fmt.Errorf("Unable to fetch 'age' attribute on cert: %v", found)
		}
		if err4 != nil {
			return fmt.Errorf("Unable to parse string to int");
		}
		// Access Control #2: Additional access control on specific attribute of the certificate.
		if intVar <= requiredAge {
			return fmt.Errorf("To sign contracts you have to be atleast %v years old", requiredAge);
		} */
		
		// We retrieve the contract that is in question.
		contractJSON, getErr := ctx.GetStub().GetState(contractID)
		if getErr != nil {
			return fmt.Errorf("failed to read from world state: %v", getErr)
		}
		if contractJSON == nil {
			return fmt.Errorf("%s does not exist", contractID)
		}
	
		var contract *Contract
		var marshalErr = json.Unmarshal(contractJSON, &contract)
		if marshalErr != nil {
			return marshalErr
		}

		// You can not sign your own contract
		if contract.OwnerId == clientId {
			return fmt.Errorf("It is not possible to sign your own contract.")
		}


		// You can't sign a contract twice
		if(contract.SignerId == clientId) {
			return fmt.Errorf("You already signed this contract!")
		}

		// Only allow signing if it is not already signed
		if(contract.SignerId != "") {
			return fmt.Errorf("Contract %s is already signed by another party", contract.ID)
		}

		// We set the signerId of the contract to the id of the client that signs it
		contract.SignerId = clientId
		// Parse the new contract to JSON
		updatedContractJSON, updatedContractErr := json.Marshal(contract)
		if updatedContractErr != nil {
		return fmt.Errorf("failed to marshal asset: %v", updatedContractErr)
		}

		// Put the updated contract in the ledger.
		return ctx.GetStub().PutState(contractID, updatedContractJSON)
	
}

// This function returns the private data of a smart contract. Only allowed by the creator of the contract.
func (s *SmartContract) GetContractPrivateData(ctx contractapi.TransactionContextInterface, contractID string) (string, error) {
	
	// Client MSPID
	var clientId, err = ctx.GetClientIdentity().GetMSPID();
	// Peer MSPID
	var peerId, err2 = shim.GetMSPID();

	if err != nil {
		return "", fmt.Errorf("Unable to retrieve Client MSPID %s", err)
	}
	if err2 != nil {
		return "", fmt.Errorf("Unable to retrieve Peer MSPID %s", err2)
	}

	// We need to verify that the client org matches the peer org.
	if clientId != peerId {
		return "", fmt.Errorf("The peer has to match your organization!")
	}
	
	// In this scenario, client is only authorized to read/write private data from its own peer.
	var collectionName = fmt.Sprintf("_implicit_org_%s", clientId)

	
	immutableProperties, err := ctx.GetStub().GetPrivateData(collectionName, contractID)
	if err != nil {
		return "", fmt.Errorf("failed to read asset private properties from client org's collection: %v", err)
	}
	if immutableProperties == nil {
		return "", fmt.Errorf("asset private details does not exist in client org's collection: %s", contractID)
	}
	return string(immutableProperties), nil
}


// Main function to run the contract.
func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		log.Panicf("Error create contract handling chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting contract handling chaincode: %v", err)
	}
}


