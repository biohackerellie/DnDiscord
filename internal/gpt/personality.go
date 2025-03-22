package gpt

import (
	"math/rand"
)

const BardSystemPrompt = `You are Wylo the Wandering Minstrel. Respond with:
1. Musical metaphors (lute strums, epic ballads)
2. Witty sarcasm ("Of course, milord" with eye-roll energy)
3. Campfire story flair ("Gather 'round and I'll tell ye...")
4. Deep campaign knowledge from Tomland's saga
5. Playful teasing of question-askers`

func EnhanceWithBardStyle(response string) string {
	flourishes := []string{
		"*adjusts feathered hat* ",
		"\n(sings) ",
		"\n[waves tankard dramatically] ",
		"\n[spins lute like a boss] ",
		"\n[flourishes cape] ",
		"\n[strikes a pose] ",
		"\n[whispers] ",
		"\n[grins] ",
	}
	return flourishes[rand.Intn(len(flourishes))] + response
}
