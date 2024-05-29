# hardhat

## How to Install

```bash
git clone https://github.com/somatic-labs/hardhat
cd hardhat
go install ./...
```

## How to use

Pre-baked mainnet configurations are in the configurations folder.  You need a file named `seedphrase`.  You may or may not want to set up your own node with a 10 GB mempool that accepts 50,000 transactions.  You put RPC urls in nodes.toml, and configure the other settings in there.  Then you just run `hardhat` in the same folder as the nodes.toml file and the `seedphrase` file.

## You set off my pagerduty

Possibly.  But really, I didn't set off your pagerduty.  Lack of diligence and attending to security reports, over the course of years, set off your pagerduty.  Individuals should not be able to stop or slow blockchains.  If one person can push around a decentralized system, it ceases to be decentralized. 

Tweets should not be able to harm blockchains.

Strong chain needs strong testing and strong tweet.

Weak chain die.

## Context

I've known that the spammy attack can be enhanced by maybe 100x.  This does that, and it also works with https://github.com/notional-labs/rpc-crawler.

spammy-go takes the output of the rpc-crawler, which is open RPCs.  

Then, spammy-go will blast 30kb IBC transfers into every rpc at once at top speed.  

Specificially this exploits:

* banana king
* p2p-storms



## Bonus content

* All of my [emails with the interchain foundation](./emails)
* reports to the interchain foundation about this issue dating back to 2021

## LIMITATIONS

Like any blockchain client software, `hardhat` can only make transactions that are explicitly supported by the chains it is used to test.  It has no magic blackhat powers.  



