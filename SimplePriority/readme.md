#### Simple Channel Prioritization Example
I'm currently (May 2021) working on building a webscraper at work, and ran into an interesting challenge. I have two output channels that send/receive structs I want to manipulate, summarize, and then document in a csv file. There's a channel of structs for successful page scrapes, and a channel for unsuccessful page scrapes. In general, they undergo the following process:

- [successfully scraped] &#8594; [processing of scraped contents] &#8594; [creation of documentation] &#8594; [write documentation to .csv line]
- [unsuccessfully scraped] &#8594; [creation of documentation] &#8594; [write documentation to .csv line]

In general, I expect significantly more volume over the successfully scraped channel. Both processes have small buffers, but I don't want to be carrying too much data at any one time during the process. I'm sending over both channels to my documentation creator, which then delivers a slice of slices of strings (each internal slice is one row) to batch write to the .csv. As a consequence, I want to manage traffic into the documentation creator, prioritizing the success channel.

`SimpleChannelPriority.go` contains a code sample mocking how I ended up handling the design challenge. I simulate two agents sending data over a channel, and prioritize one of them. The program constantly listen for communications over either channel, and wait if there is no sent data over either channel. The process cleanly exits by waiting for both channels to be empty, and then sending an exiting signal over an 'exit' channel.

Due to the Sleeps on each channel, we expect the following order of prints:
1. hp: a
2. hp: b
3. hp: c
4. lp: a
5. [Wait period]
6. hp: d
7. hp: e
8. lp: c
9. lp: d
10. [Exit]

 It's worth noting that in general, I don't expect the unsuccessful channel to fill up (sends over the success channel are sparse enough that I never expect it to be completely dominating the unsuccessful channel). If the unsuccessful channel does fill, this will ultimately lead to enough blocking of workers that the successful channel becomes empty, which leads to writes (in the example, prints) from the lower-priority channel.

 Although not an incredibly complex example, solutions like this are what continue to keep me excited about Go as I continue to learn it. Just by using core language features, one can to quickly solve the problem. Handling more than 2 channels or changing priority mid-process would take some more work, but are within sight from this solution (more on this later).
