# hardhat

## ULTIMATE PURPOSE

Decentralized file storage on all Cosmos Hub counterparties via packet forward middleware.

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

## Make my pagerduty shut up

Mitigations are [placid](github.com/notional-labs/placid).  As reporting to [ICF](https://interchain.io) has failed and resulted in retaliation, you will need to encourage chain teams to integrate the mitigations found in the placid repository.

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
* COMMENTARY ON THIS ISSUE FROM VARIOUS ICF FUNDED PEOPLE/TEAMS



**Jack Zampolin's commentary**

```
Jacob, you aren’t a security researcher. This is a long standing issue in the codebase with many existing mitigations. Your months long campaign of self aggrandizement using threats holding yourself up as the only hero helping is self destructive and transparently self serving. Your refusal to work with core teams in a productive manner is part and parcel of a pattern of destructive behavior that we as a community cannot continue to countenance. Please take this conversation to another channel.
:100:
4
:pray:
1
:heart:
1



Jacob Gadikian
  7 months ago
Security researcher is, security researcher does Jack.
I hope sincerely it is okay to discuss these matters here. 
It would be like me saying "jack you aren't a light client aficianado you're a chef"
No, you're both jack.
You can attempt to define me but ultimately I'll only be what I am.


Jacob Gadikian
  7 months ago
While doing security research pursuant to this matter, Zaki let me know that I am the only person who has reproduced it.
It is because I had been carefully watching a string of incidents that all shared "336 block parts" (edited) 


Jacob Gadikian
  7 months ago
This spanned years.


Jacob Gadikian
  7 months ago
When I see a normally operating chain suddenly begin to not make blocks right, as a validator I take note of it. When I see bandwidth consumption that is usually under one megabit per second reach one gigabit per second, I take note of it.


Jacob Gadikian
  7 months ago
The really surprising part though Jack, is that you and others have a problem with that


Jacob Gadikian
  7 months ago
It's honestly shocking


Jacob Gadikian
  7 months ago
All I did is write about 50 pages of documentation on a bunch of different issues, and work on that with a number of people on the notional team and a number of other teams.


Jacob Gadikian
  7 months ago
Why you choose to make these technical findings personal I cannot explain.


Jacob Gadikian
  7 months ago
There's no aggrandizement here jack.  The Cosmos Hub chose to fund notional for security work and we found a security issue. Then we reported it.
Almost immediately thereafter, we begin to experience various forms retaliation, but strangely only from teams that are heavily funded by the ICF like yours
Every single other team thought that it was in fact a serious problem that field length limitations in IBC Do not exist and that it's fairly cheap and trivial to stop Cosmos blockchains.


Jacob Gadikian
  7 months ago
Notably, The binary builders team does not seem to be affected by this strange funding related change to clear technical findings and has been enormously helpful.


Jacob Gadikian
  7 months ago
Lastly, 
@Jack
 -- if this is a long standing issue (I mean it is, first time I reported it to 
@Ethan Buchman
 was 2021....) Then sir, why is it not documented so that people don't hit it?
Also sent to the channel


Jack
  7 months ago
Admins please remove the spam? This is beyond ridiculous
:heart:
3
:heavy_plus_sign:
3



Jacob Gadikian
  7 months ago
if you don't put each message in the channel as you've been doing, we can talk in a nice self-contained thread Jack


Jacob Gadikian
  7 months ago
I think there are serious problems with security reporting in cosmos.


Jacob Gadikian
  7 months ago
I also think that there are issues in comet we should be discussing.


Jacob Gadikian
  7 months ago
If folks wish to remove me for that, I'm okay with it, would be a clear indication I've made some poor choices.


Mircea
  7 months ago
Your only poor choice is this continued harassment of people who just asked you to play nicely.
You've got this persecution complex + conspiracy theory going in every possible communications channel.
Like, I've personally went through this with you ages ago and now just DGAF anymore. Personally I think you should just be banned from the repo if you continue with this nonsense.
:hearts:
2



Jacob Gadikian
  7 months ago
that's acceptable to me Mircea, I'd get some answers from that.
:+1::skin-tone-3:
1



Jack
  7 months ago
Should have been a long time ago
:heart:
1



Jack
  7 months ago
Who’s the admin here?
:heart:
1



Jacob Gadikian
  7 months ago
idk.  I do know that there's a shockingly wide divergence of opinion on even if there's a security issue in comet related to the p2p storms report I wrote


Jacob Gadikian
  7 months ago
and that I have written code with a 100% success rate at making chains not make blocks


Jacob Gadikian
  7 months ago
and documented that


Jacob Gadikian
  7 months ago
I'd like to work on solutions, so that no one can do that ever


Jacob Gadikian
  7 months ago
I'm deeply concerned that has been framed as harassment, and that my company's name is still on a report we did not author


Jacob Gadikian
  7 months ago
that I did not even get an opportunity to review, and which I think does not actually describe what we reported


Jacob Gadikian
  7 months ago
Jack, I think if you were faced with those circumstances, that'd be a concern for you too


Jacob Gadikian
  7 months ago
afaik this is the correct place to discuss such concerns wrt cometbft, so here I am


Jacob Gadikian
  7 months ago
@Mircea
 I have written code with a 100% success rate at making chains not make blocks
... harassment?  conspiracy?


Mircea
  7 months ago
That was all I had to say to you Jacob. If you can't see how your overall tone and aggressiveness in communication is causing people mental anguish, I don't know what tell you.
And yes, it's perceived as threatening. You've been told this dozens of times before by many different people.


Jacob Gadikian
  7 months ago
Examples of this so-called threatening language?
Really, the software I made really and actually stops chains from making blocks reliably, on every chain I tried it on.
Actually, only one person has described my language as "threatening" and that is 
@Ethan Buchman


Mircea
  7 months ago
That's because no one else wants to even talk to you. I don't even know why I'm doing it right now.
It accomplishes nothing when you're set in your way like this.


Jacob Gadikian
  7 months ago
he did that in the same set of messages where he described me as a hysterical child and said that I was like a kindergartener


Jacob Gadikian
  7 months ago
Mircea, that's just not true.  I reviewed my work with a number of other teams, no issues whatsoever (edited) 


Mircea
  7 months ago
Anything anyone says is always turned around on them and blasted over social media and coms channels. Why would anyone want to go through that?


Jacob Gadikian
  7 months ago
That is not true Mircea.


Mircea
  7 months ago
You're literally doing it right now with private coms from Ethan


Mircea
  7 months ago
¯\_(ツ)_/¯


Jacob Gadikian
  7 months ago
Sir, there's really nothing wrong with that.  We were trying to make a security report and that was the reaction of the vidce president of the interchain foundation


Mircea
  7 months ago
I expect what I say in this thread here to also be plastered on twitter or whatever with "evil informal employee" tags


Jacob Gadikian
  7 months ago
and that's notable and sad sir


Jacob Gadikian
  7 months ago
nah


Jacob Gadikian
  7 months ago
you're.... I like you man


Jacob Gadikian
  7 months ago
:man-shrugging:


Jacob Gadikian
  7 months ago
you're def not an evil informal employee


Jacob Gadikian
  7 months ago
nor is bucky evil -- his actions are slowing down response to an issue though sir


Jacob Gadikian
  7 months ago
and that matters and has impacts way beyond informal


Jacob Gadikian
  7 months ago
you had the guts to tell me repeatedly I was wrong about missed blocks.  Took me a year to see that, but yes, I was quite wrong.


Jack
  7 months ago
Mircea, he likes you, thats when you need to watch out


Mircea
  7 months ago
So how would you feel if you were Ethan, and everyone and their dog is dumping on the company you run and the thing you've worked to build?
And I'm not even saying anything about what facts are being discussed. Just the simple fact that you're 100% of the time under mental attack


Jacob Gadikian
  7 months ago
Mircea, how about myself?


Jacob Gadikian
  7 months ago
I found a bug with billions of dollars of production systems


Jacob Gadikian
  7 months ago
the reaction was to call me a hysterical child


Jack
  7 months ago
its not a bug


Jacob Gadikian
  7 months ago
eh what?


Jack
  7 months ago
its a series of performance issues with complex and multivariate fixes


Jacob Gadikian
  7 months ago
I think that we should really classify anything that can cause unfair economic outcomes to users of systems using comet, a bug


Jack
  7 months ago
and your simplification of the issue and rabid attacking of anyone who disagrees with you slightly is driving people away from engaging with you


Jack
  7 months ago
to which you respond with more and more walls of text


Jack
  7 months ago
No one cares anymore jacob


Jack
  7 months ago
again I find myself in a position of trying to tell you this


Jack
  7 months ago
because I actually liked you


Jacob Gadikian
  7 months ago
Jack, that's the problem


Jack
  7 months ago
but you again prove you have no ability to see yourself and how you come off


Jacob Gadikian
  7 months ago
any chain can be trivially halted, the foundations security contractor published it to the world, and no one at ICFormulovet cares


Jack
  7 months ago
the problem is you thats right jacob


Mircea
  7 months ago
It's just tiring man. Like, I am going to peace out of this channel now. You've made your point. Try to make peace with the fact that sometimes your point either doesn't come through or maybe it's even not correct.


Jacob Gadikian
  7 months ago
The problem, Mircea, is that afaik, I'm correct, and I've checked this with numerous other teams who also think I'm correct


Mircea
  7 months ago
So let them push it forward then


Jacob Gadikian
  7 months ago
yes, one of them has already forked


Mircea
  7 months ago
If they agree and on the severity of this, why aren't they all pushing it as hard as you


Jack
  7 months ago
Jacob you’ve repeated misrepresented these “other teams” to the point that no one can trust you


Jacob Gadikian
  7 months ago
maybe that's not their job.  It's my job.


Mircea
  7 months ago
 yes, one of them has already forked
So why go through all this anguish instead of forking then?


Mircea
  7 months ago
It's OSS man, just fork it


Jack
  7 months ago
I’ve followed up with many of these people and had productive conversations with them where they also say you are abrasive and not correct yet you keep using people’s name


Jacob Gadikian
  7 months ago
Yeah, so as a validator, the idea is to fix it


Mircea
  7 months ago
Forking is a valid fix


Jacob Gadikian
  7 months ago
I don't believe you Jack.  But ok.


Jack
  7 months ago
As a validator you need to sign blocks.


Jack
  7 months ago
You don’t believe the truth jacob and its going to hit you in the face like a wall of bricks when you are sitting a bunch of chatrooms you created and the only messages in them are yours screaming into the void


Jacob Gadikian
  7 months ago
then why are we suddenly discussing the removal of the mempool?


Jacob Gadikian
  7 months ago
if I am incorrect?


Jack
  7 months ago
Because that conversation has been happening for years and is long overdue


Mircea
  7 months ago
This discussion has been around for at least 5 years


Jack
  7 months ago
Its been badly needed


Jacob Gadikian
  7 months ago
I don't think that's how the recent instance of it came about at all.


Jack
  7 months ago
people thinking the mempool needs replacing != agreeing with jacob


Mircea
  7 months ago
I recall this literally being talked about before the hub launch


Jack
  7 months ago
I’ve been talking with marko and zaki about this for years and SL even has a from scratch rewrite of tendermint we are working on because of this set of issues


Jacob Gadikian
  7 months ago
I'm aware of Gordian Jack, and frankly this is its strongest aspect


Jack
  7 months ago
TM perf in current code base has upper bounds
:100:
1



Jack
  7 months ago
need a rewrite to eliminate those
:man-shrugging:
1



Jacob Gadikian
  7 months ago
Now the issue was documented nowhere at all


Jack
  7 months ago
It was documented in many many places


Jacob Gadikian
  7 months ago
and thus we are handing all of the teams entering the ecosystem a footgun


Mircea
  7 months ago
Man, are you really going to gripe about docs in this ecosystem? Like, really? lol
:-1:
1



Jack
  7 months ago
what you’ve done is taken a number of disparate issues and tried to call it yours and use it to promote yourself.


Jacob Gadikian
  7 months ago
Mircea the problem is with the footguns


Jack
  7 months ago
its an abuse of the respect you’ve been given


Jacob Gadikian
  7 months ago
Jack, I took a bunch of disprate issues, and I combined them and I made a way to stop any chain and I want to fix the fact that it is trivial to halt cosmos chains


Jack
  7 months ago
and its embarrassing for the whole ecosystem to have to watch you squirm and constantly demand recgonition in every conceivable channel.


Jacob Gadikian
  7 months ago
I am not demanding recognition Jack


Jacob Gadikian
  7 months ago
I didn't want this published


Jack
  7 months ago
degrade performance != stop


Jacob Gadikian
  7 months ago
Jack, I said stop and I mean it


Jack
  7 months ago
and you are repeatedly demanding recognition. thats what the whole thing is about


Jack
  7 months ago
Jacob i’ve seen your tool deployed as a validiator on testnets a number of times


Jacob Gadikian
  7 months ago
Jack, I'm asking to have my company's name removed from the issue


Jack
  7 months ago
its degraded perf


Jack
  7 months ago
… while you are screaming in every channel about how people need to listen to you or else


Jacob Gadikian
  7 months ago
4 blocks in 30 minutes?


Jack
  7 months ago
still faster than bitcoin


Jack
  7 months ago
:wave:


Jacob Gadikian
  7 months ago
"faster than bitcoin" is not -- afaik -- what we are trying to accomplish here





Jacob Gadikian
  7 months ago
I think we're trying to build reliable global scale systems
```





## LIMITATIONS

Like any blockchain client software, `hardhat` can only make transactions that are explicitly supported by the chains it is used to test.  It has no magic blackhat powers.  



