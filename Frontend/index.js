/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const fs = require('fs');
const yaml = require('js-yaml');

const express = require('express')
const app = express()
const port = 3000



// Network Configuration
let connectionProfile1 = yaml.safeLoad(fs.readFileSync('./gateway/connection-org2.yaml', 'utf8'));
let connectionProfile2 = yaml.safeLoad(fs.readFileSync('./gateway/connection-org1.yaml', 'utf8'));

const { Gateway, Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const path = require('path');
const { buildCAClient, registerAndEnrollUser, enrollAdmin } = require('./test-application/javascript/CAUtil.js');
const { buildCCPOrg1, buildCCPOrg2, buildWallet } = require('./test-application/javascript/AppUtil.js');

const channelName = 'mychannel';
const chaincodeName = 'secured';

const org1 = 'Org1MSP';
const org2 = 'Org2MSP';
const Org1UserId = 'appUser1';
const Org2UserId = 'appUser2';

const RED = '\x1b[31m\n';
const GREEN = '\x1b[32m\n';
const RESET = '\x1b[0m';




app.get('/', (req, res) => {
    res.sendFile(__dirname + '/index.html');
  });

  app.get('/chaincode', (req, res) => {
    res.sendFile(__dirname + '/chaincode.html');
  });

app.get('/enrollAdmin', async(req, res) => {
    try {
        // Create a new CA client for interacting with the CA.
        const caInfo = connectionProfile2.certificateAuthorities['ca.org1.example.com'];
        const caTLSCACerts = caInfo.tlsCACerts.pem;
        const ca = new FabricCAServices(caInfo.url, { trustedRoots: caTLSCACerts, verify: false }, caInfo.caName);
	
		
	
		const wallet = await Wallets.newFileSystemWallet(walletPath);

        // Check to see if we've already enrolled the admin user.
        const userExists = await wallet.get('admin');
        if (userExists) {
            console.log(userExists);
            console.log('An identity for the client user "admin" already exists in the wallet');
            // res.send('An identity for the client user "user1" already exists in the wallet');
        }

        // Enroll the admin user, and import the new identity into the wallet.
        const enrollment = await ca.enroll({ enrollmentID: 'admin', enrollmentSecret: 'adminpw' });
        const x509Identity = {
            credentials: {
                certificate: enrollment.certificate,
                privateKey: enrollment.key.toBytes(),
            },
            mspId: 'Org1MSP',
            type: 'X.509',
        };
        await wallet.put('admin', x509Identity);
        console.log('Successfully enrolled client user "admin" and imported it into the wallet');
		res.send('Successfully enrolled client user "admin" and imported it into the wallet');

    } catch (error) {
        console.error(`Failed to enroll client user "admin": ${error}`);
		res.send(`Failed to enroll client user "admin": ${error}`);
        process.exit(1);
    }
})

app.get('/registerUser', async(req, res) => {
    try {
        // Create a new CA client for interacting with the CA.
        const caInfo = connectionProfile2.certificateAuthorities['ca.org1.example.com'];
        const caTLSCACerts = caInfo.tlsCACerts.pem;
        const ca = new FabricCAServices(caInfo.url, { trustedRoots: caTLSCACerts, verify: false }, caInfo.caName);
		
	
		const wallet = await Wallets.newFileSystemWallet(walletPath);

        // Register a new user
        const register = await ca.register({enrollmentID: "jonas", enrollmentSecret: "abe", role: "client"});
        console.log(register);
        
    } catch (error) {
        console.error(`Failed to enroll client user "jonas": ${error}`);
		res.send(`Failed to enroll client user "jonas": ${error}`);
        process.exit(1);
    }
})

// Initializes the certificates. If we call this endpoint a second time, it switches to org2 and 3rd time back to org1 in a loop.
app.get('/initGateway', async(req, res) => {
	console.log(`${GREEN}--> Fabric client user & Gateway init: Using Org1 identity to Org1 Peer${RESET}`);
	// build an in memory object with the network configuration (also known as a connection profile)
	const ccpOrg1 = buildCCPOrg1();

	// build an instance of the fabric ca services client based on
	// the information in the network configuration
	const caOrg1Client = buildCAClient(FabricCAServices, ccpOrg1, 'ca.org1.example.com');

	// setup the wallet to cache the credentials of the application user, on the app server locally
	const walletPathOrg1 = path.join(__dirname, 'wallet', 'org1');
	const walletOrg1 = await buildWallet(Wallets, walletPathOrg1);

	// in a real application this would be done on an administrative flow, and only once
	// stores admin identity in local wallet, if needed
	await enrollAdmin(caOrg1Client, walletOrg1, org1);
	// register & enroll application user with CA, which is used as client identify to make chaincode calls
	// and stores app user identity in local wallet
	// In a real application this would be done only when a new user was required to be added
	// and would be part of an administrative flow
	await registerAndEnrollUser(caOrg1Client, walletOrg1, org1, Org1UserId, 'org1.department1');
	
	try {
		// Create a new gateway for connecting to Org's peer node.
		const gatewayOrg1 = new Gateway();
		//connect using Discovery enabled
		await gatewayOrg1.connect(ccpOrg1,
			{ wallet: walletOrg1, identity: Org1UserId, discovery: { enabled: true, asLocalhost: false } });

		return gatewayOrg1;
	} catch (error) {
		console.error(`Error in connecting to gateway for Org1: ${error}`);
		process.exit(1);
	}	
	
})

app.get('/test', async(req, res) => {
	



})

// This works!
app.get('/readContract/:id', async(req, res) => {
		// build an in memory object with the network configuration (also known as a connection profile)
		const ccpOrg1 = buildCCPOrg1();
		// setup the wallet to cache the credentials of the application user, on the app server locally
		const walletPathOrg1 = path.join(__dirname, 'wallet', 'org1');
		const walletOrg1 = await buildWallet(Wallets, walletPathOrg1);

	try {	
	// Create a new gateway for connecting to Org's peer node.
	const gatewayOrg1 = new Gateway();

	await gatewayOrg1.connect(ccpOrg1, {
		wallet: walletOrg1,
		identity: Org1UserId,
		discovery: { enabled: true, asLocalhost: true } // using asLocalhost as this gateway is using a fabric network deployed locally
	});

	
	// Build a network instance based on the channel where the smart contract is deployed
	const network = await gatewayOrg1.getNetwork(channelName);
	// Get the contract from the network.
	const contract = network.getContract(chaincodeName);
	let result;
	result = await contract.submitTransaction('ReadContract', req.params.id);
	console.log('*** Result: ' + result.toString());
	res.send(result.toString());
	} catch(error) {
		console.log(error);
		res.send("Unable to retrieve contract " + req.params.id);
	}
});

app.post('/createContract/:id/:price/:amount/:publicDescription', async(req, res) => {
	// build an in memory object with the network configuration (also known as a connection profile)
	const ccpOrg1 = buildCCPOrg1();
	// setup the wallet to cache the credentials of the application user, on the app server locally
	const walletPathOrg1 = path.join(__dirname, 'wallet', 'org1');
	const walletOrg1 = await buildWallet(Wallets, walletPathOrg1);

	try {
		
	// Create a new gateway for connecting to Org's peer node.
	const gatewayOrg1 = new Gateway();

	await gatewayOrg1.connect(ccpOrg1, {
		wallet: walletOrg1,
		identity: Org1UserId,
		discovery: { enabled: true, asLocalhost: true } // using asLocalhost as this gateway is using a fabric network deployed locally
	});

	
	// Build a network instance based on the channel where the smart contract is deployed
	const network = await gatewayOrg1.getNetwork(channelName);
	// Get the contract from the network.
	const contract = network.getContract(chaincodeName);
	let result;
	result = await contract.submitTransaction('CreateContract', req.params.id, req.params.price, req.params.amount, req.params.publicDescription);
	console.log('*** Result: ' + result.toString());
	res.send(result.toString());
	} catch(error) {
		console.log(error);
		res.send("Unable to create contract " + req.params.id);
	}
})

app.post('/signContract/:id', async(req, res) => {
	// build an in memory object with the network configuration (also known as a connection profile)
	const ccpOrg1 = buildCCPOrg1();
	// setup the wallet to cache the credentials of the application user, on the app server locally
	const walletPathOrg1 = path.join(__dirname, 'wallet', 'org1');
	const walletOrg1 = await buildWallet(Wallets, walletPathOrg1);

	try {	
		// Create a new gateway for connecting to Org's peer node.
		const gatewayOrg1 = new Gateway();
	
		await gatewayOrg1.connect(ccpOrg1, {
			wallet: walletOrg1,
			identity: Org1UserId,
			discovery: { enabled: false, asLocalhost: true } // using asLocalhost as this gateway is using a fabric network deployed locally
		});
		
		// Build a network instance based on the channel where the smart contract is deployed
		const network = await gatewayOrg1.getNetwork(channelName);
		// Get the contract from the network.
		const contract = network.getContract(chaincodeName);
	
		let resultCreateContract;
		resultCreateContract = await contract.submitTransaction('SignContract', req.params.id);
		console.log('*** Result: ' + result.toString());
		res.send(result.toString());
		} catch(error) {
			console.log(error);
			res.send("Unable to sign contract " + req.params.id);
		}

	
		
	
	
	
	
})





 
app.listen(port, () => {
  console.log(`Example app listening on port ${port}`)
})



