// Copyright 2013 Akeda Bagus <admin@gedex.web.id>. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inflector

import (
	"testing"
)

type inflector func(string) string

type inflectorTest struct {
	in  string
	out string
}

var pluralTests = []inflectorTest{
	{"categoria", "categorias"},
	{"house", "houses"},
	{"powerhouse", "powerhouses"},
	{"Bus", "Buses"},
	{"bus", "buses"},
	{"menu", "menus"},
	{"news", "news"},
	{"food_menu", "food_menus"},
	{"Menu", "Menus"},
	{"FoodMenu", "FoodMenus"},
	{"quiz", "quizzes"},
	{"matrix_row", "matrix_rows"},
	{"matrix", "matrices"},
	{"vertex", "vertices"},
	{"index", "indices"},
	{"Alias", "Aliases"},
	{"Aliases", "Aliases"},
	{"Media", "Media"},
	{"NodeMedia", "NodeMedia"},
	{"alumnus", "alumni"},
	{"bacillus", "bacilli"},
	{"cactus", "cacti"},
	{"focus", "foci"},
	{"fungus", "fungi"},
	{"nucleus", "nuclei"},
	{"octopus", "octopuses"},
	{"radius", "radii"},
	{"stimulus", "stimuli"},
	{"syllabus", "syllabi"},
	{"terminus", "termini"},
	{"virus", "viri"},
	{"person", "people"},
	{"people", "people"},
	{"glove", "gloves"},
	{"crisis", "crises"},
	{"tax", "taxes"},
	{"wave", "waves"},
	{"bureau", "bureaus"},
	{"cafe", "cafes"},
	{"roof", "roofs"},
	{"foe", "foes"},
	{"cookie", "cookies"},
	{"wolf", "wolves"},
	{"thief", "thieves"},
	{"potato", "potatoes"},
	{"hero", "heroes"},
	{"buffalo", "buffalo"},
	{"tooth", "teeth"},
	{"goose", "geese"},
	{"foot", "feet"},
	{"objective", "objectives"},
	{"", ""},
}

var singularTests = []inflectorTest{
	{"categorias", "categoria"},
	{"menus", "menu"},
	{"news", "news"},
	{"food_menus", "food_menu"},
	{"Menus", "Menu"},
	{"FoodMenus", "FoodMenu"},
	{"houses", "house"},
	{"powerhouses", "powerhouse"},
	{"quizzes", "quiz"},
	{"Buses", "Bus"},
	{"buses", "bus"},
	{"matrix_rows", "matrix_row"},
	{"matrices", "matrix"},
	{"vertices", "vertex"},
	{"indices", "index"},
	{"Aliases", "Alias"},
	{"Alias", "Alias"},
	{"Media", "Media"},
	{"NodeMedia", "NodeMedia"},
	{"alumni", "alumnus"},
	{"bacilli", "bacillus"},
	{"cacti", "cactus"},
	{"foci", "focus"},
	{"fungi", "fungus"},
	{"nuclei", "nucleus"},
	{"octopuses", "octopus"},
	{"radii", "radius"},
	{"stimuli", "stimulus"},
	{"syllabi", "syllabus"},
	{"termini", "terminus"},
	{"viri", "virus"},
	{"people", "person"},
	{"gloves", "glove"},
	{"doves", "dove"},
	{"lives", "life"},
	{"knives", "knife"},
	{"wolves", "wolf"},
	{"slaves", "slave"},
	{"shelves", "shelf"},
	{"taxis", "taxi"},
	{"taxes", "tax"},
	{"Taxes", "Tax"},
	{"AwesomeTaxes", "AwesomeTax"},
	{"faxes", "fax"},
	{"waxes", "wax"},
	{"niches", "niche"},
	{"waves", "wave"},
	{"bureaus", "bureau"},
	{"genetic_analyses", "genetic_analysis"},
	{"doctor_diagnoses", "doctor_diagnosis"},
	{"parantheses", "paranthesis"},
	{"Causes", "Cause"},
	{"colossuses", "colossus"},
	{"diagnoses", "diagnosis"},
	{"bases", "basis"},
	{"analyses", "analysis"},
	{"curves", "curve"},
	{"cafes", "cafe"},
	{"roofs", "roof"},
	{"foes", "foe"},
	{"databases", "database"},
	{"cookies", "cookie"},
	{"thieves", "thief"},
	{"potatoes", "potato"},
	{"heroes", "hero"},
	{"buffalos", "buffalo"},
	{"babies", "baby"},
	{"teeth", "tooth"},
	{"geese", "goose"},
	{"feet", "foot"},
	{"objectives", "objective"},
	{"", ""},
}

func checker(t *testing.T, name string, fn inflector, tests []inflectorTest) {
	for _, n := range tests {
		s := fn(n.in)
		if s != n.out {
			t.Fatalf("%s(%s) = %s want %s", name, n.in, s, n.out)
		}
		// Second retrieval should return the same result.
		// This also test the cache
		s = fn(n.in)
		if s != n.out {
			t.Fatalf("%s(%s) = %s want %s", name, n.in, s, n.out)
		}
	}
}

func TestPluralize(t *testing.T) {
	checker(t, "Pluralize", Pluralize, pluralTests)
}

func TestSingularize(t *testing.T) {
	checker(t, "Singularize", Singularize, singularTests)
}
