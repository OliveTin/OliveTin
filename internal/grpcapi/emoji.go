package grpcapi

var emojis = map[string]string{
	"poop":  "&#x1f4a9;",
	"smile": "&#x1F600;",
	"ping":  "&#x1f4e1;",
}

func lookupHTMLIcon(keyToLookup string) string {
	if emoji, found := emojis[keyToLookup]; found {
		return emoji
	}

	return keyToLookup
}
