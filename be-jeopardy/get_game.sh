curl localhost:8080/jeopardy/private | jq '.testroom | del(.firstRound, .secondRound, .lastToPick, .lastToAnswer, .lastToBuzz, .finalQuestion)'
