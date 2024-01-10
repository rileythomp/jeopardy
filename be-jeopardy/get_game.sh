curl localhost:8080/jeopardy/games | jq '.testroom | del(.firstRound, .secondRound, .lastToPick, .lastToAnswer, .lastToBuzz, .finalQuestion)'
