package lobby

import "math/rand"

var adjectives = []string{
	"Swift", "Brave", "Fierce", "Sneaky", "Mighty",
	"Cunning", "Silent", "Reckless", "Clever", "Bold",
	"Ancient", "Crimson", "Golden", "Shadow", "Iron",
	"Crystal", "Blazing", "Frozen", "Thunder", "Neon",
	"Cosmic", "Phantom", "Raging", "Serene", "Witty",
	"Stealth", "Lunar", "Solar", "Toxic", "Quantum",
	"Hollow", "Gilded", "Arcane", "Savage", "Noble",
	"Jolly", "Grim", "Feral", "Radiant", "Obsidian",
	"Vexing", "Stormy", "Lucky", "Sly", "Wicked",
	"Brazen", "Sinister", "Quirky", "Daring", "Rebel",
}

var nouns = []string{
	"Walrus", "Panda", "Cobra", "Falcon", "Badger",
	"Otter", "Raven", "Fox", "Wolf", "Bear",
	"Tiger", "Lion", "Eagle", "Shark", "Viper",
	"Drake", "Lynx", "Bison", "Gecko", "Kraken",
	"Sloth", "Hyena", "Marmot", "Dingo", "Ferret",
	"Jackal", "Wombat", "Narwhal", "Capybara", "Platypus",
	"Axolotl", "Mantis", "Coyote", "Osprey", "Ibis",
	"Toucan", "Meerkat", "Tapir", "Quokka", "Newt",
	"Lemur", "Okapi", "Manatee", "Caracal", "Dugong",
	"Numbat", "Binturong", "Coati", "Kinkajou", "Fossa",
}

func randomName() string {
	return adjectives[rand.Intn(len(adjectives))] + nouns[rand.Intn(len(nouns))]
}
