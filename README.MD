This is CLI application, helps to easily make shielded IBC transfer between Namada and Osmosis

## Installation
### Docker (Recommended):
You only need docker installed.
when it is first time using this you should make some initializations, as follows:

build the image
```bash
docker compose build
```
now copy `.env.sample` to `.env` and modify as needed.    
run a temporary container and exec into it to perform initialization    
```bash
docker compose run --rm --entrypoint /bin/bash ezshield
```
now you have access to namada and osmosis binaries.    

##### PLEASE USE A BURNER MNEMONIC KEYS, BECAUSE WE ARE STORING THE PRIVATE KEY AS PLAINTEXT, IN NEXT RELEASES WE FIND A WAY TO PASS PASSWORDS TO THE CLI APP THROUGH ENV    

- initialize namada chain and add your wallets shielded accounts   

```bash
namadac utils join-network --chain-id shielded-expedition.88f17d1d14 

namadaw derive --shielded --unsafe-dont-encrypt --alias ezshield-1 

namadaw gen-payment-addr --key ezshield-1 --alias ezshield-1-addr-1 

```
- add osmosis configs and  wallets   
```bash
osmosisd config node https://rpc.testnet.osmosis.zone:443
osmosisd config chain-id osmo-test-5
osmosisd keys add acc1 --recover --keyring-backend test
```

Now fund these accounts with namada and osmo    
note that the `ezshield-1 ` account is a shielded account, so to fund it you should send NAAN to it's payment address `ezshield-1-addr-1`   

when you are done. leave the container, now run the container in background, it start to perform `shielded-sync` it background forever, to keep the state in sync

### Using the CLI    
everytime you want to use the app you should exec into this container    
```bash
docker compose exec ezshield bash
```
here you can use the CLI by running    
```bash
ezshield
```
#### Build from source
you can build the project with `go build` command.
you should also have `namada` and `osmosisd` available and accessible in the system path    
run `./ezshield` to use the app


---
