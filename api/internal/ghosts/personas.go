package ghosts

type bookSeed struct {
	OLID     string
	Title    string
	CoverURL string
	Authors  string
}

type persona struct {
	Username    string
	DisplayName string
	Email       string
	Books       []bookSeed
	Reviews     []string
	MinRating   int
	MaxRating   int
	ReviewRate  float64 // probability of reviewing a finished book
}

var personas = []persona{
	{
		Username:    "ghost-jeff",
		DisplayName: "Jeff",
		Email:       "ghost-jeff@rosslib.local",
		MinRating:   2,
		MaxRating:   4,
		ReviewRate:  0.3,
		Reviews: []string{
			"Solid read. Not life-changing but I enjoyed it.",
			"The pacing was off in the middle but the ending saved it.",
			"Would recommend to anyone who likes this genre.",
			"Interesting premise but the execution was uneven.",
			"A decent page-turner. Nothing more, nothing less.",
		},
		Books: []bookSeed{
			{OLID: "OL27448W", Title: "The Great Gatsby", CoverURL: "https://covers.openlibrary.org/b/olid/OL27448W-M.jpg", Authors: "F. Scott Fitzgerald"},
			{OLID: "OL52163W", Title: "Slaughterhouse-Five", CoverURL: "https://covers.openlibrary.org/b/olid/OL52163W-M.jpg", Authors: "Kurt Vonnegut"},
			{OLID: "OL17930368W", Title: "Project Hail Mary", CoverURL: "https://covers.openlibrary.org/b/olid/OL17930368W-M.jpg", Authors: "Andy Weir"},
			{OLID: "OL82586W", Title: "Brave New World", CoverURL: "https://covers.openlibrary.org/b/olid/OL82586W-M.jpg", Authors: "Aldous Huxley"},
			{OLID: "OL103123W", Title: "The Road", CoverURL: "https://covers.openlibrary.org/b/olid/OL103123W-M.jpg", Authors: "Cormac McCarthy"},
			{OLID: "OL17353W", Title: "Cat's Cradle", CoverURL: "https://covers.openlibrary.org/b/olid/OL17353W-M.jpg", Authors: "Kurt Vonnegut"},
			{OLID: "OL14933414W", Title: "The Martian", CoverURL: "https://covers.openlibrary.org/b/olid/OL14933414W-M.jpg", Authors: "Andy Weir"},
			{OLID: "OL1168083W", Title: "Blood Meridian", CoverURL: "https://covers.openlibrary.org/b/olid/OL1168083W-M.jpg", Authors: "Cormac McCarthy"},
			{OLID: "OL276557W", Title: "Catch-22", CoverURL: "https://covers.openlibrary.org/b/olid/OL276557W-M.jpg", Authors: "Joseph Heller"},
			{OLID: "OL82592W", Title: "1984", CoverURL: "https://covers.openlibrary.org/b/olid/OL82592W-M.jpg", Authors: "George Orwell"},
			{OLID: "OL15358691W", Title: "Dark Matter", CoverURL: "https://covers.openlibrary.org/b/olid/OL15358691W-M.jpg", Authors: "Blake Crouch"},
			{OLID: "OL46125W", Title: "Fahrenheit 451", CoverURL: "https://covers.openlibrary.org/b/olid/OL46125W-M.jpg", Authors: "Ray Bradbury"},
			{OLID: "OL2163649W", Title: "The Left Hand of Darkness", CoverURL: "https://covers.openlibrary.org/b/olid/OL2163649W-M.jpg", Authors: "Ursula K. Le Guin"},
			{OLID: "OL59864W", Title: "Neuromancer", CoverURL: "https://covers.openlibrary.org/b/olid/OL59864W-M.jpg", Authors: "William Gibson"},
			{OLID: "OL44910W", Title: "Do Androids Dream of Electric Sheep?", CoverURL: "https://covers.openlibrary.org/b/olid/OL44910W-M.jpg", Authors: "Philip K. Dick"},
		},
	},
	{
		Username:    "ghost-goob",
		DisplayName: "Goob",
		Email:       "ghost-goob@rosslib.local",
		MinRating:   3,
		MaxRating:   5,
		ReviewRate:  0.5,
		Reviews: []string{
			"Absolutely loved this! Couldn't put it down.",
			"Beautiful prose. Every sentence was a joy to read.",
			"This book changed the way I think about things.",
			"One of my favorites this year. Highly recommend.",
			"Charming and heartfelt. A real comfort read.",
			"The characters felt so real I miss them already.",
		},
		Books: []bookSeed{
			{OLID: "OL32432W", Title: "Pride and Prejudice", CoverURL: "https://covers.openlibrary.org/b/olid/OL32432W-M.jpg", Authors: "Jane Austen"},
			{OLID: "OL45883W", Title: "Jane Eyre", CoverURL: "https://covers.openlibrary.org/b/olid/OL45883W-M.jpg", Authors: "Charlotte Bronte"},
			{OLID: "OL20188805W", Title: "Circe", CoverURL: "https://covers.openlibrary.org/b/olid/OL20188805W-M.jpg", Authors: "Madeline Miller"},
			{OLID: "OL17860744W", Title: "The Song of Achilles", CoverURL: "https://covers.openlibrary.org/b/olid/OL17860744W-M.jpg", Authors: "Madeline Miller"},
			{OLID: "OL81634W", Title: "Persuasion", CoverURL: "https://covers.openlibrary.org/b/olid/OL81634W-M.jpg", Authors: "Jane Austen"},
			{OLID: "OL24313379W", Title: "Piranesi", CoverURL: "https://covers.openlibrary.org/b/olid/OL24313379W-M.jpg", Authors: "Susanna Clarke"},
			{OLID: "OL20861063W", Title: "The House in the Cerulean Sea", CoverURL: "https://covers.openlibrary.org/b/olid/OL20861063W-M.jpg", Authors: "TJ Klune"},
			{OLID: "OL82536W", Title: "Little Women", CoverURL: "https://covers.openlibrary.org/b/olid/OL82536W-M.jpg", Authors: "Louisa May Alcott"},
			{OLID: "OL14906539W", Title: "Eleanor Oliphant Is Completely Fine", CoverURL: "https://covers.openlibrary.org/b/olid/OL14906539W-M.jpg", Authors: "Gail Honeyman"},
			{OLID: "OL82560W", Title: "Anne of Green Gables", CoverURL: "https://covers.openlibrary.org/b/olid/OL82560W-M.jpg", Authors: "L.M. Montgomery"},
			{OLID: "OL15099161W", Title: "A Man Called Ove", CoverURL: "https://covers.openlibrary.org/b/olid/OL15099161W-M.jpg", Authors: "Fredrik Backman"},
			{OLID: "OL47690W", Title: "Wuthering Heights", CoverURL: "https://covers.openlibrary.org/b/olid/OL47690W-M.jpg", Authors: "Emily Bronte"},
			{OLID: "OL53908W", Title: "Rebecca", CoverURL: "https://covers.openlibrary.org/b/olid/OL53908W-M.jpg", Authors: "Daphne du Maurier"},
			{OLID: "OL27258W", Title: "To Kill a Mockingbird", CoverURL: "https://covers.openlibrary.org/b/olid/OL27258W-M.jpg", Authors: "Harper Lee"},
			{OLID: "OL8193420W", Title: "The Night Circus", CoverURL: "https://covers.openlibrary.org/b/olid/OL8193420W-M.jpg", Authors: "Erin Morgenstern"},
		},
	},
	{
		Username:    "ghost-casey",
		DisplayName: "Casey",
		Email:       "ghost-casey@rosslib.local",
		MinRating:   1,
		MaxRating:   5,
		ReviewRate:  0.4,
		Reviews: []string{
			"I get why people like this but it wasn't for me.",
			"Incredible. This shook me to my core.",
			"Overrated honestly. Expected more from the hype.",
			"Raw and unflinching. Not easy to read but worth it.",
			"The twist at the end was predictable but still fun.",
			"I've read better takes on this theme but it's fine.",
			"DNF'd halfway through the first time, glad I came back.",
		},
		Books: []bookSeed{
			{OLID: "OL17930368W", Title: "Project Hail Mary", CoverURL: "https://covers.openlibrary.org/b/olid/OL17930368W-M.jpg", Authors: "Andy Weir"},
			{OLID: "OL45804W", Title: "Dune", CoverURL: "https://covers.openlibrary.org/b/olid/OL45804W-M.jpg", Authors: "Frank Herbert"},
			{OLID: "OL27258W", Title: "To Kill a Mockingbird", CoverURL: "https://covers.openlibrary.org/b/olid/OL27258W-M.jpg", Authors: "Harper Lee"},
			{OLID: "OL20188805W", Title: "Circe", CoverURL: "https://covers.openlibrary.org/b/olid/OL20188805W-M.jpg", Authors: "Madeline Miller"},
			{OLID: "OL82592W", Title: "1984", CoverURL: "https://covers.openlibrary.org/b/olid/OL82592W-M.jpg", Authors: "George Orwell"},
			{OLID: "OL46125W", Title: "Fahrenheit 451", CoverURL: "https://covers.openlibrary.org/b/olid/OL46125W-M.jpg", Authors: "Ray Bradbury"},
			{OLID: "OL103123W", Title: "The Road", CoverURL: "https://covers.openlibrary.org/b/olid/OL103123W-M.jpg", Authors: "Cormac McCarthy"},
			{OLID: "OL24313379W", Title: "Piranesi", CoverURL: "https://covers.openlibrary.org/b/olid/OL24313379W-M.jpg", Authors: "Susanna Clarke"},
			{OLID: "OL35688830W", Title: "Klara and the Sun", CoverURL: "https://covers.openlibrary.org/b/olid/OL35688830W-M.jpg", Authors: "Kazuo Ishiguro"},
			{OLID: "OL2749304W", Title: "Never Let Me Go", CoverURL: "https://covers.openlibrary.org/b/olid/OL2749304W-M.jpg", Authors: "Kazuo Ishiguro"},
			{OLID: "OL27448W", Title: "The Great Gatsby", CoverURL: "https://covers.openlibrary.org/b/olid/OL27448W-M.jpg", Authors: "F. Scott Fitzgerald"},
			{OLID: "OL15358691W", Title: "Dark Matter", CoverURL: "https://covers.openlibrary.org/b/olid/OL15358691W-M.jpg", Authors: "Blake Crouch"},
			{OLID: "OL32432W", Title: "Pride and Prejudice", CoverURL: "https://covers.openlibrary.org/b/olid/OL32432W-M.jpg", Authors: "Jane Austen"},
			{OLID: "OL2622440W", Title: "The Handmaid's Tale", CoverURL: "https://covers.openlibrary.org/b/olid/OL2622440W-M.jpg", Authors: "Margaret Atwood"},
			{OLID: "OL47690W", Title: "Wuthering Heights", CoverURL: "https://covers.openlibrary.org/b/olid/OL47690W-M.jpg", Authors: "Emily Bronte"},
		},
	},
	{
		Username:    "ghost-bobert",
		DisplayName: "Bobert",
		Email:       "ghost-bobert@rosslib.local",
		MinRating:   4,
		MaxRating:   5,
		ReviewRate:  0.9,
		Reviews: []string{
			"Masterpiece. I've already started my reread.",
			"Every page was a gift. Truly transcendent storytelling.",
			"This is the book I'll be recommending to everyone this year.",
			"Flawless. I don't say that lightly.",
			"A triumph. Both ambitious and deeply personal.",
			"I am in awe. This belongs on every shelf.",
			"Stunning prose, unforgettable characters, perfect ending.",
			"I highlighted so many passages my Kindle ran out of colors.",
		},
		Books: []bookSeed{
			{OLID: "OL45804W", Title: "Dune", CoverURL: "https://covers.openlibrary.org/b/olid/OL45804W-M.jpg", Authors: "Frank Herbert"},
			{OLID: "OL82592W", Title: "1984", CoverURL: "https://covers.openlibrary.org/b/olid/OL82592W-M.jpg", Authors: "George Orwell"},
			{OLID: "OL27258W", Title: "To Kill a Mockingbird", CoverURL: "https://covers.openlibrary.org/b/olid/OL27258W-M.jpg", Authors: "Harper Lee"},
			{OLID: "OL20188805W", Title: "Circe", CoverURL: "https://covers.openlibrary.org/b/olid/OL20188805W-M.jpg", Authors: "Madeline Miller"},
			{OLID: "OL17860744W", Title: "The Song of Achilles", CoverURL: "https://covers.openlibrary.org/b/olid/OL17860744W-M.jpg", Authors: "Madeline Miller"},
			{OLID: "OL27448W", Title: "The Great Gatsby", CoverURL: "https://covers.openlibrary.org/b/olid/OL27448W-M.jpg", Authors: "F. Scott Fitzgerald"},
			{OLID: "OL52163W", Title: "Slaughterhouse-Five", CoverURL: "https://covers.openlibrary.org/b/olid/OL52163W-M.jpg", Authors: "Kurt Vonnegut"},
			{OLID: "OL2163649W", Title: "The Left Hand of Darkness", CoverURL: "https://covers.openlibrary.org/b/olid/OL2163649W-M.jpg", Authors: "Ursula K. Le Guin"},
			{OLID: "OL24313379W", Title: "Piranesi", CoverURL: "https://covers.openlibrary.org/b/olid/OL24313379W-M.jpg", Authors: "Susanna Clarke"},
			{OLID: "OL53908W", Title: "Rebecca", CoverURL: "https://covers.openlibrary.org/b/olid/OL53908W-M.jpg", Authors: "Daphne du Maurier"},
			{OLID: "OL35688830W", Title: "Klara and the Sun", CoverURL: "https://covers.openlibrary.org/b/olid/OL35688830W-M.jpg", Authors: "Kazuo Ishiguro"},
			{OLID: "OL2749304W", Title: "Never Let Me Go", CoverURL: "https://covers.openlibrary.org/b/olid/OL2749304W-M.jpg", Authors: "Kazuo Ishiguro"},
			{OLID: "OL276557W", Title: "Catch-22", CoverURL: "https://covers.openlibrary.org/b/olid/OL276557W-M.jpg", Authors: "Joseph Heller"},
			{OLID: "OL46125W", Title: "Fahrenheit 451", CoverURL: "https://covers.openlibrary.org/b/olid/OL46125W-M.jpg", Authors: "Ray Bradbury"},
			{OLID: "OL2622440W", Title: "The Handmaid's Tale", CoverURL: "https://covers.openlibrary.org/b/olid/OL2622440W-M.jpg", Authors: "Margaret Atwood"},
		},
	},
}
