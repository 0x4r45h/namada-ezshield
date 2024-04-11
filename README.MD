Docker:
first time :
docker compose build

then exec into a container to 
- initialize namada chain and add your wallets shielded accounts
- import you osmosis accounts

docker compose run --rm --entrypoint /bin/bash ezshield

example : 
namadac utils join-network --chain-id shielded-expedition.88f17d1d14 

# PLEASE USE A BURNER MNEMONIC KEYS, BECAUSE WE ARE STORING THE PRIVATE KEY AS PLAINTEXT, IN NEXT RELEASES WE FIND A WAY TO PASS PASSWORDS TO THE CLI APP
namadaw derive --shielded --unsafe-dont-encrypt --alias ezshield-1 

namadaw gen-payment-addr --key ezshield-1 --alias ezshield-1-addr-1 


osmosisd keys add acc1 --recover --keyring-backend test

fund these accounts with namada and osmo
note that the `ezshield-1 ` account is a shielded account, so to fund it you should send NAAN to it's payment address `ezshield-1-addr-1`

osmosisd config node https://rpc.testnet.osmosis.zone:443
osmosisd config chain-id osmo-test-5
