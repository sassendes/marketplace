package seed

type category struct {
	Slug, Name, Parent string
}

type user struct {
	Email, Name, Role string
}

type seller struct {
	Email, Store, Slug, Description string
}

type variant struct {
	SKU   string
	Price int64
	Stock int
	Attrs map[string]any
}

type product struct {
	Seller, Category, Title, Slug, Brand, Description string
	Variants                                          []variant
}

type review struct {
	Product, Author string
	Rating          int
	Title, Body     string
}

var categories = []category{
	{"developer-tools", "Developer Tools", ""},
	{"productivity", "Productivity", ""},
	{"security", "Security & Privacy", ""},
	{"design", "Design & Media", ""},
	{"games", "Games", ""},
	{"guides", "Guides & Courses", ""},
}

var users = []user{
	{"admin@example.com", "Platform Admin", "admin"},
	{"store@example.com", "sassendes", "seller"},
	{"buyer@example.com", "Sam Buyer", "buyer"},
}

var sellers = []seller{
	{"store@example.com", "sassendes", "sassendes", "Official marketplace store."},
}

var products = []product{
	{"store@example.com", "guides", "The Seller Guide: Selling Software on sassendes.shop", "seller-guide", "sassendes",
		"Everything you need to open a store and sell software here: listing licenses and downloads, setting how many keys you have, delivering to buyers after payment, handling refunds and keeping your reviews strong. Delivered as a PDF by the sassendes store after purchase.",
		[]variant{{SKU: "GUIDE-PDF", Price: 100, Stock: 100000, Attrs: map[string]any{"format": "PDF"}}}},
}

var reviews = []review{}
