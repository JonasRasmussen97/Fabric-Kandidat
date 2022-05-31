The two folders Product1 and Product2 contain two examples of chaincode that can be executed in Hyperledger Fabric. 

The folder Product1 contains files related to example of contract functionality, such as creating, signing, updating and deleting them. It also includes example use of the creation and retrieval of private data.
The chaincode can been found in the file named "contract_handling.go" located in the "chaincode-contract-go" subfolder inside the "Product1" folder.

The folder Product2 contains files related to examples of endorsement functionality, such as requiring endorsement from one organization when creating a contract and multiple organizations when signing.

The "test_network" folder contains the files neccessary to start the blockchain network, through the "network.sh" file, where the chaincode from both "Product1" and "Product2" can be installed, if desired.

Both the Frontend and the Blockchain network use some components already made by Hyperledger Fabric, which is why there is a large similarity between some of the files. (Inspiration from here: https://github.com/hyperledger/fabric-samples)