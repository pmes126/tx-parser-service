### 1. [Introduction](#introduction)
    This repository contains an implementation of a transaction parser service on the Ethereum blockchain. 
    The service exposes a REST API that allows users to subscibe Ethereum addresses and then query the transactions for each address. The service periodically polls the Ethereum blockchain for new transactions for the subscibed addresses and stores the updates.
### 2. [Installation](#installation)
    1. Clone the repository
    git clone https://github.com/pmes126/tx-parser-service.git
    3. Run the service
    ``` bash
    cd tx-parser-service
    make run
    ```
    4. The service will be running on http://localhost:8080
    default port is 8080, it can be changed in the config file config.yaml
    5. To run the tests
    ``` bash
    make test
    ```

### 3. [Usage](#usage)
    1. Subscribe to an Ethereum address
    ``` bash
    curl -X POST http://localhost:8080/subscribe -d '{"address": "0xc0ffee254729296a45a3885639AC7E10F9d54979"}'
    ```
    2. Query the transactions for the subscribed address
    ``` bash
    curl -X GET http://localhost:8080/transactions?address=0xc0ffee254729296a45a3885639AC7E10F9d54979
    ```

### 4. [Design](#design)
    The service is using an http server to expose a REST API that allows users to subscribe to Ethereum addresses and query the transactions for each address. The ethParser component of the service then periodically polls the Ethereum blockchain for new blocks, then it looks for transactions involving the subscribed addresses and stores them in the in-memory data store. 

### 5. [Future Improvements](#future-improvements)
    ***1. Store Transactions in a Database***
    Use a database to store the transactions instead of an in-memory data store. This will allow the service to scale to handle a large number of transactions.
    ***2. Add Pagination***
    Add support for pagination to the transactions API. This will allow users to query transactions in batches.
    ***3. Add Integration Tests with a test Ethereum network***
    Add integration tests that run against a test Ethereum network, either a local node or a testnet. This will allow to test the service against real Ethereum transactions.
    ***4. Add Monitoring and Alerting***
    Add monitoring and alerting to the service. This will allow to monitor the health of the service and receive alerts when there are issues.
    ***5. Add logic for historical transactions***
    Add logic to fetch historical transactions for an address. This will allow users to query transactions that were processed before the service started.
    ***6. Investigate different approaches for finding matching transactions and storing them***
    Different approaches can be used with regards to leveraging goroutines for scanning transactions and using various data structures to store transactions in memory. The current implementation uses a simple approach.
    ***6. Dockerize the service***
    Dockerize the service in order to deploy and run in various environments.
    ***7. Add support for multiple Ethereum networks***
    The service currently supports the Ethereum mainnet. It can be extended to support multiple Ethereum networks such as Ropsten, Rinkeby, Kovan, etc.

### 6. [License](#license)
This project is strictly for interviewing purposes. All rights reserved.

### 7. [Contact](#contact)
For any questions or feedback, please contact me at pmesidis@gmail.com
