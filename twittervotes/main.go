package main

import (
	"log"
)

func main() {
	twitterStreamStopCh := make(chan struct{})
	twitterVote := newTwitterVote(twitterStreamStopCh)
	go twitterVote.waitInterruptSignalToFinishTwitterStream()
	go twitterVote.closeConnectionToTwitterStreamPerSecond()

	twitterVoteDB := initTwitterVoteDB("mongodb://ballots:ballots@localhost/ballots")
	if err := twitterVoteDB.dial(); err != nil {
		log.Fatalf("could not dial to DB: %s\n", err)
	}
	defer twitterVoteDB.close()

	votes := make(chan string)
	twitterVoteNSQ := initTwitterVoteNSQ("localhost:4150")
	go twitterVoteNSQ.publishVotes(votes)
	twitterStreamStoppedCh := startTwitterStream(twitterStreamStopCh, votes, twitterVoteDB)

	<-twitterStreamStoppedCh
	close(votes)
	<-twitterVoteNSQ.stoppedCh
}
